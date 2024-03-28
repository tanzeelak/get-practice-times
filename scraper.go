// scraper.go

package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

/**
https://app.acuityscheduling.com/schedule.php?action=showCalendar&fulldate=1&owner=30525417&template=weekly
action=showCalendar&fulldate=1&owner=30525417&template=weekly
type=58324142&calendar=9651874&timezone=America%2FLos_Angeles&skip=true&options%5Bqty%5D=1&options%5BnumDays%5D=3&ignoreAppointment=&appointmentType=&calendarID=

https://app.acuityscheduling.com/schedule.php?action=showCalendar&fulldate=1&owner=30525417&template=weekly
action=showCalendar&fulldate=1&owner=30525417&template=weekly
type=58324342&calendar=9672997&timezone=America%2FLos_Angeles&skip=true&options%5Bqty%5D=1&options%5BnumDays%5D=3&ignoreAppointment=&appointmentType=&calendarID=


https://app.acuityscheduling.com/schedule.php?action=showCalendar&fulldate=1&owner=30525417&template=weekly


*/

// var typeToCalendars = new Array();

// typeToCalendars[58324142] = [[9651874, 'Studio B', '', '552 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[54155578] = [[9651830, 'Studio C', '', '552 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[54535605] = [[9672985, 'Studio D', '', '552 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[58324342] = [[9672997, 'Studio E', '', '552 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[54535629] = [[9651036, 'Cottage Studio', '', '552 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[54535652] = [[9651030, 'Recital Hall', '', '552 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[58324504] = [[9650981, 'Concert Hall', '', '544 Capp St. San Francisco, CA 94110', 'Concert Hall', 'America/Los_Angeles']  ];
// typeToCalendars[58324623] = [[9673379, 'Studio 1', '', '544 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[58324707] = [[9673424, 'Studio 2', '', '544 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[58324742] = [[9673434, 'Studio 3', '', '544 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[58324779] = [[9673444, 'Studio 4', '', '544 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[58324847] = [[9673455, 'Studio 5', '', '544 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[58324961] = [[9703524, 'Studio 7', '', '544 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[58324992] = [[9673461, 'Studio 8', '', '544 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[58325034] = [[9673482, 'Studio 9', '', '544 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[58325156] = [[9673493, 'Studio 10', '', '544 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[58325267] = [[9127354, 'Studio 12', '', '544 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];
// typeToCalendars[58325228] = [[9673015, 'Studio 11', '', '544 Capp St. San Francisco, CA 94110', '', 'America/Los_Angeles']  ];

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

func main() {

	typeToCalendars := map[int]Calendar{
		58324142: {9651874, "Studio B"},
		54155578: {9651830, "Studio C"},
		54535605: {9672985, "Studio D"},
		58324342: {9672997, "Studio E"},
		54535629: {9651036, "Cottage Studio"},
		54535652: {9651030, "Recital Hall"},
		58324504: {9650981, "Concert Hall"},
		58324623: {9673379, "Studio 1"},
		58324707: {9673424, "Studio 2"},
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
		fmt.Println("Response body:", string(r.Body))
	})

	baseURL := "https://app.acuityscheduling.com/schedule.php?action=showCalendar&fulldate=1&owner=30525417&template=weekly"
	header := http.Header{}
	header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	for typeName, calendar := range typeToCalendars {
		body := generateBody(typeName, calendar.ID)
		c.Request("POST", baseURL, body, nil, header)
	}
}
