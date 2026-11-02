package main

import (
	"encoding/json"
	"net/http"
)

func listRecords(w http.ResponseWriter, r *http.Request) {
	vin := r.URL.Query().Get("vin")
	if vin == "" {
		http.Error(w, "query parameter vin is required", http.StatusBadRequest)
		return
	}

	records := recordsByVIN[vin]
	if len(records) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	res := Response{
		Data:  records,
		Total: len(records),
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
