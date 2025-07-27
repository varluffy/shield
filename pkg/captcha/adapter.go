package captcha

import (
	"time"

	"github.com/mojocn/base64Captcha/store"
	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/pkg/redis"
	"go.uber.org/zap"
)

// NewCaptchaStoreFromConfig 从配置创建验证码存储
func NewCaptchaStoreFromConfig(cfg *config.Config, redisClient *redis.Client, logger *zap.Logger) store.Store {
	// 如果Redis配置为空或者没有设置地址，使用内存存储
	if cfg.Redis == nil || len(cfg.Redis.Addrs) == 0 {
		logger.Info("Using memory store for captcha (Redis not configured)")
		return store.NewMemoryStore(1000, 5*time.Minute)
	}

	// 设置默认值
	keyPrefix := "ultrafit:captcha:"
	if cfg.Redis.KeyPrefix != "" {
		keyPrefix = cfg.Redis.KeyPrefix
	}

	expiration := 5 * time.Minute
	if cfg.Captcha != nil && cfg.Captcha.Expiration > 0 {
		expiration = cfg.Captcha.Expiration
	}

	logger.Info("Using Redis store for captcha")
	return NewRedisCaptchaStore(redisClient, keyPrefix, expiration, logger)
}

// NewCaptchaServiceFromConfig 从配置创建验证码服务
func NewCaptchaServiceFromConfig(store store.Store, cfg *config.Config, logger *zap.Logger) CaptchaService {
	// 设置默认验证码配置
	captchaConfig := &CaptchaConfig{
		Type:       "digit",
		Width:      160,
		Height:     60,
		Length:     4,
		NoiseCount: 5,
	}

	// 如果有配置则使用配置值
	if cfg.Captcha != nil {
		captchaConfig.Type = cfg.Captcha.Type
		captchaConfig.Width = cfg.Captcha.Width
		captchaConfig.Height = cfg.Captcha.Height
		captchaConfig.Length = cfg.Captcha.Length
		captchaConfig.NoiseCount = cfg.Captcha.NoiseCount
	}

	return NewCaptchaService(store, captchaConfig, logger)
} 