package download

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
)

func TestDownloadValidKeyReturnsFilename(t *testing.T) {
	ctx := context.Background()
	bkt, err := blob.OpenBucket(ctx, "mem://")
	if err != nil {
		t.Fatalf("opening mem bucket: %v", err)
	}
	defer bkt.Close()

	const key = "forecast"
	const content = "abc"
	if err := bkt.WriteAll(ctx, key, []byte(content), nil); err != nil {
		t.Fatalf("writing object: %v", err)
	}

	baseDir := t.TempDir()
	toDir := filepath.Join(baseDir, "data")
	if err := os.Mkdir(toDir, 0o755); err != nil {
		t.Fatalf("creating target dir: %v", err)
	}

	obj := &blob.ListObject{
		Key:  key,
		Size: int64(len(content)),
	}

	filename, err := download(ctx, bkt, obj, toDir)
	if err != nil {
		t.Fatalf("download returned error: %v", err)
	}

	expected := filepath.Join(toDir, key)
	if filename != expected {
		t.Fatalf("unexpected filename: got %q, want %q", filename, expected)
	}

	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("opening downloaded file: %v", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if string(b) != content {
		t.Fatalf("unexpected file content: got %q, want %q", string(b), content)
	}
}

func TestDownloadRejectsPathTraversalKey(t *testing.T) {
	baseDir := t.TempDir()
	toDir := filepath.Join(baseDir, "data")
	if err := os.Mkdir(toDir, 0o755); err != nil {
		t.Fatalf("creating target dir: %v", err)
	}

	obj := &blob.ListObject{
		Key:  "../escape.txt",
		Size: 3,
	}

	_, err := download(context.Background(), nil, obj, toDir)
	if err == nil {
		t.Fatal("expected error for path traversal key, got nil")
	}

	escapedPath := filepath.Join(baseDir, "escape.txt")
	if _, statErr := os.Stat(escapedPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected no file to be created outside target dir, stat err: %v", statErr)
	}
}
