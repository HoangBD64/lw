package payment

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

const AirwallexAPIBaseSandBox = "https://api-demo.airwallex.com/api/v1"
const AirwallexAPIBaseLive = "https://api.airwallex.com/api/v1"

type InputAirwallexCreateOrder struct {
	MerchantOrderID string  `json:"merchant_order_id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	RequestID       string  `json:"request_id"`
}

type AirWallexInfo struct {
	ClientID string `json:"client_id"`
	APIKey   string `json:"api_key"`
	Domain   string `json:"domain"`
	Token    string `json:"token"`
}

type ReponseTokenAirWallex struct {
	// ExpiresAt *time.Time `json:"expires_at"`
	Token string `json:"token"`
}

func AirwallexGenerateAccessToken(info AirWallexInfo) (ReponseTokenAirWallex, error) {
	client := resty.New()
	response := ReponseTokenAirWallex{}
	// client.SetDebug(true)
	resp, err := client.R().SetResult(&response).SetHeader("Accept", "application/json").SetHeader("X-client-id", info.ClientID).SetHeader("X-api-key", info.APIKey).Post(info.Domain + "/authentication/login")
	if err != nil {
		return response, err
	}

	if resp.StatusCode() != http.StatusCreated {
		bodyBytes := resp.Body()
		return response, fmt.Errorf("error AirwallexGenerateAccessToken: %s", string(bodyBytes))
	}
	err2 := mapstructure.Decode(resp.Result(), &response)
	if err2 != nil {
		logrus.Debug("request GenerateAccessToken failed")
	}
	return response, nil
}

type ReponseAirWallexOrderCreate struct {
	ID                          string   `json:"id"`
	RequestID                   string   `json:"request_id"`
	Amount                      float64  `json:"amount"`
	Currency                    string   `json:"currency"`
	MerchantOrderID             string   `json:"merchant_order_id"`
	Descriptor                  string   `json:"descriptor"`
	Status                      string   `json:"status"`
	CapturedAmount              float64  `json:"captured_amount"`
	CreatedAt                   string   `json:"created_at"`
	UpdatedAt                   string   `json:"updated_at"`
	AvailablePaymentMethodTypes []string `json:"available_payment_method_types"`
	ClientSecret                string   `json:"client_secret"`
	BaseAmount                  float64  `json:"base_amount"`
	BaseCurrency                string   `json:"base_currency"`
}

func AirwallexCreateOrder(info AirWallexInfo, input InputAirwallexCreateOrder) (ReponseAirWallexOrderCreate, error) {
	client := resty.New()
	response := ReponseAirWallexOrderCreate{}
	// client.SetDebug(true)
	resp, err := client.R().SetResult(&response).SetHeader("Accept", "application/json").SetHeader("Authorization", "Bearer "+info.Token).SetBody(input).Post(info.Domain + "/pa/payment_intents/create")
	if err != nil {
		return response, err
	}

	if resp.StatusCode() != http.StatusCreated {
		bodyBytes := resp.Body()
		return response, fmt.Errorf("error AirwallexGenerateAccessToken: %s", string(bodyBytes))
	}
	err2 := mapstructure.Decode(resp.Result(), &response)
	if err2 != nil {
		logrus.Debug("request GenerateAccessToken failed")
	}
	return response, nil
}

func AirwallexGetOrderDetail(info AirWallexInfo, id string) (ReponseAirWallexOrderCreate, error) {
	client := resty.New()
	response := ReponseAirWallexOrderCreate{}
	// client.SetDebug(true)
	resp, err := client.R().SetResult(&response).SetHeader("Accept", "application/json").SetHeader("Authorization", "Bearer "+info.Token).Get(info.Domain + "/pa/payment_intents/" + id)
	if err != nil {
		return response, err
	}

	if resp.StatusCode() != http.StatusOK {
		bodyBytes := resp.Body()
		return response, fmt.Errorf("error AirwallexGetOrderDetail: %s", string(bodyBytes))
	}
	err2 := mapstructure.Decode(resp.Result(), &response)
	if err2 != nil {
		logrus.Debug("request get payment detail failed")
	}
	return response, nil
}
