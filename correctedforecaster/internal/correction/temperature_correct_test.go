package correction

import (
	"math"
	"testing"
	"time"
)

func TestDewPointTemperature(t *testing.T) {
	result := dewPointTemperature(.60060547, 17.02835)
	expected := 9.250092
	if math.Abs(result-expected) > 0.001 {
		t.Errorf("unexpected dew point temperature: %v", result)
	}
}

func testingTimesteps() map[time.Time]map[string]float32 {
	ret := make(map[time.Time]map[string]float32)

	startingTime := time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 6; i++ {
		ret[startingTime.Add(time.Duration(i)*time.Hour)] = map[string]float32{
			airTemperature2m:      4,
			airTemperature2mMin6h: 8,
			airTemperature2mMax6h: 16,
		}
	}
	for i := 6; i < 12; i++ {
		ret[startingTime.Add(time.Duration(i)*time.Hour)] = map[string]float32{
			airTemperature2m: 2,
		}
	}

	return ret
}

func TestAvgTemperature1h(t *testing.T) {
	timeSteps := testingTimesteps()
	timeStep := time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC)
	ta, ok := temperature1h(timeStep, timeSteps)
	if !ok {
		t.Error("could not get temperature")
	}
	expected := 4
	if ta != float32(expected) {
		t.Errorf("expected value %v, got %v", expected, ta)
	}

	if _, ok := temperature1h(timeStep.Add(-time.Hour), timeSteps); ok {
		t.Error("expected lookup for non-existing time to fail")
	}
}

func TestAvgTemperature6h_fromMinMax(t *testing.T) {
	timeSteps := testingTimesteps()
	timeStep := time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC)
	ta, ok := avgTemperature6h(timeStep, timeSteps)
	if !ok {
		t.Error("could not get temperature")
	}
	expected := 12
	if ta != float32(expected) {
		t.Errorf("expected value %v, got %v", expected, ta)
	}
}

func TestAvgTemperature6h_fromHourly(t *testing.T) {
	timeSteps := testingTimesteps()
	timeStep := time.Date(2024, 4, 17, 6, 0, 0, 0, time.UTC)
	ta, ok := avgTemperature6h(timeStep, timeSteps)
	if !ok {
		t.Error("could not get temperature")
	}
	expected := 2
	if ta != float32(expected) {
		t.Errorf("expected value %v, got %v", expected, ta)
	}
}

func TestAvgTemperature6h_missing(t *testing.T) {
	timeSteps := testingTimesteps()
	timeStep := time.Date(2024, 4, 16, 23, 0, 0, 0, time.UTC)

	if _, ok := avgTemperature6h(timeStep.Add(-time.Hour), timeSteps); ok {
		t.Error("expected lookup for non-existing time to fail")
	}
}

func TestAvgTemperature12h(t *testing.T) {
	timeSteps := testingTimesteps()
	timeStep := time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC)
	ta, ok := avgTemperature12h(timeStep, timeSteps)
	if !ok {
		t.Error("could not get temperature")
	}
	expected := 7
	if ta != float32(expected) {
		t.Errorf("expected value %v, got %v", expected, ta)
	}
}
