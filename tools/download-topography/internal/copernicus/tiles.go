package copernicus

import (
	"fmt"
	"math"
)

// BoundingBox defines a geographic area
type BoundingBox struct {
	LatMin float64
	LatMax float64
	LonMin float64
	LonMax float64
}

// PredefinedRegions contains common African regions
var PredefinedRegions = map[string]BoundingBox{
	"africa":       {LatMin: -35, LatMax: 37, LonMin: -18, LonMax: 52},
	"south-africa": {LatMin: -35, LatMax: -22, LonMin: 16, LonMax: 33},
	"kenya":        {LatMin: -5, LatMax: 5, LonMin: 33, LonMax: 42},
	"nigeria":      {LatMin: 4, LatMax: 14, LonMin: 2, LonMax: 15},
	"ethiopia":     {LatMin: 3, LatMax: 15, LonMin: 33, LonMax: 48},
	"tanzania":     {LatMin: -12, LatMax: -1, LonMin: 29, LonMax: 41},
	"malawi":       {LatMin: -17, LatMax: -9, LonMin: 32, LonMax: 36},
}

// GenerateTileNames generates Copernicus DEM tile names for a bounding box.
// Tiles are 1°×1° and named by their southwest corner.
// Format: Copernicus_DSM_COG_10_N00_00_E040_00_DEM
func GenerateTileNames(bbox BoundingBox) []string {
	tiles := []string{}

	latMin := int(math.Floor(bbox.LatMin))
	latMax := int(math.Floor(bbox.LatMax))
	lonMin := int(math.Floor(bbox.LonMin))
	lonMax := int(math.Floor(bbox.LonMax))

	for lat := latMin; lat <= latMax; lat++ {
		for lon := lonMin; lon <= lonMax; lon++ {
			tiles = append(tiles, formatTileName(lat, lon))
		}
	}

	return tiles
}

// formatTileName formats a tile name from lat/lon coordinates
func formatTileName(lat, lon int) string {
	var latStr, lonStr string

	// Format latitude
	if lat >= 0 {
		latStr = fmt.Sprintf("N%02d_00", lat)
	} else {
		latStr = fmt.Sprintf("S%02d_00", -lat)
	}

	// Format longitude
	if lon >= 0 {
		lonStr = fmt.Sprintf("E%03d_00", lon)
	} else {
		lonStr = fmt.Sprintf("W%03d_00", -lon)
	}

	return fmt.Sprintf("Copernicus_DSM_COG_10_%s_%s_DEM", latStr, lonStr)
}
