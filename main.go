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
	"time"

	"github.com/gorilla/mux"
)

// App Global Parameters.
const (
	// ComponentName defines the name of this component for logging.
	// Cloud Log Viewer can filter this value.
	// Omitempty.
	ComponentName = ""

	GCSPathItems     = "items.json"
	GCSPathCards     = "cards.json"
	GCSAccessTimeout = 10 * time.Second
	GCSURLDuration   = 5 * time.Minute
)

// App Global Variables.
var (
	// ProjectID defines GCP project ID.
	ProjectID string

	// onceZlogf is used in logf.
	onceZlogf sync.Once

	// onceZDownload is userd in Handler.Download.
	onceZDownload sync.Once
)

func main() {
	ProjectID = os.Getenv("DL_GCP_ID")
	appBucket := os.Getenv("DL_GCS_APP_BUCKET")
	dataBucket := os.Getenv("DL_GCS_DATA_BUCKET")
	port := os.Getenv("PORT")

	staticDIR := "static"
	address := fmt.Sprintf("0.0.0.0:%s", port)

	if ProjectID == "" || appBucket == "" || dataBucket == "" || port == "" {
		logf(EMERGENCY, "some environment variables are empty")
		os.Exit(1)
	}

	adp := NewGCSAdapter(appBucket, dataBucket)
	handler := NewHandler(adp, adp, adp, adp)

	r := mux.NewRouter()
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }).Methods(http.MethodGet)
	r.HandleFunc("/", handler.Download).Methods(http.MethodPost)
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

// Logging.

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
	Component string `json:"component,omitempty"`
	Trace     string `json:"logging.googleapis.com/trace,omitempty"`
}

func (e Entry) String() string {
	j, err := json.Marshal(e)
	if err != nil {
		log.Printf("json.Marshal: %v", err)
	}
	return string(j)
}

func logf(severity Severity, msgFmt string, args ...interface{}) {
	onceZlogf.Do(func() {
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
