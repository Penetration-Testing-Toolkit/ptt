// Package shared See https://github.com/hashicorp/go-plugin for help with the plugin system.
package shared

import (
	"context"
	"errors"
	"github.com/Penetration-Testing-Toolkit/ptt/shared/proto"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"net/http"
)

var Logger hclog.Logger

// HandshakeConfig is used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user-friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PTT_PLUGIN",
	MagicCookieValue: "HELLO_PLUGINS",
}

// PluginMap is the map of plugins that a plugin can provide to the server.
var PluginMap = map[string]plugin.Plugin{
	"module": &ModuleGRPCPlugin{},
}

// Route is the HTTP route a Module plugin wants to register.
type Route struct {
	Method string
	Path   string

	// UseSSE is an optional flag indicating an SSE route.
	UseSSE bool
}

// ModuleInfo contains a Module plugin's information to provide to the server.
type ModuleInfo struct {
	// ID should be the git URL of the plugin (same as Go mod) in order to keep plugin IDs unique.
	// DO NOT include a trailing '/'.
	// E.g. github.com/chronotrax/plugin_example
	ID string

	// Name is the "pretty" name of the Module plugin.
	// E.g. "Module Example"
	Name string

	// Version should follow semantic versioning.
	// E.g. "1.2.3"
	Version string

	// Routes is a slice of Route the Module needs to handle.
	Routes []*Route
}

// Response is an HTTP response for transmission over gRPC.
type Response struct {
	Status int
	Header http.Header

	// Body must be an HTML string to render to the page.
	// If there is no content to render, leave the Body empty.
	Body string
}

// Module is the interface all Module plugins must implement.
// A Module plugin is responsible for taking a proxied http.Request and responding with Response, containing HTML
// that will be rendered to the page.
type Module interface {
	// Register returns a ModuleInfo to the server.
	// It is the first thing called by the server to set up the Module plugin.
	Register(ctx context.Context) (*ModuleInfo, error)

	// Handle is the primary function of a Module plugin; to handle an http.Request proxied to the plugin.
	Handle(ctx context.Context, req *http.Request) (*Response, error)

	// HandleSSE is the SSE version of Handle.
	// It is optional to use it, but it should at least be declared to fulfil the interface.
	HandleSSE(ctx context.Context, req *http.Request) (chan *Response, error)
}

// ModuleGRPCClient is a gRPC client implementation of proto.ModuleClient.
// It implements the Module interface.
// Its methods are how the host app's gRPC client will call a plugin's ModuleGRPCServer.
type ModuleGRPCClient struct {
	client proto.ModuleClient
}

func (c *ModuleGRPCClient) Register(ctx context.Context) (*ModuleInfo, error) {
	resp, err := c.client.Register(ctx, &proto.Empty{})
	if err != nil {
		return nil, err
	}

	return &ModuleInfo{
		ID:      resp.GetId(),
		Name:    resp.GetName(),
		Version: resp.GetVersion(),
		Routes:  routesFromProto(resp.GetRoutes()),
	}, nil
}

func (c *ModuleGRPCClient) Handle(ctx context.Context, req *http.Request) (*Response, error) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Handle(ctx, &proto.Request{
		Method: req.Method,
		Url:    req.URL.String(),
		Header: headerToProto(req.Header),
		Body:   b,
	})
	if err != nil {
		return nil, err
	}

	return &Response{
		Status: int(resp.GetStatus()),
		Header: headerFromProto(resp.GetHeader()),
		Body:   resp.GetBody(),
	}, nil
}

