package emulator

import (
	"context"
	"testing"

	"github.com/majd/ipatool/v2/internal/sap/assets"
	"github.com/majd/ipatool/v2/internal/sap/unicorn"
	"github.com/majd/ipatool/v2/internal/store/profile"
)

func TestSwiftMainQueue(t *testing.T) {
	ctx := context.Background()
	bundle, err := assets.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := profile.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	guest, err := open(ctx, bundle.AppleMediaServices, prepared, true)
	if err != nil {
		t.Fatal(err)
	}
	defer guest.Close()
	expected := guest.dispatchMainQueue()
	for i := 0; i < 2; i++ {
		if !guest.swiftSymbol("$sSo17OS_dispatch_queueC8DispatchE4mainABvgZ") {
			t.Fatal("main queue import unhandled")
		}
		got, err := guest.engine.RegRead(unicorn.RegX0)
		if err != nil {
			t.Fatal(err)
		}
		if got != expected {
			t.Fatalf("main queue = %#x, want %#x", got, expected)
		}
		if guest.objects[got].Class != "HostQueue" {
			t.Fatal("main queue is not a host queue")
		}
	}
}
