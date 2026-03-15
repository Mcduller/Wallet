package controller

import (
	"errors"
	"net/http"
	"wallet_service/internal/domain"
	"wallet_service/internal/service"

	"github.com/gin-gonic/gin"
)

type WalletController struct {
	service service.WalletService
}

func NewWalletController(svc service.WalletService) *WalletController {
	return &WalletController{service: svc}
}

func (c *WalletController) CreateWallet(ctx *gin.Context) {
	wallet, err := c.service.CreateWallet(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, wallet)
}

func (c *WalletController) GetWallet(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "wallet id is required"})
		return
	}

	wallet, err := c.service.GetWallet(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, wallet)
}

type TransferRequest struct {
	SourceID      string `json:"source_id" binding:"required"`
	DestinationID string `json:"destination_id" binding:"required"`
	Amount        int    `json:"amount" binding:"required,gt=0"`
}

func (c *WalletController) Transfer(ctx *gin.Context) {
	var req TransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.service.Transfer(ctx.Request.Context(), req.SourceID, req.DestinationID, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidAmount):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrSameWallet):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrInsufficientFunds):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrWalletNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}