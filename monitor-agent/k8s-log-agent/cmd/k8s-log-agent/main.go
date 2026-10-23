package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/WeBankPartners/open-monitor/k8s-log-agent/internal/agent"
)

func main() {
	defer func() {
		if recover() != nil {
			log.Printf("agent panic\n%s", debug.Stack())
			os.Exit(1)
		}
	}()
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
func run() error {
	cfg, err := agent.LoadRuntimeConfig(os.Getenv)
	if err != nil {
		return err
	}
	reader, err := agent.New(cfg.Root, cfg.StateDir)
	if err != nil {
		return errors.New("cannot initialize log reader")
	}
	log.Printf("agent started: cluster=%s node=%s root=%s state_dir=%s poll_interval=%s", cfg.ClusterID, cfg.NodeName, cfg.Root, cfg.StateDir, cfg.PollInterval)
	defer reader.Close()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	var ready atomic.Bool
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if !ready.Load() {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(200)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if !ready.Load() {
			http.Error(w, "metrics unavailable", 503)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(reader.MetricsText(cfg.ClusterID, cfg.NodeName)))
	})
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	errs := make(chan error, 1)
	go func() {
		defer func() {
			if recover() != nil {
				log.Printf("HTTP worker panic\n%s", debug.Stack())
				errs <- errors.New("HTTP worker failed")
			}
		}()
		errs <- server.ListenAndServe()
	}()
	defer func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdown)
	}()
	client := &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	var lastRegistrationError string
	var lastRegistrationLog time.Time
	registrationHealthy := false
	register := func() {
		if err := agent.Register(client, cfg); err != nil {
			registrationHealthy = false
			if err.Error() != lastRegistrationError || time.Since(lastRegistrationLog) >= time.Minute {
				log.Printf("agent registration failed: %v", err)
				lastRegistrationError, lastRegistrationLog = err.Error(), time.Now()
			}
		} else if !registrationHealthy {
			log.Printf("agent registration succeeded")
			registrationHealthy = true
			lastRegistrationError = ""
		} else if lastRegistrationError != "" {
			log.Printf("agent registration recovered")
			lastRegistrationError = ""
		}
	}
	register()
	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()
	lastConfig := time.Time{}
	var lastRefreshError string
	var lastRefreshLog time.Time
	var lastUnpublishedLog time.Time
	for {
		register()
		next, fetchErr := agent.FetchConfig(client, cfg, reader.Version())
		if fetchErr == nil {
			previousVersion := reader.Version()
			fetchErr = reader.Apply(next)
			if fetchErr != nil {
				log.Printf("configuration apply failed: %v", fetchErr)
			} else if next.Version != previousVersion {
				log.Printf("configuration applied: version=%d target_count=%d", next.Version, len(next.Targets))
			}
			if lastRefreshError != "" {
				log.Printf("configuration refresh recovered")
				lastRefreshError = ""
			}
		} else if errors.Is(fetchErr, agent.ErrConfigNotPublished) {
			if time.Since(lastUnpublishedLog) >= time.Minute {
				log.Printf("configuration not published yet")
				lastUnpublishedLog = time.Now()
			}
			if lastRefreshError != "" {
				log.Printf("configuration refresh recovered")
				lastRefreshError = ""
			}
		} else {
			if fetchErr.Error() != lastRefreshError || time.Since(lastRefreshLog) >= time.Minute {
				log.Printf("configuration refresh failed: %v", fetchErr)
				lastRefreshError, lastRefreshLog = fetchErr.Error(), time.Now()
			}
		}
		if fetchErr == nil {
			lastConfig = time.Now()
		}
		if time.Since(lastConfig) > time.Minute {
			ready.Store(false)
		} else if err = reader.Scan(); err != nil {
			ready.Store(false)
			log.Printf("log scan failed: %v", err)
		} else {
			ready.Store(true)
		}
		select {
		case <-ctx.Done():
			return nil
		case err = <-errs:
			if err == http.ErrServerClosed {
				return nil
			}
			return errors.New("HTTP listener failed")
		case <-ticker.C:
		}
	}
}
