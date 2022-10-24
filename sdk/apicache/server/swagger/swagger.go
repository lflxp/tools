package swagger

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterSwaggerMiddleware(router *gin.Engine) {
	// swagger
	router.GET("/cache/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
