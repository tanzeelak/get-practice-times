// scraper.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/redis/go-redis/v9"
)

var cacheExpiration = 21600 // 6 hours

var ctx context.Context
var rdb *redis.Client

type Schedule = map[string]map[string][]string
type Calendar struct {
	ID   int
	Name string
}

var typeToCalendars = map[int]Calendar{
	58324142: {9651874, "Studio B"},
	54155578: {9651830, "Studio C"},
	54535605: {9672985, "Studio D"},
	58324342: {9672997, "Studio E"},
	54535629: {9651036, "Cottage Studio"},
	58324623: {9673379, "Studio 1"},
	58324707: {9673424, "Studio 2"},
	58324742: {9673434, "Studio 3"},
	58324779: {9673444, "Studio 4"},
	58324847: {9673455, "Studio 5"},
	58324992: {9673461, "Studio 8"},
	58325034: {9673482, "Studio 9"},
	58325156: {9673493, "Studio 10"},
	58325267: {9127354, "Studio 12"},
	58325228: {9673015, "Studio 11"},
}

func generateBody(typeName int, calendarID int) *strings.Reader {
	values := url.Values{}
	values.Set("type", strconv.Itoa(typeName))
	values.Set("calendar", strconv.Itoa(calendarID))
	values.Set("timezone", "America/Los_Angeles")
	values.Set("skip", "true")
	values.Set("options[qty]", "1")
	values.Set("options[numDays]", "3")
	values.Set("ignoreAppointment", "")
	values.Set("appointmentType", "")
	values.Set("calendarID", "")
	body := values.Encode()
	return strings.NewReader(body)
}

func parseHTML(htmlString string, roomID int, schedule Schedule) Schedule {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlString))
	if err != nil {
		log.Fatal("Error loading HTML: ", err)
	}
	// Find each time slot and print its date and time
	doc.Find(".choose-time").Each(func(i int, s *goquery.Selection) {
		// For each .choose-time, find the associated date
		date := s.Parent().Find(".date-secondary").Text()
		dayOfWeek := s.Parent().Find(".day-of-week").Text()

		fullDate := fmt.Sprintf("%s, %s", dayOfWeek, date)
		// Now, find each time within this date
		s.Find(".time-selection").Each(func(j int, timeSelection *goquery.Selection) {
			timeValue, exists := timeSelection.Attr("value")
			if exists {
				updateMap(fullDate, timeValue, roomID, schedule)
			}
		})
	})
	return schedule
}

func updateMap(fullDate string, timeValue string, roomID int, schedule Schedule) Schedule {
	militaryTime := strings.Split(timeValue, " ")[1]
	t, err := time.Parse("15:04", militaryTime)
	if err != nil {
		fmt.Println("Error parsing time:", err)
	}
	standardTime := t.Format("3:04PM")
	timeToStudios, ok := schedule[fullDate]
	if !ok {
		timeToStudios = make(map[string][]string)
		schedule[fullDate] = timeToStudios
	}
	studios, ok := timeToStudios[standardTime]
	if !ok {
		studios = []string{}
		timeToStudios[standardTime] = studios
	}
	schedule[fullDate][standardTime] = append(studios, typeToCalendars[roomID].Name)
	return schedule
}

func printSchedule(schedule Schedule) {
	// Create a sorted version of the schedule
	sortedSchedule := make(map[string]map[string][]string)

	// Get all dates and sort them
	dates := make([]string, 0, len(schedule))
	for date := range schedule {
		dates = append(dates, date)
	}

	// Parse dates and sort them
	sort.Slice(dates, func(i, j int) bool {
		// Parse the dates (format: "Day, Month Date")
		dateI, errI := time.Parse("Monday, January 2", dates[i])
		dateJ, errJ := time.Parse("Monday, January 2", dates[j])
		if errI != nil || errJ != nil {
			return dates[i] < dates[j] // Fallback to string comparison if parsing fails
		}
		return dateI.Before(dateJ)
	})

	// Create sorted schedule
	for _, date := range dates {
		sortedSchedule[date] = schedule[date]
	}

	jsonData, err := json.MarshalIndent(sortedSchedule, "", "    ")
	if err != nil {
		log.Fatalf("Error occurred during marshaling. Error: %s", err.Error())
	}
	// Print the JSON string
	stringifiedSchedule := string(jsonData)
	statusCmd := rdb.Set(ctx, "schedule", stringifiedSchedule, time.Duration(cacheExpiration)*time.Second)
	if err := statusCmd.Err(); err != nil {
		fmt.Println("Error setting key:", err)
	} else {
		result, err := statusCmd.Result()
		if err != nil {
			fmt.Println("Error getting result:", err)
		} else {
			fmt.Println("Result:", result)
		}
	}
	fmt.Println(stringifiedSchedule)
}

func main() {
	// SETUP REDIS
	ctx = context.Background()
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	status, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Error connecting to Redis:", err)
	}
	fmt.Println(status)

	// Ensure that the connection is properly closed gracefully
	defer rdb.Close()

	value, err := rdb.Get(ctx, "schedule").Result()
	if err != nil {
		fmt.Println("Error getting key: ", err)
	}

	if value != "" {
		fmt.Println("Cache Hit")
		fmt.Println(value)
	} else {
		// SETUP COLLY
		fmt.Println("Cache Miss")
		roomID := 0
		schedule := Schedule{}

		c := colly.NewCollector()
		c.OnScraped(func(r *colly.Response) {
			schedule = parseHTML(string(r.Body), roomID, schedule) // update schedule with itself
		})
		baseURL := "https://app.acuityscheduling.com/schedule.php?action=showCalendar&fulldate=1&owner=30525417&template=weekly"
		header := http.Header{}
		header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		for typeName, calendar := range typeToCalendars {
			body := generateBody(typeName, calendar.ID)
			roomID = typeName
			c.Request("POST", baseURL, body, nil, header)
		}
		printSchedule(schedule)
	}
}
