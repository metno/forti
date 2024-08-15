package check

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"

	config "gitlab.met.no/forti/f2/healthz/internal/health/config"
	"gitlab.met.no/forti/f2/jsonfrontend/pkg/jsonformat"
)

// Location runs the set of tests specified by blueprint against the given Location.
func Location(location *url.URL, expected config.Blueprint) LocationResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := getRequest(ctx, location)
	if err != nil {
		return LocationResult{
			Service: TypeLocationResult{
				OK:       false,
				Problems: []string{fmt.Sprintf("unable to initialize request for %s: %s", location.String(), err)},
			},
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return LocationResult{
			Service: TypeLocationResult{
				OK:       false,
				Problems: []string{fmt.Sprintf("forecast request for %s failed: %s", location.String(), err)},
			},
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return LocationResult{
			Service: TypeLocationResult{
				OK:       false,
				Problems: []string{fmt.Sprintf("failed response, could not decode, status code: %d", resp.StatusCode)},
			},
		}
	}

	var document jsonformat.GeoJSON
	err = json.NewDecoder(resp.Body).Decode(&document)
	if err != nil {
		return LocationResult{
			Service: TypeLocationResult{
				OK:       false,
				Problems: []string{err.Error()},
			},
		}
	}

	return LocationResult{
		Data: runDataChecks(document.Properties, expected),
	}
}

func getRequest(ctx context.Context, location *url.URL) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", location.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "forti healthz")

	username := os.Getenv("FORTI_USER")
	if username != "" {
		log.Printf("authenticating as user %s", username)
		password := os.Getenv("FORTI_PASSWORD")
		req.SetBasicAuth(username, password)
	}

	return req, nil
}

func runDataChecks(doc *jsonformat.Forecast, expected config.Blueprint) TypeLocationResult {
	if doc == nil {
		return TypeLocationResult{
			OK:       false,
			Problems: []string{"document contains no forecast data"},
		}
	}

	var problemsList []string

	problems := checkUpdatedAge(doc, expected.MaxAge)
	problemsList = append(problemsList, problems...)

	problems = checkTimeseriesDuration(doc, expected)
	problemsList = append(problemsList, problems...)

	problems = checkTimestepIntervals(doc, expected)
	problemsList = append(problemsList, problems...)

	problems = checkDetailsParameters(doc, expected)
	problemsList = append(problemsList, problems...)

	problems = checkSummaryParameters(doc, expected)
	problemsList = append(problemsList, problems...)

	sort.Strings(problemsList)

	return TypeLocationResult{
		OK:       len(problemsList) == 0,
		Problems: problemsList,
	}
}

func checkUpdatedAge(doc *jsonformat.Forecast, maxAge config.Duration) (problems []string) {
	updatedAt := doc.Meta.UpdatedAt

	if (updatedAt == time.Time{}) {
		problems = append(problems, "Updated time not specified in forecast.")
	}

	if time.Since(updatedAt) > maxAge.Duration {
		problems = append(problems, fmt.Sprintf("Forecast data too old, last updated at: %v", updatedAt))
	}

	return problems
}

// checkTimeseriesDuration checks that the total time interval for the timeseries is long enough.
func checkTimeseriesDuration(doc *jsonformat.Forecast, expected config.Blueprint) (problems []string) {
	if len(doc.Timeseries) == 0 {
		problems = append(problems, "Timeseries has zero timesteps!")
		return problems
	}

	// Check for minimum accepted duration of timeseries
	expectedTimeseriesDuration := expected.Timeseries.MinDuration.Duration
	lastStep := doc.Timeseries[len(doc.Timeseries)-1]
	firstStep := doc.Timeseries[0]
	timeseriesDuration := lastStep.Time.Sub(firstStep.Time)

	if timeseriesDuration < expectedTimeseriesDuration {
		problems = append(problems,
			fmt.Sprintf("Too short timeseries. Expected minimum %s; Got %s", expectedTimeseriesDuration, timeseriesDuration),
		)
	}

	return problems
}

