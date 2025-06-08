package server

import (
	"context"
	"fmt"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/database"
	"github.com/Penetration-Testing-Toolkit/ptt/shared"
	"github.com/Penetration-Testing-Toolkit/ptt/shared/proto"
	"google.golang.org/grpc"
	"math/rand/v2"
	"net"
)

type store struct {
	g *app.Global
}

func (s *store) Get(ctx context.Context, pluginID, key string) ([]byte, error) {
	value, err := database.StoreGet(ctx, s.g, pluginID, key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (s *store) Set(ctx context.Context, pluginID, key string, value []byte) error {
	return database.StoreSet(ctx, s.g, pluginID, key, value)
}

func startStoreServer(g *app.Global) (serv *grpc.Server, addr string, err error) {
	// TODO: use network address when on Windows
	addr = fmt.Sprintf("/tmp/store-server-%d", rand.Int())
	lis, err := net.Listen("unix", addr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create store server listener: %w", err)
	}
	serv = grpc.NewServer([]grpc.ServerOption{}...)
	proto.RegisterStoreServer(serv, &shared.StoreGRPCServer{Impl: &store{g: g}})

	g.Logger().Info("store gRPC server listening", "address", addr)
	go func() {
		err = serv.Serve(lis)
		if err != nil {
			g.Logger().Error("error serving store gRPC server", "error", err.Error())
		}
	}()

	return serv, addr, nil
}
