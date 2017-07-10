package utils

import (
	"testing"
)

func TestStripDiacritics(t *testing.T) {
	tests := []struct {
		input, output string
	}{
		{input: "Dezo", output: "dezo"},
		{input: "Dežo", output: "dezo"},
		{input: "+ěščřžýáíé", output: "+escrzyaie"},
	}
	for i, tt := range tests {
		output := StripDiacritics(tt.input)
		if output != tt.output {
			t.Errorf("%d: Input %s became %s", i, tt.input, output)
		}
	}

}

func TestDistance(t *testing.T) {
	tests := []struct {
		name, query string
		distance    int
	}{
		{name: "Dežo", query: "", distance: 4},
		{name: "Dežo", query: "dezo", distance: 0},
		{name: "Dežo", query: "Dezo", distance: 0},
		{name: "Dezo", query: "Dežo", distance: 0},
		{name: "Dezo", query: "Dezo", distance: 0},
		{name: "Dezo", query: "Deso", distance: 1},
		{name: "Dezo", query: "Dezo12", distance: 2},
		{name: "Dzeo", query: "Dezo", distance: 2},
		{name: "Džeo", query: "Dezo", distance: 2},
		{name: "abc", query: "ghj", distance: 3},
		{name: "Marciniszyn", query: "marcinisin", distance: 2},
	}
	for i, tt := range tests {
		output := Distance(tt.name, tt.query)
		if output != tt.distance {
			t.Errorf("%d: %s -> %s = %d", i, tt.name, tt.query, output)
		}
	}

}
