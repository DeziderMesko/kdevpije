package employees

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"bitbucket.org/dmesko/utils"
)

func TestGetEmployeesFromFile(t *testing.T) {
	dataFileName := "test.txt"
	putEmployeesToFile("12345", dataFileName)
	putEmployeesToFile("12345", dataFileName)

	b, e := utils.ReadFileFromWorkingOrHomeDir(dataFileName, ".kdevpije")
	if e != nil {
		t.Errorf("Error: %v", e)
	}
	employees := string(b)
	if employees == "" {
		t.Errorf("No file loaded %v", e)
	}
	if len(employees) != len("12345") {
		t.Fail()
	}
	n, _ := utils.GetHomeDirConfigFileName(dataFileName, ".kdevpije")
	os.Remove(n)
}

func TestGiveMeEmployees(t *testing.T) {
	dataFileName := "test data.json"
	emp, err := GiveMeEmployees(dataFileName, false)
	if len(emp.List) == 0 || err != nil {
		t.Errorf("No employees decoded, %v", err)
	}
	emp.Query = "marcinisn"
	sort.Sort(emp)
	fmt.Println(emp.List[:3])
}

func TestExpandAliases(t *testing.T) {
	tests := []struct {
		result, toExpand []string
		aliases          map[string][]string
	}{
		{toExpand: []string{"aBc", "d"}, aliases: map[string][]string{"abc": []string{"a", "b", "c"}}, result: []string{"a", "b", "c", "d"}},
		{toExpand: []string{"abc", "d"}, aliases: map[string][]string{"aBc": []string{"a", "b", "c"}}, result: []string{"a", "b", "c", "d"}},
		{toExpand: []string{"abc", "c"}, aliases: map[string][]string{"aBc": []string{"a", "b", "c"}}, result: []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		if fmt.Sprint(ExpandAliases(tt.toExpand, tt.aliases)) != fmt.Sprint(tt.result) {
			fmt.Println("Doesn't match:", ExpandAliases(tt.toExpand, tt.aliases), tt.result)
			t.Fail()
		}
	}
}

func TestExpandFiveTimes(t *testing.T) {
	tests := []struct {
		result, toExpand []string
		aliases          map[string][]string
	}{
		{toExpand: []string{"abcd", "g"}, aliases: map[string][]string{"aBcd": []string{"ab", "c", "d"}, "AB": []string{"a", "b"}}, result: []string{"a", "b", "c", "d", "g"}},
		{toExpand: []string{"aBc", "d"}, aliases: map[string][]string{"abc": []string{"a", "b", "c"}}, result: []string{"a", "b", "c", "d"}},
		{toExpand: []string{"abc", "d"}, aliases: map[string][]string{"aBc": []string{"a", "b", "c"}}, result: []string{"a", "b", "c", "d"}},
		{toExpand: []string{"abc", "c"}, aliases: map[string][]string{"aBc": []string{"a", "b", "c"}}, result: []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		if fmt.Sprint(ExpandFiveTimes(tt.toExpand, tt.aliases)) != fmt.Sprint(tt.result) {
			fmt.Println("Doesn't match:", ExpandFiveTimes(tt.toExpand, tt.aliases), tt.result)
			t.Fail()
		}
	}
}
