package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gitlab.met.no/forti/f2/healthz/internal/json/check"
	"gitlab.met.no/forti/f2/healthz/internal/json/config"
)

func main() {
	runServer := flag.Bool("run-server", false, "Run a server on port 8080, continously serving status.")
	configFile := flag.String("config", "/etc/forti/checks.json", "Read check configuration from the given file.")
	flag.Parse()

	conf, err := config.Read(*configFile)
	if err != nil {
		log.Fatalln(err)
	}

	if *runServer {
		http.HandleFunc("/", healthzHandler(conf))
		address := ":8080"
		log.Println("ok. Serving data on " + address)
		log.Fatal(http.ListenAndServe(address, nil))
	} else {
		result := runChecks(conf)
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			panic(err)
		}
		if !result.OK {
			os.Exit(1)
		}
	}
}

func healthzHandler(conf *config.CheckConfiguration) func(w http.ResponseWriter, r *http.Request) {

	var lastRun time.Time
	var nextRun time.Time
	var lastResult check.Result
	var lastResultOK bool

	return func(w http.ResponseWriter, r *http.Request) {

		var format func(w http.ResponseWriter, r check.Result)
		switch r.URL.Path {
		case "/":
			format = simpleOutput
		case "/full":
			format = jsonOutput
		default:
			http.NotFound(w, r)
			return
		}

		now := time.Now()
		if now.After(nextRun) {
			lastRun = now
			log.Println("running checks")
			nextRun = now.Add(time.Minute)
			result := runChecks(conf)

			log.Printf("%v\n", result)
			if !result.OK || !lastResultOK {
				log.Println(lastResult)
			}

			lastResultOK = result.OK
			lastResult = result
		}

		for key, value := range conf.Headers {
			w.Header().Add("Healthz-"+key, value)
		}
		w.Header().Add("Healthz-Date", lastRun.UTC().Format("2006-01-02T15:04:05Z"))

		if !lastResultOK {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		format(w, lastResult)
	}
}

func simpleOutput(w http.ResponseWriter, r check.Result) {
	w.Header().Add("Content-Type", "text/plain;charset=UTF-8")
	fmt.Fprintln(w, r)
}

func jsonOutput(w http.ResponseWriter, r check.Result) {
	w.Header().Add("Content-Type", "application/json;charset=UTF-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(r)
}

func runChecks(conf *config.CheckConfiguration) check.Result {
	result := check.NewResult()

	var failedRequests int
	for _, request := range conf.GetRequests() {

		problems := check.URL(request.URL, request.Blueprint)
		if !problems.OK {
			failedRequests++
			result.Locations[request.Name] = problems
		}
	}
	if failedRequests > conf.Response.MaxFailures {
		result.OK = false
	}

	return result
}
