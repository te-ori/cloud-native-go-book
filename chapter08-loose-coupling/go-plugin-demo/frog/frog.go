package main

type frog struct{}

func (d frog) Says() string {
	return "wrrak"
}

// Animal exported as a symbol
var Animal frog
