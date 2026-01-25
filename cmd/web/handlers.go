package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/cards"
	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/encryption"
	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/models"
	urlsigner "github.com/georgiev098/virtual-credit-card-terminal-go/internal/url-signer"
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
	if err := app.renderTemplate(w, r, "terminal", nil); err != nil {
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

	widget, err := app.DB.GetWidget(2)
	if err != nil {
		app.ErrorLog.Print(err)
		return
	}

	data := make(map[string]any)
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "bronze-plan", &templateData{
		Data: data,
	}, "stripe-js"); err != nil {
		app.ErrorLog.Println(err)
	}
}

func (app *Application) BronzePlanReceipt(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "receipt-plan", &templateData{}, "stripe-js"); err != nil {
		app.ErrorLog.Println(err)
	}
}

func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "login", &templateData{}, "stripe-js"); err != nil {
		app.ErrorLog.Println(err)
	}
}

func (app *Application) PostLogin(w http.ResponseWriter, r *http.Request) {
	if err := app.Session.RenewToken(r.Context()); err != nil {
		app.ErrorLog.Println(err)
	}

	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	app.InfoLog.Println(password)
	id, err := app.DB.Authenticate(email, password)
	if err != nil {
		app.ErrorLog.Println(err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	app.InfoLog.Println(id)

	app.Session.Put(r.Context(), "userID", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (app *Application) Logout(w http.ResponseWriter, r *http.Request) {
	app.Session.Destroy(r.Context())
	app.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) ForgotReset(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "forgot-password", &templateData{}, "stripe-js"); err != nil {
		app.ErrorLog.Println(err)
	}
}

func (app *Application) ShowResetPassword(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	theURL := r.RequestURI
	testURL := fmt.Sprintf("%s%s", app.Config.FrontEndURL, theURL)

	signer := urlsigner.Signer{
		Secret: []byte(app.Config.SecretKey),
	}

	valid := signer.VerifyToken(testURL)
	if !valid {
		app.ErrorLog.Println("Invalid URL. Tempering detected")
		return
	}

	// make sure not expired
	expired := signer.IsTokenExpired(testURL, 60)
	if expired {
		app.ErrorLog.Println("Link expired.")
		return
	}

	encryptor := encryption.Encryption{
		Key: []byte(app.Config.SecretKey),
	}

	encryptedEmail, err := encryptor.Encrypt(email)
	if err != nil {
		app.ErrorLog.Println("Encryption failed")
		return
	}

	app.InfoLog.Println(encryptedEmail)
	data := make(map[string]any, 0)
	data["email"] = encryptedEmail

	if err := app.renderTemplate(w, r, "reset-password", &templateData{
		Data: data,
	}, "stripe-js"); err != nil {
		app.ErrorLog.Println(err)
	}
}

func (app *Application) AllSales(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "all-sales", &templateData{}, "stripe-js"); err != nil {
		app.ErrorLog.Println(err)
	}
}

func (app *Application) OneSale(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string, 0)
	stringMap["title"] = "Sale"
	stringMap["cancel"] = "/admin/sales"
	stringMap["refund-url"] = "/api/admin/refund"
	stringMap["refund-btn"] = "Refund Order"
	stringMap["refund-badge"] = "Refunded"
	stringMap["refund-msg"] = "Charge Refunded"

	if err := app.renderTemplate(w, r, "sale", &templateData{
		StringMap: stringMap,
	}, "stripe-js"); err != nil {
		app.ErrorLog.Println(err)
	}
}

func (app *Application) OneSubscription(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string, 0)
	stringMap["title"] = "Subscription"
	stringMap["cancel"] = "/admin/subscriptions"
	stringMap["refund-url"] = "/api/admin/cancel-subscription"
	stringMap["refund-btn"] = "Cancel Subscription"
	stringMap["refund-badge"] = "Cancelled"
	stringMap["refund-msg"] = "Subscription Cancelled"
	if err := app.renderTemplate(w, r, "sale", &templateData{
		StringMap: stringMap,
	}, "stripe-js"); err != nil {
		app.ErrorLog.Println(err)
	}
}

func (app *Application) AllSubcriptions(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "all-subscriptions", &templateData{}, "stripe-js"); err != nil {
		app.ErrorLog.Println(err)
	}
}
