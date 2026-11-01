package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/majd/ipatool/v2/pkg/appstore"
	"github.com/majd/ipatool/v2/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lookup command", func() {
	var store *fakeLookupAppStore
	var output bytes.Buffer

	BeforeEach(func() {
		previous := dependencies
		DeferCleanup(func() { dependencies = previous })
		store = &fakeLookupAppStore{output: appstore.LookupOutput{
			App:               appstore.App{ID: 42, BundleID: "com.example.app", Name: "Example", Version: "1.0"},
			ExternalVersionID: "123456",
		}}
		dependencies.AppStore = store
		output.Reset()
		dependencies.Logger = log.NewLogger(log.Args{Writer: &output})
	})

	execute := func(args ...string) error {
		cmd := lookupCmd()
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		cmd.SetArgs(args)

		return cmd.Execute()
	}

	It("looks up an ID without an account and defaults to the US", func() {
		Expect(execute("-i", "42")).To(Succeed())
		Expect(store.inputs).To(Equal([]appstore.LookupInput{{AppID: 42, Country: "US"}}))
		var result map[string]any
		Expect(json.Unmarshal(output.Bytes(), &result)).To(Succeed())
		Expect(result).To(HaveKeyWithValue("success", true))
		Expect(result).NotTo(HaveKey("externalVersionID"))
		Expect(result["app"]).To(HaveKeyWithValue("bundleID", "com.example.app"))
	})

	It("passes the bundle identifier, country, platform, and version option", func() {
		Expect(execute("-b", "com.example.app", "-c", "JP", "--platform", "macos", "--get-version-id")).To(Succeed())
		Expect(store.inputs).To(Equal([]appstore.LookupInput{{
			BundleID: "com.example.app", Country: "JP", Platform: appstore.PlatformMacOS, IncludeVersionID: true,
		}}))
		Expect(output.String()).To(ContainSubstring(`"externalVersionID":"123456"`))
	})

	DescribeTable("rejects invalid arguments before requesting metadata", func(args []string) {
		Expect(execute(args...)).NotTo(Succeed())
		Expect(store.inputs).To(BeEmpty())
	},
		Entry("missing identifier", []string{}),
		Entry("negative app ID", []string{"-i", "-1"}),
		Entry("empty country", []string{"-i", "42", "-c", ""}),
		Entry("invalid platform", []string{"-i", "42", "--platform", "invalid"}),
		Entry("positional argument", []string{"-i", "42", "unexpected"}),
	)

	It("preserves lookup failures", func() {
		store.err = errors.New("lookup failed")
		Expect(execute("-i", "42")).To(MatchError(store.err))
		Expect(output.Len()).To(BeZero())
	})

	DescribeTable("reports bulk country outcomes", func(lookupErr error, status string) {
		store.err = lookupErr
		cmd := lookupCmd()
		cmd.SetContext(context.Background())
		Expect(lookupBulk(cmd, appstore.LookupInput{AppID: 42, Platform: appstore.PlatformIPad}, []string{"JP", "US"}, 0)).To(Succeed())
		Expect(store.inputs).To(Equal([]appstore.LookupInput{
			{AppID: 42, Platform: appstore.PlatformIPad, Country: "JP"},
			{AppID: 42, Platform: appstore.PlatformIPad, Country: "US"},
		}))
		lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
		Expect(lines).To(HaveLen(2))
		for index, line := range lines {
			var result map[string]any
			Expect(json.Unmarshal(line, &result)).To(Succeed())
			Expect(result).To(HaveKeyWithValue("country", []string{"JP", "US"}[index]))
			Expect(result).To(HaveKeyWithValue("availability", status))
		}
	},
		Entry("available", nil, "available"),
		Entry("not listed", appstore.ErrAppNotFound, "unavailable"),
		Entry("request failure", errors.New("rate limited"), "unknown"),
	)

	It("rejects an explicit country in bulk mode", func() {
		Expect(execute("-i", "42", "--bulk", "-c", "US")).NotTo(Succeed())
		Expect(store.inputs).To(BeEmpty())
	})

	It("stops bulk lookup when canceled", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		cmd := lookupCmd()
		cmd.SetContext(ctx)
		Expect(lookupBulk(cmd, appstore.LookupInput{AppID: 42}, []string{"US"}, 0)).To(MatchError(context.Canceled))
		Expect(store.inputs).To(BeEmpty())
	})

	It("is registered on the root command", func() {
		cmd, _, err := rootCmd().Find([]string{"lookup"})
		Expect(err).NotTo(HaveOccurred())
		Expect(cmd.Name()).To(Equal("lookup"))
	})
})

type fakeLookupAppStore struct {
	appstore.AppStore
	inputs []appstore.LookupInput
	output appstore.LookupOutput
	err    error
}

func (f *fakeLookupAppStore) Lookup(input appstore.LookupInput) (appstore.LookupOutput, error) {
	f.inputs = append(f.inputs, input)

	return f.output, f.err
}
