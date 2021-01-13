package area

import (
	"math"
	"testing"
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
	p, err := newProjector("+ellps=WGS84 +proj=utm +zone=32")
	if err != nil {
		t.Fatal(err)
	}
	defer p.Free()
	coord := p.Convert(LatLon{Longitude: 8.31257, Latitude: 61.63639})

	tolerance := 5.0

	if math.Abs(coord.X-463567) > tolerance {
		t.Errorf("expected X coordinate to be 463567, got %f", coord.X)
	}
	if math.Abs(coord.Y-6833868) > tolerance {
		t.Errorf("expected X coordinate to be 6833868, got %f", coord.Y)
	}
}
