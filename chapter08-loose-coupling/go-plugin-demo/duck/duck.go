package main

type duck struct{}

func (d duck) Says() string {
	return "quack"
}

// Animal exported as a symbol
var Animal duck
