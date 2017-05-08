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

var port = flag.String("port", "8000", "Run server on 0.0.0.0 at this port")

func main() {
	flag.Parse()
	runOneHar()
	//runTheWebInterface()
}

func runTheWebInterface() {
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
	web.Addr = "0.0.0.0"
	fmt.Printf("Starting Traffic server on https://%s:%s\n", "0.0.0.0", *port)
	err = web.ListenAndServeTLS("server/cert.pem", "server/key.pem")
	if err != nil {
		fmt.Printf("Error starting server: %#v", err.Error())
	}
}

func runOneHar() {
	var file = flag.String("harfile", "", "a .har file to replay")
	var archiveID = flag.String("archiveID", "", "the id of the archive record to replay")
	var velocityFlag = flag.String("velocity", "", "how fast to replay the archive (defaults to 1.0)")
	var concurrencyFlag = flag.String("concurrency", "", "how many threads to run in parallel")

	flag.Parse()

	if *file == "" && *archiveID == "" {
		fmt.Printf("Specify a .har file to replay\n")
		flag.PrintDefaults()
		return
	}

	var err error
	var har *model.Har
	if *file != "" {
		har, err = parser.HarFromFile(*file)
		fatalize(err)
	} else {
		db, err := persistence.NewDb()
		fatalize(err)
		id, err := strconv.Atoi(*archiveID)
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
			name := filepath.Base(*file) + " #" + num
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