// checkTimestepIntervals checks that the steps in a timeseries progresses with correct time intervals.
func checkTimestepIntervals(doc *jsonformat.Forecast, expected config.Blueprint) []string {
	timeseries := doc.Timeseries
	checker := &intervalChecker{expected.Timeseries.Timeresolution}

	for i, current := range timeseries {
		if i == 0 {
			continue
		}
		previous := timeseries[i-1]
		timestepInterval := current.Time.Sub(previous.Time)

		if !checker.correctTimeInterval(timestepInterval) {
			return []string{
				fmt.Sprintf("Wrong duration between timestep %v and %v", previous.Time, current.Time),
			}
		}
	}

	return []string{}
}

type intervalChecker struct {
	resolutions []config.Duration
}

// correctTimeInterval check the stepInterval against the current resolution.
// If current time resolution fails, try the next resolution, if it exists.
func (ic *intervalChecker) correctTimeInterval(stepInterval time.Duration) bool {
	resolutions := ic.resolutions
	if resolutions[0].Duration == stepInterval {
		return true
	}

	if len(resolutions) > 1 && resolutions[1].Duration == stepInterval {
		ic.resolutions = resolutions[1:]
		return true
	}
	return false
}

// checkDetailsParameters check each parameter for:
// minimum number of instances.
// correct values.
func checkDetailsParameters(doc *jsonformat.Forecast, expected config.Blueprint) (problems []string) {
	// Loop through all "details" parameters. E.g instant.details.air_temperature: 10.1
	for expectedPeriod, expectedStepData := range expected.Data {
		for expectedParam, spec := range expectedStepData.Details {

			var paramCount int
			for _, timeStep := range doc.Timeseries {
				value, exists := timeStep.Data[expectedPeriod].Details[expectedParam]

				if !exists {
					continue
				}

				paramCount++
				if spec.MinimumValue != 0 && float32(value) < spec.MinimumValue {
					problems = append(problems, fmt.Sprintf("Param %s at %v has too low a value: %f", expectedParam, timeStep.Time, value))
				}
				if spec.MaximumValue != 0 && float32(value) > spec.MaximumValue {
					problems = append(problems, fmt.Sprintf("Param %s at %v has too high a value: %f", expectedParam, timeStep.Time, value))
				}
			}
			if paramCount < spec.MinimumCount {
				problems = append(problems,
					fmt.Sprintf("Too few instances of parameter %s under period %s. Expected at least %d; Got %d", expectedParam, expectedPeriod, spec.MinimumCount, paramCount),
				)
			}
		}
	}
	return problems
}

func checkSummaryParameters(doc *jsonformat.Forecast, expected config.Blueprint) (problems []string) {
	for expectedPeriod, expectedStepData := range expected.Data {
		for expectedParam, spec := range expectedStepData.Summary {

			var paramCount int
			for _, timeStep := range doc.Timeseries {

				// Skip timesteps in forecast where period is missing, e.g next_6_hours
				if _, exists := timeStep.Data[expectedPeriod]; !exists {
					continue
				}

				summary := timeStep.Data[expectedPeriod].Summary
				if summary == nil {
					problems = append(problems, fmt.Sprintf("Missing summary object under period type %s for time %s",
						expectedPeriod, timeStep.Time))
					continue
				}

				var value string
				if expectedParam == "symbol_code" {
					value = summary.SymbolCode
				}
				if expectedParam == "symbol_confidence" {
					value = summary.SymbolConfidence
				}

				if value == "" {
					continue
				}
				paramCount++
			}

			if paramCount < spec.MinimumCount {
				problems = append(problems, fmt.Sprintf("Too few instances of parameter %s under period type %s. Expected at least %d; Got %d",
					expectedParam, expectedPeriod, spec.MinimumCount, paramCount),
				)
			}
		}
	}
	return problems
}
