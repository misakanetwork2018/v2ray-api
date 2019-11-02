package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/akkuman/parseConfig"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/exec"
	"strconv"
	"v2ray.com/core/app/proxyman/command"
	statsService "v2ray.com/core/app/stats/command"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/proxy/vmess"
)

const (
	// Default Settings
	ApiAddress = "127.0.0.1"
	ApiPort    = 8848
	InboundTag = "proxy"
	ALTERID    = 64
	ShellToUse = "bash"
)

var (
	accessKey string
	users     []V2RayUser
	cmdConn   *grpc.ClientConn
)

type V2RayUser struct {
	Email string `json:"email"`
	UUID  string `json:"uuid"`
	Level int    `json:"level"`
}

func main() {
	var config = parseConfig.New("/etc/v2ray/api_config.json")
	accessKeyI := config.Get("key")
	if accessKeyI == nil {
		fmt.Println("No access key set. Abort.")
		os.Exit(1)
	}
	accessKey = accessKeyI.(string)
	var address, msg string
	var err error
	success := false
	r := gin.Default()
	cmdConn, err = grpc.Dial(fmt.Sprintf("%s:%d", ApiAddress, ApiPort), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	r.Use(webMiddleware)
	r.POST("/sync", func(c *gin.Context) {
		var newList []V2RayUser
		var result []string
		err2 := json.Unmarshal([]byte(c.PostForm("data")), &newList)
		if err2 != nil {
			respondWithError(200, "Couldn't parse JSON data", c)
			return
		}
		if len(users) > 0 {
			result = sync(users, newList)
		} else {
			for i := 0; i < len(newList); i++ {
				resp, err3 := addUser(cmdConn, &newList[i])
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
		users = newList
		c.JSON(200, gin.H{
			"success": true,
			"msg":     result,
		})
	})
	r.POST("/addUser", func(c *gin.Context) {
		success = false
		level, _ := strconv.Atoi(c.PostForm("level"))
		email := c.PostForm("email")
		if email == "" {
			respondWithError(200, "Email Required", c)
			return
		}
		UUID := c.PostForm("email")
		if UUID == "" {
			respondWithError(200, "UUID Required", c)
			return
		}
		resp, err2 := addUser(cmdConn, &V2RayUser{email, UUID, level})
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
	})
	r.POST("/delUser", func(c *gin.Context) {
		success = false
		email := c.PostForm("email")
		if email == "" {
			respondWithError(200, "Email Required", c)
			return
		}
		resp, err2 := removeUser(cmdConn, &V2RayUser{Email: email})
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
	})
	r.POST("/traffic", func(c *gin.Context) {
		reset, _ := strconv.ParseBool(c.DefaultPostForm("reset", "false"))
		resp, err2 := traffic(cmdConn, reset)
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
	})
	r.POST("/reboot", func(c *gin.Context) {
		err, okS, errS := shell("systemctl restart v2ray")
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
	})
	addressI := config.Get("address")
	if addressI == nil {
		address = "127.0.0.1:8080"
	} else {
		address = addressI.(string)
	}
	_ = r.Run(address)
}

func sync(oldList []V2RayUser, newList []V2RayUser) []string {
	var base []V2RayUser
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
			resp, err := removeUser(cmdConn, &oldList[i])
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
			resp, err := addUser(cmdConn, &newList[i])
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

func webMiddleware(c *gin.Context) {
	if c.PostForm("key") == "" {
		respondWithError(401, "API token required", c)
		return
	}
	if c.PostForm("key") != accessKey {
		respondWithError(403, "API token incorrect", c)
		return
	}
	c.Next()
}

func respondWithError(code int, message string, c *gin.Context) {
	c.JSON(code, gin.H{
		"success": false,
		"msg":     message,
	})
	c.Abort()
}

func shell(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}

func addUser(cmdConn *grpc.ClientConn, user *V2RayUser) (*command.AlterInboundResponse, error) {
	c := command.NewHandlerServiceClient(cmdConn)
	resp, err := c.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag: InboundTag,
		Operation: serial.ToTypedMessage(&command.AddUserOperation{
			User: &protocol.User{
				Level: uint32(user.Level),
				Email: user.Email,
				Account: serial.ToTypedMessage(&vmess.Account{
					Id:               user.UUID,
					AlterId:          ALTERID,
					SecuritySettings: &protocol.SecurityConfig{Type: protocol.SecurityType_AUTO},
				}),
			},
		}),
	})
	if err != nil {
		log.Printf("failed to call grpc command: %v", err)
	} else {
		log.Printf("ok: %v", resp)
	}
	return resp, err
}
func removeUser(cmdConn *grpc.ClientConn, user *V2RayUser) (*command.AlterInboundResponse, error) {
	c := command.NewHandlerServiceClient(cmdConn)
	resp, err := c.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag: InboundTag,
		Operation: serial.ToTypedMessage(&command.RemoveUserOperation{
			Email: user.Email,
		}),
	})
	if err != nil {
		log.Printf("failed to call grpc command: %v", err)
	} else {
		log.Printf("ok: %v", resp)
	}
	return resp, err
}

func traffic(cmdConn *grpc.ClientConn, reset bool) (*statsService.QueryStatsResponse, error) {
	c := statsService.NewStatsServiceClient(cmdConn)
	r := &statsService.QueryStatsRequest{}
	r.Reset_ = reset
	resp, err := c.QueryStats(context.Background(), r)
	if err != nil {
		log.Printf("failed to call grpc command: %v", err)
	} else {
		log.Printf("ok: %v", resp)
	}
	return resp, err
}
