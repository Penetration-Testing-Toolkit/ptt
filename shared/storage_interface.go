package shared

import (
	"context"
	"github.com/Penetration-Testing-Toolkit/ptt/shared/proto"
)

// Store is the interface PTT must implement to provide key/value storage
// to plugins.
type Store interface {
	Get(ctx context.Context, pluginID, key string) ([]byte, error)
	Set(ctx context.Context, pluginID, key string, value []byte) error
}

type StoreGRPCClient struct {
	client proto.StoreClient
}

func (c *StoreGRPCClient) Get(ctx context.Context, pluginID, key string) ([]byte, error) {
	resp, err := c.client.Get(ctx, &proto.GetRequest{
		PluginId: pluginID,
		Key:      key,
	})
	if err != nil {
		return nil, err
	}

	return resp.Value, nil
}

func (c *StoreGRPCClient) Set(ctx context.Context, pluginID, key string, value []byte) error {
	_, err := c.client.Set(ctx, &proto.SetRequest{
		PluginId: pluginID,
		Key:      key,
		Value:    value,
	})
	if err != nil {
		return err
	}

	return nil
}

type StoreGRPCServer struct {
	proto.UnimplementedStoreServer

	Impl Store
}

func (s *StoreGRPCServer) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	resp, err := s.Impl.Get(ctx, req.PluginId, req.Key)
	if err != nil {
		return nil, err
	}

	return &proto.GetResponse{Value: resp}, nil
}
func (s *StoreGRPCServer) Set(ctx context.Context, req *proto.SetRequest) (*proto.Empty, error) {
	err := s.Impl.Set(ctx, req.PluginId, req.Key, req.Value)
	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}
