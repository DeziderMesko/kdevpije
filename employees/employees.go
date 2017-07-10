package employees

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"

	"bitbucket.org/dmesko/utils"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
)

//var dataFileName = "Employees data.json"

func GiveMeEmployees(dataFileName string, reloadData bool) (e Employees, err error) {
	bytes, err := []byte(nil), error(nil)
	fileString := ""
	if !reloadData {
		bytes, err = utils.ReadFileFromWorkingOrHomeDir(dataFileName, ".kdevpije")
		fileString = string(bytes)
	}
	if err != nil || reloadData {
		fmt.Println("Loading employees data from GDrive...")
		emp, err := getEmployeesFromDrive()
		if err != nil {
			fmt.Errorf("Employees Loading Error!\n%s", err.Error())
			return e, err
		}
		putEmployeesToFile(emp, dataFileName)
		fileString = emp
	}

	err = json.NewDecoder(strings.NewReader(fileString)).Decode(&e.List)
	if err != nil {
		log.Printf("Decoding Error!\n%s", err.Error())
		return e, err
	}
	return
}

/*
EmployeesFromList gets list of names/abreviations specified by user on command line and try to map them on employees.
Returns list of mapped employees
*/
func EmployeesFromCLIArgument(list []string, dataFileName string, reloadData bool) ([]*Employee, error) {
	// result := (make(map[string]*Employee))
	result := (make([]*Employee, 0, len(list)))
	employees, err := GiveMeEmployees(dataFileName, reloadData)
	if err != nil {
		return result, err
	}
	for _, uq := range list {
		if len(uq) <= 3 {
			res, err := employees.FindByShort(uq)
			candidate := res
			if err != nil {
				employees.CalculateDistances(uq)
				sort.Sort(employees)
				firstEmployee := employees.List[0]
				if firstEmployee.Metadata.QueryDistance != 0 {
					res.Metadata.Error = true
					res.Metadata.Message = err.Error()
				} else {
					// overwrite with better match
					candidate = &firstEmployee
				}
			}
			result = append(result, candidate)

		} else {

			employees.CalculateDistances(uq)
			sort.Sort(employees)
			firstEmployee := employees.List[0]
			result = append(result, &firstEmployee)
			if firstEmployee.Metadata.QueryDistance != 0 {
				firstEmployee.Metadata.Message = fmt.Sprintf("No employee with name %s found, spinning up crystal ball:", uq)
			}
		}
	}
	log.Printf("Employees on command line %d, mapped %d\n", len(list), len(result))
	return result, nil
}

func ExpandAliases(empList []string, aliases map[string][]string) (expanded []string) {
	empList = utils.SliceToLower(empList)
	aliases = utils.KeysToLower(aliases)
	set := make(map[string]bool)

	for _, item := range empList {
		semiExpanded, ok := aliases[item]
		if ok {
			for _, item := range semiExpanded {
				set[item] = true
			}
		} else {
			set[item] = true
		}
	}
	for k, _ := range set {
		expanded = append(expanded, k)
	}
	sort.Strings(expanded)
	return
}

func ExpandFiveTimes(empList []string, aliases map[string][]string) (expanded []string) {
	for i := 0; i < 5; i++ {
		empList = ExpandAliases(empList, aliases)
	}
	return empList
}

func putEmployeesToFile(content, dataFileName string) error {
	path, err := utils.GetHomeDirConfigFileName(dataFileName, ".kdevpije")
	if err != nil {
		log.Printf("File write error!: %v\n", err)
		return err
	}
	log.Println("Writing data to: ", path)
	err = ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		log.Printf("File write error!: %v\n", err)
	}
	return err
}

func getEmployeesFromDrive() (string, error) {
	ctx := context.Background()
	b, err := utils.ReadFileFromWorkingOrHomeDir("client_secret.json", ".kdevpije")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := utils.GetClient(ctx, config, "drive-go-quickstart.json")
	resp, err := client.Get("https://www.googleapis.com/drive/v2/files/" + "0B7qM3vgXOhNyU25fOEt4N05fSzQ" + "?alt=media")
	if err != nil {
		log.Fatalf("Unable to get data: %v", err)
	}
	defer resp.Body.Close()
	body, bodyError := ioutil.ReadAll(resp.Body)
	if bodyError != nil {
		log.Fatalf("Unable to get body: %v", err)
	}
	log.Printf("Downloaded: %d", len(body))
	return string(body), nil
}
