package api

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"v2ray.com/core/app/proxyman/command"
	statsService "v2ray.com/core/app/stats/command"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/proxy/vmess"
)

const (
	// Default Settings
	apiAddress = "127.0.0.1"
	apiPort    = 8848
	InboundTag = "proxy"
	ALTERID    = 0
)

var (
	cmdConn *grpc.ClientConn
)

type V2RayUser struct {
	Email string `json:"email"`
	UUID  string `json:"uuid"`
	Level int    `json:"level"`
}

func InitGRPC() {
	var err error
	cmdConn, err = grpc.Dial(fmt.Sprintf("%s:%d", apiAddress, apiPort), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
}

func AddUser(user V2RayUser) (*command.AlterInboundResponse, error) {
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

func RemoveUser(user V2RayUser) (*command.AlterInboundResponse, error) {
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

func Traffic(reset bool) (*statsService.QueryStatsResponse, error) {
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
