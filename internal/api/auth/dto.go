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
	AccessToken string `json:"access_token"` // Future milestone (SC-023)
	Email       string `json:"email"`
	Status      string `json:"status"`
}
