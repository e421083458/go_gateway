package http_proxy_middleware

import (
	"fmt"
	"github.com/e421083458/go_gateway/dao"
	"github.com/e421083458/go_gateway/middleware"
	"github.com/e421083458/go_gateway/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func HTTPJwtFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		appInterface, ok := c.Get("app")
		if !ok {
			c.Next()
			return
		}
		appInfo := appInterface.(*dao.App)
		appCounter, err := public.FlowCounterHandler.GetCounter(public.FlowAppPrefix + appInfo.AppID)
		if err != nil {
			middleware.ResponseError(c, 2002, err)
			c.Abort()
			return
		}
		appCounter.Increase()
		if appInfo.Qpd>0 && appCounter.TotalCount>appInfo.Qpd{
			middleware.ResponseError(c, 2003, errors.New(fmt.Sprintf("租户日请求量限流 limit:%v current:%v",appInfo.Qpd,appCounter.TotalCount)))
			c.Abort()
			return
		}
		c.Next()
	}
}
