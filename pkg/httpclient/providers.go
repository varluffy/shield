// Package httpclient provides Wire providers for HTTP client components.
package httpclient

import (
	"github.com/google/wire"
	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/pkg/logger"
)

// ProviderSet HTTP客户端相关的Wire Provider集合
var ProviderSet = wire.NewSet(
	ProvideHTTPClient,
)

// ProvideHTTPClient 提供HTTP客户端实例
func ProvideHTTPClient(cfg *config.Config, logger *logger.Logger) HTTPClient {
	if cfg.HTTPClient == nil {
		// 如果没有配置HTTPClient，使用默认配置
		defaultConfig := &config.HTTPClientConfig{
			Timeout:          30,
			RetryCount:       3,
			RetryWaitTime:    1,
			RetryMaxWaitTime: 10,
			EnableTrace:      true,
			EnableLog:        true,
			MaxLogBodySize:   10240,
			UserAgent:        "UltraFit-HTTP-Client/1.0",
		}
		return NewHTTPClient(defaultConfig, logger)
	}
	return NewHTTPClient(cfg.HTTPClient, logger)
}
