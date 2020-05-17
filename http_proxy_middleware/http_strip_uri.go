package http_proxy_middleware

import (
	"github.com/e421083458/go_gateway/dao"
	"github.com/e421083458/go_gateway/middleware"
	"github.com/e421083458/go_gateway/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

//匹配接入方式 基于请求信息
func HTTPStripUriMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serverInterface.(*dao.ServiceDetail)

		if serviceDetail.HTTPRule.RuleType==public.HTTPRuleTypePrefixURL && serviceDetail.HTTPRule.NeedStripUri==1{
			//fmt.Println("c.Request.URL.Path",c.Request.URL.Path)
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path,serviceDetail.HTTPRule.Rule,"",1)
			//fmt.Println("c.Request.URL.Path",c.Request.URL.Path)
		}
		//http://127.0.0.1:8080/test_http_string/abbb
		//http://127.0.0.1:2004/abbb

		c.Next()
	}
}
