package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// Datastore defines an datastore.
type Datastore string

// Supported datastores.
const (
	GCS = "GCS"
)

// Item defines a downloadable item.
type Item struct {
	Code  string    `json:"code"`
	Name  string    `json:"name"`
	Store Datastore `json:"store"`
	URI   string    `json:"uri"`
}

// DownloadCard defines a download card.
type DownloadCard struct {
	Key      string `json:"key"`
	ItemCode string `json:"item_code"`
	CountNow int    `json:"count_now"`
	CoundMax int    `json:"count_max"`
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
	UpdateCards(cards map[string]*DownloadCard) error
}

// URLIssuer issues a download link.
type URLIssuer interface {
	IssueURL(filename string) (string, error)
}

// GCSAdapter provides ItemsLoader, CardsLoader, CardsUpdater, URLIssuer.
type GCSAdapter struct {
	appBucket  string
	dataBucket string
	client     *storage.Client
}

var (
	_ ItemsLoader  = (*GCSAdapter)(nil)
	_ CardsLoader  = (*GCSAdapter)(nil)
	_ CardsUpdater = (*GCSAdapter)(nil)
	_ URLIssuer    = (*GCSAdapter)(nil)
)

// NewGCSAdapter initializes a GCSAdapter.
func NewGCSAdapter(appBucket, dataBucket string) *GCSAdapter {
	a := &GCSAdapter{appBucket: appBucket, dataBucket: dataBucket}
	return a
}

// LoadItems loads a list of items from GCS.
func (a *GCSAdapter) LoadItems() (map[string]*Item, error) {
	if err := a.initClientIfNeeded(); err != nil {
		return nil, err
	}
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, GCSAccessTimeout)
	defer cancelFunc()
	logf(DEBUG, "LoadItems: read GCS")
	obj := a.client.Bucket(a.appBucket).Object(GCSPathItems)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	items := make(map[string]*Item)
	if err := json.NewDecoder(reader).Decode(&items); err != nil {
		return nil, err
	}
	if err := reader.Close(); err != nil {
		return nil, err
	}
	return items, nil
}

// LoadCards loads download cards from GCS.
func (a *GCSAdapter) LoadCards() (map[string]*DownloadCard, error) {
	if err := a.initClientIfNeeded(); err != nil {
		return nil, err
	}
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, GCSAccessTimeout)
	defer cancelFunc()
	logf(DEBUG, "LoadCards: read GCS")
	obj := a.client.Bucket(a.appBucket).Object(GCSPathCards)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	cards := make(map[string]*DownloadCard)
	if err := json.NewDecoder(reader).Decode(&cards); err != nil {
		return nil, err
	}
	return cards, nil
}

// UpdateCards updates download cards and uploads them to GCS.
func (a *GCSAdapter) UpdateCards(cards map[string]*DownloadCard) error {
	if err := a.initClientIfNeeded(); err != nil {
		return err
	}
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, GCSAccessTimeout)
	defer cancelFunc()
	logf(DEBUG, "UpdateCards: write GCS")
	obj := a.client.Bucket(a.appBucket).Object(GCSPathCards)
	writer := obj.NewWriter(ctx)
	if err := json.NewEncoder(writer).Encode(&cards); err != nil {
		return err
	}
	// Writes happen asynchronously!
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}

// IssueURL issues a GCS presigned URL.
func (a *GCSAdapter) IssueURL(filename string) (string, error) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, GCSAccessTimeout)
	defer cancelFunc()
	creds, err := google.FindDefaultCredentials(ctx, storage.ScopeReadOnly)
	if err != nil {
		return "", err
	}
	conf, err := google.JWTConfigFromJSON(creds.JSON, storage.ScopeReadOnly)
	if err != nil {
		return "", err
	}
	logf(INFO, "issue a SignedURL for %v", filename)
	opts := &storage.SignedURLOptions{
		GoogleAccessID: conf.Email,
		PrivateKey:     conf.PrivateKey,
		Method:         http.MethodGet,
		Expires:        time.Now().Add(GCSURLDuration),
	}
	url, err := storage.SignedURL(a.dataBucket, filename, opts)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (a *GCSAdapter) initClientIfNeeded() error {
	if a.client != nil {
		return nil
	}
	logf(INFO, "initialize GCS client")
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, GCSAccessTimeout)
	defer cancelFunc()
	creds, err := google.FindDefaultCredentials(ctx, storage.ScopeReadWrite)
	if err != nil {
		return err
	}
	client, err := storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return err
	}
	a.client = client
	return nil
}
