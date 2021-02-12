package lookup

import (
	"testing"

	"gitlab.met.no/forti/f2/rawdataforecaster/internal/server/pointdata"
)

func TestLookup(t *testing.T) {
	geo, err := New([]float32{60, 60, 61, 61}, []float32{10, 11, 10, 11})
	if err != nil {
		t.Fatalf("Error when creating lookup")
	}
	defer geo.Free()
	gr, err := geo.Nearest(61, 11)
	if err != nil {
		t.Error(err)
	}

	if gr.Idx != 3 {
		t.Errorf("unexpected idx: %d", gr.Idx)
	}
	if gr.Distance != 0 {
		t.Errorf("unexpected distance: %d", gr.Distance)
	}
	expectedPoint := pointdata.Point{Lat: 61, Long: 11}
	if gr.Point != expectedPoint {
		t.Errorf("unexpected point: %v", gr.Point)
	}
}

func TestCorrectLatLon(t *testing.T) {
	geo, err := New([]float32{60, 60, 61, 61}, []float32{10, 11, 10, 11})
	if err != nil {
		t.Fatalf("Error when creating lookup")
	}
	defer geo.Free()
	gr, err := geo.Nearest(61, 11.1)
	if err != nil {
		t.Error(err)
	}

	if gr.Idx != 3 {
		t.Errorf("unexpected idx: %d", gr.Idx)
	}
	expectedPoint := pointdata.Point{Lat: 61, Long: 11}
	if gr.Point != expectedPoint {
		t.Errorf("unexpected point: %v", gr.Point)
	}
}

func TestLookup2(t *testing.T) {
	geo, err := New([]float32{60, 60, 61, 61}, []float32{10, 11, 10, 11})
	if err != nil {
		t.Fatalf("Error when creating lookup")
	}
	defer geo.Free()
	gr, err := geo.Nearest(60.1, 10.1)
	if err != nil {
		t.Error(err)
	}
	expected := uint32(0)
	if expected != gr.Idx {
		t.Errorf("expected %v, got %v", expected, gr.Idx)
	}
}

func TestInvalidLatLonInNew(t *testing.T) {
	if _, err := New([]float32{91}, []float32{10}); err == nil {
		t.Error("Failed to return error due to invalid positive latitude")
	}
	if _, err := New([]float32{-91}, []float32{10}); err == nil {
		t.Error("Failed to return error due to invalid negative latitude")
	}
	if _, err := New([]float32{0}, []float32{181}); err == nil {
		t.Error("Failed to return error due to invalid positive longitude")
	}
	if _, err := New([]float32{0}, []float32{-181}); err == nil {
		t.Error("Failed to return error due to invalid negative longitude")
	}
}

func TestInvalidLatLonInNearest(t *testing.T) {
	geo, err := New([]float32{60, 60, 61, 61}, []float32{10, 11, 10, 11})
	if err != nil {
		t.Fatalf("Error when creating lookup")
	}
	defer geo.Free()
	if _, err := geo.Nearest(91, 0); err == nil {
		t.Error("Failed to return error due to invalid positive latitude")
	}
	if _, err := geo.Nearest(-91, 0); err == nil {
		t.Error("Failed to return error due to invalid negative latitude")
	}
	if _, err := geo.Nearest(0, 181); err == nil {
		t.Error("Failed to return error due to invalid positive longitude")
	}
	if _, err := geo.Nearest(0, -181); err == nil {
		t.Error("Failed to return error due to invalid negative longitude")
	}
}
