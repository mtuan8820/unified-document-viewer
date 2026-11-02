package main

import "time"

// recordsByVIN is the mock's "database".
//
//	1HGCM82633A004352  3 records; ID "1001" also exists in the sales mock
//	5YJSA1E26HF000337  2 records; this VIN is not in the sales mock
//	3VWFE21C04M000001  deliberately absent (no service records)
var recordsByVIN = map[string][]Record{
	"1HGCM82633A004352": {
		{
			ID:          "1001",
			Title:       "Oil Change Invoice",
			DocType:     "invoice",
			CreatedAt:   time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 10, 20, 10, 5, 0, 0, time.UTC),
			Description: "Engine oil and filter replacement at 15,000 km",
			Uri:         "https://service.example.com/files/1001.pdf",
		},
		{
			ID:          "1002",
			Title:       "Annual Safety Inspection",
			DocType:     "inspection",
			CreatedAt:   time.Date(2026, 6, 5, 14, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 6, 7, 9, 30, 0, 0, time.UTC),
			Description: "Yearly roadworthiness inspection, passed",
			Uri:         "https://service.example.com/files/1002.pdf",
		},
		{
			ID:          "1003",
			Title:       "Warranty Repair: Brake Pads",
			DocType:     "warranty",
			CreatedAt:   time.Date(2026, 8, 11, 8, 45, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 8, 12, 16, 20, 0, 0, time.UTC),
			Description: "Front brake pads replaced under warranty",
			Uri:         "https://service.example.com/files/1003.pdf",
		},
	},
	"5YJSA1E26HF000337": {
		{
			ID:          "3001",
			Title:       "Tire Rotation Invoice",
			DocType:     "invoice",
			CreatedAt:   time.Date(2026, 3, 14, 11, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 3, 14, 11, 0, 0, 0, time.UTC),
			Description: "Four-tire rotation and pressure check",
			Uri:         "https://service.example.com/files/3001.pdf",
		},
		{
			ID:          "3002",
			Title:       "Recall Notice: Seat Belt Sensor",
			DocType:     "recall",
			CreatedAt:   time.Date(2026, 9, 1, 7, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 9, 28, 13, 10, 0, 0, time.UTC),
			Description: "Manufacturer recall, sensor replaced",
			Uri:         "https://service.example.com/files/3002.pdf",
		},
	},
}
