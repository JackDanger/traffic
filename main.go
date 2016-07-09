package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/runner"
	"os"
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
		stdout := bufio.NewWriter(os.Stdout)
		fmt.Fprintf(stdout, "failed with %s", err)
	}

	r := &runner.Runner{}
	r.Play(&har.Entries[0])
}
