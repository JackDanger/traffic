package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/runner"
	"github.com/JackDanger/traffic/transforms"
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

	executor := runner.NewHTTPExecutor(os.Stdout)
	transforms := []transforms.RequestTransform{
		&transforms.ConstantTransform{
			Search:  "https?://.*/",
			Replace: "http://localhost:8000/",
		},
	}
	// To test against localhost:
	// $ python -m SimpleHTTPServer

	runner := runner.Run(har, executor, transforms)
	fmt.Println("started runner")
	<-runner.DoneChannel
	fmt.Println("Done. Exiting.")
}
