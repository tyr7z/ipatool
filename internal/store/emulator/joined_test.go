package emulator

import (
	"context"
	"strings"
	"testing"

	"github.com/majd/ipatool/v2/internal/sap/assets"
	"github.com/majd/ipatool/v2/internal/sap/unicorn"
	"github.com/majd/ipatool/v2/internal/store/profile"
)

func TestSwiftStringArrayJoined(t *testing.T) {
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
	metadata := guest.alloc(24)
	guest.write64(metadata+16, guest.profile.Symbols["$sSSN"])
	for _, test := range []struct {
		values    []string
		separator string
	}{
		{nil, ", "}, {[]string{"RBSAssertion"}, ", "},
		{[]string{"RBSAssertion", "os.Transaction"}, ", "},
		{[]string{"", "café", "a longer native Swift string"}, " / "},
		{[]string{"a", "b"}, ""},
	} {
		array := guest.alloc(uint64(32 + 16*len(test.values)))
		guest.write64(array+16, uint64(len(test.values)))
		for i, value := range test.values {
			guest.returnSwiftString(value)
			first, _ := guest.engine.RegRead(unicorn.RegX0)
			second, _ := guest.engine.RegRead(unicorn.RegX1)
			guest.write64(array+32+uint64(i)*16, first)
			guest.write64(array+40+uint64(i)*16, second)
		}
		self := guest.alloc(8)
		guest.write64(self, array)
		guest.returnSwiftString(test.separator)
		if err := guest.engine.RegWrite(unicorn.RegX0+20, self); err != nil {
			t.Fatal(err)
		}
		if err := guest.engine.RegWrite(unicorn.RegX0+2, metadata); err != nil {
			t.Fatal(err)
		}
		if !guest.swiftSymbol("$sSKsSS7ElementRtzrlE6joined9separatorS2S_tF") {
			t.Fatal("joined import unhandled")
		}
		first, _ := guest.engine.RegRead(unicorn.RegX0)
		second, _ := guest.engine.RegRead(unicorn.RegX1)
		if got, want := guest.swiftString(first, second), strings.Join(test.values, test.separator); got != want {
			t.Fatalf("joined = %q, want %q", got, want)
		}
	}
}
