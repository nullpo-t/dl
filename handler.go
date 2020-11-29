package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
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

	mu sync.Mutex
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
	// POST only.
	if r.Method != http.MethodPost {
		logf(ERROR, "the request is not POST: %v", r)
		writeInternalError(w)
		return
	}

	// Cache data.
	isReturn := false
	onceZDownload.Do(func() {
		items, err := h.IL.LoadItems()
		if err != nil {
			logf(ERROR, "could not load item list: %v", err)
			writeInternalError(w)
			isReturn = true
			return
		}
		h.items = items

		cards, err := h.CL.LoadCards()
		if err != nil {
			logf(ERROR, "could not load card list: %v", err)
			writeInternalError(w)
			isReturn = true
			return
		}
		h.cards = cards
	})
	if isReturn {
		return
	}

	key := sanitizeDLKey(r.FormValue("dlkey"))
	logf(INFO, "POST key=%v", key)

	card, ok := h.cards[key]
	if !ok {
		logf(INFO, "invalid key: %+v", key)
		writeResponse(w, http.StatusOK, "invalid download key; if you are sure this is an error, please contact us")
		return
	}

	// Critical section.
	// FIXME: Too big and impacts performance.
	h.mu.Lock()
	defer h.mu.Unlock()

	// Reload card list as the cache may be old.
	cards, err := h.CL.LoadCards()
	if err != nil {
		logf(ERROR, "could not load card list: %v", err)
		writeInternalError(w)
		return
	}
	h.cards = cards
	card, _ = h.cards[key]

	// Check download count.
	if card.CountNow >= card.CoundMax {
		logf(INFO, "count is max: %+v", card)
		writeResponse(w, http.StatusOK, "download count exceeded; please contact us if you want to do it")
		return
	}

	// Increment download count.
	card.CountNow++
	if err := h.CU.UpdateCards(h.cards); err != nil {
		logf(ERROR, "could not upload card list: %v", err)
		writeInternalError(w)
		return
	}

	// Issue URL.
	item, ok := h.items[card.ItemCode]
	if !ok {
		logf(ERROR, "invalid ItemCode in the card: %+v", card)
		writeInternalError(w)
		return
	}
	dlURL, err := h.UI.IssueURL(item.URI)
	if err != nil {
		logf(ERROR, "could not issue URL: %v", err)
		writeInternalError(w)
		return
	}

	writeResponse(w, http.StatusOK, `<a href="%s" target="_blank">%s</a>`, dlURL, item.Name)
}

// sanitizeDLkey checks k is only including [0-9A-Za-z] and its length is 0-8,
// then convert it to uppercase and return.
// If failed, just return "".
// TODO: do some tests
func sanitizeDLKey(k string) string {
	re := regexp.MustCompile(`^[0-9A-Za-z]{0,8}$`)
	if !re.MatchString(k) {
		return ""
	}
	return strings.ToUpper(k)
}

// Close closes the Handler.
func (h *Handler) Close() error {
	return nil
}

const htmlFmt = `<!DOCTYPE html>
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
  %s
  <br>
  <button type="button" onclick="history.back()">Back</button>
</body>
</html>`

func writeResponse(w http.ResponseWriter, statusCode int, htmlBodyFmt string, args ...interface{}) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, htmlFmt, fmt.Sprintf(htmlBodyFmt, args...))
}

func writeInternalError(w http.ResponseWriter) {
	writeResponse(w, http.StatusInternalServerError, "internal error; please contact us; %v", time.Now())
}
