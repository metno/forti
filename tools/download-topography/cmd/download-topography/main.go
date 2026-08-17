package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/metno/forti/tools/download-topography/internal/copernicus"
)

func main() {
	region := flag.String("region", "", "Predefined region (africa, kenya, malawi, nigeria, ethiopia, tanzania, south-africa)")
	latMin := flag.Float64("lat-min", 0, "Minimum latitude (south)")
	latMax := flag.Float64("lat-max", 0, "Maximum latitude (north)")
	lonMin := flag.Float64("lon-min", 0, "Minimum longitude (west)")
	lonMax := flag.Float64("lon-max", 0, "Maximum longitude (east)")
	output := flag.String("output", "", "Output directory for topography data (required)")
	dryRun := flag.Bool("dry-run", false, "Show what would be downloaded without downloading")
	workers := flag.Int("workers", 4, "Number of parallel download workers")
	flag.Parse()

	if *output == "" {
		fmt.Println("Error: --output is required")
		flag.Usage()
		os.Exit(1)
	}

	// Determine bounding box
	var bbox copernicus.BoundingBox
	var regionName string

	if *region != "" {
		r, ok := copernicus.PredefinedRegions[*region]
		if !ok {
			fmt.Printf("Error: unknown region '%s'\n", *region)
			fmt.Println("\nAvailable regions:")
			for name, r := range copernicus.PredefinedRegions {
				fmt.Printf("  %-20s lat: [%3.0f, %3.0f], lon: [%3.0f, %3.0f]\n",
					name, r.LatMin, r.LatMax, r.LonMin, r.LonMax)
			}
			os.Exit(1)
		}
		bbox = r
		regionName = *region
		fmt.Printf("Using predefined region: %s\n", *region)
	} else if *latMin != 0 || *latMax != 0 || *lonMin != 0 || *lonMax != 0 {
		bbox = copernicus.BoundingBox{
			LatMin: *latMin,
			LatMax: *latMax,
			LonMin: *lonMin,
			LonMax: *lonMax,
		}
		regionName = "custom"
		fmt.Println("Using custom bounding box")
	} else {
		fmt.Println("Error: Either --region or all of --lat-min, --lat-max, --lon-min, --lon-max must be specified")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Bounding box: lat [%.1f, %.1f], lon [%.1f, %.1f]\n",
		bbox.LatMin, bbox.LatMax, bbox.LonMin, bbox.LonMax)

	// Generate tile list
	tiles := copernicus.GenerateTileNames(bbox)
	fmt.Printf("\nFound %d tiles to download\n", len(tiles))

	if *dryRun {
		fmt.Println("\nDRY RUN - would download:")
		for _, tile := range tiles {
			fmt.Printf("  %s\n", tile)
		}
		fmt.Printf("\nOutput directory: %s\n", *output)
		return
	}

	// Create output directory
	if err := os.MkdirAll(*output, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Download tiles
	fmt.Printf("\nDownloading to: %s\n", *output)
	fmt.Printf("Using %d parallel workers\n\n", *workers)

	downloader := copernicus.NewDownloader(*output, *workers)
	stats, err := downloader.DownloadTiles(tiles)
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	// Print summary
	fmt.Println()
	fmt.Printf("Region:          %s\n", regionName)
	fmt.Printf("Total tiles:     %d\n", stats.Total)
	fmt.Printf("Downloaded:      %d\n", stats.Downloaded)
	fmt.Printf("Not available:   %d (ocean tiles)\n", stats.NotAvailable)
	fmt.Printf("Failed:          %d\n", stats.Failed)
	fmt.Printf("Output:          %s\n", *output)

	if stats.Failed > 0 {
		os.Exit(1)
	}
}
