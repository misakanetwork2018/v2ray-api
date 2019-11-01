package main

import (
	"bytes"
	"context"
	"fmt"
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

func main() {
	var address, msg string
	success := false
	r := gin.Default()
	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", ApiAddress, ApiPort), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	r.POST("/addUser", func(c *gin.Context) {
		success = false
		level, _ := strconv.Atoi(c.PostForm("level"))
		resp, _ := addUser(cmdConn, uint32(level), c.PostForm("email"), c.PostForm("uuid"))
		msg = resp.String()
		c.JSON(200, gin.H{
			"success": success,
			"msg":     msg,
		})
	})
	r.POST("/delUser", func(c *gin.Context) {
		success = false
		resp, _ := removeUser(cmdConn, c.PostForm("email"))
		msg = resp.String()
		c.JSON(200, gin.H{
			"success": success,
			"msg":     msg,
		})
	})
	r.POST("/traffic", func(c *gin.Context) {
		reset, _ := strconv.ParseBool(c.DefaultPostForm("reset", "false"))
		resp, _ := traffic(cmdConn, reset)
		msg = resp.String()
		c.JSON(200, gin.H{
			"success": success,
			"msg":     msg,
		})
	})
	args := os.Args[1:]
	if len(args) >= 1 {
		address = args[0]
	} else {
		address = "127.0.0.1:8080"
	}
	r.Use(webMiddleware)
	_ = r.Run(address)
}

func webMiddleware(c *gin.Context) {
	if c.PostForm("key") == "" {

	}
	c.Next()
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

func addUser(cmdConn *grpc.ClientConn, level uint32, email string, UUID string) (*command.AlterInboundResponse, error) {
	c := command.NewHandlerServiceClient(cmdConn)
	resp, err := c.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag: InboundTag,
		Operation: serial.ToTypedMessage(&command.AddUserOperation{
			User: &protocol.User{
				Level: level,
				Email: email,
				Account: serial.ToTypedMessage(&vmess.Account{
					Id:               UUID,
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
func removeUser(cmdConn *grpc.ClientConn, email string) (*command.AlterInboundResponse, error) {
	c := command.NewHandlerServiceClient(cmdConn)
	resp, err := c.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag: InboundTag,
		Operation: serial.ToTypedMessage(&command.RemoveUserOperation{
			Email: email,
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
