package emulator

import (
	"context"
	"testing"

	"github.com/majd/ipatool/v2/internal/sap/assets"
	"github.com/majd/ipatool/v2/internal/sap/unicorn"
	"github.com/majd/ipatool/v2/internal/store/profile"
)

func TestSignpostIDOptionalTag(t *testing.T) {
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
	address := guest.alloc(16)
	guest.write64(address, 0x123456789abcdef0)
	for _, tag := range []uint64{1, 0} {
		for i, value := range []uint64{address, tag, 1} {
			if err := guest.engine.RegWrite(unicorn.RegX0+i, value); err != nil {
				t.Fatal(err)
			}
		}
		if !guest.swiftSymbol("$s2os12OSSignpostIDVwst") {
			t.Fatal("store tag import unhandled")
		}
		if got := guest.read64(address); got != 0x123456789abcdef0 {
			t.Fatalf("tag store changed payload: %#x", got)
		}
		for i, value := range []uint64{address, 1} {
			if err := guest.engine.RegWrite(unicorn.RegX0+i, value); err != nil {
				t.Fatal(err)
			}
		}
		if !guest.swiftSymbol("$s2os12OSSignpostIDVwet") {
			t.Fatal("get tag import unhandled")
		}
		got, err := guest.engine.RegRead(unicorn.RegX0)
		if err != nil {
			t.Fatal(err)
		}
		if got != tag {
			t.Fatalf("optional tag = %d, want %d", got, tag)
		}
	}
}
