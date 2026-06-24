package fortihttp

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/metno/forti/tools/benchmarker/internal/location"
)

type FortiHTTP struct {
	address   string
	userAgent string
	client    *http.Client
	readDelay time.Duration
}

func NewClient(address string, readDelay time.Duration) (*FortiHTTP, error) {

	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}

	return &FortiHTTP{
		address:   address,
		userAgent: "forti-benchamrker",
		client: &http.Client{
			Timeout:   5 * time.Second,
			Transport: &http.Transport{},
		},
		readDelay: readDelay,
	}, nil
}

func (f *FortiHTTP) Close() error {
	return nil
}

func (f *FortiHTTP) RandomRequest() (time.Duration, error) {
	l := location.Pick()

	url := fmt.Sprintf("%s?lat=%.4f&lon=%.4f", f.address, l.Latitude, l.Longitude)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("User-Agent", f.userAgent)
	req.Header.Add("Accept-Encoding", "gzip")

	start := time.Now()
	response, err := f.client.Do(req)
	if err == nil {
		if f.readDelay != 0 {
			time.Sleep(f.readDelay)
		}
		_, err = io.ReadAll(response.Body)
		response.Body.Close()

		if response.StatusCode != 200 {
			err = errors.New(response.Status)
		}
	}
	end := time.Now()

	return end.Sub(start), err
}
