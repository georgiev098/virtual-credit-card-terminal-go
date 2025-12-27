package card

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
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
