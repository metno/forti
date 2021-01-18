package encode

import (
	"testing"
	"time"
)

func getForecast() []parsedForecast {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)

	return []parsedForecast{
		{
			Time: start.Add(-time.Hour),
		},
		{
			Time: start,
		},
		{
			Time: start.Add(time.Hour),
		},
		{
			Time: start.Add(2 * time.Hour),
		},
	}
}

func TestCutForecastKeepCurrent(t *testing.T) {
	result := cutForecast(getForecast(), true)
	if len(result) != 3 {
		t.Errorf("unexpected length: %d", len(result))
	}
}

func TestCutForecastNotKeepCurrent(t *testing.T) {
	result := cutForecast(getForecast(), false)
	if len(result) != 2 {
		t.Errorf("unexpected length: %d", len(result))
	}
}
