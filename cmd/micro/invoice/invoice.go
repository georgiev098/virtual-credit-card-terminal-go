package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

type Config struct {
	Port        int
	FrontEndURL string
}

type Application struct {
	Config   Config
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	version  string
}

func (app *Application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.Config.Port),
		Handler:           app.enableCORS(app.routes()),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.InfoLog.Printf("Starting invoice microservice on port %d", app.Config.Port)
	return srv.ListenAndServe()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var cfg Config

	flag.IntVar(&cfg.Port, "port", 5000, "Server port to listen on")
	flag.StringVar(&cfg.FrontEndURL, "front-end-url", os.Getenv("FRONT_END_URL"), "Front End URL")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &Application{
		Config:   cfg,
		InfoLog:  infoLog,
		ErrorLog: errLog,
	}

	err = app.serve()
	if err != nil {
		app.ErrorLog.Println(err)
		log.Fatal(err)
	}
}
