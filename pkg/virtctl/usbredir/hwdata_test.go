package usbredir

import (
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/v2"
)

//go:embed testdata/hwdata_usb.ids
var hwdata string

var _ = Describe("usbredir on hwdata usb.ids", func() {

	Context("Parsing the hwdata/usb.ids", func() {
		DescribeTable("Should match expected entries", func(
			vendor, product string,
			expectedVendorName, expectedProductName string,
			expectToFail bool,
		) {
			vendorName, productName, success := MetadataLookup(hwdata, vendor, product)
			Expect(vendorName).To(Equal(expectedVendorName))
			Expect(productName).To(Equal(expectedProductName))
			Expect(success).To(Equal(!expectToFail))
		},
			Entry("Find top", "03eb", "2002", "Atmel Corp.", "Mass Storage Device", false),
			Entry("Find bottom", "03f0", "3005", "HP, Inc", "ScanJet 4670v", false),
			Entry("Only product", "03ee", "beef", "Mitsumi", "", false),
			Entry("Not found", "dead", "beef", "", "", true),
		)
	})
})
