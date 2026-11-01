package cmd

import (
	"errors"
	"io"

	"github.com/majd/ipatool/v2/pkg/appstore"
	"github.com/majd/ipatool/v2/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Search country override", func() {
	var store *fakeSearchAppStore
	BeforeEach(func() {
		previous := dependencies
		DeferCleanup(func() { dependencies = previous })
		store = &fakeSearchAppStore{}
		dependencies.AppStore = store
		dependencies.Logger = log.NewLogger(log.Args{Writer: io.Discard})
	})
	execute := func(args ...string) error {
		cmd := searchCmd()
		cmd.SetArgs(args)
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		return cmd.Execute()
	}
	It("searches an explicit country without loading an account", func() {
		store.accountError = errors.New("not signed in")
		Expect(execute("minecraft", "-c", "JP", "--platform", "macos", "--limit", "10")).To(Succeed())
		Expect(store.accountCalls).To(BeZero())
		Expect(store.inputs).To(Equal([]appstore.SearchInput{{Term: "minecraft", Country: "JP", Platform: appstore.PlatformMacOS, Limit: 10}}))
	})
	It("keeps the account storefront as the default", func() {
		Expect(execute("minecraft")).To(Succeed())
		Expect(store.accountCalls).To(Equal(1))
		Expect(store.inputs).To(Equal([]appstore.SearchInput{{Term: "minecraft", Limit: 5, Account: appstore.Account{StoreFront: "143441"}}}))
	})
	It("defaults to US when no account is available", func() {
		store.accountError = errors.New("not signed in")
		Expect(execute("minecraft")).To(Succeed())
		Expect(store.inputs).To(Equal([]appstore.SearchInput{{Term: "minecraft", Country: "US", Limit: 5}}))
	})
	It("rejects an empty explicit country", func() {
		Expect(execute("minecraft", "-c", " ")).NotTo(Succeed())
		Expect(store.inputs).To(BeEmpty())
		Expect(store.accountCalls).To(BeZero())
	})
	It("preserves search errors", func() {
		store.searchError = errors.New("catalog unavailable")
		Expect(execute("minecraft", "-c", "JP")).To(MatchError(store.searchError))
	})
})

type fakeSearchAppStore struct {
	appstore.AppStore
	accountCalls int
	accountError error
	searchError  error
	inputs       []appstore.SearchInput
}

func (f *fakeSearchAppStore) AccountInfo() (appstore.AccountInfoOutput, error) {
	f.accountCalls++

	return appstore.AccountInfoOutput{Account: appstore.Account{StoreFront: "143441"}}, f.accountError
}

func (f *fakeSearchAppStore) Search(input appstore.SearchInput) (appstore.SearchOutput, error) {
	f.inputs = append(f.inputs, input)

	return appstore.SearchOutput{}, f.searchError
}
