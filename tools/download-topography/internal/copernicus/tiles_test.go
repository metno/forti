package copernicus

import (
	"testing"
)

func TestGenerateTileNames(t *testing.T) {
	tests := []struct {
		name     string
		bbox     BoundingBox
		expected int // number of tiles
	}{
		{
			name:     "single tile",
			bbox:     BoundingBox{LatMin: 0, LatMax: 0, LonMin: 0, LonMax: 0},
			expected: 1,
		},
		{
			name:     "2x2 grid",
			bbox:     BoundingBox{LatMin: 0, LatMax: 1, LonMin: 0, LonMax: 1},
			expected: 4,
		},
		{
			name:     "malawi",
			bbox:     BoundingBox{LatMin: -17, LatMax: -9, LonMin: 32, LonMax: 36},
			expected: 45, // 9 lat * 5 lon
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tiles := GenerateTileNames(tt.bbox)
			if len(tiles) != tt.expected {
				t.Errorf("GenerateTileNames() got %d tiles, want %d", len(tiles), tt.expected)
			}
		})
	}
}

func TestFormatTileName(t *testing.T) {
	tests := []struct {
		lat      int
		lon      int
		expected string
	}{
		{0, 0, "Copernicus_DSM_COG_10_N00_00_E000_00_DEM"},
		{10, 20, "Copernicus_DSM_COG_10_N10_00_E020_00_DEM"},
		{-10, -20, "Copernicus_DSM_COG_10_S10_00_W020_00_DEM"},
		{59, 11, "Copernicus_DSM_COG_10_N59_00_E011_00_DEM"},
		{-17, 32, "Copernicus_DSM_COG_10_S17_00_E032_00_DEM"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatTileName(tt.lat, tt.lon)
			if result != tt.expected {
				t.Errorf("formatTileName(%d, %d) = %s, want %s", tt.lat, tt.lon, result, tt.expected)
			}
		})
	}
}

func TestPredefinedRegions(t *testing.T) {
	// Verify all predefined regions exist and have valid bounds
	requiredRegions := []string{"africa", "kenya", "malawi", "nigeria", "ethiopia", "tanzania", "south-africa"}

	for _, name := range requiredRegions {
		region, ok := PredefinedRegions[name]
		if !ok {
			t.Errorf("Missing predefined region: %s", name)
			continue
		}

		if region.LatMin >= region.LatMax {
			t.Errorf("Region %s has invalid latitude bounds: min=%f, max=%f", name, region.LatMin, region.LatMax)
		}
		if region.LonMin >= region.LonMax {
			t.Errorf("Region %s has invalid longitude bounds: min=%f, max=%f", name, region.LonMin, region.LonMax)
		}
	}
}
