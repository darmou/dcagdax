package exchanges

import (
	"context"
	"errors"
	"fmt"
	"github.com/coinbase-samples/advanced-trade-sdk-go/accounts"
	"github.com/coinbase-samples/advanced-trade-sdk-go/client"
	"github.com/coinbase-samples/advanced-trade-sdk-go/model"
	"github.com/coinbase-samples/advanced-trade-sdk-go/orders"
	"github.com/coinbase-samples/advanced-trade-sdk-go/paymentmethods"
	"github.com/coinbase-samples/advanced-trade-sdk-go/portfolios"
	"github.com/coinbase-samples/advanced-trade-sdk-go/products"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	exchange "github.com/sberserker/dcagdax/clients/coinbase"
	"github.com/sberserker/dcagdax/clients/coinbasev3"
	"github.com/shopspring/decimal"
)

type CoinbaseV3 struct {
	portfolio       portfolios.PortfoliosService
	products        products.ProductsService
	payment         paymentmethods.PaymentMethodsService
	accountsService accounts.AccountsService
	orders          orders.OrdersService
	client3         client.RestClient
	client          *exchange.Client
	accounts        map[string]*account
}

type account struct {
	Id        string  `json:"id"`
	Hold      float64 `json:"hold,string"`
	Available float64 `json:"available,string"`
	Currency  string  `json:"currency"`
}

func NewCoinbaseV3() (*CoinbaseV3, error) {
	secret := os.Getenv("COINBASE_SECRET")
	key := os.Getenv("COINBASE_KEY")
	portfolioId := os.Getenv("PORTFOLIO_ID")

	if secret == "" {
		return nil, errors.New("COINBASE_SECRET environment variable is required")
	}

	if key == "" {
		return nil, errors.New("COINBASE_KEY environment variable is required")
	}

	if portfolioId == "" {
		return nil, errors.New("PORTFOLIO_ID environment variable is required")
	}

	// to allow new token pem format be passed via .env file
	secret = strings.ReplaceAll(secret, `\n`, "\n")

	client := exchange.NewClient(secret, key, "")
	client3 := coinbasev3.NewApiClient(key, secret, portfolioId)
	portfolio := portfolios.NewPortfoliosService(client3.GetClient())
	products := products.NewProductsService(client3.GetClient())
	payment := paymentmethods.NewPaymentMethodsService(client3.GetClient())
	orders := orders.NewOrdersService(client3.GetClient())
	accountsService := accounts.NewAccountsService(client3.GetClient())

	return &CoinbaseV3{
		accounts:        map[string]*account{},
		portfolio:       portfolio,
		products:        products,
		accountsService: accountsService,
		payment:         payment,
		orders:          orders,
		client3:         client3.GetClient(),
		client:          client,
	}, nil
}

func (c *CoinbaseV3) CreateOrder(ctx context.Context, productId string, amount float64, orderType OrderTypeType, limitOrderFunc CalcLimitOrder) (*Order, error) {

	var orderReq orders.CreateOrderRequest

	if orderType == Limit {
		marketTradeRequest := products.GetMarketTradesRequest{
			ProductId: productId,
			Limit:     "10",
		}
		trades, err := c.products.GetMarketTrades(ctx, &marketTradeRequest)
		if err != nil {
			return nil, err
		}

		bestAsk, err := decimal.NewFromString(trades.BestAsk)
		if err != nil {
			return nil, err
		}
		orderPrice, orderSize := limitOrderFunc(bestAsk, decimal.NewFromFloat(amount))

		orderConfig := model.OrderConfiguration{
			LimitLimitGtc: &model.LimitGtc{
				BaseSize:   orderPrice.String(),
				LimitPrice: orderSize.String(),
			},
		}
		orderReq = orders.CreateOrderRequest{
			ProductId:          productId,
			OrderConfiguration: orderConfig,
			Side:               coinbasev3.OrderSideBuy,
			ClientOrderId:      uuid.NewString(),
		}
	} else {
		orderConfig := model.OrderConfiguration{
			MarketMarketIoc: &model.MarketIoc{
				QuoteSize: decimal.NewFromFloat(amount).StringFixedBank(2),
			},
		}
		orderReq = orders.CreateOrderRequest{
			ProductId:          productId,
			OrderConfiguration: orderConfig,
			Side:               coinbasev3.OrderSideBuy,
			ClientOrderId:      uuid.NewString(),
		}
	}
	orderConfig := model.OrderConfiguration{
		MarketMarketIoc: &model.MarketIoc{
			QuoteSize: decimal.NewFromFloat(amount).StringFixedBank(2),
		},
	}
	orderReq = orders.CreateOrderRequest{
		ProductId:          productId,
		OrderConfiguration: orderConfig,
		Side:               coinbasev3.OrderSideBuy,
		ClientOrderId:      uuid.NewString(),
	}

	order, err := c.orders.CreateOrder(ctx, &orderReq)

	if err != nil {
		return nil, err
	}

	if !order.Success {
		return nil, errors.New(fmt.Sprintf("order failed with %s, %s", order.FailureReason, order.ErrorResponse.Message))
	}

	return &Order{
		Symbol:  order.SuccessResponse.ProductId,
		OrderID: order.SuccessResponse.OrderId,
	}, nil
}

func (c *CoinbaseV3) GetTickerSymbol(baseCurrency string, quoteCurrency string) string {
	return baseCurrency + "-" + quoteCurrency
}

