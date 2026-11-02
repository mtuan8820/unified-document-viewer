package main

import "time"

// Record is one document in the Service System.
type Record struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	DocType     string    `json:"doc_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description string    `json:"description"`
	Uri         string    `json:"uri"`
}

// Response is the body of GET /service/records.
type Response struct {
	Data  []Record `json:"data"`
	Total int      `json:"total"`
}
