package http_proxy_middleware

import (
	"github.com/e421083458/go_gateway/dao"
	"github.com/e421083458/go_gateway/middleware"
	"github.com/gin-gonic/gin"
)

func HttpAccessModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceDetail, err := dao.ServiceHandler.MatchAccessMode(c)
		if err != nil {
			middleware.ResponseError(c, 1001, err)
			c.Abort()
			return
		}
		c.Set("service_detail", serviceDetail)
		c.Next()
	}
}
