package cmd

import (
	"github.com/majd/ipatool/v2/pkg/appstore"
	"github.com/majd/ipatool/v2/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Country flags", func() {
	It("exposes country with the c shorthand on all requested commands", func() {
		for _, name := range []string{"download", "get-version-metadata", "list-purchases", "list-versions", "lookup", "search"} {
			cmd, _, err := rootCmd().Find([]string{name})
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd.Flags().Lookup("country")).NotTo(BeNil(), name)
			Expect(cmd.Flags().Lookup("country").Shorthand).To(Equal("c"), name)
		}
	})
	DescribeTable("uses the requested country after authentication refresh", func(factory func() *cobra.Command, args []string) {
		for _, flag := range []string{"-c", "--country"} {
			store := &countryRetryStore{}
			previous := dependencies
			DeferCleanup(func() { dependencies = previous })
			dependencies.AppStore = store
			dependencies.Logger = log.NewLogger(log.Args{})
			cmd := factory()
			cmd.SetArgs(append(append([]string{}, args...), flag, " jp "))
			Expect(cmd.Execute()).To(Succeed())
			Expect(store.accounts).To(HaveLen(2))
			for _, account := range store.accounts {
				Expect(account.StoreFront).To(Equal("143462"))
			}
			for _, account := range store.lookupAccounts {
				Expect(account.StoreFront).To(Equal("143462"))
			}
			Expect(store.accounts[1].PasswordToken).To(Equal("fresh"))
		}
	},
		Entry("metadata", getVersionMetadataCmd, []string{"-b", "com.example.app", "--external-version-id", "123"}),
		Entry("purchases", listPurchasesCmd, []string{}),
	)
	DescribeTable("rejects invalid countries before reading the account", func(factory func() *cobra.Command, args []string) {
		previous := dependencies
		DeferCleanup(func() { dependencies = previous })
		dependencies.AppStore = nil
		cmd := factory()
		cmd.SetArgs(append(args, "-c", "ZZ"))
		Expect(cmd.Execute()).To(MatchError(`unsupported country code "ZZ"`))
	},
		Entry("metadata", getVersionMetadataCmd, []string{"-i", "42", "--external-version-id", "123"}),
		Entry("purchases", listPurchasesCmd, []string{}),
	)
})

type countryRetryStore struct {
	appstore.AppStore
	accounts       []appstore.Account
	lookupAccounts []appstore.Account
}

func (*countryRetryStore) AccountInfo() (appstore.AccountInfoOutput, error) {
	return appstore.AccountInfoOutput{Account: appstore.Account{StoreFront: "143454"}}, nil
}
func (*countryRetryStore) Login(appstore.LoginInput) (appstore.LoginOutput, error) {
	return appstore.LoginOutput{Account: appstore.Account{StoreFront: "143454", PasswordToken: "fresh"}}, nil
}
func (s *countryRetryStore) Lookup(input appstore.LookupInput) (appstore.LookupOutput, error) {
	s.lookupAccounts = append(s.lookupAccounts, input.Account)
	return appstore.LookupOutput{App: appstore.App{ID: 42}}, nil
}
func (s *countryRetryStore) record(account appstore.Account) error {
	s.accounts = append(s.accounts, account)
	if len(s.accounts) == 1 {
		return appstore.ErrPasswordTokenExpired
	}
	return nil
}
func (s *countryRetryStore) GetVersionMetadata(input appstore.GetVersionMetadataInput) (appstore.GetVersionMetadataOutput, error) {
	return appstore.GetVersionMetadataOutput{}, s.record(input.Account)
}
func (s *countryRetryStore) OwnedApps(input appstore.OwnedAppsInput) (appstore.OwnedAppsOutput, error) {
	return appstore.OwnedAppsOutput{}, s.record(input.Account)
}
