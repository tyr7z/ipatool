package emulator

import (
	"context"
	"testing"

	"github.com/majd/ipatool/v2/internal/sap/assets"
	"github.com/majd/ipatool/v2/internal/sap/unicorn"
	"github.com/majd/ipatool/v2/internal/store/profile"
)

func TestHostKeepAliveLifecycle(t *testing.T) {
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
	name := guest.New("NSString", "purchase test")
	for _, selector := range []string{"keepAliveWithName:", "keepAliveWithName:style:"} {
		if err := guest.engine.RegWrite(unicorn.RegX0+2, name); err != nil {
			t.Fatal(err)
		}
		if err := guest.engine.RegWrite(unicorn.RegX0+3, 2); err != nil {
			t.Fatal(err)
		}
		guest.message(guest.ClassAddress("AMSRBSKeepAlive"), selector, false)
		object, _ := guest.engine.RegRead(unicorn.RegX0)
		if guest.objects[object] == nil {
			t.Fatal("factory did not return a host assertion")
		}
		guest.message(object, "name", false)
		got, _ := guest.engine.RegRead(unicorn.RegX0)
		if got != name {
			t.Fatal("name was not preserved")
		}
		guest.message(object, "style", false)
		got, _ = guest.engine.RegRead(unicorn.RegX0)
		wantStyle := uint64(0)
		if selector == "keepAliveWithName:style:" {
			wantStyle = 2
		}
		if got != wantStyle {
			t.Fatalf("style = %d, want %d", got, wantStyle)
		}
		guest.message(object, "invalidate", false)
		guest.message(object, "invalidate", false)
		if guest.objects[object].Properties["invalidated"] != 1 {
			t.Fatal("assertion was not invalidated")
		}
	}
	if guest.keepAliveMessage(0, "AMSBuyTask", "run", true) {
		t.Fatal("intercepted purchase logic")
	}
}
