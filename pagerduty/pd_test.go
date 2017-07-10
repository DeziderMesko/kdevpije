package pagerduty

import (
	"flag"
	"os"
	"testing"

	"time"

	"bitbucket.org/dmesko/kdevpije/calendar"
	emp "bitbucket.org/dmesko/kdevpije/employees"
)

var emps []*emp.Employee
var pds []PDUser

func TestMain(m *testing.M) {
	flag.Parse()

	emp1 := &emp.Employee{Email: "dezider.mesko@gooddata.com"}
	emp2 := &emp.Employee{Email: "petr.ohnutek@gooddata.com"}
	emp3 := &emp.Employee{Email: "jarda.pulchart@gooddata.com"}
	emps = append(emps, emp1)
	emps = append(emps, emp2)
	emps = append(emps, emp3)

	pd1 := PDUser{Email: "dezider.mesko@gooddata.com", ID: "xclkj33"}
	pd2 := PDUser{Email: "petr.ohnutek@gooddata.com", ID: "gfjg14b"}
	pd3 := PDUser{Email: "pavol.gressa@gooddata.com", ID: "95dfsdf"}
	pds = append(pds, pd1)
	pds = append(pds, pd2)
	pds = append(pds, pd3)
	os.Exit(m.Run())
}

func TestGetIDsForEmails(t *testing.T) {
	ids, _, err := MapPDIDsToEmails(pds, emps)
	if err != nil {
		t.Fail()
	}
	val := ids["dezider.mesko@gooddata.com"]
	if val != "xclkj33" {
		t.Fail()
	}
	val = ids["petr.ohnutek@gooddata.com"]
	if val != "gfjg14b" {
		t.Fail()
	}
}

type FillTimeTestStruct struct {
	StartDate       time.Time
	EndDate         time.Time
	Now             time.Time
	CorrectTimeline *[]calendar.Timeline
}

var parseFormat = "2006-01-02 15:04:05 -0700"
var nowString = "2016-09-30 13:50:00 +0000"
var timeLineLen = 5

/*
2016/09/30 16:29:20 Mapping this call {{Radek Smidl  PT5NZ5D} 2016-09-29 08:00:00 +0000 UTC 2016-10-01 17:00:00 +0000 UTC}
2016/09/30 16:29:20 Mapping this call {{Martin Surovčák  PVKGOHR} 2016-10-01 17:00:00 +0000 UTC 2016-10-02 08:00:00 +0000 UTC}
2016/09/30 16:29:20 Mapping this call {{Radek Smidl  PT5NZ5D} 2016-10-02 08:00:00 +0000 UTC 2016-10-03 08:00:00 +0000 UTC}
*/

func case1() FillTimeTestStruct {
	start, _ := time.Parse(parseFormat, "2016-09-29 08:00:00 +0000")
	end, _ := time.Parse(parseFormat, "2016-10-01 17:00:00 +0000")
	now, _ := time.Parse(parseFormat, nowString)
	result := []calendar.Timeline{
		calendar.Timeline{Desc: []string{"PD"}}, // 30
		calendar.Timeline{Desc: []string{"PD"}}, // 01
		calendar.Timeline{Desc: []string{""}},   // 02
		calendar.Timeline{Desc: []string{""}},   // 03
		calendar.Timeline{Desc: []string{""}},   // 04
	}
	return FillTimeTestStruct{start, end, now, &result}
}

func case2() FillTimeTestStruct {
	start, _ := time.Parse(parseFormat, "2016-10-01 17:00:00 +0000")
	end, _ := time.Parse(parseFormat, "2016-10-02 08:00:00 +0000")
	now, _ := time.Parse(parseFormat, nowString)
	result := []calendar.Timeline{
		calendar.Timeline{Desc: []string{""}}, // 30
		calendar.Timeline{Desc: []string{""}}, // 01
		calendar.Timeline{Desc: []string{""}}, // 02
		calendar.Timeline{Desc: []string{""}}, // 03
		calendar.Timeline{Desc: []string{""}}, // 04
	}
	return FillTimeTestStruct{start, end, now, &result}
}

func case3() FillTimeTestStruct {
	start, _ := time.Parse(parseFormat, "2016-10-02 08:00:00 +0000")
	end, _ := time.Parse(parseFormat, "2016-10-03 08:00:00 +0000")
	now, _ := time.Parse(parseFormat, nowString)
	result := []calendar.Timeline{
		calendar.Timeline{Desc: []string{""}},   // 30
		calendar.Timeline{Desc: []string{""}},   // 01
		calendar.Timeline{Desc: []string{"PD"}}, // 02
		calendar.Timeline{Desc: []string{""}},   // 03
		calendar.Timeline{Desc: []string{""}},   // 04
	}
	return FillTimeTestStruct{start, end, now, &result}
}

var ftts = []FillTimeTestStruct{case1(), case2(), case3()}

func TestFillTimeLine(t *testing.T) {
	for _, tcase := range ftts {
		inputTimeLine := make([]calendar.Timeline, timeLineLen)
		fillTimeLine(tcase.StartDate, tcase.EndDate, tcase.Now, &inputTimeLine)
		correct := *tcase.CorrectTimeline
		real := inputTimeLine

		for idx := range correct {
			if correct[idx].Desc[0] != real[idx].Desc[0] {
				t.Errorf("\nGot      %s, \nexpected %s", real, correct)
				break
			}
		}

	}
}
