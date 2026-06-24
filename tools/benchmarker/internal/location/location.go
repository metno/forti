package location

import (
	"math/rand"
)

type Location struct {
	Latitude  float32
	Longitude float32
}

func Pick() Location {
	if rand.Float32() > 0.5 {
		// around Norway
		return Location{
			Longitude: (rand.Float32() * 5) + 4,
			Latitude:  (rand.Float32() * 10) + 59,
		}
	}
	return Location{
		Longitude: (rand.Float32() * 360) - 180,
		Latitude:  (rand.Float32() * 180) - 90,
	}
}
