package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/cards"
	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/models"
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
	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	email := r.Form.Get("cardholder_email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	widgetId, _ := strconv.Atoi(r.Form.Get("product_id"))

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
	bankReturnCode := pi.Charges.Data[0].ID

	// create new customer
	customerID, err := app.SaveCustomer(firstName, lastName, email)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	// create a transaction
	amount, _ := strconv.Atoi(paymentAmount)
	txn := models.Transaction{
		Amount:              amount,
		LastFour:            lastFour,
		Currency:            paymentCurrency,
		ExpiryMonth:         int(expMonth),
		ExpiryYear:          int(expYear),
		BankReturnCode:      bankReturnCode,
		TransactionStatusID: 2,
		CreatedAt:           time.Now(),
		UpdateddAt:          time.Now(),
	}
	app.InfoLog.Println(customerID)

	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}
	app.InfoLog.Println(txnID)

	// create order
	order := models.Order{
		WidgetID:      widgetId,
		TransactionID: txnID,
		CustomerID:    customerID,
		StatusID:      1,
		Quantity:      1,
		Amount:        amount,
		CreatedAt:     time.Now(),
		UpdateddAt:    time.Now(),
	}

	newOrderID, err := app.SaveOrder(order)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}
	app.InfoLog.Println(newOrderID)

	data := make(map[string]any)
	data["cardholder"] = cardholder
	data["first_name"] = firstName
	data["last_name"] = lastName
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

func (app *Application) SaveCustomer(firstName string, lastName string, email string) (int, error) {
	customer := models.Customer{
		FirstName:  firstName,
		LastName:   lastName,
		Email:      email,
		CreatedAt:  time.Now(),
		UpdateddAt: time.Now(),
	}

	newCustomerID, err := app.DB.InsertNewCustomer(customer)
	if err != nil {
		return 0, err
	}

	return newCustomerID, nil
}

func (app *Application) SaveTransaction(txn models.Transaction) (int, error) {
	newTransactionID, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}

	return newTransactionID, nil
}

func (app *Application) SaveOrder(order models.Order) (int, error) {
	newOrderID, err := app.DB.InsertNewOrder(order)
	if err != nil {
		return 0, err
	}

	return newOrderID, nil
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
