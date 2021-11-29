package controller

import (
	"github.com/gin-gonic/gin"
	"mini-seckill/service"
	"net/http"
	"strconv"
)

func CreateWrongOrder(c *gin.Context) {
	sid, err := strconv.Atoi(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	//log.Println("buying..., sid:", sid)
	id := service.CreateWrongOrder(sid)
	//log.Println("create order, id:", id)
	c.JSON(http.StatusOK, gin.H{"message": "ok", "id": id})
}
