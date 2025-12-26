package handler

import (
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/pkg/payment"
	"core/utils"
	"core/utils/base"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/plutov/paypal"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

// FOR tessting
// var CLIENT_ID = "AYr4vHe7MM1OaDzHsd3aVsNoTDYJxnKkK6rBdApEYIXiPC3Yxoam0s7c1U5W2yMYTe1L5OcPns3jmWCR"
// var CLIENT_SECRET = "EPsrhZ-GzkBvki01P1PsEgL0p0BkNcx_ncpL0LPmXdv-nRY6-HIps_3jcLs3-ZejMxRNAfzl12409UlX"

type PaymentInfo struct {
	ClientID     string `json:"client_id"`
	APIKey       string `json:"api_key"`
	ClientSecret string `json:"client_secret"`
	Domain       string `json:"domain"`
	APIBase      string `json:"api_base"`
}

type js = map[string]any

// for logic payment in checkout page
func GetPaymentInfo(c *fiber.Ctx) error {
	// _, ok := c.Locals("user").(model.User)
	// if !ok {
	// 	return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	// }

	var query repo.Query
	query.Parse(c)

	setting, err := repo.GetSettingSystemByType(model.PAYMENT_INFO)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get GetSettingByType Failed", Error: err})
	}
	data := []model.PaymentInfoPublic{}
	if err := json.Unmarshal(setting.Metadata, &data); err != nil {
		slog.Debug("Error marshaling", "err", err)
	}
	payment_info := []model.PaymentInfoPublic{}
	lo.ForEach(data, func(payment model.PaymentInfoPublic, _ int) {
		if payment.IsActive {
			payment_info = append(payment_info, payment)
		}
	})
	jsonBytes, err := json.Marshal(payment_info)
	if err != nil {
		slog.Debug("Lỗi khi mã hóa JSON", "err", err)
	} else {
		setting.Metadata = jsonBytes
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: setting})

}

