package main

import (
	"fmt"

	"bitbucket.org/dmesko/kdevpije/employees"
	"bitbucket.org/dmesko/kdevpije/pagerduty"
)

var token = "s22zntFmLdfadjYK99XH"
var testToken = "w_8PcNuhHa-y8xdsfdc1x"

func main() {
	Integrated()
	// GetCalls()

}

func Integrated() {
	cfg := pagerduty.NewPDConfiguration()
	cfg.Token = token
	cfg.MaxPages = 4

	var emps []*employees.Employee
	rsm := &employees.Employee{Email: "radek.smidl@gooddata.com"}
	msu := &employees.Employee{Email: "martin.surovcak@gooddata.com"}
	tdu := &employees.Employee{Email: "tomas.dubec@gooddata.com"}
	emps = *new([]*employees.Employee)
	emps = append(emps, rsm, msu, tdu)

	pagerduty := pagerduty.GetPDEventsForEmployeesList(emps, cfg)
	fmt.Println("bah", pagerduty)
	for k, v := range pagerduty {
		fmt.Println(k, v)
	}
	// fmt.Println(pagerduty["PT5NZ5D"])
	// fmt.Println(pagerduty["PVKGOHR"])

}

func GetCalls() []pagerduty.Oncall {
	idsMap := map[string]string{
		// "a": "P5T36BU",
		// "b": "PZ7JFQ7",
		"c": "PT5NZ5D",
		"D": "PVKGOHR",
	}

	cfg := pagerduty.NewPDConfiguration()
	cfg.Token = token
	calls, err := pagerduty.GetPaginatedOncallsForUsers(0, idsMap, cfg)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(calls)
	return calls
}

func GetUsers() []pagerduty.PDUser {
	cfg := pagerduty.NewPDConfiguration()
	cfg.Token = token
	users, err := pagerduty.GetPaginatedPDUsers(0, cfg)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(users)
	return users
}
