package weathersymbol

import "testing"

func TestString(t *testing.T) {
	if id := Fair.String(); id != "Fair" {
		t.Errorf("unexpected value for Fair: %s", id)
	}
	if id := Fair.WithSun(Down).String(); id != "Fair" {
		t.Errorf("unexpected value for Fair with sun down: %s", id)
	}
	if id := Fair.WithSun(PolarTwighlight).String(); id != "Fair" {
		t.Errorf("unexpected value for Fair with polar twighlight: %s", id)
	}
}

func TestInvalidSymbol(t *testing.T) {
	invalid := FromValue(1241241)
	if invalid.IsValid() {
		t.Error("invalid symbol is marked as valid")
	}

	if s := invalid.String(); s != "<error>" {
		t.Errorf("unexpected value for invalid symbol: %s", s)
	}
	invalid = invalid.WithSun(Down)
	if s := invalid.String(); s != "<error>" {
		t.Errorf("unexpected value for invalid symbol: %s", s)
	}

	invalid = invalid.WithSun(PolarTwighlight)
	if s := invalid.String(); s != "<error>" {
		t.Errorf("unexpected value for invalid symbol: %s", s)
	}
}

func TestIdentifier(t *testing.T) {
	if id := Fair.Identifier(); id != "fair_day" {
		t.Errorf("unexpected value for Fair: %s", id)
	}
	if id := Fair.WithSun(Down).Identifier(); id != "fair_night" {
		t.Errorf("unexpected value for Fair with day down: %s", id)
	}
	if id := Fair.WithSun(PolarTwighlight).Identifier(); id != "fair_polartwilight" {
		t.Errorf("unexpected value for Fair with polar twighlight: %s", id)
	}
}

func TestSunlesIdentifier(t *testing.T) {
	if id := Cloudy.Identifier(); id != "cloudy" {
		t.Errorf("unexpected value for Cloudy: %s", id)
	}
	if id := Cloudy.WithSun(Down).Identifier(); id != "cloudy" {
		t.Errorf("unexpected value for Cloudy with sun down: %s", id)
	}
	if id := Cloudy.WithSun(PolarTwighlight).Identifier(); id != "cloudy" {
		t.Errorf("unexpected value for Cloudy with polar twighlight: %s", id)
	}
}

func TestExtractSunState(t *testing.T) {
	if state := LightRainShowers.WithSun(Down).SunState(); state != Down {
		t.Errorf("unexpected sun state: %s", state)
	}
}

func TestMultipleSunStates(t *testing.T) {
	s := ClearSky
	if state := s.SunState(); state != Up {
		t.Errorf("unvalid state for sun %d", int(state))
	}

	s = s.WithSun(Down)
	if state := s.SunState(); state != Down {
		t.Errorf("unvalid state for sun %d", int(state))
	}

	s = s.WithSun(PolarTwighlight)
	if state := s.SunState(); state != PolarTwighlight {
		t.Errorf("unvalid state for sun %d", int(state))
	}

	s = s.WithSun(Up)
	if state := s.SunState(); state != Up {
		t.Errorf("unvalid state for sun %d", int(state))
	}
}
