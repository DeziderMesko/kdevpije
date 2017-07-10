package calendar

import ()

type ParsedEvents struct {
	FullName  string
	ShortName string
	Original  string
	Timeline  []Timeline
}
type Timeline struct {
	Desc []string
}
