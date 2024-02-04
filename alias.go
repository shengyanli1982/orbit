package orbit

import "github.com/gin-gonic/gin"

// 相关 gin 类型的别名。
// Aliases of the related gin types.
type (
	RouterGroup   = gin.RouterGroup
	Context       = gin.Context
	HandlerFunc   = gin.HandlerFunc
	HandlersChain = gin.HandlersChain
	Accounts      = gin.Accounts
)
