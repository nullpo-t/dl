package main

import (
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
	w.WriteHeader(http.StatusOK)
}

// Close closes the Handler.
func (h *Handler) Close() error {
	return nil
}
