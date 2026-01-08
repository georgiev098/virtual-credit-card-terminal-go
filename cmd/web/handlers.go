package main

import (
	"net/http"
	"strconv"

	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/cards"
	"github.com/go-chi/chi/v5"
)

func (app *Application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "terminal", nil, "stripe-js"); err != nil {
		app.ErrorLog.Println(err)
	}
}

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", nil); err != nil {
		app.ErrorLog.Println(err)
	}
}

func (app *Application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	cardholder := r.Form.Get("cardholder_name")
	email := r.Form.Get("cardholder_email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")

	card := cards.Card{
		Secret: app.Config.Stripe.Secret,
		Key:    app.Config.Stripe.Key,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	lastFour := pm.Card.Last4
	expMonth := pm.Card.ExpMonth
	expYear := pm.Card.ExpYear

	data := make(map[string]any)
	data["cardholder"] = cardholder
	data["email"] = email
	data["pi"] = paymentIntent
	data["pm"] = paymentMethod
	data["pa"] = paymentAmount
	data["pc"] = paymentCurrency
	data["last_four"] = lastFour
	data["expiry_month"] = expMonth
	data["expiry_year"] = expYear
	data["bank_return_code"] = pi.Charges.Data[0].ID

	if err := app.renderTemplate(w, r, "succeeded", &templateData{
		Data: data,
	}); err != nil {
		app.ErrorLog.Println(err)
		return
	}

}

func (app *Application) ChargeOnce(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	widgetId, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetId)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	data := make(map[string]any)
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "buy-once", &templateData{
		Data: data,
	}, "stripe-js"); err != nil {
		app.ErrorLog.Println(err)
	}
}
