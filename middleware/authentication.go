package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"mini-seckill/rpc"
	"net/http"
)

//AuthorizeJWT -> to authorize JWT Token
func Authentication(obj string, act string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sub, exist := ctx.Get("userID")
		if !exist {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No userID found"})
		}
		pass, err := rpc.Authentication(fmt.Sprint(sub), obj, act)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		if !pass {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		}
		ctx.Next()
	}
}
