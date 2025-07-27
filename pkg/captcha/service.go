package captcha

import (
	"context"
	"fmt"

	"github.com/mojocn/base64Captcha"
	"github.com/mojocn/base64Captcha/store"
	"go.uber.org/zap"
)

// CaptchaService interface defines captcha service operations
type CaptchaService interface {
	GenerateCaptcha(ctx context.Context) (*CaptchaResponse, error)
	VerifyCaptcha(ctx context.Context, captchaID, answer string) error
}

// CaptchaResponse represents the captcha generation response
type CaptchaResponse struct {
	CaptchaID    string `json:"captcha_id"`
	CaptchaImage string `json:"captcha_image"`
}

// captchaService implements CaptchaService interface
type captchaService struct {
	store  store.Store
	config *CaptchaConfig
	logger *zap.Logger
}

// NewCaptchaService creates a new captcha service
func NewCaptchaService(store store.Store, config *CaptchaConfig, logger *zap.Logger) CaptchaService {
	// 设置自定义存储
	base64Captcha.SetCustomStore(store)
	
	return &captchaService{
		store:  store,
		config: config,
		logger: logger,
	}
}

// GenerateCaptcha generates a new captcha
func (s *captchaService) GenerateCaptcha(ctx context.Context) (*CaptchaResponse, error) {
	var configuration interface{}
	
	// 根据配置创建不同类型的验证码配置
	switch s.config.Type {
	case "digit":
		configuration = base64Captcha.ConfigDigit{
			Height:     s.config.Height,
			Width:      s.config.Width,
			MaxSkew:    0.7,
			DotCount:   s.config.NoiseCount,
			CaptchaLen: s.config.Length,
		}
	case "string":
		configuration = base64Captcha.ConfigCharacter{
			Height:             s.config.Height,
			Width:              s.config.Width,
			IsUseSimpleFont:    true,
			IsShowHollowLine:   false,
			IsShowNoiseDot:     false,
			IsShowNoiseText:    false,
			IsShowSlimeLine:    false,
			IsShowSineLine:     false,
			CaptchaLen:         s.config.Length,
			ComplexOfNoiseText: base64Captcha.CaptchaComplexLower,
			ComplexOfNoiseDot:  base64Captcha.CaptchaComplexLower,
		}
	case "math":
		// 数学验证码使用数字验证码的配置
		configuration = base64Captcha.ConfigDigit{
			Height:     s.config.Height,
			Width:      s.config.Width,
			MaxSkew:    0.7,
			DotCount:   s.config.NoiseCount,
			CaptchaLen: s.config.Length,
		}
	default:
		// 默认使用数字验证码
		configuration = base64Captcha.ConfigDigit{
			Height:     s.config.Height,
			Width:      s.config.Width,
			MaxSkew:    0.7,
			DotCount:   s.config.NoiseCount,
			CaptchaLen: s.config.Length,
		}
	}
	
	// 生成验证码
	id, captchaInstance := base64Captcha.GenerateCaptcha("", configuration)
	if captchaInstance == nil {
		s.logger.Error("Failed to generate captcha instance")
		return nil, fmt.Errorf("failed to generate captcha")
	}
	
	// 转换为base64字符串
	b64s := base64Captcha.CaptchaWriteToBase64Encoding(captchaInstance)
	
	s.logger.Debug("Captcha generated successfully", 
		zap.String("id", id))
	
	return &CaptchaResponse{
		CaptchaID:    id,
		CaptchaImage: b64s,
	}, nil
}

// VerifyCaptcha verifies the captcha answer
func (s *captchaService) VerifyCaptcha(ctx context.Context, captchaID, answer string) error {
	if captchaID == "" {
		s.logger.Warn("Empty captcha ID provided")
		return fmt.Errorf("captcha ID is required")
	}
	
	if answer == "" {
		s.logger.Warn("Empty captcha answer provided", 
			zap.String("captcha_id", captchaID))
		return fmt.Errorf("captcha answer is required")
	}
	

	// 验证验证码（验证后自动清除）
	isValid := base64Captcha.VerifyCaptcha(captchaID, answer)
	
	if !isValid {
		s.logger.Warn("Invalid captcha answer", 
			zap.String("captcha_id", captchaID))
		return fmt.Errorf("invalid captcha")
	}
	
	s.logger.Debug("Captcha verified successfully", 
		zap.String("captcha_id", captchaID))
	
	return nil
}

// CaptchaConfig represents captcha configuration
type CaptchaConfig struct {
	Type       string `json:"type"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Length     int    `json:"length"`
	NoiseCount int    `json:"noise_count"`
} 