package memory

import (
	"context"
	"math"
	"testing"

	"github.com/metno/forti/fortiup/pkg/fortiblob/sampleblob"
	"github.com/metno/forti/rawdataforecaster/internal/server/forecast/dataset/values"
)

func TestDownload(t *testing.T) {
	reader, err := getBlobClient()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	for loc := 0; loc < 100; loc++ {
		forecast, err := reader.Read(loc)
		if err != nil {
			t.Fatal(err)
		}
		if len(forecast.Data) != 5 {
			t.Errorf("unexpected length: %d", len(forecast.Data))
		}

		for i := 0; i < 5; i++ {
			expected := (0.1 * float32(i)) + float32(loc)
			diff := math.Abs(float64(forecast.Data[i] - expected))
			if diff > 0.0001 {
				t.Fatalf("unexpected value: %f", forecast.Data[i])
			}
		}
	}
	if _, err := reader.Read(100); err == nil {
		t.Error("expected error")
	}
}

func getBlobClient() (values.Reader, error) {
	source := sampleblob.Get()
	meta, err := source.GetMeta(context.Background(), "a", 1)
	if err != nil {
		return nil, err
	}
	config := make(map[string]interface{})

	reader, err := Download(
		context.Background(),
		source,
		meta,
		"gridid",
		config,
	)
	if err != nil {
		return nil, err
	}
	return reader, nil
}
