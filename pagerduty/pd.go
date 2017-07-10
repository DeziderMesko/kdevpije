package pagerduty

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/dmesko/kdevpije/calendar"
	emp "bitbucket.org/dmesko/kdevpije/employees"
)

var token = "E7px6VVr3PVHZPJq51oa"
var testToken = "E7px6VVr3PVHZPJq51oa"

type PDConfiguration struct {
	PerPage    int
	MaxPages   int
	TimeFrame  int
	Token      string
	OncallsURI string
	UsersURI   string
}

func NewPDConfiguration() *PDConfiguration {
	return &PDConfiguration{
		PerPage:    100,
		MaxPages:   3,
		TimeFrame:  7,
		Token:      "E7px6VVr3PVHZPJq51oa",
		OncallsURI: "https://api.pagerduty.com/oncalls",
		UsersURI:   "https://api.pagerduty.com/users",
	}
}

// from list of emaployees (parsed from GD employee list json) creates map of calendar events
// map[string]*calendar.ParsedEvents = map[email]calendarEvents
func GetPDEventsForEmployeesList(employees []*emp.Employee, cfg *PDConfiguration) (keyShortNameValueEvents map[string]*calendar.ParsedEvents) {

	users, err := GetPaginatedPDUsers(0, cfg)
	if err != nil {
		fmt.Println(err)
	}

	userIDs, userEmails, error := MapPDIDsToEmails(users, employees)
	if error != nil {
		log.Println("Error while mapping PD IDs on emails", error)
		return keyShortNameValueEvents
	}

	oncalls, error := GetPaginatedOncallsForUsers(0, userIDs, cfg)
	if error != nil {
		log.Println("Error while getting PD calls", error)
		return keyShortNameValueEvents
	}

	keyShortNameValueEvents = make(map[string]*calendar.ParsedEvents)
	keyPDIDValueEvents := MergeOncallsToEvent(oncalls, cfg.TimeFrame)
	log.Println("MergeOncallsToEvent returns", keyPDIDValueEvents)

	for pdid := range keyPDIDValueEvents {
		for _, emp := range employees {
			if emp.Email == userEmails[pdid] {
				keyShortNameValueEvents[emp.Short] = keyPDIDValueEvents[pdid]
				break
			}
		}
	}
	return keyShortNameValueEvents
}

// convert many oncalls to one event structure per user
// returns  map[string]*calendar.ParsedEvents = map[PagerdutyId]Event(name, email, timeline)
func MergeOncallsToEvent(oncalls []Oncall, timeFrame int) (events map[string]*calendar.ParsedEvents) {
	events = make(map[string]*calendar.ParsedEvents)
	for _, oncall := range oncalls {
		log.Println("Mapping this call", oncall)
		var event *calendar.ParsedEvents

		if _, ok := events[oncall.User.ID]; !ok {
			event = new(calendar.ParsedEvents)
			event.FullName = oncall.User.Name
			event.Timeline = make([]calendar.Timeline, timeFrame)
			events[oncall.User.ID] = event

		} else {
			event = events[oncall.User.ID]
		}

		startDate := oncall.Start
		endDate := oncall.End
		truncNow := time.Now().Truncate(time.Duration(24) * time.Hour)

		fillTimeLine(startDate, endDate, truncNow, &event.Timeline)

	}
	return events
}

func fillTimeLine(pdStartDate, pdEndDate, now time.Time, timeline *[]calendar.Timeline) {

	tl := *timeline

	for ti := 0; ti < len(tl); ti++ {
		givenDay := now.Truncate(time.Hour * 24).Add(time.Duration((ti)*24) * time.Hour)
		//givenDayEnd := givenDay.Add(time.Duration(24) * time.Hour)
		truncStartDate := pdStartDate.Truncate(time.Hour * 24)
		truncEndDate := pdEndDate.Truncate(time.Hour * 24)
		_, _, hg := givenDay.Date()
		nopd := strconv.Itoa(hg)
		nopd = ""
		if (truncStartDate.Before(givenDay) || truncStartDate.Equal(givenDay)) &&
			(truncEndDate.Equal(givenDay) || truncEndDate.After(givenDay)) {
			hs, _, _ := pdStartDate.Clock()
			he, _, _ := pdEndDate.Clock()
			if truncStartDate.Equal(givenDay) && hs > 16 {
				tl[ti] = calendar.Timeline{Desc: []string{nopd}}
				continue
			}
			if truncEndDate.Equal(givenDay) && he < 9 {
				tl[ti] = calendar.Timeline{Desc: []string{nopd}}
				continue
			}
			tl[ti] = calendar.Timeline{Desc: []string{"PD"}}

		} else {
			tl[ti] = calendar.Timeline{Desc: []string{nopd}}

		}

		// fmt.Println("given", givenDay, "given end", givenDayEnd, "start", pdStartDate, "end", pdEndDate)

	}
}

