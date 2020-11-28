package main

import "time"

// Datastore defines an datastore.
type Datastore string

// Supported datastores.
const (
	GCS = "GCS"
)

// Item defines a downloadable item.
type Item struct {
	Code  string
	Name  string
	Store Datastore
	URI   string
}

// DownloadCard defines a download card.
type DownloadCard struct {
	Key       string
	ItemCode  string
	CountNow  int
	CoundMax  int
	StartedAt time.Time
	ExpiredAt time.Time
}

// ItemsLoader loads a list of items.
type ItemsLoader interface {
	LoadItems() (map[string]*Item, error)
}

// CardsLoader loads download cards.
type CardsLoader interface {
	LoadCards() (map[string]*DownloadCard, error)
}

// CardsUpdater updates download cards.
type CardsUpdater interface {
	UpdateCards() error
}

// URLIssuer issues a download link.
type URLIssuer interface {
	IssueURL(itemCode string) (string, error)
}

// GCSAdapter provides ItemsLoader, CardsLoader, CardsUpdater, URLIssuer.
type GCSAdapter struct {
}

var (
	_ ItemsLoader  = (*GCSAdapter)(nil)
	_ CardsLoader  = (*GCSAdapter)(nil)
	_ CardsUpdater = (*GCSAdapter)(nil)
	_ URLIssuer    = (*GCSAdapter)(nil)
)

// NewGCSAdapter initializes a GCSAdapter.
func NewGCSAdapter() *GCSAdapter {
	a := &GCSAdapter{}
	return a
}

// LoadItems loads a list of items from GCS.
func (a *GCSAdapter) LoadItems() (map[string]*Item, error) {
	panic("not implemented")
}

// LoadCards loads download cards from GCS.
func (a *GCSAdapter) LoadCards() (map[string]*DownloadCard, error) {
	panic("not implemented")
}

// UpdateCards updates download cards and uploads them to GCS.
func (a *GCSAdapter) UpdateCards() error {
	panic("not implemented")
}

// IssueURL issues a GCS presigned URL.
func (a *GCSAdapter) IssueURL(itemCode string) (string, error) {
	panic("not implemented")
}
