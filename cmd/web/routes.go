package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *Application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(app.SessionLoad)

	mux.Get("/", app.Home)
	mux.Get("ws", app.WsEndPoint)

	mux.Route("/admin", func(r chi.Router) {
		r.Use(app.Auth)
		r.Get("/virtual-terminal", app.VirtualTerminal)

		r.Get("/sales", app.AllSales)
		r.Get("/sales/{id}", app.OneSale)

		r.Get("/subscriptions", app.AllSubcriptions)
		r.Get("/subscriptions/{id}", app.OneSubscription)

		r.Get("/all-users", app.AllSUsers)
		r.Get("/all-users/{id}", app.OneSUser)
	})

	mux.Get("/receipt", app.Receipt)
	mux.Post("/payment-succeeded", app.PaymentSucceeded)

	mux.Get("/plans/bronze", app.BronzePlan)
	mux.Get("/receipt/bronze", app.BronzePlanReceipt)

	// auth routes
	mux.Get("/login", app.Login)
	mux.Post("/login", app.PostLogin)
	mux.Get("/logout", app.Logout)
	mux.Get("/forgot-password", app.ForgotReset)
	mux.Get("/reset-password", app.ShowResetPassword)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Get("/widget/{id}", app.ChargeOnce)
	return mux
}