// MapPDIDsToEmails returns maps PDID-Email maps for list of employees in second parameter
// IDs map[string]string = map[email]=pagerdutyId
// Emails map[string]string = map[pagerdutyId]=email
func MapPDIDsToEmails(PDUsers []PDUser, employees []*emp.Employee) (IDs, Emails map[string]string, err error) {
	IDs = make(map[string]string)
	Emails = make(map[string]string)
	for _, emp := range employees {
		e := emp.Email
		for _, user := range PDUsers {
			if e != user.Email {
				continue
			}
			Emails[user.ID] = user.Email
			IDs[user.Email] = user.ID
		}
	}
	return
}

/*
GetPaginatedUsers collect users through several userlist pages served by PagerDuty
page - zerobased index of first page to fetch
perPage - how many users fetch with one request (PagerDuty default 25, max 100)
maxPages - limit how much pages to fetch maximum
It returns single slice of all users fetched from pages range
*/
func GetPaginatedPDUsers(page int, cfg *PDConfiguration) (users []PDUser, err error) {
	if cfg.MaxPages < page+1 {
		return users, errors.New("Max Pages limit reached, PD users listing cancelled")
	}

	// build query params
	queryParams := collectUsersQueryArguments(page, cfg)

	body, err := executeHTTPRequest(queryParams, cfg.UsersURI, cfg.Token)
	if err != nil {
		return users, err
	}

	// demarshal JSON
	usersStruct := new(UsersReply)
	json, err := demarshallJSON(body, usersStruct)
	if err != nil {
		return users, err
	}
	usersReply := json.(*UsersReply)

	if usersReply.More {
		anotherUsers, err := GetPaginatedPDUsers(page+1, cfg)
		if err != nil {
			return users, err
		}
		users = append(users, anotherUsers...)
	}
	users = append(users, usersReply.Users...)
	return users, nil
}

func collectUsersQueryArguments(page int, cfg *PDConfiguration) (queryParams [][2]string) {
	queryParams = append(queryParams,
		[2]string{"offset", strconv.Itoa(page * cfg.PerPage)},
		[2]string{"limit", strconv.Itoa(cfg.PerPage)},
	)
	return
}

func executeHTTPRequest(queryParams []([2]string), uri string, token string) (body io.ReadCloser, err error) {
	request, _ := http.NewRequest("GET", uri, nil)
	q := request.URL.Query()
	for _, param := range queryParams {
		q.Add(param[0], param[1])
	}
	request.URL.RawQuery = q.Encode()
	log.Println("HTTP Query: ", uri, request.URL.Query(), q.Encode())

	request.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
	request.Header.Set("Authorization", "Token token="+token)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return resp.Body, err
	}

	if resp.StatusCode == 429 {
		log.Println("PagerDuty server is throttling you, wait one minute and try again. WAIT!")
		return resp.Body, errors.New("PagerDuty throttling is active")
	}

	return resp.Body, nil
}

func GetPaginatedOncallsForUsers(page int, userIDs map[string]string, cfg *PDConfiguration) (oncalls []Oncall, err error) {
	if cfg.MaxPages < page+1 {
		return oncalls, errors.New("Max Pages limit reached, PD calls listing canceled")
	}

	// build query params
	queryParams := collectOncallsQueryArguments(page, userIDs, cfg)

	// execute query
	body, err := executeHTTPRequest(queryParams, cfg.OncallsURI, cfg.Token)
	if err != nil {
		return oncalls, err
	}

	// demarshal JSON
	oncallsStruct := new(OncallsReply)
	json, err := demarshallJSON(body, oncallsStruct)
	if err != nil {
		return oncalls, err
	}
	oncallsReply := json.(*OncallsReply)
	log.Println("OncallsReply len:", len(oncallsReply.Oncalls))
	// agregate data from pages and decide if continue to another page
	if oncallsReply.More {
		anotherCalls, err := GetPaginatedOncallsForUsers(page+1, userIDs, cfg)
		if err != nil {
			return oncalls, err
		}
		oncalls = append(oncalls, anotherCalls...)
	}
	oncalls = append(oncalls, oncallsReply.Oncalls...)
	return oncalls, nil
}

func collectOncallsQueryArguments(page int, userIDs map[string]string, cfg *PDConfiguration) (queryParams [][2]string) {
	queryParams = append(queryParams,
		[2]string{"offset", strconv.Itoa(page * cfg.PerPage)},
		[2]string{"limit", strconv.Itoa(cfg.PerPage)},
		[2]string{"time-zone", "CET"},
		[2]string{"until", time.Now().Add(time.Duration(24*cfg.TimeFrame) * time.Hour).Format("2006-01-02")},
	)
	for _, v := range userIDs {
		queryParams = append(queryParams, [2]string{"user_ids[]", v})
	}
	return
}

func demarshallJSON(body io.ReadCloser, mappingStructure interface{}) (interface{}, error) {
	defer body.Close()
	err := json.NewDecoder(body).Decode(&mappingStructure)

	if err != nil {
		log.Printf("Users listing JSON decoding error!\n%s", err.Error())
		return mappingStructure, err
	}
	return mappingStructure, err
}
