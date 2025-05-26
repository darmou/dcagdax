package exchanges

//go:generate mockgen -destination=../mocks/mock_exchange.go -package=mocks github.com/sberserker/dcagdax/exchanges Exchange

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type CalcLimitOrder func(askPrice decimal.Decimal, fiatAmount decimal.Decimal) (orderPrice decimal.Decimal, orderSize decimal.Decimal)

type Exchange interface {
	GetTickerSymbol(baseCurrency string, quoteCurrency string) string

	GetTicker(ctx context.Context, productId string) (*Ticker, error)

	GetProduct(ctx context.Context, productId string) (*Product, error)

	Deposit(ctx context.Context, currency string, amount float64) (*time.Time, error)

	CreateOrder(ctx context.Context, productId string, amount float64, orderType OrderTypeType, limitOrderFunc CalcLimitOrder) (*Order, error)

	LastPurchaseTime(ctx context.Context, coin string, currency string, since time.Time) (*time.Time, error)

	GetFiatAccount(ctx context.Context, currency string) (*Account, error)

	GetPendingTransfers(currency string) ([]PendingTransfer, error)
}

type OrderTypeType int32

const (
	Market OrderTypeType = 0
	Limit  OrderTypeType = 1
)

type Order struct {
	Symbol  string
	OrderID string
}

type Ticker struct {
	Price float64
}

type Product struct {
	QuoteCurrency string
	BaseCurrency  string
	BaseMinSize   float64
}

type Account struct {
	Available float64
}

type PendingTransfer struct {
	Amount float64
}
