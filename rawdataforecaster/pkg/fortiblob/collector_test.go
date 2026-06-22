package fortiblob

import (
	"encoding/json"
	"fmt"
)

// ExampleMetaCollection shows how to use MetaCollection to interpret raw data.
func ExampleMetaCollection() {
	sampleData := []int16{
		142, 139, 92, 0, 1,
	}
	metaJSON := `{
  "parameters": {
    "temperature": {
      "units": "c",
      "times": ["2021-03-03T00:00:00Z", "2021-03-03T01:00:00Z"],
      "slice_from": 0,
      "scale_factor": 0.1
    },
    "altitude": {
      "units": "m",
      "times": ["0001-01-01T00:00:00Z"],
      "slice_from": 2,
      "scale_factor": 1
    },
    "precipitation": {
      "units": "kg/m²",
      "times": ["2021-03-03T00:00:00Z", "2021-03-03T01:00:00Z"],
      "slice_from": 3,
      "scale_factor": 0.1
    }
  },
  "number_of_points": 5
}
`

	var meta MetaCollection
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		panic(err)
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
