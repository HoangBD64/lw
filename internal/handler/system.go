package handler

import (
	"context"
	"core/internal/domain_model/domain"
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/utils/base"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PaymentInfoSetting struct {
	Type           string         `json:"type"`
	APIBase        string         `json:"api_base,omitempty"`
	ClientID       string         `json:"client_id,omitempty"`
	APIKey         string         `json:"api_key,omitempty"`
	ClientSecret   string         `json:"client_secret,omitempty"`
	IsActive       bool           `json:"is_active"`
	Metadata       datatypes.JSON `json:"metadata,omitempty"`
	TransactionFee float64        `json:"transaction_fee"`
}

func GetSystemSetting(c *fiber.Ctx) error {
	var query repo.Query
	query.Parse(c)

	admin, err := checkCurrentSystemAdmin(c)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "unauthorized", Error: err})
	}

	if err := isSystemSettingRoleAllowed(query.Type, admin); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "an error occurred", Error: err})
	}

	setting, err := repo.GetSystemSetting(context.TODO(), query.Type)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(base.BaseResponse{Status: false, Message: "failed to get data", Error: err})
		}

		newSetting := model.SystemSetting{Type: query.Type}

		createdSetting, err := repo.CreateSystemSetting(newSetting)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "failed to create default setting", Error: err})
		}

		setting = createdSetting
	}

	if setting.Type == model.STORAGE_CREDENTIAL {
		credential := model.StorageCredential{}
		if err := json.Unmarshal(setting.Metadata, &credential); err != nil {
			slog.ErrorContext(context.TODO(), "failed to unmarshal storage credential data", "error", err)
		}

		// mask secret key
		credential.SecretKey = ""

		credentialJson, err := json.Marshal(credential)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "could not process fetched data", Error: err})
		}

		setting.Metadata = credentialJson
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "ok", Data: setting})
}

func AdminUpdateSetting(c *fiber.Ctx) error {
	admin, err := checkCurrentSystemAdmin(c)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "unauthorized", Error: err})
	}

	var input domain.AdminUpdateSettingRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "bad request", Error: err})
	}

	if err := appUtils.Validate(context.TODO(), &input); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "bad request", Error: err})
	}

	setting_found, err := repo.GetSystemSettingByID(input.ID)
	if err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "Get GetSystemSettingByID Failed", Error: err})
	}

	if err := isSystemSettingRoleAllowed(setting_found.Type, admin); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "an error occurred", Error: err})
	}

	if setting_found.Type == model.PAYMENT_INFO {
		old_data := []PaymentInfoSetting{}
		if err := json.Unmarshal(setting_found.Metadata, &old_data); err != nil {
			slog.Debug("Lỗi khi chuyển đổi cấu trúc thành JSON:", "error", err)
		}
		new_data := []PaymentInfoSetting{}
		if err := json.Unmarshal(input.Metadata, &new_data); err != nil {
			slog.Debug("Lỗi khi chuyển đổi cấu trúc thành JSON:", "error", err)
		}
		old_data_update := []PaymentInfoSetting{}
		for _, new_setting := range new_data {
			for _, payment_info := range old_data {
				if new_setting.Type == payment_info.Type && (new_setting.Type == model.PAYMENT_PAYPAL || new_setting.Type == model.PAYMENT_AIRWALLEX) {
					if new_setting.APIKey == "" {
						new_setting.APIKey = payment_info.APIKey
					}
					if new_setting.ClientSecret == "" {
						new_setting.ClientSecret = payment_info.ClientSecret
					}
				}
			}

			old_data_update = append(old_data_update, new_setting)
		}
		setting_new, err := json.Marshal(old_data_update)
		if err != nil {
			return c.JSON(fiber.Map{"status": false, "message": "json.Marshal(old_data) Error", "error": err.Error()})
		}

		input.Metadata = setting_new
	}

	if setting_found.Type == model.STORAGE_CREDENTIAL {
		newData := model.StorageCredential{}
		if err := json.Unmarshal(input.Metadata, &newData); err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "could not process requested storage config", Error: err})
		}

		oldData := model.StorageCredential{}
		if err := json.Unmarshal(setting_found.Metadata, &oldData); err != nil {
			slog.ErrorContext(context.TODO(), "failed to unmarshal existing storage credential data", "error", err)
		}

		// keep old access key and secret key if new ones are empty
		if newData.SecretKey == "" {
			newData.SecretKey = oldData.SecretKey
		}

		newSetting, err := json.Marshal(newData)
		if err != nil {
			return c.JSON(base.BaseResponse{Status: false, Message: "could not process requested data", Error: err})
		}

		input.Metadata = newSetting
	}
	setting_found.Resource = input.Resource
	setting_found.Metadata = input.Metadata
	setting_found.Contant = input.Contant

	if _, err := repo.UpdateSystemSetting(setting_found); err != nil {
		return c.JSON(base.BaseResponse{Status: false, Message: "failed to update data", Error: err})
	}

	return c.JSON(base.BaseResponse{Status: true, Message: "ok"})
}

func checkCurrentSystemAdmin(c *fiber.Ctx) (model.SystemAdmin, error) {
	if admin, ok := c.Locals("admin").(model.Admin); ok {
		return admin, nil
	}

	if superAdmin, ok := c.Locals("super_admin").(model.SuperAdmin); ok {
		return superAdmin, nil
	}

	return nil, errors.New("unauthorized access")
}

func isSystemSettingRoleAllowed(typeName string, sysAdmin model.SystemAdmin) error {
	dualAccess := "both"

	src := map[string]string{
		// NOTE: TEMPORARY
		// model.STORAGE_CREDENTIAL: model.SUPER_ADMIN,
		model.STORAGE_CREDENTIAL: model.RoleAdmin,
		model.PAYMENT_INFO:       model.RoleAdmin,
	}

	perm, exists := src[typeName]
	if !exists {
		return fmt.Errorf("setting type not recognized")
	}

	// allow both admin and super admin
	if perm == dualAccess {
		return nil
	}

	if perm != sysAdmin.AdminType() {
		return fmt.Errorf("insufficient permissions to access this setting")
	}

	return nil
}
