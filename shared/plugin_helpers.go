package shared

import (
	"bytes"
	"context"
	"github.com/Penetration-Testing-Toolkit/ptt/shared/proto"
	"github.com/hashicorp/go-hclog"
	"net/http"
	"os"
)

const (
	PTTUsername    = "PTT-Username"
	PTTUserID      = "PTT-User-ID"
	PTTProjectName = "PTT-Project-Name"
	PTTProjectID   = "PTT-Project-ID"
)

// LoggerOptions are common options that server & plugins can use as defaults.
var LoggerOptions = &hclog.LoggerOptions{
	Level:       hclog.Info,
	Output:      os.Stderr,
	JSONFormat:  true,
	DisableTime: true,
}

// stdRequest converts a proto.Request to an http.Request with context.Context.
func stdRequest(ctx context.Context, req *proto.Request) (*http.Request, error) {
	h, err := http.NewRequestWithContext(ctx, req.GetMethod(), req.GetUrl(), bytes.NewReader(req.GetBody()))
	if err != nil {
		return nil, err
	}
	h.Header = headerFromProto(req.GetHeader())
	return h, nil
}

// headerToProto prepares http.Header for transmission over gRPC proto.Header.
func headerToProto(headers http.Header) *proto.Header {
	h := make(map[string]*proto.Header_Value)
	for k, v := range headers {
		h[k] = &proto.Header_Value{Values: v}
	}
	return &proto.Header{Header: h}
}

// headerFromProto converts gRPC proto.Header back into an http.Header.
func headerFromProto(headers *proto.Header) http.Header {
	h := make(map[string][]string)
	for k, v := range headers.Header {
		h[k] = v.GetValues()
	}
	return h
}

// routesToProto prepares Route for transmission over gRPC proto.RegisterResponse_Route.
func routesToProto(routes []*Route) []*proto.RegisterResponse_Route {
	r := make([]*proto.RegisterResponse_Route, 0)
	for _, route := range routes {
		r = append(r, &proto.RegisterResponse_Route{
			Method: route.Method,
			Path:   route.Path,
			UseSse: route.UseSSE,
		})
	}
	return r
}

// routesFromProto converts gTPC proto.RegisterResponse_Route back into a Route.
func routesFromProto(routes []*proto.RegisterResponse_Route) []*Route {
	r := make([]*Route, 0)
	for _, route := range routes {
		r = append(r, &Route{
			Method: route.GetMethod(),
			Path:   route.GetPath(),
			UseSSE: route.GetUseSse(),
		})
	}
	return r
}