func (c *CoinbaseV3) GetTicker(ctx context.Context, productId string) (*Ticker, error) {
	marketTradeRequest := products.GetMarketTradesRequest{
		ProductId: productId,
		Limit:     "10",
	}
	ticker, err := c.products.GetMarketTrades(ctx, &marketTradeRequest)
	if err != nil {
		return nil, err
	}

	bestAsk, err := strconv.ParseFloat(ticker.BestAsk, 64)
	if err != nil {
		return nil, err
	}

	return &Ticker{Price: bestAsk}, nil
}

func (c *CoinbaseV3) GetProduct(ctx context.Context, productId string) (*Product, error) {
	productRequest := products.GetProductRequest{
		ProductId: productId,
	}
	product, err := c.products.GetProduct(ctx, &productRequest)
	if err != nil {
		return nil, err
	}

	price, err := strconv.ParseFloat(product.BaseMinSize, 64)
	if err != nil {
		return nil, err
	}

	return &Product{
		QuoteCurrency: product.QuoteCurrencyId,
		BaseCurrency:  product.BaseCurrencyId,
		BaseMinSize:   price,
	}, nil
}

func (c *CoinbaseV3) Deposit(ctx context.Context, currency string, amount float64) (*time.Time, error) {
	account, err := c.accountFor(ctx, currency) //taking the first coins a marker, make sure to put your main coin first
	if err != nil {
		return nil, err
	}
	paymentMethodRequest := paymentmethods.ListPaymentMethodsRequest{}
	paymentMethods, err := c.payment.ListPaymentMethods(ctx, &paymentMethodRequest)

	if err != nil {
		return nil, err
	}

	var bankAccount *model.PaymentMethod = nil

	for i := range paymentMethods.PaymentMethods {
		if paymentMethods.PaymentMethods[i].Type == "ACH" {
			bankAccount = paymentMethods.PaymentMethods[i]
		}
	}

	if bankAccount == nil {
		return nil, errors.New("No ACH bank account found on this account")
	}

	depositResponse, err := c.client.Deposit(account.Id, exchange.DepositParams{
		Amount:          amount,
		Currency:        currency,
		PaymentMethodID: bankAccount.Id,
		Commit:          true,
	})

	if err != nil {
		return nil, err
	}

	payoutAt := depositResponse.Data.PayoutAt
	return &payoutAt, nil
}

func (c *CoinbaseV3) LastPurchaseTime(ctx context.Context, coin string, currency string, since time.Time) (*time.Time, error) {

	productIds := make([]string, 1)
	productIds[0] = c.GetTickerSymbol(coin, currency)
	orderList, err := c.orders.ListOrders(ctx, &orders.ListOrdersRequest{
		ProductIds:  productIds,
		StartDate:   since.Format("2006-01-02T15:04:05.999999999Z07:00"),
		OrderStatus: []string{"FILLED"},
	})

	if err != nil {
		return nil, err
	}

	if len(orderList.Orders) > 0 {
		layout := time.RFC3339 // This is the layout matching the string format

		t, err := time.Parse(layout, orderList.Orders[0].CreatedTime)
		if err != nil {
			fmt.Println("Error:", err)
			return nil, err
		}
		return &t, nil
	}

	return nil, nil
}

func (c *CoinbaseV3) GetFiatAccount(ctx context.Context, currency string) (*Account, error) {

	account, err := c.accountFor(ctx, currency)
	if err != nil {
		return nil, err
	}

	return &Account{Available: account.Available}, nil
}

func (c *CoinbaseV3) GetPendingTransfers(currency string) ([]PendingTransfer, error) {
	pendingTransfers := []PendingTransfer{}
	// // Dang, we don't have enough funds. Let's see if money is on the way.
	// var transfers []exchange.Transfer
	// cursor := c.client.ListAccountTransfers(account.Id)

	// for cursor.HasMore {
	// 	if err := cursor.NextPage(&transfers); err != nil {
	// 		return pendingTransfers, err
	// 	}

	// 	for _, t := range transfers {
	// 		unprocessed := (t.ProcessedAt.Time() == time.Time{})
	// 		notCanceled := (t.CanceledAt.Time() == time.Time{})
	// 		//if it's pending for more than 1 day consider it stuck
	// 		//coinbase sometimes have those issues which support is unable to resolve
	// 		stuck := t.CreatedAt.Time().Before(time.Now().AddDate(0, 0, -1))

	// 		// This transfer is stil pending, so count it.
	// 		if unprocessed && notCanceled && !stuck {
	// 			pendingTransfers = append(pendingTransfers, PendingTransfer{Amount: t.Amount})
	// 		}
	// 	}
	// }
	return pendingTransfers, nil
}

func (c *CoinbaseV3) accountFor(ctx context.Context, currencyCode string) (*account, error) {

	// cache accounts
	if a, found := c.accounts[currencyCode]; found {
		return a, nil
	}
	listAccountsRequest := accounts.ListAccountsRequest{
		Pagination: &model.PaginationParams{
			Cursor: "",
			Limit:  "100",
		},
	}
	accounts, err := c.accountsService.ListAccounts(ctx, &listAccountsRequest)
	if err != nil {
		return nil, err
	}

	for _, a := range accounts.Accounts {
		available, err := strconv.ParseFloat(a.AvailableBalance.Value, 64)
		if err != nil {
			return nil, err
		}

		hold, err := strconv.ParseFloat(a.Hold.Value, 64)
		if err != nil {
			return nil, err
		}

		if a.Currency == currencyCode {
			acct := &account{
				Id:        a.Uuid,
				Currency:  a.Currency,
				Available: available,
				Hold:      hold,
			}

			c.accounts[currencyCode] = acct
			return acct, nil
		}
	}

	return nil, fmt.Errorf("No %s wallet on this account", currencyCode)
}
