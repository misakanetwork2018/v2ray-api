package action

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"v2ray-api"
	"v2ray-api/utils"
)

func AddUser(users *[]api.V2RayUser) func(c *gin.Context) {
	return func(c *gin.Context) {
		var success = false
		var msg string
		level, _ := strconv.Atoi(c.PostForm("level"))
		email := c.PostForm("email")
		if email == "" {
			utils.RespondWithError(403, "Email Required", c)
			return
		}
		UUID := c.PostForm("email")
		if UUID == "" {
			utils.RespondWithError(403, "UUID Required", c)
			return
		}
		var user = api.V2RayUser{Email: email, UUID: UUID, Level: level}
		resp, err2 := api.AddUser(user)
		if err2 == nil {
			success = true
			msg = resp.String()
			*users = append(*users, user)
		} else {
			msg = err2.Error()
		}
		c.JSON(200, gin.H{
			"success": success,
			"msg":     msg,
		})
	}
}
