package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/gorilla/mux"
)

// ProjectID defines GCP project ID.
var ProjectID string

func main() {
	ProjectID = os.Getenv("DL_GCP_ID")
	itemsURI := os.Getenv("DL_GCS_ITEMS")
	cardsURI := os.Getenv("DL_GCS_CARDS")
	port := os.Getenv("PORT")

	staticDIR := "static"
	address := fmt.Sprintf("0.0.0.0:%s", port)

	if ProjectID == "" || cardsURI == "" || itemsURI == "" || port == "" {
		logf(EMERGENCY, "some environment variables are empty")
		os.Exit(1)
	}

	adp := NewGCSAdapter()
	handler := NewHandler(adp, adp, adp, adp)

	r := mux.NewRouter()
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }).Methods(http.MethodGet)
	r.HandleFunc("/download", handler.Download).Methods(http.MethodPost)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(staticDIR)))).Methods(http.MethodGet)

	srv := &http.Server{Addr: address, Handler: r}
	srv.RegisterOnShutdown(func() {
		err := handler.Close()
		if err != nil {
			logf(ERROR, "handler.Close: %v", err)
		}
	})

	go func() {
		logf(NOTICE, "listening on %v", address)
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				logf(NOTICE, "http server is closed")
			} else {
				logf(EMERGENCY, "http server error: %v", err)
				os.Exit(1)
			}
		}
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
	logf(NOTICE, "SIGINT received")
	ctx := context.Background()
	err := srv.Shutdown(ctx)
	if err != nil {
		logf(ERROR, "srv.Shutdown: %v", err)
	}
	logf(NOTICE, "exit(0)")
}

// Severity defines a log severity.
type Severity string

// A part of Cloud Logging's severity.
const (
	EMERGENCY = "EMERGENCY"
	CRITICAL  = "CRITICAL"
	ERROR     = "ERROR"
	NOTICE    = "NOTICE"
	INFO      = "INFO"
	DEBUG     = "DEBUG"
)

// Entry defines a log entry.
type Entry struct {
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Component string `json:"component"`
	Trace     string `json:"logging.googleapis.com/trace,omitempty"`
}

func (e Entry) String() string {
	j, err := json.Marshal(e)
	if err != nil {
		log.Printf("json.Marshal: %v", err)
	}
	return string(j)
}

// ComponentName defines the name of this component for logging.
// We can filter this value in Cloud Log Viewer.
const ComponentName = "dl.nullpo-t.net"

var once sync.Once

func logf(severity Severity, msgFmt string, args ...interface{}) {
	once.Do(func() {
		// Disable adding log prefix as it will prevent jsonify.
		log.SetFlags(0)
	})
	log.Println(Entry{
		Severity:  string(severity),
		Message:   fmt.Sprintf(msgFmt, args...),
		Component: ComponentName,
		Trace:     ProjectID,
	})
}
