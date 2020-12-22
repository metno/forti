package simpledatagroup

// import (
// 	"context"
// 	"testing"

// 	"gitlab.met.no/forti/f2/simpleforecaster/internal/server/fortiblob"
// )

// func Test(t *testing.T) {
// 	store := fortiblob.MakeTestingBlob()
// 	fortiblob.AddToBlob(store,
// 		"group", 1, "hash",
// 		map[string]int{"foo": 2, "bar": 1},
// 		[]float32{59, 59, 60, 60},
// 		[]float32{10, 11, 10, 11},
// 	)

// 	d := NewDownloader(store, "")

// 	ctx := context.Background()
// 	reader, err := d.Get(ctx, "group", 1, "hash")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer reader.Close()

// 	if _, err := reader.Read(5); err == nil {
// 		t.Error("expected an out-of-bounds error")
// 	}

// 	data, err := reader.Read(0)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if len(data.Data) != 3 {
// 		t.Errorf("invalid length: %d", len(data.Data))
// 	}
// 	if data.Data[0] != 0.0 {
// 		t.Errorf("unexpected data: %f", data.Data[0])
// 	}
// 	if data.Data[1] != 1.0 {
// 		t.Errorf("unexpected data: %f", data.Data[1])
// 	}
// 	if data.Data[2] != 20.0 {
// 		t.Errorf("unexpected data: %f", data.Data[2])
// 	}

// }
