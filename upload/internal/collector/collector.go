package collector

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"

	"gitlab.met.no/forti/f2/upload/internal/collector/hash"
	"gitlab.met.no/forti/f2/upload/internal/collector/upload"
	"gitlab.met.no/forti/f2/upload/internal/modelprovider"
	"gitlab.met.no/forti/f2/upload/pkg/collector"
	"gocloud.dev/blob"
)

func Get(ctx context.Context, blobIn, blobOut, group string, version int) error {
	in, err := modelprovider.NewBlobClient(blobIn)
	if err != nil {
		return fmt.Errorf("unable to connect to input source: %w", err)
	}
	defer in.Close()

	bucket, err := blob.OpenBucket(ctx, blobOut)
	if err != nil {
		return fmt.Errorf("unable to connect to output source: %w", err)
	}
	defer bucket.Close()
	out := upload.New(bucket)

	dataset, err := in.GetDataSet(ctx, group, version)
	if err != nil {
		return err
	}

	hashes, err := getHashes(ctx, in, &dataset)
	if err != nil {
		return err
	}
	hashCollector := hash.New(ctx, in, out, group, version)

	for hash, parameters := range hashes {
		log.Println(hash)

		var rg []modelprovider.Meta
		for _, parameter := range parameters {
			id := modelprovider.ID{
				Group:   group,
				Version: version,
				Param:   parameter,
			}
			meta, err := in.Meta(ctx, id)
			if err != nil {
				return err
			}
			rg = append(rg, meta)
		}

		if err := hashCollector.Collect(ctx, hash, rg); err != nil {
			return err
		}
	}

	meta := collector.DatasetMeta{
		Group:         group,
		Version:       version,
		TimeUntilNext: dataset.TimeUntilNext,
	}
	return out.SetDatasetMeta(ctx, &meta)
}

func getHashes(ctx context.Context, in *modelprovider.Client, dataset *modelprovider.DataSet) (map[string][]string, error) {
	geos := make(map[string][]string)

	for _, parameter := range dataset.Parameters {
		id := modelprovider.ID{
			Group:   dataset.Group,
			Version: dataset.Version,
			Param:   parameter,
		}
		lat, err := in.Latitude(ctx, id)
		if err != nil {
			return nil, err
		}
		defer lat.Close()

		checksumStream := md5.New()
		if _, err := io.Copy(checksumStream, lat); err != nil {
			return nil, err
		}

		checksum := hex.EncodeToString(checksumStream.Sum(nil))
		geos[checksum] = append(geos[checksum], parameter)
	}
	return geos, nil
}
