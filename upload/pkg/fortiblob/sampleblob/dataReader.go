package sampleblob

import (
	"bytes"
	"encoding/binary"
	"io"
)

type dataReader struct {
	reader io.Reader
	size   int64
}

func sampleDataReader(data interface{}, elements, size int) *dataReader {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, data); err != nil {
		panic(err)
	}
	return &dataReader{
		reader: bytes.NewReader(buf.Bytes()),
		size:   int64(elements * size),
	}
}

func (r *dataReader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *dataReader) Close() error {
	r.reader = nil
	return nil
}

func (r *dataReader) Size() int64 {
	return r.size
}
