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
	STORAGE_CREDENTIAL           = "storage_credential"
)

type StorageCredential struct {
	Bucket      string `json:"bucket"`
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
	EndPointApi string `json:"end_point_api"`
	EndPointCdn string `json:"end_point_cdn"`
	IsSSL       bool   `json:"is_ssl"`
}
