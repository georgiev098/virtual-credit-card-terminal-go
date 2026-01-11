package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *Application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(app.SessionLoad)

	mux.Get("/", app.Home)
	mux.Get("/virtual-terminal", app.VirtualTerminal)
	mux.Get("/receipt", app.Receipt)
	mux.Post("/payment-succeeded", app.PaymentSucceeded)

	fileServer := http.FileServer(http.Dir("./static"))

	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Get("/widget/{id}", app.ChargeOnce)
	return mux
}
