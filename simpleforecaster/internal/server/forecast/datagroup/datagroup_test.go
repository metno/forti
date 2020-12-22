package datagroup

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"gitlab.met.no/forti/f2/simpleforecaster/internal/server/fortiblob"
// )

// func Test(t *testing.T) {
// 	store := fortiblob.MakeTestingBlob()
// 	fortiblob.AddToBlob(store,
// 		"grup", 2, "hashA",
// 		map[string]int{"foo": 2, "bar": 1},
// 		[]float32{59, 59, 60, 60},
// 		[]float32{10, 11, 10, 11},
// 	)
// 	fortiblob.AddToBlob(store,
// 		"grup", 2, "hashB",
// 		map[string]int{"bik": 4, "bok": 0},
// 		[]float32{59, 59, 60},
// 		[]float32{10, 11, 10},
// 	)

// 	ctx := context.Background()
// 	dataset, err := Download(ctx, store, "grup", 2, "")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer dataset.Close()

// 	pointdata, err := dataset.Read(59, 11)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	foo := pointdata.GetData("foo")
// 	if foo == nil {
// 		t.Fatal("could not find foo")
// 	}
// 	if len(foo) != 2 {
// 		t.Fatalf("invalid size: %d", len(foo))
// 	}

// 	if val, ok := foo[time.Date(2020, 12, 24, 0, 0, 0, 0, time.UTC)]; val != 100 {
// 		if !ok {
// 			t.Error("could not find time")
// 		} else {
// 			t.Errorf("unexpected value: %f", val)
// 		}
// 	}
// 	if val, ok := foo[time.Date(2020, 12, 24, 1, 0, 0, 0, time.UTC)]; val != 101 {
// 		if !ok {
// 			t.Error("could not find time")
// 		} else {
// 			t.Errorf("unexpected value: %f", val)
// 		}
// 	}
// }
