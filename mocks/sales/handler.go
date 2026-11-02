package main

import (
	"encoding/json"
	"net/http"
)

type MyHandler struct {
}

// get all documents with id
// if len(docs) > 0 then return 200 with documents (can be empty)
// return 404 if unknow VIN
func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("vin")

	documents, ok := GetAllDocuments(id)
	if !ok {
		msg := map[string]string{"error": "vehicle not found"}
		writeJSON(w, http.StatusNotFound, msg)
		return
	}

	res := &Response{
		VIN:       id,
		Documents: documents,
	}
	writeJSON(w, http.StatusOK, res)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}
