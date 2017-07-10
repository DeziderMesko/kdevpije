package main

import (
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"bitbucket.org/dmesko/kdevpije/calendar"
	"bitbucket.org/dmesko/kdevpije/config"
	"bitbucket.org/dmesko/kdevpije/employees"
	"bitbucket.org/dmesko/kdevpije/pagerduty"
)

var cfg *config.Config

func main() {
	// log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags)
	var err error
	cfg, err = config.ProcessConfig()
	if err != nil {
		fmt.Println("Config can't be loaded: ", err)
		return
	}
	u := config.ProcessArgs(cfg)
	if u == nil {
		log.Fatal("Arguments parsing failed")
	}
	fmt.Println("Getting Google Calendar events...")
	calendarEvents := getCalendarsEvents()
	employeesList, err := employees.EmployeesFromCLIArgument(u, cfg.EmployeesFile, cfg.ReloadData)

	fmt.Println("Getting PagerDuty OnCalls...")
	pdEvents := pagerduty.GetPDEventsForEmployeesList(employeesList, cfg.PDConfig)
	log.Println("pd events", pdEvents)

	allEvents := mergeMappedEvents(calendarEvents, pdEvents)
	log.Println(allEvents)
	if err != nil {
		fmt.Println(err)
	} else {
		formatOutput(employeesList, u, allEvents)
	}
}

func mergeMappedEvents(maps ...map[string]*calendar.ParsedEvents) (merged map[string]*calendar.ParsedEvents) {
	merged = make(map[string]*calendar.ParsedEvents)
	for _, themap := range maps {
		log.Println("Maps timeline", themap)
		for k, v := range themap {
			k = strings.ToLower(k)
			if _, ok := merged[k]; ok {
				log.Println("!!!Key found:", k)
				merged[k] = mergeTwoEvents(merged[k], v)
			} else {
				log.Println("Key not found:", k)
				merged[k] = v
			}
		}
	}
	return merged
}

func mergeTwoEvents(eventA, eventB *calendar.ParsedEvents) *calendar.ParsedEvents {
	for idx, tli := range eventB.Timeline {
		log.Println("Merging: ", eventA.Timeline[idx].Desc, "and", tli.Desc)
		if len(tli.Desc) == 1 && tli.Desc[0] == "" {
			continue
		}
		if len(eventA.Timeline[idx].Desc) == 1 && eventA.Timeline[idx].Desc[0] == "" {
			eventA.Timeline[idx].Desc = tli.Desc
			continue
		}

		eventA.Timeline[idx].Desc = append(eventA.Timeline[idx].Desc, tli.Desc...)
	}
	return eventA
}

func getCalendarsEvents() (aggregate map[string]*calendar.ParsedEvents) {
	p1 := calendar.ObtainAndParseEvents(cfg.VacationCalendarID, cfg.TimeFrame)
	p2 := calendar.ObtainAndParseEvents(cfg.TripsCalendarID, cfg.TimeFrame)
	aggregate = calendar.AggregateEventsByShortName(append(p1, p2...))
	log.Println("Aggregated:", len(aggregate))
	return
}

func formatOutput(empMap []*employees.Employee, queryList []string, events map[string]*calendar.ParsedEvents) {
	longestDesc := 3
	longestName := 3
	for _, item := range empMap {
		e := item
		if e.Metadata.Error {
			continue
		} else {
			newNameLen := utf8.RuneCountInString(fmt.Sprintf("%s %s - %-3s: ", e.FirstName, e.Surname, e.Short))
			if newNameLen > longestName {
				longestName = newNameLen
			}
			if timeline, ok := events[strings.ToLower(e.Short)]; !ok {
			} else {
				for _, tl := range timeline.Timeline {
					newLen := len(strings.Join(tl.Desc, ","))
					if newLen > longestDesc {
						longestDesc = newLen
					}
				}
			}
		}
	}
	now := time.Now()
	weekendIndexes := make(map[int]bool)
	const weekend = "×|"
	if cfg.TimeFrame != 1 {
		fmt.Printf(fmt.Sprintf("%%%ds", longestName), "")
		for i := 0; i < cfg.TimeFrame; i++ {
			if now.Weekday() == time.Sunday || now.Weekday() == time.Saturday {
				fmt.Printf(weekend)
				weekendIndexes[i] = true
			} else {
				fmt.Printf(fmt.Sprintf("%%-%ds|", longestDesc), now.Format("Mon"))
			}
			now = now.Add(time.Duration(24) * time.Hour)
		}
	}
	fmt.Println()
	for _, item := range empMap {
		e := item
		if e.Metadata.Error {
			fmt.Println(e.Metadata.Message)
			continue
		} else {
			name := fmt.Sprintf("%s %s - %-3s: ", e.FirstName, e.Surname, e.Short)
			fmt.Printf(fmt.Sprintf("%%%ds", longestName), name)
			cellFormat := "%%%ds|"
			if cfg.TimeFrame == 1 {
				cellFormat = "%%%ds"
			}
			if timeline, ok := events[strings.ToLower(e.Short)]; !ok {
				for i := 0; i < cfg.TimeFrame; i++ {
					if _, ok := weekendIndexes[i]; ok {
						fmt.Print(weekend)
					} else {
						fmt.Printf(fmt.Sprintf(cellFormat, longestDesc), "")
					}

				}
			} else {
				for i, tl := range timeline.Timeline {
					if _, ok := weekendIndexes[i]; ok {
						fmt.Print(weekend)
					} else {
						fmt.Printf(fmt.Sprintf(cellFormat, longestDesc), strings.Join(tl.Desc, ","))
					}
				}
			}
			fmt.Println()
		}
	}

}
