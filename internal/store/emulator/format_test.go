package emulator

import (
	"context"
	"testing"

	"github.com/majd/ipatool/v2/internal/store/profile"
)

func TestFormatObject(t *testing.T) {
	guest := &Runtime{
		classNames: map[uint64]string{1: "AMSTask"},
		objects:    map[uint64]*Object{2: {Class: "NSString", Value: "example"}, 3: {Class: "NSNumber", Value: int64(42)}},
	}
	for _, test := range []struct {
		address uint64
		want    string
	}{{0, "(null)"}, {1, "AMSTask"}, {2, "example"}, {3, "42"}} {
		if got := guest.formatObject(test.address); got != test.want {
			t.Errorf("formatObject(%#x) = %q, want %q", test.address, got, test.want)
		}
	}
}

func TestPurchaseDialogClassFormatting(t *testing.T) {
	data, err := profile.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	cache := metadataCache{gate: make(chan struct{}, 1)}
	metadata, err := cache.load(context.Background(), data.Metadata)
	if err != nil {
		t.Fatal(err)
	}
	guest := metadata.newRuntime()
	address := guest.classes["AMSHandleDialogResultTask"].Address
	if got := guest.formatObject(address); got != "AMSHandleDialogResultTask" {
		t.Fatalf("class description = %q", got)
	}
}
