package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *Application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Post("/api/payment-intent", app.GetPaymentIntent)
	mux.Get("/api/widget/{id}", app.GetWidgetById)
	return mux
}
