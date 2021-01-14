package fortiblob

import (
	"fmt"
)

func path(area string, version int, hash, name string) string {
	return fmt.Sprintf("%s/%d/%s/%s", area, version, hash, name)
}

// func SetAvailable(ctx context.Context, bucket *blob.Bucket, area string, version int) error {
// 	path := fmt.Sprintf("%s/%v/meta.json", area, version)
// 	w, err := bucket.NewWriter(ctx, path, nil)
// 	if err != nil {
// 		return err
// 	}
// 	if _, err := fmt.Fprintf(w, "%s\n", "hello"); err != nil {
// 		return err
// 	}
// 	return w.Close()
// }
