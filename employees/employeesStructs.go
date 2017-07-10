package employees

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"bitbucket.org/dmesko/utils"
)

type Employee struct {
	Short     string
	FirstName string `json:"First Name"`
	Surname   string
	StartDate string `json:"Start Date"`
	TermDate  string `json:"Term Date"`
	Email     string `json:"Bussiness E-mail"`
	Metadata  struct {
		QueryDistance int
		Message       string
		Error         bool
	}
}

type Employees struct {
	Query string
	List  []Employee
}

func (a *Employees) FindByShort(query string) (*Employee, error) {
	for _, one := range a.List {
		if strings.ToLower(one.Short) == strings.ToLower(query) {
			return &one, nil
		}
	}
	errorMessage := fmt.Sprintf("No employee with shortcut %s, found.", query)
	return new(Employee), errors.New(errorMessage)
}

func (a *Employees) CalculateDistances(query string) {
	a.Query = query
	for idx, emp := range a.List {
		a.List[idx].Metadata.QueryDistance = utils.Distance(emp.Surname, query)
	}
}
func (a Employees) Len() int      { return len(a.List) }
func (a Employees) Swap(i, j int) { a.List[i], a.List[j] = a.List[j], a.List[i] }
func (a Employees) Less(i, j int) bool {
	distI := a.List[i].Metadata.QueryDistance
	distJ := a.List[j].Metadata.QueryDistance
	switch {
	case distI < distJ:
		return true
	case distI > distJ:
		return false
	case distI == distJ:
		if utf8.RuneCountInString(a.List[i].Surname) >= utf8.RuneCountInString(a.List[j].Surname) {
			return true
		} else {
			return false
		}
	}
	return false
}
