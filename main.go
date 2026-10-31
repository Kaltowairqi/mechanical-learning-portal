package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	check := flag.String("check-data", "", "validate a progress backup without writing")
	restore := flag.String("restore", "", "validate and restore a backup to DATA_FILE; stop other instances first")
	flag.Parse()
	if *check != "" && *restore != "" {
		log.Fatal("choose -check-data or -restore")
	}
	path := os.Getenv("DATA_FILE")
	if path == "" {
		path = "data/progress.json"
	}
	input := *check
	if *restore != "" {
		input = *restore
	}
	if input != "" {
		if _, err := os.Stat(input); err != nil {
			log.Fatal(err)
		}
		backup, err := newServer("curriculum/helicopter.json", input)
		if err != nil {
			log.Fatal(err)
		}
		if *restore != "" {
			backup.path = path
			if err := backup.save(backup.state); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Restored validated progress to %s\n", path)
		} else {
			fmt.Println("Progress backup is valid (legacy records are normalized on load).")
		}
		return
	}
	s, err := newServer("curriculum/helicopter.json", path)
	if err != nil {
		log.Fatal(err)
	}
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:3000"
	}
	server := &http.Server{Addr: addr, Handler: s.handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()
	log.Printf("Helicopter portal listening on http://%s (data: %s)", addr, path)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	if ctx.Err() != nil {
		<-shutdownDone
	}
}
