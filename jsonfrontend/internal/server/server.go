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

	"gitlab.met.no/forti/f2/internalprotocol"
	"gitlab.met.no/forti/f2/jsonfrontend/internal/server/config"
	"gitlab.met.no/forti/f2/jsonfrontend/internal/server/encode"
	"gitlab.met.no/forti/f2/jsonfrontend/internal/server/metrics"
	"gitlab.met.no/forti/f2/jsonfrontend/pkg/jsonformat"
	"google.golang.org/grpc"
)

type Server struct {
	conn   *grpc.ClientConn
	client internalprotocol.ForecasterClient
}

func New(upstream string) (*Server, error) {
	conn, err := grpc.Dial(upstream, grpc.WithInsecure(), grpc.WithBlock())
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
	startTime := time.Now()

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

	if !config.Configuration.LocationFromGrid || data.ForecastMeta.GridLocation == nil {
		data.ForecastMeta.GridLocation = &internalprotocol.Location{
			Latitude:  forecastRequest.Latitude,
			Longitude: forecastRequest.Longitude,
		}
	}

	var doc *jsonformat.GeoJSON
	switch data.ForecastStatus {
	case internalprotocol.ForecastStatus_OutsideAllGrids:
		http.Error(w, "Outside of coverage area", http.StatusNotFound)
		metrics.OutsideAllGrids.Add(1)
		return
	case internalprotocol.ForecastStatus_PointTooFarAway:
		doc = encode.EncodeError(data, "no data at the given location")
		metrics.PointTooFarAway.Add(1)
	default:
		var err error
		doc, err = encode.Encode(data)
		if err != nil {
			log.Println(err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
	}

	jsonDoc, err := json.Marshal(doc)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	metrics.UpstreamProcessingDuration.Observe(time.Since(startTime).Seconds())

	addHttpHeaders(w)

	if _, err := w.Write(jsonDoc); err != nil {
		log.Println(err)
	}

	metrics.TotalProcessingDuration.Observe(time.Since(startTime).Seconds())
}

func addHttpHeaders(w http.ResponseWriter) {
	for _, header := range config.Configuration.HTTPHeaders {
		w.Header().Add(header.Key, header.Value)
	}
	now := time.Now()
	w.Header().Add("Last-Modified", now.Format(http.TimeFormat))
	if config.Configuration.DataExpiryOffset != 0 {
		expiry := now.Add(time.Duration(config.Configuration.DataExpiryOffset) * time.Second)
		w.Header().Add("Expires", expiry.Format(http.TimeFormat))
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

	if !config.Configuration.SkipAltitude {
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
