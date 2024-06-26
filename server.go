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
	"github.com/go-chi/chi/v5"
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
	ws, err := fgweb.New(p.RootDir, cfgpath)
	if err != nil {
		log.Fatalf("HTTP server error: %v", err)
		return
	}

	port := ws.Configuration.Port
	log.Println("Service running on port", port)
	httpserver := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	go func() {

		ws.Mux = fgweb.CreateRequestHandler(func(mux *chi.Mux) error {
			return Router(mux)
		})

		httpserver.Handler = ws.Mux

		if err := httpserver.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}

		time.Sleep(1 * time.Millisecond)

		log.Println("Stoppend serving new connection")
		shutdownChannel <- true
		s.wg.Done()

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
