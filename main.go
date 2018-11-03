package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/persistence"
	"github.com/JackDanger/traffic/runner"
	"github.com/JackDanger/traffic/server"
	"github.com/JackDanger/traffic/transforms"
)

var serverFlags = flag.NewFlagSet("server", flag.ExitOnError)
var workerFlags = flag.NewFlagSet("worker", flag.ExitOnError)
var runnerFlags = flag.NewFlagSet("runner", flag.ExitOnError)

// Server flags
var port = serverFlags.String("port", "8000", "Run server on <hostname> at this port")
var hostname = serverFlags.String("hostname", "0.0.0.0", "Bind to a network interface (if you want this to be accessible outside of a Docker container this must be '0.0.0.0')")

// Worker flags
var queue = workerFlags.String("queue", "localhost:7000", "The address of the queue that this worker should read from (not yet implemented)")

// One-off HAR file runner flags
var fileFlag = runnerFlags.String("harfile", "", "a .har file to replay")
var archiveIDFlag = runnerFlags.String("archiveID", "", "the id of the archive record to replay")
var velocityFlag = runnerFlags.String("velocity", "", "how fast to replay the archive (defaults to 1.0)")
var concurrencyFlag = runnerFlags.String("concurrency", "", "how many threads to run in parallel")

func main() {
	// If there's just one argument then assume we need to print usage
	if len(os.Args) == 1 {
		fmt.Println("usage: traffic [server|runner] [args]")
		return
	}

	switch os.Args[1] {
	case "server":
		serverFlags.Parse(os.Args[2:])
		runServer()
	case "worker":
		workerFlags.Parse(os.Args[2:])
	case "runner":
		runnerFlags.Parse(os.Args[2:])
		runOneHar()
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}
}

func runServer() {
	_, err := persistence.NewDb()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	web, err := server.NewServer(*port)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	web.Addr = fmt.Sprintf("%s:%s", *hostname, *port)
	fmt.Printf("Starting Traffic server on https://%s\n", web.Addr)
	err = web.ListenAndServeTLS("server/cert.pem", "server/key.pem")
	if err != nil {
		fmt.Printf("Error starting server: %#v", err.Error())
	}
}

func runOneHar() {

	if *fileFlag == "" && *archiveIDFlag == "" {
		fmt.Printf("Specify a .har file to replay\n")
		runnerFlags.PrintDefaults()
		os.Exit(1)
	}

	var err error
	var har *model.Har
	if *fileFlag != "" {
		har, err = parser.HarFromFile(*fileFlag)
		fatalize(err)
	} else {
		db, err := persistence.NewDb()
		fatalize(err)
		id, err := strconv.Atoi(*archiveIDFlag)
		fatalize(err)
		archive, err := db.GetArchive(id)
		fatalize(err)
		har, err = archive.Model()
		fatalize(err)
	}

	if *velocityFlag == "" {
		*velocityFlag = "1.0"
	}
	velocity, err := strconv.ParseFloat(*velocityFlag, 64)
	fatalize(err)

	if *concurrencyFlag == "" {
		*concurrencyFlag = "3"
	}
	concurrency, err := strconv.Atoi(*concurrencyFlag)
	fatalize(err)

	//// Turn every host in every URL into localhost
	//transforms := []transforms.RequestTransform{
	//	&transforms.ConstantTransform{
	//		Search:  "https?://.*/",
	//		Replace: "http://localhost:8000/",
	//	},
	//}
	transforms := []transforms.RequestTransform{}

	startLocalhostServerOnPort("8000")
	waitForRunners := sync.WaitGroup{}
	waitForRunners.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		num := string('0' + i)
		go func() {
			name := filepath.Base(*fileFlag) + " #" + num
			runner := runner.NewHarRunner(har, runner.NewHTTPExecutor(name, os.Stdout), transforms, velocity)
			runner.Run()
			<-runner.GetDoneChannel()
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

func fatalize(err error) {
	if err != nil {
		fmt.Printf("failed with %s", err)
		os.Exit(1)
	}
}
