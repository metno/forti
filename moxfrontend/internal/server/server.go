package server

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/metno/forti/internal/internalprotocol"
	"github.com/metno/forti/moxfrontend/internal/mox"
	"github.com/metno/forti/moxfrontend/internal/server/encode"
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
		return nil, fmt.Errorf("could not connect to upstream: %w", err)
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
	forecastRequest, err := getForecastRequest(r)
	if err != nil {
		//log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
	defer cancel()

	data, err := s.client.GetForecast(ctx, forecastRequest)
	if err != nil {
		log.Printf("location lat = %f lon = %f: %s", forecastRequest.Latitude, forecastRequest.Longitude, err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	var output *mox.ForecastDocument
	if data.ForecastStatus == internalprotocol.ForecastStatus_OK {
		// Use coordinates from query, rather than returned coordinates
		data.ForecastMeta.GridLocation.Latitude = forecastRequest.Latitude
		data.ForecastMeta.GridLocation.Longitude = forecastRequest.Longitude

		output, err = encode.Encode(data)
		if err != nil {
			log.Println(err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
	} else {
		output = encode.EncodeNoData()
	}

	now := time.Now()
	w.Header().Add("Last-Modified", now.Format(http.TimeFormat))
	w.Header().Add("Expires", now.Add(time.Hour).Format(http.TimeFormat))
	w.Header().Add("Content-Type", "application/xml")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Add("Access-Control-Allow-Headers", "Origin")

	if _, err := fmt.Fprintln(w, `<?xml version="1.0" encoding="UTF-8" ?>`); err != nil {
		log.Println(err)
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(output); err != nil {
		log.Println(err)
	}
}

func getForecastRequest(r *http.Request) (*internalprotocol.GetForecastRequest, error) {
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
