package payment

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/plutov/paypal"
	"github.com/sirupsen/logrus"
)

type PurchaseUnitAmount struct {
	ReferenceID string `json:"reference_id"`
	Value       string `json:"value"`
	Currency    string `json:"currency"`
}

func PaypalCreateOrder(p *PaypalInfo, data PurchaseUnitAmount) (*paypal.Order, error) {
	// sanbox only
	c, err := paypal.NewClient(p.ClientID, p.ClientSecret, p.Domain)
	logrus.Debug(err)
	accessToken, err := c.GetAccessToken()
	logrus.Debug(accessToken)
	logrus.Debug(err)
	ApplicationContext := &paypal.ApplicationContext{
		BrandName: "legaswift.com",
		Locale:    "es-MX",
	}
	order, err := c.CreateOrder(paypal.OrderIntentCapture,
		[]paypal.PurchaseUnitRequest{
			{
				ReferenceID: data.ReferenceID,
				Amount:      &paypal.PurchaseUnitAmount{Value: data.Value, Currency: data.Currency},
			},
		}, &paypal.CreateOrderPayer{}, ApplicationContext,
	)
	if err != nil {
		logrus.Debug(err)
		return nil, err
	}
	return order, nil
}

func PaypalCapturePayment(p *PaypalInfo, order_id string) (paypal.CaptureOrderResponse, error) {
	c, err := paypal.NewClient(p.ClientID, p.ClientSecret, p.Domain)
	logrus.Debug(err)
	accessToken, err := c.GetAccessToken()
	logrus.Debug(accessToken, err)
	capture, err := c.CaptureOrder(order_id, paypal.CaptureOrderRequest{})
	if err != nil {
		return paypal.CaptureOrderResponse{}, err
	}
	return *capture, nil
}

type PaypalInfo struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Domain       string `json:"domain"`
}
type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency_code"`
}
type RefundCapturedPaymentInput struct {
	Amount      Amount `json:"amount"`
	InvoiceID   string `json:"invoice_id"`
	NoteToPayer string `json:"note_to_payer"`
}

type RefundCapturedPaymentResponse struct {
	ID     string `json:"id"`
	Amount struct {
		CurrencyCode string `json:"currency_code"`
		Value        string `json:"value"`
	} `json:"amount"`
	NoteToPayer string `json:"note_to_payer"`
	InvoiceID   string `json:"invoice_id"`
	Status      string `json:"status"`
	CreateTime  string `json:"create_time"`
	UpdateTime  string `json:"update_time"`
	Links       []struct {
		Href   string `json:"href"`
		Rel    string `json:"rel"`
		Method string `json:"method"`
	} `json:"links"`
}

func RefundCapturedPayment(p *PaypalInfo, refund RefundCapturedPaymentInput, transaction_id string) (RefundCapturedPaymentResponse, error) {
	client := resty.New()
	encoded := base64.StdEncoding.EncodeToString([]byte(p.ClientID + ":" + p.ClientSecret))
	client.SetHeader("Accept", "application/json").SetHeader("Authorization", "Basic "+encoded)
	response := RefundCapturedPaymentResponse{}
	// client.SetDebug(true)
	domain := "https://api-m.paypal.com"
	url := fmt.Sprintf(domain+"/v2/payments/captures/%s/refund", transaction_id)
	resp, err := client.R().SetResult(&response).SetHeader("Accept", "application/json").SetBody(&refund).Post(url)
	if err != nil {
		return response, err
	}
	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return response, fmt.Errorf("error creating refund: %s", resp.Body())
	}
	if err := mapstructure.Decode(resp.Result(), &response); err != nil {
		logrus.Debug(fmt.Errorf("error proccessing refund: %s", err))
	}
	return response, nil
}
