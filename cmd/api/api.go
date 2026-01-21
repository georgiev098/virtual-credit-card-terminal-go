package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/driver"
	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/models"
	"github.com/joho/godotenv"
)

type Config struct {
	Port int
	Env  string
	Db   struct {
		Dsn string
	}
	Stripe struct {
		Secret string
		Key    string
	}
	SMTP struct {
		Username string
		Password string
		Port     int
		Host     string
	}
}

type Application struct {
	Config   Config
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	DB       models.DBModel
}

func (app *Application) serve() error {
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", app.Config.Port),
		// Handler:           app.routes(),
		Handler:           app.enableCORS(app.routes()),
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
	flag.StringVar(&cfg.Db.Dsn, "dsn", fmt.Sprintf("%s:%s@tcp(localhost:3306)/widgets?parseTime=true&tls=false", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")), "DSN")
	// SMTP
	flag.StringVar(&cfg.SMTP.Host, "smtp-host", os.Getenv("MAILTRAP_HOST"), "SMTP Host")
	flag.StringVar(&cfg.SMTP.Username, "smtp-username", os.Getenv("MAILTRAP_USER"), "SMTP Username")
	flag.StringVar(&cfg.SMTP.Password, "smtp-password", os.Getenv("MAILTRAP_PASSWORD"), "SMTP Password")

	smtpPort := 2525
	if p, err := strconv.Atoi(os.Getenv("MAILTRAP_PORT")); err == nil {
		smtpPort = p
	}
	flag.IntVar(&cfg.SMTP.Port, "port", smtpPort, "SMTP port")

	flag.Parse()

	cfg.Stripe.Key = os.Getenv("STRIPE_KEY")
	cfg.Stripe.Secret = os.Getenv("STRIPE_SECRET")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	con, err := driver.OpenDB(cfg.Db.Dsn)
	if err != nil {
		errLog.Fatal(err)
	}

	defer con.Close()

	app := &Application{
		Config:   cfg,
		InfoLog:  infoLog,
		ErrorLog: errLog,
		DB: models.DBModel{
			DB: con,
		},
	}

	err = app.serve()
	if err != nil {
		errLog.Fatalf("server error: %v", err)
	}
}
