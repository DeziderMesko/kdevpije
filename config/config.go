package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"bitbucket.org/dmesko/kdevpije/employees"
	"bitbucket.org/dmesko/kdevpije/pagerduty"
	"bitbucket.org/dmesko/utils"
)

// Config struct is app wide configuration structure
type Config struct {
	Aliases            map[string][]string
	Intervals          map[string]int
	VacationCalendarID string
	TripsCalendarID    string
	TimeFrame          int
	EmployeesFile      string
	ReloadData         bool
	PDConfig           *pagerduty.PDConfiguration
}

// ProcessConfig loads config.json to app wide structure
func ProcessConfig() (c *Config, err error) {
	file, err := os.Open("config.json")
	if err != nil {
		log.Println("Config not found in working directory, trying user's home")
		cfgName, err2 := utils.GetHomeDirConfigFileName("config.json", ".kdevpije")
		if err2 != nil {
			log.Println("Error obtaining home directory:", err2)
			return c, err2
		}
		var err3 error
		file, err3 = os.Open(cfgName)
		if err3 != nil {
			log.Println("Config not found in ~/.kdevpije/")
			return c, err3
		}
	}
	decoder := json.NewDecoder(file)
	c = &Config{}
	err = decoder.Decode(c)
	if err != nil {
		log.Fatalln("Config decode error")
		return c, err
	}
	if c.Aliases == nil {
		c.Aliases = make(map[string][]string)
	}
	if c.PDConfig == nil {
		c.PDConfig = pagerduty.NewPDConfiguration()
		c.PDConfig.Token = "s23zntFxLYp9NjYK99XH"
	}
	return
}

// ProcessArgs process arguments - like splitting, or assigning default values
func ProcessArgs(cfg *Config) (u []string) {
	flag.Usage = func() {
		fmt.Println("Usage:\nkdevpije user1,user2,alias3... [default|week|sprint]")
		flag.PrintDefaults()
	}
	var debugFlag = flag.Bool("debug", false, "Print logs to stderr")
	var reloadData = flag.Bool("reloadData", false, "Download list of employees again")
	flag.Parse() // Scan the arguments list

	if !*debugFlag {
		log.SetOutput(ioutil.Discard)
	}
	log.Println("Processing arguments")
	cfg.ReloadData = *reloadData
	emps := flag.Arg(0)
	if emps == "" {
		flag.PrintDefaults()
		return
	}
	u = strings.Split(emps, ",")
	u = employees.ExpandFiveTimes(u, cfg.Aliases)

	timeframe := flag.Arg(1)
	if timeframe == "" {
		timeframe = "default"
	}
	tf, ok := cfg.Intervals[timeframe]
	if !ok {
		tf = 1
	}
	cfg.TimeFrame = tf
	cfg.PDConfig.TimeFrame = cfg.TimeFrame
	log.Println("Processed config:", cfg)
	return
}
