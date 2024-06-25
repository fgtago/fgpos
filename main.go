package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/judwhite/go-svc"
)

type Program struct {
	LogFile        *os.File
	svr            *Server
	ctx            context.Context
	ConfigFileName string
	RootDir        string
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	prg := Program{
		svr: &Server{},
		ctx: ctx,
	}

	defer func() {
		if prg.LogFile != nil {
			err := prg.LogFile.Close()
			if err != nil {
				log.Printf("err '%s' : %v\n", prg.LogFile.Name(), err)
			}
		}
	}()

	err := svc.Run(&prg)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Program) Context() context.Context {
	return p.ctx
}

func (p *Program) Init(env svc.Environment) error {

	rd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	p.RootDir = rd

	log.Printf("is win service? %v\n", env.IsWindowsService())
	if env.IsWindowsService() {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return err
		}

		logPath := filepath.Join(dir, "testservice.log")
		f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		p.LogFile = f
		p.ConfigFileName = "config.yml"

		log.Println("logfile", p.LogFile)
		log.SetOutput(f)
	} else {
		p.ConfigFileName = "config-dev.yml"
	}

	return nil
}

func (p *Program) Start() error {
	log.Printf("Starting...\n")

	go p.svr.Start(p)
	return nil
}

func (p *Program) Stop() error {
	log.Printf("Stopping...\n")
	err := p.svr.Stop()
	if err != nil {
		log.Println("Error")
		log.Println(err.Error())
		return err
	}
	log.Printf("Stopped.\n")
	return nil
}
