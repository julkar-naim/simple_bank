package api

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	db "github.com/julkar-naim/simple-bank/db/sqlc"
	"net/http"
)

type CreateTransferRequest struct {
	FromAccountID int64  `json:"from_account_id"`
	ToAccountId   int64  `json:"to_account_id"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req CreateTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !server.validateCurrency(ctx, req) {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountId,
		Amount:        req.Amount,
	}
	result, err := server.store.TransferTx(context.Background(), arg)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validateCurrency(ctx *gin.Context, req CreateTransferRequest) bool {
	account1, err := server.store.GetAccount(context.Background(), req.FromAccountID)
	account2, err := server.store.GetAccount(context.Background(), req.ToAccountId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}
	if account1.Currency != req.Currency || account2.Currency != req.Currency {
		err := errors.New("currency mismatch")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}
	return true
}
