package appstore

import (
	"errors"
	"net/url"

	"github.com/majd/ipatool/v2/pkg/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Lookup country and version selection", func() {
	var store *appstore
	var search *http.MockClient[searchResult]
	var catalog *http.MockClient[platformVersionLookupResult]
	app := App{ID: 42, BundleID: "com.example.app"}

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		search = http.NewMockClient[searchResult](ctrl)
		catalog = http.NewMockClient[platformVersionLookupResult](ctrl)
		store = &appstore{searchClient: search, platformClient: catalog}
	})

	expectSearch := func(country string) {
		search.EXPECT().Send(gomock.Any()).Do(func(req http.Request) {
			parsed, err := url.Parse(req.URL)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Query().Get("country")).To(Equal(country))
			Expect(parsed.Query().Get("id")).To(Equal("42"))
		}).Return(http.Result[searchResult]{StatusCode: 200, Data: searchResult{Results: []App{app}}}, nil)
	}

	It("normalizes an explicit country and overrides the account storefront", func() {
		expectSearch("JP")
		output, err := store.Lookup(LookupInput{AppID: 42, Country: " jp ", Account: Account{StoreFront: "143441"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(output.App).To(Equal(app))
		Expect(output.ExternalVersionID).To(BeEmpty())
	})

	It("supports an explicit country without account credentials", func() {
		expectSearch("US")
		output, err := store.Lookup(LookupInput{AppID: 42, Country: "US"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output.App).To(Equal(app))
	})

	It("rejects unsupported countries before sending a request", func() {
		_, err := store.Lookup(LookupInput{AppID: 42, Country: "ZZ"})
		Expect(err).To(MatchError(ContainSubstring("unsupported country code")))
	})

	It("uses the selected country and default iPhone catalog for version IDs", func() {
		expectSearch("JP")
		catalog.EXPECT().Send(gomock.Any()).Do(func(req http.Request) {
			parsed, err := url.Parse(req.URL)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Query().Get("cc")).To(Equal("jp"))
			Expect(parsed.Query().Get("platform")).To(Equal("enterprisestore"))
		}).Return(http.Result[platformVersionLookupResult]{StatusCode: 200, Data: platformVersionLookupResult{
			Results: map[string]platformVersionLookupItem{"42": {Offers: []platformVersionLookupOffer{
				{Version: platformVersionLookupVersion{ExternalID: "123456"}},
			}}},
		}}, nil)
		output, err := store.Lookup(LookupInput{AppID: 42, Country: "JP", IncludeVersionID: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(output.ExternalVersionID).To(Equal("123456"))
	})

	It("preserves errors from optional version lookup", func() {
		expectSearch("US")
		failure := errors.New("catalog unavailable")
		catalog.EXPECT().Send(gomock.Any()).Return(http.Result[platformVersionLookupResult]{}, failure)
		_, err := store.Lookup(LookupInput{AppID: 42, Country: "US", IncludeVersionID: true})
		Expect(errors.Is(err, failure)).To(BeTrue())
	})

	It("selects the macOS version reader for Mac apps", func() {
		expectSearch("US")
		storefront := http.NewMockClient[[]byte](gomock.NewController(GinkgoT()))
		store.storefrontClient = storefront
		storefront.EXPECT().Send(gomock.Any()).Do(func(req http.Request) {
			Expect(req.URL).To(Equal("https://apps.apple.com/us/app/id42?platform=mac"))
		}).Return(http.Result[[]byte]{StatusCode: 200, Data: macVersionPage(`{"purchaseConfiguration":{"bundleId":"com.example.app","appPlatforms":["mac"],"buyParams":"salableAdamId=42&appExtVrsId=654321"}}`)}, nil)
		output, err := store.Lookup(LookupInput{AppID: 42, Country: "US", Platform: PlatformMacOS, IncludeVersionID: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(output.ExternalVersionID).To(Equal("654321"))
	})
})
