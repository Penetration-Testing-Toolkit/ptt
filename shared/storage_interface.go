package shared

import (
	"context"
	"github.com/Penetration-Testing-Toolkit/ptt/shared/proto"
)

// Store is the interface PTT must implement to provide key/value storage
// to plugins.
type Store interface {
	// Get retrieves the []byte value associated with the provided IDs and key.
	// userID and projectID may be empty (string: "").
	// If userID has a value and projectID is empty, the value is associated with that user but no specific project
	// context.
	// If userID is empty and projectID has a value, the value is associated with that project but not a specific user
	// context.
	// If userID is empty and projectID is empty, the value is only associated with the plugin, not any specific
	// user or project context.
	Get(ctx context.Context, pluginID, userID, projectID, key string) ([]byte, error)

	// Set updates the provided IDs and key pairing with the provided []byte value.
	// userID and projectID may be empty (string: "").
	// If userID has a value and projectID is empty, the value is associated with that user but no specific project
	// context.
	// If userID is empty and projectID has a value, the value is associated with that project but not a specific user
	// context.
	// If userID is empty and projectID is empty, the value is only associated with the plugin, not any specific
	// user or project context.
	Set(ctx context.Context, pluginID, userID, projectID, key string, value []byte) error

	// Delete deletes the []byte value associated with the provided IDs and key.
	// userID and projectID may be empty (string: "").
	Delete(ctx context.Context, pluginID, userID, projectID, key string) error
}

type StoreGRPCClient struct {
	client proto.StoreClient
}

func (c *StoreGRPCClient) Get(ctx context.Context, pluginID, userID, projectID, key string) ([]byte, error) {
	resp, err := c.client.Get(ctx, &proto.GetRequest{
		PluginId:  pluginID,
		UserId:    userID,
		ProjectId: projectID,
		Key:       key,
	})
	if err != nil {
		return nil, err
	}

	return resp.Value, nil
}

func (c *StoreGRPCClient) Set(ctx context.Context, pluginID, userID, projectID, key string, value []byte) error {
	_, err := c.client.Set(ctx, &proto.SetRequest{
		PluginId:  pluginID,
		UserId:    userID,
		ProjectId: projectID,
		Key:       key,
		Value:     value,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *StoreGRPCClient) Delete(ctx context.Context, pluginID, userID, projectID, key string) error {
	_, err := c.client.Delete(ctx, &proto.DeleteRequest{
		PluginId:  pluginID,
		UserId:    userID,
		ProjectId: projectID,
		Key:       key,
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
	resp, err := s.Impl.Get(ctx, req.PluginId, req.UserId, req.ProjectId, req.Key)
	if err != nil {
		return nil, err
	}

	return &proto.GetResponse{Value: resp}, nil
}

func (s *StoreGRPCServer) Set(ctx context.Context, req *proto.SetRequest) (*proto.Empty, error) {
	err := s.Impl.Set(ctx, req.PluginId, req.UserId, req.ProjectId, req.Key, req.Value)
	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *StoreGRPCServer) Delete(ctx context.Context, req *proto.DeleteRequest) (*proto.Empty, error) {
	err := s.Impl.Delete(ctx, req.PluginId, req.UserId, req.ProjectId, req.Key)
	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}
