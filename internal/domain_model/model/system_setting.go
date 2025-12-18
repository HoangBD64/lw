package model

import (
	"gorm.io/datatypes"
)

type SystemSetting struct {
	Model    `gorm:"embedded"`
	Type     string         `json:"type,omitempty"  gorm:"index"`
	Resource string         `json:"resource,omitempty"`
	Contant  float64        `json:"contant,omitempty"`
	Metadata datatypes.JSON `json:"metadata,omitempty"`
}

const ( // type
	STORAGE_CREDENTIAL = "storage_credential"
	PAYMENT_INFO       = "payment"
)

const ( //payment type
	PAYMENT_PAYPAL    = "paypal"
	PAYMENT_AIRWALLEX = "airwallex"
	PAYMENT_CUSTOM    = "custom"
)

type StorageCredential struct {
	Bucket      string `json:"bucket"`
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
	EndPointApi string `json:"end_point_api"`
	EndPointCdn string `json:"end_point_cdn"`
	IsSSL       bool   `json:"is_ssl"`
}

type PaymentInfo struct {
	TransactionFee float64        `json:"transaction_fee"`
	Type           string         `json:"type"`
	ClientID       string         `json:"client_id"`
	APIKey         string         `json:"api_key"`
	ClientSecret   string         `json:"client_secret"`
	Domain         string         `json:"domain"`
	APIBase        string         `json:"api_base"`
	IsActive       bool           `json:"is_active"`
	Metadata       datatypes.JSON `json:"metadata,omitempty"`
}
type PaymentInfoPublic struct {
	TransactionFee float64        `json:"transaction_fee"`
	Type           string         `json:"type"`
	ClientID       string         `json:"client_id,omitempty"`
	APIBase        string         `json:"api_base,omitempty"`
	IsActive       bool           `json:"is_active"`
	Metadata       datatypes.JSON `json:"metadata,omitempty"`
}
