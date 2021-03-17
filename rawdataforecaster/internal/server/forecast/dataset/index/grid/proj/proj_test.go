package proj

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// func TestConstructObjectCode(t *testing.T) {
// 	// According to docs, this should work, but it does not. Could be proj version.
// 	_, err := newProjector("urn:ogc:def:crs:EPSG::4326")
// 	if err == nil {
// 		t.Error("expected failure")
// 	}
// }

// func TestConstructCRSFromWKT(t *testing.T) {
// 	// According to docs, this should work, but it does not. Could be proj version.
// 	srsWKT := `GEOGCS["WGS 84",
//     DATUM["WGS_1984",
//         SPHEROID["WGS 84",6378137,298.257223563,
//             AUTHORITY["EPSG","7030"]],
//         AUTHORITY["EPSG","6326"]],
//     PRIMEM["Greenwich",0,
//         AUTHORITY["EPSG","8901"]],
//     UNIT["degree",0.01745329251994328,
//         AUTHORITY["EPSG","9122"]],
// 	AUTHORITY["EPSG","4326"]]`
// 	_, err := newProjector(srsWKT)
// 	if err == nil {
// 		t.Error("expected failure")
// 	}
// }

func TestConvert(t *testing.T) {
	p, err := Get("+ellps=WGS84 +proj=utm +zone=32")
	if err != nil {
		t.Fatal(err)
	}
	defer p.Return()
	coord := p.Convert(8.31257, 61.63639)

	tolerance := 5.0

	if math.Abs(coord.X-463567) > tolerance {
		t.Errorf("expected X coordinate to be 463567, got %f", coord.X)
	}
	if math.Abs(coord.Y-6833868) > tolerance {
		t.Errorf("expected X coordinate to be 6833868, got %f", coord.Y)
	}
}

func TestMultithreadConvert(t *testing.T) {

	goroutines := 100
	projDef := "+ellps=WGS84 +proj=utm +zone=32"

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			for ctx.Err() == nil {
				p, err := Get(projDef)
				if err != nil {
					panic(err)
				}
				p.Convert(rand.Float64()*360-180, rand.Float64()*180-90)
				p.Return()
			}
			wg.Done()
		}()
	}

	wg.Wait()
	myPool := pool[projDef]
	if len(myPool) > goroutines {
		t.Errorf("too many entries in pool: %d", len(myPool))
	}
}
