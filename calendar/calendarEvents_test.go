package calendar

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"google.golang.org/api/calendar/v3"
)

var testCase1 string = `
{"items":[{
		"summary": "MPA - OOO(1/2am)",
		"start": {"date": "2015-10-15"},
		"end": {"date": "2015-10-16"}
}]}
`
var testCase2 string = `
{"items":[{
		"summary": "TDU - PTO",
		"start": {"date": "2015-10-17"},
		"end": {"date": "2015-10-25"}
}]}
`
var testCase3 string = `
{"items":[{
		"summary": "JMI - PTO",
		"start": {"dateTime": "2015-10-16T16:00:00+02:00"},
		"end": {"dateTime": "2015-10-16T18:00:00+02:00"}
}]}
`

var testCase4 string = `
{"items":[{
		"summary": "MSU - PTO",
		"start": {"date": "2015-10-19"},
		"end": {"date": "2015-10-20"}
}]}
`

func TestParseEvents(t *testing.T) {
	tests := []struct {
		input  string
		output []string
	}{
		{input: testCase1, output: []string{"OOO(1/2am", "", "", "", ""}},
		{input: testCase2, output: []string{"", "", "PTO", "PTO", "PTO"}},
		{input: testCase3, output: []string{"", "PTO", "", "", ""}},
		{input: testCase4, output: []string{"", "", "", "", "PTO"}},
	}

	for _, tt := range tests {
		ev := &calendar.Events{}
		json.Unmarshal([]byte(tt.input), ev)
		fakeNow, _ := time.Parse("2006-01-02", "2015-10-15")
		parsed, err := ParseEvents(ev, 5, fakeNow)
		if err != nil {
			t.Fail()
		}
		for _, oneEvent := range parsed {
			for idx, value := range oneEvent.Timeline {
				if value.Desc[0] != tt.output[idx] {
					fmt.Printf("Value on index %d '%s' doesn't match '%s'\n", idx, value.Desc, tt.output[idx])
					t.Fail()
				}
			}
		}
	}
}

func TestAggregateEvents(t *testing.T) {
	events := []ParsedEvents{
		generateEvent("DME", "", "", "SL3"),
		generateEvent("DME", "WFH", "", ""),
		generateEvent("DME", "PTO", "PTO", ""),
		generateEvent("MSC", "", "SL3", ""),
	}
	merged := AggregateEventsByShortName(events)
	result := fmt.Sprint(merged["DME"], merged["MSC"])
	expected := "&{ DME  [{[WFH PTO]} {[PTO]} {[SL3]}]} &{ MSC  [{[]} {[SL3]} {[]}]}"
	if result != expected {
		fmt.Println("Result", result)
		fmt.Println("Wanted", expected)
		t.Fail()
	}
}

func generateEvent(name, desc1, desc2, desc3 string) (pe ParsedEvents) {
	pe1 := new(ParsedEvents)
	pe1.ShortName = name
	pe1.Timeline = []Timeline{*new(Timeline), *new(Timeline), *new(Timeline)}
	pe1.Timeline[0].Desc = []string{desc1}
	pe1.Timeline[1].Desc = []string{desc2}
	pe1.Timeline[2].Desc = []string{desc3}
	return *pe1
}
