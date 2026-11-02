package main

import "time"

// documentsByVIN is the mock's "database".
//
//	1HGCM82633A004352  3 documents
//	3VWFE21C04M000001  known vehicle, zero documents
var documentsByVIN = map[string][]Document{
	"1HGCM82633A004352": {
		{
			ID:           "1001",
			Name:         "Purchase Agreement",
			Category:     CategoryContract,
			LastModified: time.Date(2026, 10, 1, 9, 30, 0, 0, time.UTC),
		},
		{
			ID:           "1002",
			Name:         "Sales Invoice",
			Category:     CategoryBill,
			LastModified: time.Date(2026, 10, 2, 14, 15, 0, 0, time.UTC),
		},
		{
			ID:           "1003",
			Name:         "Certificate of Title",
			Category:     CategoryVehicleTitle,
			LastModified: time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC),
		},
	},
	"3VWFE21C04M000001": {},
}

// GetAllDocuments returns the documents for a VIN. ok is false when the VIN
// is unknown, so callers can tell it apart from a known VIN with zero docs.
func GetAllDocuments(id string) (docs []Document, ok bool) {
	docs, ok = documentsByVIN[id]
	return docs, ok
}