func (c *ModuleGRPCClient) HandleSSE(ctx context.Context, req *http.Request) (chan *Response, error) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	// Get gRPC server streaming client & send request
	streamClient, err := c.client.HandleSSE(ctx, &proto.Request{
		Method: req.Method,
		Url:    req.URL.String(),
		Header: headerToProto(req.Header),
		Body:   b,
	})
	if err != nil {
		return nil, err
	}
	Logger.Debug("sent HandleSSE request")

	// Listen for responses
	ch := make(chan *Response, 1)
	go func() {
		Logger.Debug("SSE response listener started")
		for {
			select {
			case <-ctx.Done():
				// Close connection
				Logger.Debug("closing SSE gRPC streamClient", "cause", "ctx.Done() received")
				err = streamClient.CloseSend()
				if err != nil {
					Logger.Error("error from SSE gRPC streamClient.CloseSend()", "error", err.Error())
				}

				Logger.Debug("closing server's SSE channel", "cause", "ctx.Done() received")
				close(ch)
				return
			default:
				res, err := streamClient.Recv()
				if err != nil {
					st, ok := status.FromError(err)
					if errors.Is(err, io.EOF) {
						Logger.Debug("closing server's SSE channel",
							"cause", "SSE gRPC streamClient received EOF")
					} else if ok && st.Code() == codes.Canceled {
						Logger.Debug("closing server's SSE channel",
							"cause", "SSE canceled",
							"code", st.Code(), "message", st.Message())
					} else {
						Logger.Debug("closing server's SSE channel",
							"cause", "error from SSE gRPC streamClient.Recv()",
							"error", err.Error())
					}

					close(ch)
					return
				}

				Logger.Debug("SSE response received from stream client")

				ch <- &Response{
					Status: int(res.GetStatus()),
					Header: headerFromProto(res.GetHeader()),
					Body:   res.GetBody(),
				}

				Logger.Debug("SSE response sent to server's channel")
			}
		}
	}()

	return ch, nil
}

// ModuleGRPCServer is a gRPC server implementation of proto.ModuleServer.
// It uses the same method names as the Module interface, but does not actually implement the interface.
// Its methods are how the plugin's gRPC server will call the plugin's concrete implementation,
// then respond back to the host app's ModuleGRPCClient.
type ModuleGRPCServer struct {
	proto.UnimplementedModuleServer

	// Impl is the concrete implementation of Module.
	Impl Module
}

func (s *ModuleGRPCServer) Register(ctx context.Context, _ *proto.Empty) (*proto.RegisterResponse, error) {
	info, err := s.Impl.Register(ctx)
	if err != nil {
		return nil, err
	}

	return &proto.RegisterResponse{
		Id:      info.ID,
		Name:    info.Name,
		Version: info.Version,
		Routes:  routesToProto(info.Routes),
	}, nil
}

func (s *ModuleGRPCServer) Handle(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	r, err := stdRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := s.Impl.Handle(ctx, r)
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Status: int32(resp.Status),
		Header: headerToProto(resp.Header),
		Body:   resp.Body,
	}, nil
}

func (s *ModuleGRPCServer) HandleSSE(req *proto.Request, streamServer grpc.ServerStreamingServer[proto.Response]) error {
	Logger.Debug("HandleSSE called")

	// streamServer is the gRPC server streamer
	ctx := streamServer.Context()

	r, err := stdRequest(ctx, req)
	if err != nil {
		return err
	}

	Logger.Debug("sending SSE request to plugin implementation")
	ch, err := s.Impl.HandleSSE(ctx, r)
	if err != nil {
		return err
	}

	for resp := range ch {
		Logger.Debug("SSE response came through plugin's channel, relaying to app server")
		err = streamServer.Send(&proto.Response{
			Status: int32(resp.Status),
			Header: headerToProto(resp.Header),
			Body:   resp.Body,
		})
		if err != nil {
			return err
		}
	}

	Logger.Debug("done reading SSE responses from implementation")

	return nil
}

// ModuleGRPCPlugin is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type ModuleGRPCPlugin struct {
	// Not supporting net/rpc, only gRPC.
	plugin.NetRPCUnsupportedPlugin

	// Impl is the Concrete implementation of Module, written in Go. This is only used for plugins
	// that are written in Go.
	Impl Module
}

func (p *ModuleGRPCPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterModuleServer(s, &ModuleGRPCServer{Impl: p.Impl})
	return nil
}

func (p *ModuleGRPCPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &ModuleGRPCClient{client: proto.NewModuleClient(c)}, nil
}
