package cards

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/refund"
	"github.com/stripe/stripe-go/v72/sub"
)

type Card struct {
	Secret   string
	Key      string
	Currency string
}

type Transaction struct {
	TransactionStatusID int
	Amount              int
	Currency            string
	LastFour            string
	BankReturnCode      string
}

func (card *Card) Charge(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return card.CreatePaymentIntent(currency, amount)
}

func (card *Card) CreatePaymentIntent(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	stripe.Key = card.Secret
	// craete payment intent

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount)),
		Currency: stripe.String(currency),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrMsg(stripeErr.Code)
		}
		return nil, msg, err
	}

	return pi, "", nil
}

func (c *Card) GetPaymentMethod(s string) (*stripe.PaymentMethod, error) {
	stripe.Key = c.Secret
	pm, err := paymentmethod.Get(s, nil)
	if err != nil {
		return nil, err
	}

	return pm, nil
}

func (c *Card) RetrievePaymentIntent(s string) (*stripe.PaymentIntent, error) {
	stripe.Key = c.Secret

	pi, err := paymentintent.Get(s, nil)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

func (c *Card) CraeteCustomer(pm string, email string) (*stripe.Customer, string, error) {
	stripe.Key = c.Secret

	customerParams := &stripe.CustomerParams{
		PaymentMethod: stripe.String(pm),
		Email:         stripe.String(email),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm),
		},
	}

	cust, err := customer.New(customerParams)
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrMsg(stripeErr.Code)
		}
		return nil, msg, err
	}

	return cust, "", nil
}

func (c *Card) SubscribeToPlan(cust *stripe.Customer, plan string, email string, last4 string, cardType string) (*stripe.Subscription, error) {
	stripeCustomerID := cust.ID
	items := []*stripe.SubscriptionItemsParams{
		{
			Plan: stripe.String(plan),
		},
	}
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(stripeCustomerID),
		Items:    items,
		Currency: stripe.String("eur"),
	}

	params.AddMetadata("last_four", last4)
	params.AddMetadata("card_type", cardType)
	params.AddExpand("lastest_invoice.payment_intent")

	subscription, err := sub.New(params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

func (c *Card) Refund(chargeID string, amount int) error {
	stripe.Key = c.Secret

	amountToRefund := int64(amount)

	params := &stripe.RefundParams{
		Charge: stripe.String(chargeID),
		Amount: stripe.Int64(amountToRefund),
	}

	_, err := refund.New(params)
	return err
}

func cardErrMsg(code stripe.ErrorCode) string {
	var msg string

	switch code {
	case stripe.ErrorCodeCardDeclined:
		msg = "Your card was declined."
	case stripe.ErrorCodeExpiredCard:
		msg = "Your card is expired."
	case stripe.ErrorCodeIncorrectCVC:
		msg = "Incorrect CVV code."
	case stripe.ErrorCodeIncorrectZip:
		msg = "Incorrect zip/postal code."
	case stripe.ErrorCodeAmountTooLarge:
		msg = "The amount is too large to charge your card."
	case stripe.ErrorCodeAmountTooSmall:
		msg = "The amount is too small to charge your card."
	case stripe.ErrorCodeBalanceInsufficient:
		msg = "Insufficient balance."
	default:
		msg = "Your card was declined "
	}

	return msg
}
