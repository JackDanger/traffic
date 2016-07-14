package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

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

	har, err := parser.HarFrom(*file, filepath.Base(*file))

	if err != nil {
		fmt.Printf("failed with %s", err)
	}

	// Turn every host in every URL into localhost
	transforms := []transforms.RequestTransform{
		&transforms.ConstantTransform{
			Search:  "https?://.*/",
			Replace: "http://localhost:8000/",
		},
	}

	startLocalhostServerOnPort("8000")

	numRunners := 2
	waitForRunners := sync.WaitGroup{}
	waitForRunners.Add(numRunners)

	for i := 0; i <= numRunners; i++ {
		num := string('0' + i)
		go func() {
			runner := runner.Run(har, runner.NewHTTPExecutor(har.Name+" #"+num, os.Stdout), transforms)
			fmt.Println("started runner")
			<-runner.DoneChannel
			fmt.Println("Done. Exiting.")
			waitForRunners.Done()
		}()
	}

	waitForRunners.Wait()
	fmt.Println("All runners completed")
}

type handler struct{}

// ServeHTTP is a little local server that we can replay our HAR files against
func (h *handler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	w.Write([]byte("nice work!"))
}

// This starts a server and immediately backgrounds it via a goroutine
func startLocalhostServerOnPort(port string) {
	server := http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: &handler{},
	}
	go server.ListenAndServe()
	// Wait a moment so the server can boot
	time.Sleep(100 * time.Millisecond)
	fmt.Println("server is running on ", port)
}
