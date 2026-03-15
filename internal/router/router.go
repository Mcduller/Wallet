package router

import (
	"wallet_service/internal/controller"

	"github.com/gin-gonic/gin"
)

func Setup(walletController *controller.WalletController) *gin.Engine {
	r := gin.Default()

	// Wallet routes
	r.POST("/wallets", walletController.CreateWallet)
	r.GET("/wallets/:id", walletController.GetWallet)

	// Transfer route
	r.POST("/transfer", walletController.Transfer)

	return r
}