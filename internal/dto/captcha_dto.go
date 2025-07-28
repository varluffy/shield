package dto

// CaptchaGenerateResponse 验证码生成响应
type CaptchaGenerateResponse struct {
	CaptchaID    string `json:"captcha_id" example:"abc123"`
	CaptchaImage string `json:"captcha_image" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."`
}

// CaptchaVerifyRequest 验证码验证请求
type CaptchaVerifyRequest struct {
	CaptchaID string `json:"captcha_id" binding:"required" example:"abc123"`
	Answer    string `json:"answer" binding:"required" example:"1234"`
}

// LoginWithCaptchaRequest 带验证码的登录请求
type LoginWithCaptchaRequest struct {
	Email     string `json:"email" binding:"required,email" example:"admin@example.com"`
	Password  string `json:"password" binding:"required" example:"123456"`
	CaptchaID string `json:"captcha_id" binding:"required" example:"abc123"`
	Answer    string `json:"answer" binding:"required" example:"1234"`
}
