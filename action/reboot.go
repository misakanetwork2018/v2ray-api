package action

import (
	"github.com/gin-gonic/gin"
	"v2ray-api/utils"
)

func Reboot() func(c *gin.Context) {
	return func(c *gin.Context) {
		err, okS, errS := utils.Shell("systemctl restart v2ray")
		var success = false
		var msg string
		if err == nil {
			msg = okS
			success = true
		} else {
			msg = errS
			success = false
		}
		c.JSON(200, gin.H{
			"success": success,
			"msg":     msg,
		})
		_, _, _ = utils.Shell("systemctl restart v2ray-proxy")
	}
}
