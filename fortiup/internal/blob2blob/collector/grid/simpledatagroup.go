package grid

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"gitlab.met.no/forti/f2/upload/internal/blob2blob/modelprovider"
	"gitlab.met.no/forti/f2/upload/pkg/fortiblob"
)

func collectSimpleDataGroup(ctx context.Context, out io.Writer, source *modelprovider.Client, group string, version int, meta []modelprovider.Meta) (*fortiblob.MetaCollection, error) {
	readers := make(map[string]io.Reader)

	for _, m := range meta {
		id := modelprovider.ID{
			Group:   group,
			Version: version,
			Param:   m.Parameter,
		}
		r, err := source.Data(ctx, id)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		readers[m.Parameter] = bufio.NewReader(r)
	}

	bufferedOut := bufio.NewWriter(out)

	if err := collect(ctx, bufferedOut, readers, meta); err != nil {
		return nil, err
	}

	if err := bufferedOut.Flush(); err != nil {
		return nil, err
	}

	pMeta := make(map[string]fortiblob.ParameterMeta)
	var elements int
	for _, m := range meta {
		if len(m.Times) == 0 {
			m.Times = append(m.Times, time.Time{})
		}
		idx := fortiblob.ParameterMeta{
			SliceFrom: elements,
			Times:     m.Times,
			Units:     m.Units,
		}
		pMeta[m.Parameter] = idx
		elements += len(m.Times)
	}

	return &fortiblob.MetaCollection{
		Parameters:    pMeta,
		LocationCount: elements,
	}, nil
}

func collect(ctx context.Context, out io.Writer, readers map[string]io.Reader, meta []modelprovider.Meta) error {
	for {
		for _, m := range meta {
			reader := readers[m.Parameter]

			count := len(m.Times)
			if count == 0 {
				count = 1
			}

			uncompressed := make([]float32, count)
			if err := binary.Read(reader, binary.LittleEndian, &uncompressed); err != nil {
				if errors.Is(err, io.EOF) {
					return verifyAllAtEnd(meta, readers)
				}
				return err
			}

			compressed := make([]int16, count)
			for i, val := range uncompressed {
				compressed[i] = int16(math.Round(float64(val * 10)))
			}

			if err := binary.Write(out, binary.LittleEndian, compressed); err != nil {
				return err
			}
		}
	}
}

func verifyAllAtEnd(meta []modelprovider.Meta, inputs map[string]io.Reader) error {
	buf := make([]byte, 1)
	for _, parameter := range meta {
		_, err := inputs[parameter.Parameter].Read(buf)
		if err == nil {
			return fmt.Errorf("%s is not at end of input", parameter.Parameter)
		}
		if !errors.Is(err, io.EOF) {
			return err
		}
	}
	return nil
}
