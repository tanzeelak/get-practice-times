// scraper.go

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type Calendar struct {
	ID   int
	Name string
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

func parseHTML(htmlString string, calendarID int, schedule map[string]map[string][]int) map[string]map[string][]int {
	/**
	get all of the date.date-heading.date-secondary: March 28
	get all of date.choose-time.form-inline.label: 10am
	*/

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlString))
	if err != nil {
		log.Fatal("Error loading HTML: ", err)
	}
	// Find each time slot and print its date and time
	doc.Find(".choose-time").Each(func(i int, s *goquery.Selection) {
		// For each .choose-time, find the associated date
		date := s.Parent().Find(".date-secondary").Text()
		dayOfWeek := s.Parent().Find(".day-of-week").Text()

		fullDate := fmt.Sprintf("Date: %s, %s\n", dayOfWeek, date)
		fmt.Println(fullDate)
		// Now, find each time within this date
		s.Find(".time-selection").Each(func(j int, timeSelection *goquery.Selection) {
			timeValue, exists := timeSelection.Attr("value")
			if exists {
				fmt.Println("Time:", timeValue)
				updateMap(fullDate, timeValue, calendarID, schedule)
			}
		})
	})
	return schedule
}

func updateMap(fullDate string, timeValue string, calendarID int, schedule map[string]map[string][]int) map[string]map[string][]int {
	time := strings.Split(timeValue, " ")[1]
	timeToStudios, ok := schedule[fullDate]
	if !ok {
		timeToStudios = make(map[string][]int)
		schedule[fullDate] = timeToStudios
	}
	studios, ok := timeToStudios[time]
	if !ok {
		studios = []int{}
		timeToStudios[time] = studios
	}
	schedule[fullDate][time] = append(studios, calendarID)
	return schedule
}

func printSchedule(schedule map[string]map[string][]int) {
	jsonData, err := json.MarshalIndent(schedule, "", "    ")
	if err != nil {
		log.Fatalf("Error occurred during marshaling. Error: %s", err.Error())
	}
	// Print the JSON string
	fmt.Println(string(jsonData))
}

func main() {
	calendarID := 0

	/*
		list by day in ascending order
		day
			list of times
				list of studios

		map[day]map[time][]string

		iterate over each day
			print the day
			iterate over each time print
				print the time
				print the studios


	*/

	schedule := map[string]map[string][]int{}

	typeToCalendars := map[int]Calendar{
		58324142: {9651874, "Studio B"},
		// 54155578: {9651830, "Studio C"},
		// 54535605: {9672985, "Studio D"},
		// 58324342: {9672997, "Studio E"},
		// 54535629: {9651036, "Cottage Studio"},
		// 54535652: {9651030, "Recital Hall"},
		// 58324504: {9650981, "Concert Hall"},
		// 58324623: {9673379, "Studio 1"},
		// 58324707: {9673424, "Studio 2"},
		// 58324742: {9673434, "Studio 3"},
		// 58324779: {9673444, "Studio 4"},
		// 58324847: {9673455, "Studio 5"},
		// 58324961: {9703524, "Studio 7"},
		// 58324992: {9673461, "Studio 8"},
		// 58325034: {9673482, "Studio 9"},
		// 58325156: {9673493, "Studio 10"},
		// 58325267: {9127354, "Studio 12"},
		// 58325228: {9673015, "Studio 11"},
	}

	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong: ", err)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Page visited: ", r.Request.URL)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println(r.Request.URL, " scraped!")
		// fmt.Println(r.Request.Body)
		schedule := parseHTML(string(r.Body), calendarID, schedule)
		printSchedule(schedule)
	})
	baseURL := "https://app.acuityscheduling.com/schedule.php?action=showCalendar&fulldate=1&owner=30525417&template=weekly"
	header := http.Header{}
	header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	for typeName, calendar := range typeToCalendars {
		body := generateBody(typeName, calendar.ID)
		calendarID = calendar.ID
		c.Request("POST", baseURL, body, nil, header)
	}
}
