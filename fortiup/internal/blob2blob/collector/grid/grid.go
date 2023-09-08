package grid

import (
	"context"
	"io"
	"log"

	"gitlab.met.no/forti/f2/upload/internal/blob2blob/modelprovider"
	"gitlab.met.no/forti/f2/upload/internal/upload"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

type Collector struct {
	source  *modelprovider.Client
	sink    *upload.Uploader
	group   string
	version int
}

func New(ctx context.Context, source *modelprovider.Client, sink *upload.Uploader, group string, version int) *Collector {
	return &Collector{
		source:  source,
		sink:    sink,
		group:   group,
		version: version,
	}
}

func (c *Collector) Collect(ctx context.Context, gridid string, rg []modelprovider.Meta) error {
	meta, err := c.collectData(ctx, c.group, c.version, gridid, rg)
	if err != nil {
		return err
	}
	if err := c.sink.SetGridMeta(ctx, meta, c.group, c.version, gridid); err != nil {
		return err
	}

	id := modelprovider.ID{
		Group:   c.group,
		Version: c.version,
		Param:   rg[0].Parameter,
	}

	latIn, err := c.source.Latitude(ctx, id)
	if err != nil {
		return err
	}
	defer latIn.Close()
	latOut, err := c.sink.GetLatitudeStream(ctx, c.group, c.version, gridid)
	if err != nil {
		return err
	}
	if _, err := io.Copy(latOut, latIn); err != nil {
		return err
	}
	if err := latOut.Close(); err != nil {
		return err
	}

	lonIn, err := c.source.Longitude(ctx, id)
	if err != nil {
		return err
	}
	defer lonIn.Close()
	lonOut, err := c.sink.GetLongitudeStream(ctx, c.group, c.version, gridid)
	if err != nil {
		return err
	}
	if _, err := io.Copy(lonOut, lonIn); err != nil {
		return err
	}
	if err := lonOut.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Collector) collectData(ctx context.Context, group string, version int, gridid string, rg []modelprovider.Meta) (*fortiblob.MetaCollection, error) {
	out, err := c.sink.GetDataStream(ctx, group, version, gridid)
	if err != nil {
		log.Fatalln(err)
	}
	meta, err := collectSimpleDataGroup(ctx, out, c.source, c.group, c.version, rg)
	if err != nil {
		log.Fatalln(err)
	}
	return meta, out.Close()
}
