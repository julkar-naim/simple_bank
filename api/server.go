package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/julkar-naim/simple-bank/db/sqlc"
)

type Server struct {
	store  *db.Store
	router *gin.Engine
}

func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)
	router.POST("/accounts/update", server.updateAccount)
	router.GET("/accounts", server.getAccountList)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts/delete/:id", server.deleteAccount)

	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}
