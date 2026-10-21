package appstore

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storefront country code resolution", func() {
	DescribeTable("resolves missing storefronts", func(storeFront, expectedCountryCode string) {
		for _, suffix := range []string{"", "-2,34"} {
			countryCode, err := countryCodeFromStoreFront(storeFront + suffix)

			Expect(err).ToNot(HaveOccurred())
			Expect(countryCode).To(Equal(expectedCountryCode))
		}
	},
		Entry("Afghanistan", "143610", "AF"),
		Entry("Bosnia and Herzegovina", "143612", "BA"),
		Entry("Burkina Faso", "143578", "BF"),
		Entry("Benin", "143576", "BJ"),
		Entry("Bhutan", "143577", "BT"),
		Entry("Democratic Republic of the Congo", "143613", "CD"),
		Entry("Republic of the Congo", "143582", "CG"),
		Entry("Cameroon", "143574", "CM"),
		Entry("Cape Verde", "143580", "CV"),
		Entry("Fiji", "143583", "FJ"),
		Entry("Micronesia", "143591", "FM"),
		Entry("Gabon", "143614", "GA"),
		Entry("Gambia", "143584", "GM"),
		Entry("Guinea-Bissau", "143585", "GW"),
		Entry("Kyrgyzstan", "143586", "KG"),
		Entry("Cambodia", "143579", "KH"),
		Entry("Laos", "143587", "LA"),
		Entry("Liberia", "143588", "LR"),
		Entry("Libya", "143567", "LY"),
		Entry("Morocco", "143620", "MA"),
		Entry("Montenegro", "143619", "ME"),
		Entry("Myanmar", "143570", "MM"),
		Entry("Mauritania", "143590", "MR"),
		Entry("Malawi", "143589", "MW"),
		Entry("Mozambique", "143593", "MZ"),
		Entry("Namibia", "143594", "NA"),
		Entry("Nauru", "143606", "NR"),
		Entry("Papua New Guinea", "143597", "PG"),
		Entry("Palau", "143595", "PW"),
		Entry("Rwanda", "143621", "RW"),
		Entry("Solomon Islands", "143601", "SB"),
		Entry("Seychelles", "143599", "SC"),
		Entry("Sierra Leone", "143600", "SL"),
		Entry("São Tomé and Príncipe", "143598", "ST"),
		Entry("Eswatini", "143602", "SZ"),
		Entry("Chad", "143581", "TD"),
		Entry("Tajikistan", "143603", "TJ"),
		Entry("Turkmenistan", "143604", "TM"),
		Entry("Tonga", "143608", "TO"),
		Entry("Vanuatu", "143609", "VU"),
		Entry("Kosovo", "143624", "XK"),
		Entry("Zambia", "143622", "ZM"),
		Entry("Zimbabwe", "143605", "ZW"),
	)
})
