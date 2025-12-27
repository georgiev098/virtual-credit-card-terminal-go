package main

import (
	"encoding/json"
	"net/http"
)

type StripePayload struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

type JSONResp struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Content string `json:"content"`
	ID      int    `json:"id"`
}

func (app *Application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	j := JSONResp{
		OK: true,
	}

	out, err := json.MarshalIndent(j, "", "   ")
	if err != nil {
		app.ErrorLog.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}
