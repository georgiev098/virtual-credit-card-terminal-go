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
	mux.Post("/api/forgot-password", app.SendResetPasswordLink)
	mux.Post("/api/password-reset", app.ResetPassword)

	mux.Route("/api/admin", func(mux chi.Router) {
		mux.Use(app.Auth)

		mux.Post("/virtual-temrinal-succeeded", app.VirtualTerminalPaymentSucceeded)
		mux.Post("/all-sales", app.AllSales)
		mux.Post("/all-subscriptions", app.AllSubscriptions)

		mux.Post("/sales/{id}", app.GetOneSale)
	})

	return mux
}
