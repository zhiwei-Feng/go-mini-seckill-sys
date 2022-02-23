package rpc

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

const addr = "seckill-user:50051"

func CheckToken(token string) (*CheckTokenReply, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error().Err(err).Msg("did not connect")
		return nil, err
	}
	defer conn.Close()
	c := NewUserClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.CheckToken(ctx, &CheckTokenReq{Token: token})
	if err != nil {
		log.Error().Err(err).Msg("could not CheckToken")
		return nil, err
	}
	return r, nil
}

func Authentication(sub, obj, act string) (bool, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error().Err(err).Msg("did not connect")
		return false, err
	}
	defer conn.Close()
	c := NewUserClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Authentication(ctx, &AuthReq{Sub: sub, Obj: obj, Act: act})
	if err != nil {
		log.Error().Err(err).Msg("could not Authentication")
		return false, err
	}
	return r.Pass, nil
}
