package check

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"gitlab.met.no/forti/f2/healthz/internal/health/xml/config"
	"gitlab.met.no/forti/f2/xmlfrontend/pkg/xmlformat"
)

// URL runs the set of tests specified by blueprint against the given URL.
func URL(location *url.URL, expected config.Blueprint) LocationResult {
	resp, err := http.Get(location.String())
	if err != nil {
		return LocationResult{
			OK:       false,
			Problems: []string{fmt.Sprintf("forecast request for %s failed: %s", location.String(), err)},
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return LocationResult{
			OK:       false,
			Problems: []string{fmt.Sprintf("failed response, could not decode, status code: %d", resp.StatusCode)},
		}
	}

	var document *xmlformat.ForecastDocument
	err = xml.NewDecoder(resp.Body).Decode(document)
	if err != nil {
		return LocationResult{
			OK:       false,
			Problems: []string{err.Error()},
		}
	}

	return runChecks(document, expected)
}

func runChecks(doc *xmlformat.ForecastDocument, expected config.Blueprint) LocationResult {
	problems := make(map[string]bool)

	if maxAge, err := time.ParseDuration(expected.MaxAge); err != nil {
		problems[fmt.Sprintf("unable to parse duration in check specification: %s", err.Error())] = true
	} else if doc.Meta == nil {
		problems["forecast is missing meta element"] = true
	} else {
		for _, meta := range doc.Meta.Models {
			if time.Since(meta.Runended) > maxAge {
				problems[fmt.Sprintf("forecast is too told - model %s was updated at %s", meta.Name, meta.Runended)] = true
			}
		}
	}

	for parameter, checks := range expected.Parameters {
		if len(checks.Durations) == 0 {
			checks.Durations = []int{0}
		}
		for _, duration := range checks.Durations {
			parameterIdentifier := mkParameterIdentifier(parameter, duration)

			forecast, err := extractData(doc, parameter, checks.Attribute, time.Duration(duration)*time.Hour)
			if err != nil {
				problems[fmt.Sprintf("unable to get data for %s: %s", parameterIdentifier, err)] = true
			}

			hasData := false
			for _, f := range forecast {
				if f.HasData {
					hasData = true
					break
				}
			}
			if !hasData && checks.Optional {
				log.Printf("missing all data for <%s>, but this is an optional parameter", parameterIdentifier)
				continue
			}

			subset := forecast
			if checks.MinimumCount != nil && *checks.MinimumCount < uint(len(subset)) {
				subset = subset[:*checks.MinimumCount]
			}
			var missingTimestep []time.Time
			for _, f := range subset {
				if !f.HasData {
					missingTimestep = append(missingTimestep, f.From)
				}
			}
			if len(missingTimestep) != 0 {
				problems[fmt.Sprintf("missing %d timesteps for %s for forecast - first missing time was %s", len(missingTimestep), parameterIdentifier, missingTimestep[0])] = true
			}

			if checks.MinimumValue != nil {
				for _, f := range forecast {
					if f.Value < *checks.MinimumValue {
						problems[fmt.Sprintf("too low value for %s: %v", parameterIdentifier, f.Value)] = true
						break
					}
				}
			}
			if checks.MaximumValue != nil {
				for _, f := range forecast {
					if f.Value > *checks.MaximumValue {
						problems[fmt.Sprintf("too high value for %s: %v", parameterIdentifier, f.Value)] = true
						break
					}
				}
			}
		}
	}

	var problemsList []string
	for p := range problems {
		problemsList = append(problemsList, p)
	}
	sort.Strings(problemsList)

	return LocationResult{
		OK:       len(problems) == 0,
		Problems: problemsList,
	}
}

func mkParameterIdentifier(parameter string, duration int) string {
	if duration == 0 {
		return parameter
	}
	return fmt.Sprintf("%s (%dh)", parameter, duration)
}

type forecastData struct {
	From    time.Time
	To      time.Time
	HasData bool
	Value   float32
}

func extractData(doc *xmlformat.ForecastDocument, parameter, attribute string, duration time.Duration) ([]forecastData, error) {
	var ret []forecastData
	for _, t := range forecastsWithDuration(doc, duration) {
		d := forecastData{
			From:    t.From,
			To:      t.To,
			HasData: false,
		}
		for _, f := range t.Location.Forecast {
			if f.XMLName.Local == parameter {
				val, err := valueOf(&f, attribute)
				if err != nil {
					return nil, fmt.Errorf("unable to find a value for %s: %w", parameter, err)
				}
				d.HasData = true
				d.Value = val
				break
			}
		}
		ret = append(ret, d)
	}
	return ret, nil
}

func forecastsWithDuration(doc *xmlformat.ForecastDocument, duration time.Duration) []xmlformat.TimeElement {
	var ret []xmlformat.TimeElement

	for _, t := range doc.Product.Time {
		forecastDuration := t.To.Sub(t.From)
		if forecastDuration == duration {
			ret = append(ret, t)
		}
	}

	return ret
}

func valueOf(data *xmlformat.DataElement, attribute string) (float32, error) {
	var strValue string
	for _, attr := range data.Attr {
		if attr.Name.Local == attribute {
			strValue = attr.Value
			break
		}
	}
	if strValue == "" {
		return 0, fmt.Errorf("unable to find attribute %s", attribute)
	}
	val, err := strconv.ParseFloat(strValue, 32)
	return float32(val), err
}
