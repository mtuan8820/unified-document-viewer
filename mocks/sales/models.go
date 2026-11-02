package main

import "time"

type Category string

const (
	CategoryContract     Category = "CONTRACT"
	CategoryBill         Category = "BILL"
	CategoryVehicleTitle Category = "TITLE" //gorvernment proof of who owns the vehicle
)

type Document struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Category     Category  `json:"category"`
	LastModified time.Time `json:"lastModified"`
}

type Response struct {
	VIN       string     `json:"vin"`
	Documents []Document `json:"documents"`
}
