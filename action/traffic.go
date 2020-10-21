package action

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"v2ray-api"
)

func Traffic() func(c *gin.Context) {
	return func(c *gin.Context) {
		reset, _ := strconv.ParseBool(c.DefaultPostForm("reset", "false"))
		resp, err2 := api.Traffic(reset)
		var success = false
		var msg string
		if err2 == nil {
			success = true
			msg = resp.String()
		} else {
			msg = err2.Error()
		}
		c.JSON(200, gin.H{
			"success": success,
			"msg":     msg,
		})
	}
}
