package memory

import (
	"testing"
	"time"

	"github.com/metno/forti/rawdataforecaster/pkg/fortiblob"
	"github.com/metno/forti/rawdataforecaster/internal/server/forecast/dataset/values"
)

func TestRead(t *testing.T) {
	r := getSampleReader()
	defer r.Close()

	val, err := r.Read(3)
	if err != nil {
		t.Fatal(err)
	}
	foo := getData(val, "foo")
	if len(foo) != 1 {
		t.Fatalf("unexpected data size: %d", len(foo))
	}
	fooValue := foo[time.Time{}]
	if fooValue != 10000 {
		t.Errorf("unexpected value: %f", fooValue)
	}

	bar := getData(val, "bar")
	if len(bar) != 2 {
		t.Fatalf("unexpected data size: %d", len(bar))
	}
	barValue := bar[time.Unix(0, 0)]
	if barValue != 0.1 {
		t.Errorf("unexpected value: %f", barValue)
	}

}

func getSampleReader() *MemoryReader {
	meta := fortiblob.MetaCollection{
		Parameters: map[string]fortiblob.ParameterMeta{
			"foo": {
				Units:       "m",
				Times:       nil,
				SliceFrom:   0,
				ScaleFactor: 1,
			},
			"bar": {
				Units: "celsius",
				Times: []time.Time{
					time.Unix(0, 0),
					time.Unix(3600, 0),
				},
				SliceFrom:   1,
				ScaleFactor: 0.1,
			},
		},
		NumberOfPoints: 3,
	}

	mad := allocate(3 * 4)
	mad.Values = []int16{
		-10, 1, 1,
		0, 1, 1,
		100, 1, 1,
		10000, 1, 1,
	}

	return &MemoryReader{
		MetaCollection: meta,
		mad:            mad,
	}

}

func getData(c *values.LocationDataCollection, parameter string) map[time.Time]float32 {
	meta, ok := c.ParameterMeta[parameter]
	if !ok {
		return nil
	}
	ret := make(map[time.Time]float32)
	if len(meta.Times) == 0 {
		ret[time.Time{}] = c.Data[meta.SliceFrom]
		return ret
	}

	for i, t := range meta.Times {
		ret[t] = c.Data[meta.SliceFrom+i]
	}

	return ret
}
