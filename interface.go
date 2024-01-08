package orbit

import "github.com/gin-gonic/gin"

type Service interface {
	RegisterGroup(g *gin.RouterGroup)
}
