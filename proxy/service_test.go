package proxy

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_NewHttpProxyByGinCustom(t *testing.T) {
	host := "localhost:8000"
	springboot := "localhost:9000"

	router := gin.New()

	proxyGroup := router.Group("")
	// proxyGroup.Use(jwts.NewGinJwtMiddlewares(jwtservice.PaasAuthorizator).MiddlewareFunc())
	//proxyGroup.Use(jwts.GetMiddleware().MiddlewareFunc())
	// 代理
	{
		proxyGroup.Any("/dashboards/*action", NewHttpProxyByGinCustom(fmt.Sprintf("%s/dashboard", host), nil))
		proxyGroup.Any("/apis/*action", NewHttpProxyByGinCustom(fmt.Sprintf("%s/api", host), nil))
		proxyGroup.Any("/greeting/*action", NewHttpProxyByGinCustom(fmt.Sprintf("%s/greeting", springboot), nil))
		// proxyGroup.Any("/api/*action", NewHttpProxyByGinCustomRaw(fmt.Sprintf("%s/api", host), nil))
		// proxyGroup.Any("/dashboard/*action", NewHttpProxyByGinCustomRaw(fmt.Sprintf("%s/dashboard", host), nil))
	}

	// TODO: test it
}
