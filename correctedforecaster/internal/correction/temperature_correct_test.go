package correction

import (
	"math"
	"testing"
)

func TestDewPointTemperature(t *testing.T) {
	result := dewPointTemperature(.60060547, 17.02835)
	expected := 9.250092
	if math.Abs(result-expected) > 0.001 {
		t.Errorf("unexpected dew point temperature: %v", result)
	}
}
