package captcha

import (
	"github.com/google/wire"
)

// ProviderSet is captcha providers
var ProviderSet = wire.NewSet(
	NewCaptchaStoreFromConfig,
	NewCaptchaServiceFromConfig,
) 