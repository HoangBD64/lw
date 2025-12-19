package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Transaction struct {
	Model           `gorm:"embedded"`
	Type            string         `json:"type" gorm:"index"`
	PaymentType     string         `json:"payment_type"`
	Amount          float64        `json:"amount"`         // total amount procedure need to pay
	AmountDeposit   float64        `json:"amount_deposit"` // amount user deposit
	OrderPaymentID  string         `json:"order_payment_id"`
	Status          string         `json:"status,omitempty" gorm:"index"`
	UserID          *uuid.UUID     `json:"user_id,omitempty"`
	User            *User          `json:"user,omitempty"`
	AdminID         *uuid.UUID     `json:"admin_id,omitempty"`
	Admin           *Admin         `json:"admin,omitempty"`
	Metadata        datatypes.JSON `json:"metadata"`
	Note            string         `json:"note"`
	UserProcedureID *int           `json:"user_procedure_id,omitempty"`
	UserProcedure   *UserProcedure `json:"user_procedure,omitempty"`
}

const (
	USER_DEPOSIT = "user_deposit"
)

const ( //status
	COMPLETED = "completed"
)

type PaymentHistory struct {
	Model    `gorm:"embedded"`
	OrderID  string         `json:"order_id"`
	Status   string         `json:"status,omitempty"`
	Metadata datatypes.JSON `json:"metadata"`
}

type HistoryJson struct {
	Status string         `json:"status"`
	Data   datatypes.JSON `json:"data"`
}
