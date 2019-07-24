package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func httpResponse(w http.ResponseWriter, code int, contentType string, body string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	w.Write([]byte(body))
}

func jsonResponse(w http.ResponseWriter, code int, value interface{}) {
	body, err := json.Marshal(value)
	if err != nil {
		fmt.Fprintf(w, "{ \"message\": \"cannot marshall JSON\" }")
		return
	}

	httpResponse(w, code, "application/json", string(body))
}

func messageResponse(w http.ResponseWriter, message string) {
	jsonResponse(w, 200, MessageResponse{message})
}

func errorResponse(w http.ResponseWriter, code int, message string) {
	jsonResponse(w, code, ErrorResponse{message})
}
