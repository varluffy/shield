package transaction

import (
	"github.com/google/wire"
)

// ProviderSet 事务管理的依赖注入Provider集合
var ProviderSet = wire.NewSet(
	NewTransactionManager,
)
