package fortiblob

import (
	"fmt"
	"time"
)

// ExampleMetaCollection shows how to use MetaCollection to interpret raw data.
func ExampleMetaCollection() {
	sampleData := []int16{
		142, 139, 92, 0, 1,
	}
	meta := MetaCollection{
		Parameters: map[string]ParameterMeta{
			"temperature": {
				Units: "c",
				Times: []time.Time{
					time.Date(2021, 3, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2021, 3, 3, 1, 0, 0, 0, time.UTC),
				},
				SliceFrom:   0,
				ScaleFactor: 0.1,
			},
			"altitude": {
				Units: "m",
				Times: []time.Time{
					{},
				},
				SliceFrom:   2,
				ScaleFactor: 1,
			},
			"precipitation": {
				Units: "kg/m²",
				Times: []time.Time{
					time.Date(2021, 3, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2021, 3, 3, 1, 0, 0, 0, time.UTC),
				},
				SliceFrom:   3,
				ScaleFactor: 0.1,
			},
		},
		LocationCount: len(sampleData),
	}

	for parameter, meta := range meta.Parameters {
		fmt.Print(parameter)
		for i := range meta.Times {
			idx := meta.SliceFrom + i
			value := float32(sampleData[idx]) * meta.ScaleFactor
			fmt.Printf(" %.1f", value)
		}
		fmt.Println()
	}
	// Unordered output:
	// temperature 14.2 13.9
	// altitude 92.0
	// precipitation 0.0 0.1
}
