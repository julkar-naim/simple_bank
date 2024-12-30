package api

import (
	"github.com/gin-gonic/gin"
)

type CreateTransferRequest struct {
	FromAccountID int64  `json:"from_account_id"`
	ToAccountId   int64  `json:"to_account_id"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,oneof=USD EUR CAD"`
}

func (server *Server) createTransfer(ctx *gin.Context) {

}
