package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port int
	Env  string
	Api  string
	Db   struct {
		Dsn string
	}
	Stripe struct {
		Secret string
		Key    string
	}
}

type Application struct {
	Config        Config
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	TemplateCache map[string]*template.Template
	version       string
}

func (app *Application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.Config.Port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.InfoLog.Printf("Starting bakcend server in %s on port %d", app.Config.Env, app.Config.Port)
	return srv.ListenAndServe()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var cfg Config

	flag.IntVar(&cfg.Port, "port", 4001, "Server port to listen on")
	flag.StringVar(&cfg.Env, "dev", "dev", "Application environment {dev | prod}")

	flag.Parse()

	cfg.Stripe.Key = os.Getenv("STRIPE_KEY")
	cfg.Stripe.Secret = os.Getenv("STRIPE_SECRET")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &Application{
		Config:   cfg,
		InfoLog:  infoLog,
		ErrorLog: errLog,
	}

	err = app.serve()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

}
