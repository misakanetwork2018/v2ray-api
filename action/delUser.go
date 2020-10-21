package action

import (
	"github.com/gin-gonic/gin"
	"sort"
	"v2ray-api"
	"v2ray-api/utils"
)

func DelUser(users *[]api.V2RayUser) func(c *gin.Context) {
	return func(c *gin.Context) {
		var success = false
		var msg string
		email := c.PostForm("email")
		if email == "" {
			utils.RespondWithError(403, "Email Required", c)
			return
		}
		var user = api.V2RayUser{Email: email}
		resp, err2 := api.RemoveUser(user)
		if err2 == nil {
			success = true
			msg = resp.String()
			var tmpUsers = *users
			index := sort.Search(len(tmpUsers), func(i int) bool {
				return tmpUsers[i].Email == email
			})
			*users = append(tmpUsers[:0], tmpUsers[index+1])
		} else {
			msg = err2.Error()
		}
		c.JSON(200, gin.H{
			"success": success,
			"msg":     msg,
		})
	}
}
