package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/driver"
	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/models"
	"github.com/joho/godotenv"
)

var session *scs.SessionManager

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
	DB            models.DBModel
	Session       *scs.SessionManager
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

	app.InfoLog.Printf("Starting HTTP server in %s on port %d", app.Config.Env, app.Config.Port)
	return srv.ListenAndServe()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var cfg Config

	flag.IntVar(&cfg.Port, "port", 4000, "Server port to listen on")
	flag.StringVar(&cfg.Env, "dev", "dev", "Application environment {dev | prod}")
	flag.StringVar(&cfg.Db.Dsn, "dsn", fmt.Sprintf("%s:%s@tcp(localhost:3306)/widgets?parseTime=true&tls=false", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")), "DSN")
	flag.StringVar(&cfg.Api, "api", "http://localhost:4000", "URL to Api")

	flag.Parse()

	cfg.Stripe.Key = os.Getenv("STRIPE_KEY")
	cfg.Stripe.Secret = os.Getenv("STRIPE_SECRET")

	if cfg.Stripe.Key == "" {
		log.Fatal("STRIPE_KEY not set in environment")
	}

	if cfg.Stripe.Secret == "" {
		log.Fatal("STRIPE_SECRET not set in environment")
	}

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Create a new Session
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	// Connect to DB
	con, err := driver.OpenDB(cfg.Db.Dsn)
	if err != nil {
		errLog.Fatal(err)
	}

	defer con.Close()

	tc := make(map[string]*template.Template)

	app := &Application{
		Config:        cfg,
		InfoLog:       infoLog,
		ErrorLog:      errLog,
		TemplateCache: tc,
		DB: models.DBModel{
			DB: con,
		},
		Session: sessionManager,
	}

	err = app.serve()
	if err != nil {
		app.ErrorLog.Println(err)
		log.Fatal(err)
	}

}
