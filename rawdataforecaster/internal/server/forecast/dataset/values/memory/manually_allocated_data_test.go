package memory

import "testing"

func TestManuallyAllocatedData(t *testing.T) {
	data := allocate(10)
	if len(data.Values) != 10 {
		t.Fatalf("invalid slice size: %d", len(data.Values))
	}
	for i := 0; i < 10; i++ {
		data.Values[i] = int16(i * 10)
	}
	if data.Values[5] != 50 {
		t.Error("invalid value")
	}
	data.Free()
	if data.Values != nil {
		t.Errorf("data was not freed: %v", data.Values)
	}
}
