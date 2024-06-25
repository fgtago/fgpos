package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fgtago/fgweb"
)

type Server struct {
	exit chan struct{}
	wg   sync.WaitGroup
}

func (s *Server) Start(p *Program) {
	shutdownChannel := make(chan bool, 1)

	s.exit = make(chan struct{})
	s.wg.Add(1)

	cfgpath := filepath.Join(p.RootDir, p.ConfigFileName)

	// start jalankan web
	ws, err := fgweb.New(p.RootDir, cfgpath)
	if err != nil {
		log.Fatalf("HTTP server error: %v", err)
		return
	}

	go func() {
		if err := httpserver.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}

		time.Sleep(1 * time.Millisecond)

		log.Println("Stoppend serving new connection")
		shutdownChannel <- true

	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := httpserver.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error : %v", err)
	}

	<-shutdownChannel
	log.Println("Shutdown complete")

}

func (s *Server) Stop() error {
	close(s.exit)
	s.wg.Wait()
	return nil
}

func (s *Server) CreateHandler() {
	defer s.wg.Done()

	// buat service http yang simple
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		<html>
			<head>
				<title>GS</title>
			</head>
			<body>
				golang service home<br>
				<a href="/about">About</a>
			</body>
		</html>
		`)
	})

	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		<html>
			<head>
				<title>GS - About</title>
			</head>
			<body>
				About Page<br>
				back to <a href="/">Home</a>
			</body>
		</html>			
		`)
	})

}
