package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"mini-seckill/rpc"
	"net/http"
)

//AuthorizeJWT -> to authorize JWT Token
func AuthorizeJWT() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "No Authorization header found"})

		}
		reply, err := rpc.CheckToken(authHeader)
		if err != nil || !reply.Pass {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		} else {
			log.Info().Msgf("userId:%v", reply.UserId)
			ctx.Set("userID", reply.UserId)
		}
	}

}
