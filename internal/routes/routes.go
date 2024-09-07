package routes

import (
	"uniswap-v4-rpc/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	router.POST("/approve", handlers.ApproveTokens)
	router.POST("/initialize", handlers.Initialize)
	router.POST("/addLiquidity", handlers.AddLiquidity)
	router.POST("/addLiquidityPermit", handlers.AddLiquidityPermit)
	router.POST("/performSwap", handlers.Swap)
	router.POST("/performSwapWithPermit", handlers.SwapPermit)

}
