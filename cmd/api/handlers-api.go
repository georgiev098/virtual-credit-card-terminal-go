package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/cards"
	"github.com/go-chi/chi/v5"
)

type StripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Email         string `json:"email"`
	LastFour      string `json:"last_four"`
	Plan          string `json:"plan"`
}

type JSONResp struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	ID      int    `json:"id,omitempty"`
}

func (app *Application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {

	var payload StripePayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.ErrorLog.Println(err)
	}

	amount, err := strconv.Atoi(payload.Amount)
	if err != nil {
		app.ErrorLog.Println(err)
	}

	card := cards.Card{
		Secret:   app.Config.Stripe.Secret,
		Key:      app.Config.Stripe.Key,
		Currency: payload.Currency,
	}

	ok := true

	pi, msg, err := card.Charge(payload.Currency, amount)
	if err != nil {
		ok = false
	}

	if ok {
		out, err := json.MarshalIndent(pi, "", "   ")
		if err != nil {
			app.ErrorLog.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	} else {

		j := JSONResp{
			OK:      false,
			Message: msg,
			Content: "",
		}

		out, err := json.MarshalIndent(j, "", "   ")
		if err != nil {
			app.ErrorLog.Println(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	}

}

func (app *Application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *Application) GetWidgetById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	widgetId, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetId)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	out, err := json.Marshal(widget)

	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.Write(out)
}

func (app *Application) CraeteCustomerAndSubscribeToPlan(w http.ResponseWriter, r *http.Request) {
	var data StripePayload
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	card := cards.Card{
		Secret:   app.Config.Stripe.Secret,
		Key:      app.Config.Stripe.Key,
		Currency: data.Currency,
	}

	stripeCustomer, msg, err := card.CraeteCustomer(data.PaymentMethod, data.Email)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	subscribtionID, err := card.SubscribeToPlan(stripeCustomer, data.Plan, data.Email, data.LastFour, "")
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	app.InfoLog.Println("subscribtion ID is:", subscribtionID)

	okay := true
	// msg := ""

	jsonResp := JSONResp{
		OK:      okay,
		Message: msg,
	}

	out, err := json.Marshal(jsonResp)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}
