package action

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"v2ray-api"
	"v2ray-api/utils"
)

func Sync(users *[]api.V2RayUser) func(c *gin.Context) {
	return func(c *gin.Context) {
		var newList []api.V2RayUser
		var result []string
		err2 := json.Unmarshal([]byte(c.PostForm("data")), &newList)
		if err2 != nil {
			utils.RespondWithError(500, "Couldn't parse JSON data", c)
			return
		}
		if len(*users) > 0 {
			result = sync(*users, newList)
		} else {
			for i := 0; i < len(newList); i++ {
				resp, err3 := api.AddUser(newList[i])
				if err3 == nil {
					msg := "Add User [" + newList[i].Email + "] OK: " + resp.String()
					fmt.Println(msg)
					result = append(result, msg)
				} else {
					msg := "Add User [" + newList[i].Email + "] ERR: " + err3.Error()
					fmt.Println(msg)
					result = append(result, msg)
				}
			}
		}
		*users = newList
		c.JSON(200, gin.H{
			"success": true,
			"msg":     result,
		})
	}
}

func sync(oldList []api.V2RayUser, newList []api.V2RayUser) []string {
	var base []api.V2RayUser
	var result []string
	for i := 0; i < len(oldList); i++ {
		found := false
		for j := 0; j < len(newList); j++ {
			if oldList[i] == newList[j] {
				found = true
				break
			}
		}
		if found {
			base = append(base, oldList[i])
		} else {
			resp, err := api.RemoveUser(oldList[i])
			if err == nil {
				msg := "Remove User [" + oldList[i].Email + "] OK: " + resp.String()
				fmt.Println(msg)
				result = append(result, msg)
			} else {
				msg := "Remove User [" + oldList[i].Email + "] ERR: " + err.Error()
				fmt.Println(msg)
				result = append(result, msg)
			}
		}
	}
	for i := 0; i < len(newList); i++ {
		found := false
		for j := 0; j < len(base); j++ {
			if newList[i] == base[j] {
				found = true
				break
			}
		}
		if !found {
			resp, err := api.AddUser(newList[i])
			if err == nil {
				msg := "Add User [" + newList[i].Email + "] OK: " + resp.String()
				fmt.Println(msg)
				result = append(result, msg)
			} else {
				msg := "Add User [" + newList[i].Email + "] ERR: " + err.Error()
				fmt.Println(msg)
				result = append(result, msg)
			}
		}
	}
	return result
}
