// Package auth provides Wire dependency injection providers for authentication services.
package auth

import (
	"github.com/google/wire"
	"github.com/varluffy/shield/internal/config"
)

// AuthProviderSet JWT服务的Wire ProviderSet
var AuthProviderSet = wire.NewSet(
	ProvideJWTService,
)

// ProvideJWTService 提供JWT服务实例
func ProvideJWTService(cfg *config.Config) JWTService {
	return NewJWTService(
		cfg.Auth.JWT.Secret,
		cfg.Auth.JWT.Issuer,
		cfg.Auth.JWT.ExpiresIn,
		cfg.Auth.JWT.RefreshExpires,
	)
} 