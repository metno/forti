package forecast

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"gitlab.met.no/forti/f2/simpleforecaster/internal/server/fortiblob"
// )

// func Test(t *testing.T) {
// 	store := fortiblob.MakeTestingBlob()
// 	fortiblob.AddToBlob(store,
// 		"groupA", 2, "hashA",
// 		map[string]int{"foo": 2, "bar": 1},
// 		[]float32{59, 59, 60, 60},
// 		[]float32{10, 11, 10, 11},
// 	)
// 	fortiblob.AddToBlob(store,
// 		"groupB", 2, "hashB",
// 		map[string]int{"foo": 2, "bar": 1},
// 		[]float32{5, 5, 6, 6},
// 		[]float32{1, 1, 1, 1},
// 	)
// 	if err := fortiblob.SetAvailable(context.TODO(), store, "groupA", 2); err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := fortiblob.SetAvailable(context.TODO(), store, "groupB", 2); err != nil {
// 		t.Fatal(err)
// 	}

// 	forecast, err := New(store, []string{"groupA", "groupB"}, "")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	pointdata, err := forecast.Get(59, 11)
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
