package main

import (
	"fmt"
	"io"
	"net/http"
)

// Handler handles download functionality.
type Handler struct {
	io.Closer
	items map[string]*Item
	cards map[string]*DownloadCard

	IL ItemsLoader
	CL CardsLoader
	CU CardsUpdater
	UI URLIssuer
}

// NewHandler initializes a Handler.
func NewHandler(il ItemsLoader, cl CardsLoader, cu CardsUpdater, ui URLIssuer) *Handler {
	handler := &Handler{
		IL: il,
		CL: cl,
		CU: cu,
		UI: ui,
	}
	return handler
}

// Download do everything related download functionality.
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logf(ERROR, "the request is not POST: %v", r)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	inputKey := r.FormValue("dlkey")
	logf(INFO, "POST key=%v", inputKey)
	w.WriteHeader(http.StatusOK)
	dlURL := "https://example.com"
	htmlFmt := `<!DOCTYPE html>
	<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1">
	</head>
	<body>
	<a href="%s" target="_blank">Download</a><br><button type="button" onclick="history.back()">Back</button>
	</body></html>`
	fmt.Fprintf(w, htmlFmt, dlURL)
}

// Close closes the Handler.
func (h *Handler) Close() error {
	return nil
}
