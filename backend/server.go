// server.go

package main

import (
	"context"
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

func getScheduleJSON() string {
	// Check cache first
	value, err := rdb.Get(ctx, "schedule").Result()
	if err == nil && value != "" {
		fmt.Println("Cache Hit")
		return value
	}

	// Cache miss - scrape data
	fmt.Println("Cache Miss")
	roomID := 0
	schedule := Schedule{}

	c := colly.NewCollector()
	c.OnScraped(func(r *colly.Response) {
		schedule = parseHTML(string(r.Body), roomID, schedule)
	})
	baseURL := "https://app.acuityscheduling.com/schedule.php?action=showCalendar&fulldate=1&owner=30525417&template=weekly"
	header := http.Header{}
	header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	for typeName, calendar := range typeToCalendars {
		body := generateBody(typeName, calendar.ID)
		roomID = typeName
		c.Request("POST", baseURL, body, nil, header)
	}

	// Get all dates and sort them
	dates := make([]string, 0, len(schedule))
	for date := range schedule {
		dates = append(dates, date)
	}

	// Parse dates and sort them
	sort.Slice(dates, func(i, j int) bool {
		// Add current year to dates for proper parsing (format: "Day, Month Date")
		currentYear := time.Now().Year()
		dateWithYearI := fmt.Sprintf("%s, %d", dates[i], currentYear)
		dateWithYearJ := fmt.Sprintf("%s, %d", dates[j], currentYear)

		dateI, errI := time.Parse("Monday, January _2, 2006", dateWithYearI)
		dateJ, errJ := time.Parse("Monday, January _2, 2006", dateWithYearJ)
		if errI != nil || errJ != nil {
			return dates[i] < dates[j] // Fallback to string comparison if parsing fails
		}
		return dateI.Before(dateJ)
	})

	// Build JSON manually to preserve order
	var jsonBuilder strings.Builder
	jsonBuilder.WriteString("{\n")

	for i, date := range dates {
		if i > 0 {
			jsonBuilder.WriteString(",\n")
		}

		// Add the date key
		jsonBuilder.WriteString(fmt.Sprintf("    \"%s\": {\n", date))

		// Get all times for this date and sort them
		times := make([]string, 0, len(schedule[date]))
		for timeStr := range schedule[date] {
			times = append(times, timeStr)
		}

		// Sort times chronologically
		sort.Slice(times, func(i, j int) bool {
			timeI, errI := time.Parse("3:04PM", times[i])
			timeJ, errJ := time.Parse("3:04PM", times[j])
			if errI != nil || errJ != nil {
				return times[i] < times[j]
			}
			return timeI.Before(timeJ)
		})

		// Add each time slot
		for j, timeStr := range times {
			if j > 0 {
				jsonBuilder.WriteString(",\n")
			}

			jsonBuilder.WriteString(fmt.Sprintf("        \"%s\": [", timeStr))

			studios := schedule[date][timeStr]
			for k, studio := range studios {
				if k > 0 {
					jsonBuilder.WriteString(", ")
				}
				jsonBuilder.WriteString(fmt.Sprintf("\"%s\"", studio))
			}

			jsonBuilder.WriteString("]")
		}

		jsonBuilder.WriteString("\n    }")
	}

	jsonBuilder.WriteString("\n}")

	stringifiedSchedule := jsonBuilder.String()
	
	// Cache the result
	statusCmd := rdb.Set(ctx, "schedule", stringifiedSchedule, time.Duration(cacheExpiration)*time.Second)
	if err := statusCmd.Err(); err != nil {
		fmt.Println("Error setting key:", err)
	}

	return stringifiedSchedule
}

func rehearsalsHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scheduleJSON := getScheduleJSON()
	w.Write([]byte(scheduleJSON))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok", "service": "rehearsal-scraper"}`))
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
	fmt.Println("Redis status:", status)

	// Ensure that the connection is properly closed gracefully
	defer rdb.Close()

	// Setup HTTP routes
	http.HandleFunc("/api/rehearsals", rehearsalsHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Println("API endpoints:")
	fmt.Println("  GET /api/rehearsals - Get available rehearsal slots")
	fmt.Println("  GET /health - Health check")
	
	log.Fatal(http.ListenAndServe(port, nil))
}