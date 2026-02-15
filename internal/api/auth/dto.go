package auth

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type LoginRequest struct {
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.TenantID, validation.Required, is.UUID),
		validation.Field(&r.Email, validation.Required, is.Email),
		validation.Field(&r.Password, validation.Required, validation.Length(6, 72)),
	)
}

type LoginResponse struct {
	AccessToken string `json:"access_token,omitempty"`
	MFAToken    string `json:"mfa_token,omitempty"`
	Email       string `json:"email"`
	Status      string `json:"status"` // "success", "mfa_required"
}

type MFASetupResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"` // Provisioning URI
}

type MFAActivateRequest struct {
	Code string `json:"code"`
}

func (r MFAActivateRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Code, validation.Required, validation.Length(6, 6)),
	)
}

type MFAVerifyRequest struct {
	MFAToken string `json:"mfa_token"`
	Code     string `json:"code"`
}

func (r MFAVerifyRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.MFAToken, validation.Required),
		validation.Field(&r.Code, validation.Required, validation.Length(6, 6)),
	)
}
