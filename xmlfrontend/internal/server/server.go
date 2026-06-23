package server

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	rand "math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/metno/forti/internal/internalprotocol"
	"github.com/metno/forti/parameters/radar"
	"github.com/metno/forti/xmlfrontend/internal/server/config"
	"github.com/metno/forti/xmlfrontend/internal/server/encode"
	"github.com/metno/forti/xmlfrontend/pkg/xmlformat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	conn   *grpc.ClientConn
	client internalprotocol.ForecasterClient
}

func New(upstream string) (*Server, error) {
	conn, err := grpc.Dial(upstream, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("could not to upstream: %w", err)
	}

	return &Server{
		conn:   conn,
		client: internalprotocol.NewForecasterClient(conn),
	}, nil
}

func (s *Server) Close() error {
	return s.conn.Close()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	location, err := getLocation(r)
	if err != nil {
		//log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
	defer cancel()

	data, err := s.client.GetForecast(ctx, location)
	if err != nil {
		log.Printf("location lat = %f lon = %f: %s", location.Latitude, location.Longitude, err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	if data.ForecastStatus != internalprotocol.ForecastStatus_OK {
		http.NotFound(w, r)
		return
	}

	for _, header := range config.Configuration.HTTPHeaders {
		w.Header().Add(header.Key, header.Value)
	}

	output := encode.Encode(location, data)
	if len(output.Product.Time) == 0 {
		handleEmptyForecast(w, output, data)
		return
	}

	now := time.Now()
	w.Header().Add("Last-Modified", now.Format(http.TimeFormat))
	expiry := expires(now, config.Configuration.DataExpiryOffset)
	w.Header().Add("Expires", expiry.Format(http.TimeFormat))

	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")

	if err := enc.Encode(output); err != nil {
		log.Println(err)
	}
}

func handleEmptyForecast(w http.ResponseWriter, doc *xmlformat.ForecastDocument, data *internalprotocol.Forecast) {
	// Special handling for nowcast's radar temporarily unavailable
	for _, d := range data.ParameterMeta {
		if d.Parameter == "precipitation_status" {
			status := radar.Coverage(data.Data[int(d.SliceFrom)])
			if status == radar.TemporarilyUnavailable {
				doc.Product = nil
				if err := xml.NewEncoder(w).Encode(doc); err != nil {
					log.Println(err)
				}
				return
			}
			break
		}
	}

	http.Error(w, "404 page not found", http.StatusNotFound)
}

func getLocation(r *http.Request) (*internalprotocol.GetForecastRequest, error) {
	q := r.URL.Query()
	latitude, err := getParam(q, "lat", -90, 90)
	if err != nil {
		return nil, err
	}
	longitude, err := getParam(q, "lon", -180, 180)
	if err != nil {
		return nil, err
	}

	location := internalprotocol.GetForecastRequest{
		Latitude:  latitude,
		Longitude: longitude,
	}

	if _, ok := q["altitude"]; ok {
		altitude, err := getParam(q, "altitude", -500, 9000)
		if err != nil {
			return nil, err
		}
		location.Altitude = &internalprotocol.Altitude{
			Value:    altitude,
			Override: true,
		}
	}

	return &location, nil
}

func getParam(q url.Values, name string, from float32, to float32) (float32, error) {
	val, ok := q[name]
	if !ok || len(val) == 0 {
		return 0, errors.New("Missing value for " + name)
	}
	if len(val) > 1 {
		return 0, errors.New("Too many values for " + name)
	}
	ret, err := strconv.ParseFloat(val[0], 32)
	if err != nil {
		return 0, errors.New("Value for " + name + " cannot be parsed as float")
	}
	value := float32(ret)
	if value < from || value > to {
		return 0, fmt.Errorf("%s must be in range %v to %v", name, from, to)
	}

	return value, nil
}

// expires returns a randomized time.Time for use as a value to the Expires header.
// Maximum value is current + maxOffset seconds.
// Returns current time + 1 minute if maxOffset <=0.
func expires(current time.Time, maxOffset int) time.Time {
	if maxOffset > 0 {
		baseOffset := maxOffset - (maxOffset / 5)
		randomOffset := rand.IntN(maxOffset / 5)

		return current.Add(time.Duration(baseOffset+randomOffset) * time.Second)
	} else {
		return current.Add(60 * time.Second)
	}
}
