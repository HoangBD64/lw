package domain

type VerifyAdminSSORequest struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state"`
}
