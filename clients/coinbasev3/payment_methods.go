package coinbasev3

import (
	"context"
	"github.com/coinbase-samples/advanced-trade-sdk-go/model"
	"github.com/coinbase-samples/advanced-trade-sdk-go/paymentmethods"
)

// GetPaymentMethods get payment methods.
func (c *ApiClient) GetPaymentMethods(ctx context.Context) ([]*model.PaymentMethod, error) {
	paymentMethodsService := paymentmethods.NewPaymentMethodsService(c.restClient)
	paymentMethods, err := paymentMethodsService.ListPaymentMethods(ctx, &paymentmethods.ListPaymentMethodsRequest{})
	if err != nil {
		return []*model.PaymentMethod{}, err
	}
	return paymentMethods.PaymentMethods, nil
}

// PaymentMethods represents the payment methods.
type PaymentMethods struct {
	PaymentMethods []PaymentMethod `json:"payment_methods"`
}

// PaymentMethod represents the payment method.
type PaymentMethod struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Currency string `json:"currency"`
}
