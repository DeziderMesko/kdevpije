package calendar

import (
	"fmt"
	"log"
	"strings"
	"time"

	"bitbucket.org/dmesko/utils"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

func GiveMeEvents(calendarId string, timeMin string, timeMax string) (*calendar.Events, error) {
	ctx := context.Background()
	b, err := utils.ReadFileFromWorkingOrHomeDir("client_secret.json", ".kdevpije")
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := utils.GetClient(ctx, config, "calendar-api-quickstart.json")
	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
	}

	events, err := srv.Events.List(calendarId).ShowDeleted(false).MaxResults(250).
		SingleEvents(true).TimeMin(timeMin).TimeMax(timeMax).OrderBy("startTime").Do()

	return events, err
}

func ObtainAndParseEvents(calId string, timeFrame int) []ParsedEvents {
	log.Println("Getting events for calendar Id:", calId)
	calendarId := calId
	tmin := time.Now().Format(time.RFC3339)
	tmax := time.Now().Add(time.Duration(24*timeFrame) * time.Hour).Format(time.RFC3339)
	events, err := GiveMeEvents(calendarId, tmin, tmax)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar events. %v", err)
	}
	log.Println("Events:", len(events.Items))

	parsedEvents, err := ParseEvents(events, timeFrame, *new(time.Time))
	if err != nil {
		log.Println("Can't parse calendar events", err)
	}
	log.Println("Parsed Events:", len(parsedEvents))
	return parsedEvents
}

func PrintEvents(events *calendar.Events) {
	fmt.Println("Upcoming events:")
	if len(events.Items) > 0 {
		for _, i := range events.Items {
			startDate := ParseTimeStructure(i.Start)
			endDate := ParseTimeStructure(i.End)
			duration := endDate.Sub(startDate)
			fmt.Printf("%9s (%s - %s = %d)\n", i.Summary, startDate, endDate, int(duration.Hours()/24))
		}
	} else {
		fmt.Printf("No upcoming events found.\n")
	}
}

func ParseEvents(events *calendar.Events, interval int, now time.Time) (pe []ParsedEvents, err error) {
	if now.IsZero() {
		now = time.Now()
	}
	errors := 0
	if len(events.Items) > 0 {
		pe = make([]ParsedEvents, len(events.Items))
		for idx, i := range events.Items {
			noSpaces := strings.Replace(i.Summary, " ", "", -1)
			split := strings.Split(noSpaces, "-")
			if len(split) < 2 {
				errors++
				log.Println("Error vole, je to blbe:", noSpaces)
				continue
			}
			desc := split[1]
			if len(desc) > 9 {
				desc = string([]rune(desc)[:9])
			}
			tline := make([]Timeline, interval)
			pe[idx].Timeline = tline
			pe[idx].ShortName = strings.ToLower(split[0])
			pe[idx].Original = noSpaces
			startDate := ParseTimeStructure(i.Start)
			endDate := ParseTimeStructure(i.End)
			truncNow := now.Truncate(time.Duration(24) * time.Hour)
			for ti := 0; ti < interval; ti++ {
				givenDay := truncNow.Add(time.Duration((ti)*24) * time.Hour)
				givenDayEnd := givenDay.Add(time.Duration(24) * time.Hour)
				if startDate.Before(givenDayEnd) && endDate.After(givenDay) {
					pe[idx].Timeline[ti] = Timeline{Desc: []string{desc}}
				} else {
					pe[idx].Timeline[ti] = Timeline{Desc: []string{""}}
				}
			}
		}
	}
	return pe[:len(pe)-errors], nil
}

func AggregateEventsByShortName(events []ParsedEvents) (eventsMap map[string]*ParsedEvents) {
	eventsMap = make(map[string]*ParsedEvents)
	for idx, event := range events {
		if mapEvent, ok := eventsMap[event.ShortName]; !ok {
			//before addition of first event to map, remove items with empty string
			for clearIndex := range events[idx].Timeline {
				if event.Timeline[clearIndex].Desc[0] == "" {
					event.Timeline[clearIndex].Desc = []string{}
				}
			}
			eventsMap[event.ShortName] = &events[idx]

		} else {
			for tlineIndex, tlineItem := range event.Timeline {
				for _, descItem := range tlineItem.Desc {
					if descItem == "" {
						continue
					}
					mapEvent.Timeline[tlineIndex].Desc = append(mapEvent.Timeline[tlineIndex].Desc, descItem)
				}
			}
		}
	}
	return
}

func ParseTimeStructure(timeStruct *calendar.EventDateTime) (parsedTime time.Time) {
	var err error
	if timeStruct.DateTime != "" {
		parsedTime, err = time.Parse(time.RFC3339, timeStruct.DateTime)
	} else {
		parsedTime, err = time.Parse("2006-01-02", timeStruct.Date)
	}
	if err != nil {
		log.Println(err)
	}
	return
}

func main() {
	calendarId := "gooddata.com_2tcf3ksksm054fb7vi8oh7nhrg@group.calendar.google.com"
	calendarId = "gooddata.com_94kdisadm2giqg2odrlsui6mgs@group.calendar.google.com"
	tmin := time.Now().Format(time.RFC3339)
	tmax := time.Now().Add(time.Hour * 24 * 8).Format(time.RFC3339)
	events, err := GiveMeEvents(calendarId, tmin, tmax)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar events. %v", err)
	}
	PrintEvents(events)
}
