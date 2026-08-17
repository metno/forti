package copernicus

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

// DownloadStats tracks download statistics
type DownloadStats struct {
	Total        int
	Downloaded   int
	NotAvailable int
	Failed       int
}

// Downloader handles parallel downloading of Copernicus DEM tiles
type Downloader struct {
	outputDir string
	workers   int
}

// NewDownloader creates a new downloader
func NewDownloader(outputDir string, workers int) *Downloader {
	return &Downloader{
		outputDir: outputDir,
		workers:   workers,
	}
}

// DownloadTiles downloads all tiles using parallel workers
func (d *Downloader) DownloadTiles(tiles []string) (*DownloadStats, error) {
	// Check for AWS CLI
	if err := checkAWSCLI(); err != nil {
		return nil, err
	}

	stats := &DownloadStats{Total: len(tiles)}
	var downloaded, notAvailable, failed atomic.Int32

	// Create job queue
	jobs := make(chan string, len(tiles))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < d.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for tile := range jobs {
				status, err := d.downloadTile(tile)
				if err != nil {
					fmt.Printf("  [%d] ✗ %s: %v\n", workerID, tile, err)
					failed.Add(1)
				} else {
					switch status {
					case StatusDownloaded:
						fmt.Printf("  [%d] ✓ %s\n", workerID, tile)
						downloaded.Add(1)
					case StatusNotAvailable:
						fmt.Printf("  [%d] ⊘ %s (not available - likely ocean)\n", workerID, tile)
						notAvailable.Add(1)
					}
				}
			}
		}(i)
	}

	// Queue all jobs
	for _, tile := range tiles {
		jobs <- tile
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()

	stats.Downloaded = int(downloaded.Load())
	stats.NotAvailable = int(notAvailable.Load())
	stats.Failed = int(failed.Load())

	return stats, nil
}

// DownloadStatus represents the result of a download attempt
type DownloadStatus int

const (
	StatusDownloaded DownloadStatus = iota
	StatusNotAvailable
	StatusFailed
)

// downloadTile downloads a single tile from S3
// Downloads only the main DEM .tif file directly to the output directory
func (d *Downloader) downloadTile(tileName string) (DownloadStatus, error) {
	// The main DEM file in each tile directory
	mainFile := tileName + ".tif"
	s3Path := fmt.Sprintf("s3://copernicus-dem-30m/%s/%s", tileName, mainFile)
	destPath := filepath.Join(d.outputDir, mainFile)

	// Create destination directory
	if err := os.MkdirAll(d.outputDir, 0755); err != nil {
		return StatusFailed, fmt.Errorf("creating destination directory: %w", err)
	}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "aws", "s3", "cp",
		"--no-sign-request",
		s3Path,
		destPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if tile doesn't exist (ocean tiles)
		outputStr := string(output)
		if strings.Contains(outputStr, "does not exist") ||
			strings.Contains(outputStr, "NoSuchKey") ||
			strings.Contains(outputStr, "The specified key does not exist") {
			return StatusNotAvailable, nil
		}
		return StatusFailed, fmt.Errorf("%w: %s", err, outputStr)
	}

	return StatusDownloaded, nil
}

// checkAWSCLI verifies that AWS CLI is installed
func checkAWSCLI() error {
	cmd := exec.Command("aws", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("AWS CLI not found. Please install it:\n" +
			"  macOS:   brew install awscli\n" +
			"  Linux:   apt-get install awscli  or  yum install aws-cli\n" +
			"  pip:     pip install awscli")
	}
	return nil
}
