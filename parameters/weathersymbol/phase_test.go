package weathersymbol

import (
	"fmt"
	"testing"
)

func msg(from, expected WeatherSymbol, phase PrecipitationPhase) string {
	got := from.WithPhase(phase)
	return fmt.Sprintf("conversion from %s to phase %s failed. Got %s, expected %s", from, phase, got, expected)
}

func TestConvertToRain(t *testing.T) {
	phase := PhaseRain
	if ClearSky.WithPhase(phase) != ClearSky {
		t.Error(msg(ClearSky, ClearSky, phase))
	}
	if LightRainAndThunder.WithPhase(phase) != LightRainAndThunder {
		t.Error(msg(LightRainAndThunder, LightRainAndThunder, phase))
	}
	if SnowShowers.WithPhase(phase) != RainShowers {
		t.Error(msg(SnowShowers, RainShowers, phase))
	}
}

func TestConvertToSleet(t *testing.T) {
	phase := PhaseSleet
	if Cloudy.WithPhase(phase) != Cloudy {
		t.Error(msg(Cloudy, Cloudy, phase))
	}
	if LightRainAndThunder.WithPhase(phase) != LightSleetAndThunder {
		t.Error(msg(LightRainAndThunder, LightSleetAndThunder, phase))
	}
	if SnowShowers.WithPhase(phase) != SleetShowers {
		t.Error(msg(SnowShowers, SleetShowers, phase))
	}
}

func TestConvertToSnow(t *testing.T) {
	phase := PhaseSnow
	if Fair.WithPhase(phase) != Fair {
		t.Error(msg(Fair, Fair, phase))
	}
	if HeavyRainAndThunder.WithPhase(phase) != HeavySnowAndThunder {
		t.Error(msg(RainAndThunder, SleetAndThunder, phase))
	}
	if SnowShowers.WithPhase(phase) != SnowShowers {
		t.Error(msg(SnowShowers, SnowShowers, phase))
	}
}

func TestConvertNight(t *testing.T) {
	s1 := LightRainShowers.WithSun(Down)
	s2 := s1.WithPhase(PhaseSnow)
	if s2 != LightSnowShowers.WithSun(Down) {
		t.Errorf("night symbol not preserved: Got %s (%d), expected %s", s2.Identifier(), int(s2), (SnowShowers.WithSun(Down)).Identifier())
	}
}

func TestConvertPolarTwighlight(t *testing.T) {
	s1 := RainShowers.WithSun(PolarTwighlight)
	s2 := s1.WithPhase(PhaseSnow)
	if s2 != SnowShowers.WithSun(PolarTwighlight) {
		t.Errorf("night symbol not preserved: Got %s (%d), expected %s", s2.Identifier(), int(s2), (SnowShowers.WithSun(Down)).Identifier())
	}
}
