package main

import (
	"fmt"
	"log"
	"os"
	"plugin"
)

type Sayer interface {
	Says() string
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: run/main.go animal")
	}

	name := os.Args[1]
	module := fmt.Sprintf("./%s/%s.so", name, name)

	p, err := plugin.Open(module)
	if err != nil {
		log.Fatal(err)
	}

	symbol, err := p.Lookup("Animal")
	if err != nil {
		log.Fatal(err)
	}

	animal, ok := symbol.(Sayer)
	if !ok {
		log.Fatal("that isnot a Sayer")
	}

	fmt.Printf("A %s says: %q\n", name, animal.Says())
}
