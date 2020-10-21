package action

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"strings"
)

func Status() func(c *gin.Context) {
	return func(c *gin.Context) {
		var loadavg, uptime string
		var success = false
		data, err := ioutil.ReadFile("/proc/loadavg")
		if err != nil {
			fmt.Println("File reading error", err)
			goto end
		}
		loadavg = strings.Replace(string(data), "\n", "", 1)
		data, err = ioutil.ReadFile("/proc/uptime")
		if err != nil {
			fmt.Println("File reading error", err)
			goto end
		}
		uptime = strings.Split(string(data), " ")[0]
		success = true
	end:
		c.JSON(200, gin.H{
			"success": success,
			"data": gin.H{
				"loadavg": loadavg,
				"uptime":  uptime,
			},
		})
	}
}
