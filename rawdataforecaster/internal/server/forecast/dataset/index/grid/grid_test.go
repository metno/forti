package grid

import (
	"testing"

	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

func TestConstructNotPolygon(t *testing.T) {
	wkt := "POINT (-897442.359961731 -1104322.21784516)"
	proj4 := "+proj=lcc +lat_1=63 +lat_2=63 +lat_0=63 +lon_0=15 +x_0=0 +y_0=0 +a=6371000 +b=6371000 +units=m +no_defs"

	_, err := New(fortiblob.GeographicArea{wkt, proj4})
	if err == nil {
		t.Error("expected failure")
	}
}

func TestConstructInvalidWKT(t *testing.T) {
	wkt := "rubbish"
	proj4 := "+proj=lcc +lat_1=63 +lat_2=63 +lat_0=63 +lon_0=15 +x_0=0 +y_0=0 +a=6371000 +b=6371000 +units=m +no_defs"

	_, err := New(fortiblob.GeographicArea{wkt, proj4})
	if err == nil {
		t.Error("expected failure")
	}
}

func TestConstructInvalidProj(t *testing.T) {
	wkt := "POLYGON ((-897442.359961731 -1104322.21784516,-897442.359961731 1215678.42873958,897558.002775708 1215678.42873958,897558.002775708 -1104322.21784516,-897442.359961731 -1104322.21784516))"
	proj4 := "invalid"

	_, err := New(fortiblob.GeographicArea{wkt, proj4})
	if err == nil {
		t.Error("expected failure")
	}
}

func TestLookup(t *testing.T) {
	wkt := "POLYGON ((-897442.359961731 -1104322.21784516,-897442.359961731 1215678.42873958,897558.002775708 1215678.42873958,897558.002775708 -1104322.21784516,-897442.359961731 -1104322.21784516))"
	proj4 := "+proj=lcc +lat_1=63 +lat_2=63 +lat_0=63 +lon_0=15 +x_0=0 +y_0=0 +a=6371000 +b=6371000 +units=m +no_defs"

	area, err := New(fortiblob.GeographicArea{wkt, proj4})
	if err != nil {
		t.Fatal(err)
	}
	defer area.Free()

	ll := LatLon{Latitude: 59, Longitude: 11}
	if !area.Contains(ll) {
		t.Errorf("%s should have been within area", ll)
	}
	ll = LatLon{Latitude: 71, Longitude: 25}
	if !area.Contains(ll) {
		t.Errorf("%s should have been within area", ll)
	}

	ll = LatLon{Latitude: 65, Longitude: 35}
	if area.Contains(ll) {
		t.Errorf("%s should not have been within area", ll)
	}
}
