package auth

type RegisterRequest struct {
	Name        string `json:"name" binding:"required,min=3"`
	Email       string `json:"email" binding:"required,email"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,e164"`
	Password    string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SocialLoginRequest struct {
	Provider string `json:"provider" binding:"required,oneof=google facebook apple"`
	Token    string `json:"token"`
	IDToken  string `json:"id_token"`
}

type RequestOTPRequest struct {
	Channel string `json:"channel" binding:"required,oneof=whatsapp email"`
	Target  string `json:"target" binding:"required"`
}

type VerifyOTPRequest struct {
	Channel       string `json:"channel" binding:"required,oneof=whatsapp email"`
	Target        string `json:"target" binding:"required"`
	OTP           string `json:"otp" binding:"required,len=6,numeric"`
	DeviceName    string `json:"device_name"`
	TrustedDevice bool   `json:"trusted_device"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
