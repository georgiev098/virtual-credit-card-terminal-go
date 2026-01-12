package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/cards"
	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/models"
	"github.com/go-chi/chi/v5"
)

type TransactionData struct {
	FirstName       string
	LastName        string
	Email           string
	PaymentIntentID string
	PaymentMethodID string
	Amount          int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

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

func (app *Application) GetTransactionData(r *http.Request) (TransactionData, error) {
	var txnData TransactionData

	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
		return txnData, err
	}

	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	email := r.Form.Get("cardholder_email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	amount, _ := strconv.Atoi(paymentAmount)
	card := cards.Card{
		Secret: app.Config.Stripe.Secret,
		Key:    app.Config.Stripe.Key,
	}
	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.ErrorLog.Println(err)
		return txnData, err
	}
	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.ErrorLog.Println(err)
		return txnData, err
	}
	lastFour := pm.Card.Last4
	expMonth := pm.Card.ExpMonth
	expYear := pm.Card.ExpYear
	bankReturnCode := pi.Charges.Data[0].ID

	txnData = TransactionData{
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		Amount:          amount,
		PaymentIntentID: paymentIntent,
		PaymentMethodID: paymentMethod,
		PaymentCurrency: paymentCurrency,
		LastFour:        lastFour,
		ExpiryMonth:     int(expMonth),
		ExpiryYear:      int(expYear),
		BankReturnCode:  bankReturnCode,
	}

	return txnData, nil
}

func (app *Application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	// create a transaction
	txn := models.Transaction{
		Amount:              txnData.Amount,
		LastFour:            txnData.LastFour,
		Currency:            txnData.PaymentCurrency,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		TransactionStatusID: 2,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		CreatedAt:           time.Now(),
		UpdateddAt:          time.Now(),
	}

	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	// write to session and redirect page
	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w, r, "/virtual-terminal-receipt", http.StatusSeeOther)
}

func (app *Application) VirtualTerminalReceipt(w http.ResponseWriter, r *http.Request) {

	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]any)
	data["txn"] = txn
	app.Session.Remove(r.Context(), "receipt")

	if err := app.renderTemplate(w, r, "virtual-terminal-receipt", &templateData{
		Data: data,
	}); err != nil {
		app.ErrorLog.Println(err)
		return
	}
}

func (app *Application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	widgetId, _ := strconv.Atoi(r.Form.Get("product_id"))

	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	// create new customer
	customerID, err := app.SaveCustomer(txnData.FirstName, txnData.LastName, txnData.Email)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	// create a transaction
	txn := models.Transaction{
		Amount:              txnData.Amount,
		LastFour:            txnData.LastFour,
		Currency:            txnData.PaymentCurrency,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		TransactionStatusID: 2,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		CreatedAt:           time.Now(),
		UpdateddAt:          time.Now(),
	}

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
		Amount:        txn.Amount,
		CreatedAt:     time.Now(),
		UpdateddAt:    time.Now(),
	}

	_, err = app.SaveOrder(order)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	// write to session and redirect page
	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)
}

func (app *Application) Receipt(w http.ResponseWriter, r *http.Request) {

	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]any)
	data["txn"] = txn
	app.Session.Remove(r.Context(), "receipt")

	if err := app.renderTemplate(w, r, "receipt", &templateData{
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

func (app *Application) BronzePlan(w http.ResponseWriter, r *http.Request) {
	intMap := make(map[string]int)
	intMap["plan_id"] = 1

	if err := app.renderTemplate(w, r, "bronze-plan", &templateData{
		IntMap: intMap,
	}); err != nil {
		app.ErrorLog.Print(err)
	}
}
