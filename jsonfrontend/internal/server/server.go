package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"gitlab.met.no/forti/f2/jsonfrontend/internal/server/config"
	"gitlab.met.no/forti/f2/jsonfrontend/internal/server/encode"
	"gitlab.met.no/forti/f2/simpleforecaster/pkg/forecaster"
	"google.golang.org/grpc"
)

type Server struct {
	conn   *grpc.ClientConn
	client forecaster.ForecasterClient
}

func New(upstream string) (*Server, error) {
	conn, err := grpc.Dial(upstream, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("could not connect to upstream: %w", err)
	}

	return &Server{
		conn:   conn,
		client: forecaster.NewForecasterClient(conn),
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
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
	defer cancel()

	data, err := s.client.GetForecast(ctx, location)
	if err != nil {
		log.Printf("location lat = %f lon = %f: %s", location.Latitude, location.Longitude, err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	for _, header := range config.Configuration.HTTPHeaders {
		w.Header().Add(header.Key, header.Value)
	}

	// output := encode.GetForecast(data)
	output, err := encode.Encode(location, data)
	if err != nil {
		log.Println(err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	if err := json.NewEncoder(w).Encode(output); err != nil {
		log.Println(err)
	}
}

func getLocation(r *http.Request) (*forecaster.Location, error) {
	q := r.URL.Query()
	latitude, err := getParam(q, "lat", -90, 90)
	if err != nil {
		return nil, err
	}
	longitude, err := getParam(q, "lon", -180, 180)
	if err != nil {
		return nil, err
	}

	location := forecaster.Location{
		Latitude:  latitude,
		Longitude: longitude,
	}

	if _, ok := q["altitude"]; ok {
		altitude, err := getParam(q, "altitude", -500, 9000)
		if err != nil {
			return nil, err
		}
		location.Altitude = &forecaster.Altitude{
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
