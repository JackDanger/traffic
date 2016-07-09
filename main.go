package main

import (
	"flag"
	"fmt"
	"github.com/JackDanger/traffic/parser"
)

func main() {
	var file = flag.String("harfile", "", "a .har file to replay")
	flag.Parse()
	if *file == "" {
		fmt.Printf("Speficy a .har file to replay")
		flag.PrintDefaults()
		return
	}

	har, err := parser.HarFrom(*file)

	if err != nil {
		fmt.Errorf("failed with %s", err)
	}

	fmt.Println(har)
}
