package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/cards"
	"github.com/georgiev098/virtual-credit-card-terminal-go/internal/models"
	urlsigner "github.com/georgiev098/virtual-credit-card-terminal-go/internal/url-signer"
	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v72"
	"golang.org/x/crypto/bcrypt"
)

type StripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Email         string `json:"email"`
	LastFour      string `json:"last_four"`
	Plan          string `json:"plan"`
	CardBrand     string `json:"card_brand"`
	ExpiryMonth   int    `json:"expiry_month"`
	ExpiryYear    int    `json:"expiry_year"`
	ProductID     string `json:"product_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
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
	okay := true
	txnMsg := "Transactoin succesfull"
	var subscribtion *stripe.Subscription

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	if data.Currency == "" {
		data.Currency = "eur"
	}

	app.InfoLog.Printf("DEBUG: Received currency: %s, amount: %s", data.Currency, data.Amount)

	card := cards.Card{
		Secret:   app.Config.Stripe.Secret,
		Key:      app.Config.Stripe.Key,
		Currency: data.Currency,
	}

	stripeCustomer, msg, err := card.CraeteCustomer(data.PaymentMethod, data.Email)
	if err != nil {
		app.ErrorLog.Println(err)
		okay = false
		txnMsg = msg
	}

	if okay {
		subscribtion, err = card.SubscribeToPlan(stripeCustomer, data.Plan, data.Email, data.LastFour, "")
		if err != nil {
			app.ErrorLog.Println(err)
			okay = false
			txnMsg = "Error subscribing customer."
		}

		app.InfoLog.Println("subscribtion ID is:", subscribtion.ID)
	}

	if okay {
		productID, _ := strconv.Atoi(data.ProductID)
		customerID, err := app.SaveCustomer(data.FirstName, data.LastName, data.Email)
		if err != nil {
			app.ErrorLog.Println(err)
			return
		}

		// craete a new transactoin
		amount, _ := strconv.Atoi(data.Amount)
		// expityMonth, _ := strconv.Atoi(data.ExpiryMonth)
		// expiryYear, _ := strconv.Atoi(data.ExpiryYear)
		txn := models.Transaction{
			Amount:              amount,
			Currency:            "USD",
			LastFour:            data.LastFour,
			ExpiryMonth:         data.ExpiryMonth,
			ExpiryYear:          data.ExpiryYear,
			TransactionStatusID: 2,
		}

		txnID, err := app.SaveTransaction(txn)
		if err != nil {
			app.ErrorLog.Println(err)
			return
		}

		order := models.Order{
			WidgetID:      productID,
			TransactionID: txnID,
			CustomerID:    customerID,
			StatusID:      1,
			Quantity:      1,
			Amount:        amount,
			CreatedAt:     time.Now(),
			UpdateddAt:    time.Now(),
		}

		_, err = app.SaveOrder(order)
		if err != nil {
			app.ErrorLog.Println(err)
			return
		}

	}

	jsonResp := JSONResp{
		OK:      okay,
		Message: txnMsg,
	}

	out, err := json.Marshal(jsonResp)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (app *Application) CraeteAuthToken(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.ReadJSON(w, r, &userInput)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	// get user from db
	user, err := app.DB.GetUserByEmail(userInput.Email)
	if err != nil {
		app.InvalidCredentials(w)
		app.ErrorLog.Println(err)
		return
	}

	// validate passsword
	validPass, err := app.ValidatePassword(user.Password, userInput.Password)
	if err != nil {
		app.InvalidCredentials(w)
		return
	}

	if !validPass {
		app.InvalidCredentials(w)
		return
	}

	// generate token
	token, err := models.GenerateToken(user.ID, 24*time.Hour, models.ScopeAuthentication)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	// save token to DB
	err = app.DB.InsertToken(token, user)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}
	// send resp

	var payload struct {
		Error   bool          `json:"error"`
		Message string        `json:"message"`
		Token   *models.Token `json:"authentication_token"`
	}

	payload.Error = false
	payload.Message = fmt.Sprintf("Token for %s craeted!", user.FirstName)
	payload.Token = token

	_ = app.WriteJSON(w, http.StatusOK, payload)
}

func (app *Application) AuthenticateToken(r *http.Request) (*models.User, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return nil, errors.New("No authorization header received.")
	}

	headerParts := strings.Split(authHeader, " ")

	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("No authorization header received.")
	}

	headerToken := headerParts[1]

	if len(headerToken) != 26 {
		return nil, errors.New("Authentication token wrong length.")
	}

	// get user from tokens table
	user, err := app.DB.GetUserByToken(headerToken)
	if err != nil {
		return nil, errors.New("No matching users found")
	}

	return user, nil
}

func (app *Application) CheckIsAuthenticated(w http.ResponseWriter, r *http.Request) {

	// validate token and get user
	user, err := app.AuthenticateToken(r)
	if err != nil {
		app.InvalidCredentials(w)
		return
	}

	// valid user
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = false
	payload.Message = fmt.Sprintf("Authenticated user %s", user.Email)

	app.WriteJSON(w, http.StatusOK, payload)
}

func (app *Application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	var txnData struct {
		PaymentAmount   int    `json:"payment_amount"`
		PaymentCurrency string `json:"currency"`
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		Email           string `json:"email"`
		PaymentIntent   string `json:"payment_intent"`
		PaymentMethod   string `json:"payment_method"`
		BankReturnCode  string `json:"bank_return_code"`
		ExpiryMonth     int    `json:"expiry_month"`
		ExpiryYear      int    `json:"expiry_year"`
		LastFour        string `json:"last_four"`
	}

	err := app.ReadJSON(w, r, &txnData)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	card := cards.Card{
		Secret: app.Config.Stripe.Secret,
		Key:    app.Config.Stripe.Key,
	}

	pi, err := card.RetrievePaymentIntent(txnData.PaymentIntent)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	pm, err := card.GetPaymentMethod(txnData.PaymentMethod)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	txnData.LastFour = pm.Card.Last4
	txnData.ExpiryMonth = int(pm.Card.ExpMonth)
	txnData.ExpiryYear = int(pm.Card.ExpYear)

	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		PaymentIntent:       txnData.PaymentIntent,
		PaymentMethod:       txnData.PaymentMethod,
		BankReturnCode:      pi.Charges.Data[0].ID,
		TransactionStatusID: 2,
	}

	app.InfoLog.Println("TXN:", txn)

	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	app.WriteJSON(w, http.StatusOK, txn)

}

func (app *Application) SendResetPasswordLink(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}

	err := app.ReadJSON(w, r, &payload)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	// verify that email exists
	_, err = app.DB.GetUserByEmail(payload.Email)
	if err != nil {
		var resp struct {
			Error   bool   `json:"error"`
			Message string `json:"message"`
		}

		resp.Error = true
		resp.Message = "No matching email found on our system."

		app.WriteJSON(w, http.StatusNotAcceptable, resp)
		return
	}

	link := fmt.Sprintf("%s/reset-password?email=%s", app.Config.FrontEndURL, payload.Email)

	sign := urlsigner.Signer{
		Secret: []byte(app.Config.SecretKey),
	}

	signedLink := sign.GenerateTokenFromString(link)

	var data struct {
		Link string
	}

	data.Link = signedLink

	// send email
	err = app.SendEmail("info@widgets.com", payload.Email, "password reset request", "forgot-password", data)
	if err != nil {
		app.ErrorLog.Println(err)
		app.BadRequest(w, r, err)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false

	app.WriteJSON(w, http.StatusCreated, resp)
}

func (app *Application) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.ReadJSON(w, r, &payload)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	user, err := app.DB.GetUserByEmail(payload.Email)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	err = app.DB.UpdateUserPassword(user, string(newHash))
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	resp := struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}{
		Error:   false,
		Message: "Password Changed",
	}

	app.WriteJSON(w, http.StatusCreated, resp)

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
