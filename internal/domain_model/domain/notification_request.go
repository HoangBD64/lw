package domain

type UserUpsertNotificationPreferenceRequest struct {
	EmailEnabled bool `json:"email_enabled"`
}