// for logic payment in checkout page
func InitPayment(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	type Input struct {
		Deposit     float64 `json:"deposit" validate:"required"`
		Currency    string  `json:"currency"`
		Type        string  `json:"type" validate:"required"`
		ProcedureID int     `json:"procedure_id" validate:"required"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body"})
	}
	if err := appUtils.Validate(c.Context(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Validation failed", Error: err})
	}

	check := slices.Contains([]string{
		model.PAYMENT_AIRWALLEX,
		model.PAYMENT_PAYPAL,
	}, input.Type)
	if !check {
		return c.JSON(base.BaseResponse{Status: false, Message: "payment type is invalid"})
	}

	if input.Deposit <= 0 {
		return c.JSON(base.BaseResponse{Status: false, Message: "Amount deposit must be greater than 0", ErrorCode: "DEPOSIT_INVALID"})
	}

	setting, err := repo.GetSettingSystemByType(model.PAYMENT_INFO)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get GetSettingByType Failed", Error: err})
	}
	info_payment := []model.PaymentInfo{}
	if err := json.Unmarshal(setting.Metadata, &info_payment); err != nil {
		slog.Debug("Error marshaling", "err", err)
		return c.JSON(base.BaseResponse{Status: false, Message: "Problem checking payment method", Error: err})
	}

	payment_method, ok := lo.Find(info_payment, func(info model.PaymentInfo) bool {
		return info.Type == input.Type && info.IsActive
	})
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "This payment method is not available yet.", ErrorCode: "PAYMENT_METHOD_DISABLE"})
	}

	up, err := repo.GetProcedureByID(input.ProcedureID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Procedure not found", Error: err})
	}

	switch payment_method.Type {
	case model.PAYMENT_PAYPAL:
		purchase_string := strconv.FormatFloat(input.Deposit, 'f', 2, 64)
		input_purchase := payment.PurchaseUnitAmount{
			Value:    purchase_string,
			Currency: "USD",
		}
		if payment_method.APIBase == "live" {
			payment_method.Domain = paypal.APIBaseLive
		} else {
			payment_method.Domain = paypal.APIBaseSandBox
		}
		net := calculatePrice(input.Deposit, payment_method.TransactionFee)
		required := lo.FromPtr(up.Cost)
		if net < required {
			return c.JSON(base.BaseResponse{Status: false, Message: "Insufficient amount for this UserProcedure", ErrorCode: "INSUFFICIENT_FUNDS"})
		}
		new_trans := model.Transaction{
			UserID:        &user.ID,
			Type:          model.USER_DEPOSIT,
			Amount:        required,
			AmountDeposit: input.Deposit,
			PaymentType:   model.PAYMENT_PAYPAL,
			Status:        "pending",
			ProcedureID:   &input.ProcedureID,
		}
		transaction, err := repo.CreateTransaction(c.Context(), new_trans)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "can't create transaction", Error: err})
		}
		input_purchase.ReferenceID = transaction.ID.String()
		system_pp_info := payment.PaypalInfo{ClientID: payment_method.ClientID, ClientSecret: payment_method.ClientSecret, Domain: payment_method.Domain}
		order, err := payment.PaypalCreateOrder(&system_pp_info, input_purchase)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "can't CreateOrder", Error: err})
		}

		go paypalUpdateTransaction(transaction, order)
		return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: order})
	case model.PAYMENT_AIRWALLEX:
		if payment_method.APIBase == "live" {
			payment_method.Domain = payment.AirwallexAPIBaseLive
		} else {
			payment_method.Domain = payment.AirwallexAPIBaseSandBox
		}
		airwallex_info := payment.AirWallexInfo{
			ClientID: payment_method.ClientID,
			APIKey:   payment_method.APIKey,
			Domain:   payment_method.Domain,
		}
		token, err := payment.AirwallexGenerateAccessToken(airwallex_info)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "can't Generate Access token", Error: err})
		}
		airwallex_info.Token = token.Token
		net := calculatePrice(input.Deposit, payment_method.TransactionFee)
		required := lo.FromPtr(up.Cost)
		if net < required {
			return c.JSON(base.BaseResponse{Status: false, Message: "Insufficient amount for this UserProcedure", ErrorCode: "INSUFFICIENT_FUNDS"})
		}
		new_trans := model.Transaction{
			UserID:        &user.ID,
			Type:          model.USER_DEPOSIT,
			PaymentType:   model.PAYMENT_AIRWALLEX,
			Amount:        required,
			AmountDeposit: input.Deposit,
			Status:        "pending",
			ProcedureID:   &input.ProcedureID,
		}
		transaction, err := repo.CreateTransaction(c.Context(), new_trans)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "can't create transaction", Error: err})
		}
		data, err := payment.AirwallexCreateOrder(airwallex_info, payment.InputAirwallexCreateOrder{
			MerchantOrderID: fmt.Sprint(transaction.ID),
			Amount:          input.Deposit,
			Currency:        "USD",
			RequestID:       "LegaSwift" + uuid.NewString(),
		})
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "can't create order airwallex", Error: err})
		}

		go airwallexUpdateTransaction(transaction, data)
		return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: data})
	default:
		return c.JSON(base.BaseResponse{Status: false, Message: "No valid payment method found", ErrorCode: "MISSING_PAYMENT"})
	}
}

type FinishBody struct {
	OrderID     string `json:"order_id"`
	State       string `json:"state"`
	CompanyName string `json:"company_name"`
	CompanyType string `json:"company_type"`
}

func FinishPayment(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}

	var fnBody FinishBody
	if err := c.BodyParser(&fnBody); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Invalid request body", Error: err})
	}
	fnBody.State = strings.ToLower(fnBody.State)

	transaction, err := repo.GetTransactionByOrderID(fnBody.OrderID, user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Order not found", Error: err})
	}

	setting, err := repo.GetSettingSystemByType(model.PAYMENT_INFO)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get GetSettingByType Failed", Error: err})
	}

	info_payment := []model.PaymentInfo{}
	if err := json.Unmarshal(setting.Metadata, &info_payment); err != nil {
		slog.Debug("Error marshaling", "err", err)
		return c.JSON(base.BaseResponse{Status: false, Message: "Problem checking payment method", Error: err})
	}
	payment_method, ok := lo.Find(info_payment, func(info model.PaymentInfo) bool {
		return info.Type == transaction.PaymentType
	})
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "This payment method is not available yet.", ErrorCode: "PAYMENT_METHOD_DISABLE"})
	}

	_, err = repo.GetProcedureByID(lo.FromPtr(transaction.ProcedureID))
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Procedure not found", Error: err})
	}

	var up model.UserProcedure
	switch transaction.PaymentType {
	case model.PAYMENT_PAYPAL:
		payment_info := PaymentInfo{
			ClientID:     payment_method.ClientID,
			ClientSecret: payment_method.ClientSecret,
		}
		if payment_method.APIBase == "live" {
			payment_info.Domain = paypal.APIBaseLive
		} else {
			payment_info.Domain = paypal.APIBaseSandBox
		}
		system_pp_info := payment.PaypalInfo{ClientID: payment_info.ClientID, ClientSecret: payment_info.ClientSecret, Domain: payment_info.Domain}

		capture, err := payment.PaypalCapturePayment(&system_pp_info, fnBody.OrderID)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Error CapturePayment", Error: err})
		}

		up, err = utils.Retry(3, 2*time.Second, func() (model.UserProcedure, error) {
			return UserCreateUserProcedureAfterPayment(transaction, fnBody)
		})
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Error creating user procedure", Error: err})
		}

		transaction.UserProcedureID = &up.ID
		err = paypalUpdateBalanceOrg(transaction, capture)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Error UpdateBalanceOrg", Error: err})
		}
	case model.PAYMENT_AIRWALLEX:
		if payment_method.APIBase == "live" {
			payment_method.Domain = payment.AirwallexAPIBaseLive
		} else {
			payment_method.Domain = payment.AirwallexAPIBaseSandBox
		}
		airwallex_info := payment.AirWallexInfo{
			ClientID: payment_method.ClientID,
			APIKey:   payment_method.APIKey,
			Domain:   payment_method.Domain,
		}
		token, err := payment.AirwallexGenerateAccessToken(airwallex_info)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "can't Generate Access token", Error: err})
		}
		airwallex_info.Token = token.Token
		data, err := payment.AirwallexGetOrderDetail(airwallex_info, transaction.OrderPaymentID)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "can't AirwallexGetOrderDetail", Error: err})
		}
		if data.Status != "SUCCEEDED" {
			return c.JSON(base.BaseResponse{Status: false, Message: "Status Order payment not Success", ErrorCode: "ORDER_NOT_COMPLETED"})
		}

		up, err = utils.Retry(3, 2*time.Second, func() (model.UserProcedure, error) {
			return UserCreateUserProcedureAfterPayment(transaction, fnBody)
		})
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Error creating user procedure", Error: err})
		}
		transaction.UserProcedureID = &up.ID
		err = airwallexUpdateBalanceOrg(transaction, data)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "Error airwallexUpdateTransaction", Error: err})
		}
	default:
		return c.JSON(base.BaseResponse{Status: false, Message: "Unable to determine order payment type"})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: up.ID})
}

func UserCreateUserProcedureAfterPayment(transaction model.Transaction, body FinishBody) (model.UserProcedure, error) {
	procedure_template, err := repo.GetProcedureByID(lo.FromPtr(transaction.ProcedureID))
	if err != nil {
		slog.Debug("Error get ProcedureTemplateByID", "err", err)
		return model.UserProcedure{}, err
	}

	status := model.USER_PROCEDURE_PENDING
	title := "Untitled"
	isHaveTicket := false
	new_user_procedure := model.UserProcedure{
		UserID:       transaction.UserID,
		ProcedureID:  procedure_template.ID,
		TemplateName: procedure_template.Name,
		Cost:         procedure_template.Cost,
		DueDate:      procedure_template.DueDate,
		Status:       &status,
		Title:        &title,
		IsHaveTicket: &isHaveTicket,
		State:        &body.State,
		CompanyName:  &body.CompanyName,
		CompanyType:  &body.CompanyType,
	}
	up, err := repo.CreateUserProcedure(new_user_procedure)
	if err != nil {
		slog.Debug("Error creating UserProcedure", "err", err)
		return model.UserProcedure{}, err
	}

	ups := MappingProcedureStepsToUserProcedureSteps(procedure_template.ProcedureSteps, up.ID)
	err = repo.UserCreateBulkProcedureStep(&ups)
	if err != nil {
		slog.Debug("Error creating UserProcedureStepsBatch", "err", err)
		return model.UserProcedure{}, err
	}

	transaction.UserProcedureID = &up.ID
	_, err = repo.UpdateTransaction(transaction)
	if err != nil {
		slog.Debug("Error UpdateTransaction:", "err", err)
		return model.UserProcedure{}, err
	}
	return up, nil
}

func paypalUpdateTransaction(transaction model.Transaction, order_payment *paypal.Order) error {
	transaction.OrderPaymentID = order_payment.ID
	jsonData, err := json.Marshal(order_payment)
	if err != nil {
		slog.Debug("Error marshaling", "err", err)
		return err
	}
	new_his := model.HistoryJson{
		Status: order_payment.Status,
		Data:   jsonData,
	}
	list_history := []model.HistoryJson{new_his}
	json_history_json, err := json.Marshal(list_history)
	if err != nil {
		slog.Debug("Error marshaling", "err", err)
		return err
	}
	transaction.Metadata = json_history_json
	_, err = repo.UpdateTransaction(transaction)
	if err != nil {
		slog.Debug("Lỗi UpdateTransaction:", "err", err)
		return err
	}
	return nil
}
func airwallexUpdateTransaction(transaction model.Transaction, order_payment payment.ReponseAirWallexOrderCreate) error {
	transaction.OrderPaymentID = order_payment.ID
	jsonData, err := json.Marshal(order_payment)
	if err != nil {
		slog.Debug("Error marshaling", "err", err)
		return err
	}
	new_his := model.HistoryJson{
		Status: order_payment.Status,
		Data:   jsonData,
	}
	list_history := []model.HistoryJson{new_his}
	json_history_json, err := json.Marshal(list_history)
	if err != nil {
		slog.Debug("Error marshaling", "err", err)
		return err
	}
	transaction.Metadata = json_history_json
	_, err = repo.UpdateTransaction(transaction)
	if err != nil {
		slog.Debug("Lỗi UpdateTransaction:", "err", err)
		return err
	}
	return nil
}

func airwallexUpdateBalanceOrg(transaction model.Transaction, order_payment payment.ReponseAirWallexOrderCreate) error {
	jsonData, err := json.Marshal(order_payment)
	if err != nil {
		slog.Debug("Error marshaling", "err", err)
		return err
	}

	list_history := []model.HistoryJson{}
	err = json.Unmarshal(transaction.Metadata, &list_history)
	if err != nil {
		slog.Debug("Error marshaling", "err", err)
		return err
	}
	new_his := model.HistoryJson{
		Status: order_payment.Status,
		Data:   jsonData,
	}
	list_history = append(list_history, new_his)
	json_history_json, err := json.Marshal(list_history)
	if err != nil {
		slog.Debug("Error marshaling", "err", err)
		return err
	}
	transaction.Metadata = json_history_json
	transaction.Status = model.COMPLETED
	_, err = repo.UpdateTransaction(transaction)
	if err != nil {
		return err
	}

	if transaction.UserProcedureID != nil {
		go SendMailToUser([]int{*transaction.UserProcedureID}, model.NOTI_INVOICE_COMPLETED)
	}
	return nil
}
func paypalUpdateBalanceOrg(transaction model.Transaction, capture paypal.CaptureOrderResponse) error {
	var top_up float64 = 0
	for i := 0; i < len(capture.PurchaseUnits); i++ {
		if len(capture.PurchaseUnits[i].Payments.Captures) > 0 {
			for j := 0; j < len(capture.PurchaseUnits[i].Payments.Captures); j++ {
				floatValue, err := strconv.ParseFloat(capture.PurchaseUnits[i].Payments.Captures[j].Amount.Value, 64)
				if err != nil {
					return err
				}
				top_up += floatValue
			}
		}
	}
	if capture.Status == "COMPLETED" {
		transaction.Status = model.COMPLETED
	} else {
		transaction.Status = capture.Status
	}
	jsonData, err := json.Marshal(capture)
	if err != nil {
		slog.Debug("Error marshaling", "err", err)
		return err
	}

	list_history := []model.HistoryJson{}
	err = json.Unmarshal(transaction.Metadata, &list_history)
	if err != nil {
		slog.Debug("Error marshaling", "err", err)
		return err
	}
	new_his := model.HistoryJson{
		Status: capture.Status,
		Data:   jsonData,
	}
	list_history = append(list_history, new_his)
	json_history_json, err := json.Marshal(list_history)
	if err != nil {
		slog.Debug("Error marshaling", "err", err)
		return err
	}
	transaction.Metadata = json_history_json
	_, err = repo.UpdateTransaction(transaction)
	if err != nil {
		return err
	}

	if transaction.Status == model.COMPLETED && transaction.UserProcedureID != nil {
		go SendMailToUser([]int{*transaction.UserProcedureID}, model.NOTI_INVOICE_COMPLETED)
	}
	return nil
}

func GetTransactions(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	transactions, total, err := repo.GetTransaction(query, user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTransaction", Error: err})
	}
	for i := range transactions {
		transactions[i].AdminID = nil
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: transactions, Total: &total})
}

func GetUserBalance(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	balance, err := repo.GetTransactionBalance(c.Context(), user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTransactionBalance", Error: err})
	}
	balance = math.Round(balance*1000) / 1000
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: balance})
}

func GetTransactionDetails(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	var query repo.Query
	query.Parse(c)
	transactions, err := repo.GetTransactionDetails(*query.ID, user)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetTransaction", Error: err})
	}

	transactions.AdminID = nil
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: transactions})
}

func AdminGetTransaction(c *fiber.Ctx) error {
	_, ok := c.Locals("admin").(model.Admin)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}
	var query repo.Query
	query.Parse(c)

	transactions, total, err := repo.AdminGetTransaction(query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Can't AdminGetTransaction", Error: err})
	}
	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: transactions, Total: &total})
}

func GetBillingAnalytics(c *fiber.Ctx) error {
	_, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(base.BaseResponse{Status: false, Message: "Please login again"})
	}

	var query repo.Query
	query.Parse(c)

	if query.StartTime == nil || query.EndTime == nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "must specify time range"})
	}

	totalFulfillmentPaid, err := repo.GetAnalyticsBilling(c.Context(), query, model.USER_DEPOSIT)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "failed to fetch data", Error: err})
	}

	totalDeposit, err := repo.GetAnalyticsDeposit(c.Context(), query)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "failed to fetch data", Error: err})
	}

	response := js{
		"fulfillment_cost_paid": math.Abs(totalFulfillmentPaid),
		"deposits":              totalDeposit,
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "Success", Data: response})
}

func ExportTransactions(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "error": "Please login again"})
	}

	var query repo.Query
	query.Parse(c)
	query.UserID = &user.ID
	transactions, err := repo.GetTransactionForExports(c.Context(), query)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Can't GetTransaction", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": true, "message": "Success", "data": transactions})
}

func mathCeilPrice(number float64) float64 {
	d := decimal.NewFromFloat(number)
	rounded := d.Mul(decimal.NewFromInt(100)).Ceil().Div(decimal.NewFromInt(100))
	result, _ := rounded.Float64()
	return result
}

func mathFloorPrice(number float64) float64 {
	d := decimal.NewFromFloat(number)
	rounded := d.Truncate(2)
	result, _ := rounded.Float64()
	return result
}

func calculatePrice(price float64, percent float64) float64 {
	if percent <= 0 {
		return price
	}
	if percent >= 100 {
		return 0
	}
	price = mathFloorPrice(price - mathCeilPrice((price*percent)/100))
	return price
}
