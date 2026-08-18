package copernicus

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	outputDir  string
	workers    int
	httpClient *http.Client
}

// NewDownloader creates a new downloader
func NewDownloader(outputDir string, workers int) *Downloader {
	return &Downloader{
		outputDir:  outputDir,
		workers:    workers,
		httpClient: &http.Client{},
	}
}

// DownloadTiles downloads all tiles using parallel workers
func (d *Downloader) DownloadTiles(tiles []string) (*DownloadStats, error) {

	// Create destination directory
	if err := os.MkdirAll(d.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating destination directory: %w", err)
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
				status, err := d.downloadTile(context.Background(), tile)
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

// downloadTile downloads a single tile via HTTP
// Downloads only the main DEM .tif file directly to the output directory
func (d *Downloader) downloadTile(ctx context.Context, tileName string) (DownloadStatus, error) {
	// The main DEM file
	mainFile := tileName + ".tif"
	// Public HTTP URL for Copernicus DEM
	url := fmt.Sprintf("https://copernicus-dem-30m.s3.eu-central-1.amazonaws.com/%s/%s", tileName, mainFile)
	destPath := filepath.Join(d.outputDir, mainFile)

	// Check if file already exists by getting HEAD
	headReq, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return StatusFailed, fmt.Errorf("creating HEAD request: %w", err)
	}

	headResp, err := d.httpClient.Do(headReq)
	if err != nil {
		return StatusFailed, fmt.Errorf("HEAD request failed: %w", err)
	}
	headResp.Body.Close()

	if headResp.StatusCode == http.StatusNotFound || headResp.StatusCode == http.StatusForbidden {
		// Tile doesn't exist (ocean area)
		return StatusNotAvailable, nil
	}
	if headResp.StatusCode != http.StatusOK {
		return StatusFailed, fmt.Errorf("unexpected status code: %d", headResp.StatusCode)
	}

	// Check if local file exists and matches size (idempotent)
	if stat, err := os.Stat(destPath); err == nil && stat.Size() == headResp.ContentLength {
		// File already downloaded and size matches - skip
		return StatusDownloaded, nil
	}

	// Download the file
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return StatusFailed, fmt.Errorf("creating GET request: %w", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return StatusFailed, fmt.Errorf("GET request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
		return StatusNotAvailable, nil
	}
	if resp.StatusCode != http.StatusOK {
		return StatusFailed, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create the file
	file, err := os.Create(destPath)
	if err != nil {
		return StatusFailed, fmt.Errorf("creating local file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, resp.Body); err != nil {
		return StatusFailed, fmt.Errorf("copying data: %w", err)
	}

	return StatusDownloaded, nil
}
