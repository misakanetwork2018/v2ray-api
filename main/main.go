package main

import (
	"fmt"
	"github.com/akkuman/parseConfig"
	"github.com/gin-gonic/gin"
	"os"
	"v2ray-api"
	"v2ray-api/action"
	"v2ray-api/utils"
)

var (
	accessKey string
	users     []api.V2RayUser
)

func main() {
	var config = parseConfig.New("/etc/v2ray/api_config.json")
	accessKeyI := config.Get("key")
	if accessKeyI == nil {
		fmt.Println("No access key set. Abort.")
		os.Exit(1)
	}
	accessKey = accessKeyI.(string)

	api.InitGRPC()

	r := gin.Default()
	r.Use(webMiddleware)
	r.POST("/sync", action.Sync(&users))
	r.POST("/addUser", action.AddUser(&users))
	r.POST("/delUser", action.DelUser(&users))
	r.POST("/traffic", action.Traffic())
	r.POST("/reboot", action.Reboot())
	r.POST("/status", action.Status())

	var address string

	addressI := config.Get("address")
	if addressI == nil {
		address = "127.0.0.1:8080"
	} else {
		address = addressI.(string)
	}

	_ = r.Run(address)
}

func webMiddleware(c *gin.Context) {
	if c.PostForm("key") == "" {
		utils.RespondWithError(401, "API token required", c)
		return
	}
	if c.PostForm("key") != accessKey {
		utils.RespondWithError(403, "API token incorrect", c)
		return
	}
	c.Next()
}
