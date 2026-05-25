package download

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gocloud.dev/blob"
)

// Get downloads all files from the given blob storage into the given folder.
func Get(ctx context.Context, fromBlob, toDir string) ([]string, error) {
	if fromBlob == "" {
		log.Printf("no blob address given - just using contents of folder %s as data source", toDir)
		return ListDir(toDir)
	}

	bkt, err := blob.OpenBucket(ctx, fromBlob)
	if err != nil {
		return nil, err
	}
	defer bkt.Close()

	var files []string
	it := bkt.List(nil)
	for {
		obj, err := it.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		f, err := download(ctx, bkt, obj, toDir)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	sort.Strings(files)
	return files, nil
}

func ListDir(dir string) ([]string, error) {
	finfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, f := range finfo {
		if !f.IsDir() {
			fname := filepath.Join(dir, f.Name())
			log.Println("serve " + fname)
			files = append(files, fname)
		}
	}

	sort.Strings(files)
	return files, nil
}

func download(ctx context.Context, bkt *blob.Bucket, obj *blob.ListObject, toDir string) (string, error) {
	filename := filepath.Join(toDir, obj.Key)

	// Ensure that the filename is within the target directory, to prevent path traversal attacks.
	absDir, err := filepath.Abs(toDir)
	if err != nil {
		return "", fmt.Errorf("resolving target directory: %w", err)
	}
	absFile, err := filepath.Abs(filename)
	if err != nil {
		return "", fmt.Errorf("resolving target filename: %w", err)
	}
	if !strings.HasPrefix(absFile, absDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("blob key %q escapes target directory", obj.Key)
	}

	stat, err := os.Stat(filename)
	if os.IsNotExist(err) || stat.Size() != obj.Size {
		log.Println("download " + filename)
		f, err := os.Create(filename)
		if err != nil {
			return "", err
		}
		r, err := bkt.NewReader(ctx, obj.Key, nil)
		if err != nil {
			return "", err
		}
		defer r.Close()
		if _, err := io.Copy(f, r); err != nil {
			return "", err
		}
		return filename, nil
	}

	log.Println(filename + " already exists")
	return filename, err
}
