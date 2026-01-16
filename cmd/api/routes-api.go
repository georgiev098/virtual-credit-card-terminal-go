package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *Application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Post("/api/payment-intent", app.GetPaymentIntent)
	mux.Post("/api/craete-customer-and-subscribe-to-plan", app.CraeteCustomerAndSubscribeToPlan)

	mux.Get("/api/widget/{id}", app.GetWidgetById)

	// auth
	mux.Post("/api/authenticate", app.CraeteAuthToken)
	mux.Post("/api/is-authenticated", app.CheckIsAuthenticated)

	return mux
}
