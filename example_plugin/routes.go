package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/template/example_plugin"
	"github.com/Penetration-Testing-Toolkit/ptt/shared"
	"github.com/Penetration-Testing-Toolkit/ptt/shared/proto"
	"github.com/a-h/templ"
	"github.com/hashicorp/go-hclog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (m *ModuleExample) setupRoutes() {
	// Only these methods are needed for now
	// If more methods are needed, add to this list
	router.routes[http.MethodGet] = make(map[string]HandlerFunc)
	router.routes[http.MethodPost] = make(map[string]HandlerFunc)
	router.sseRoutes[http.MethodGet] = make(map[string]SSEHandlerFunc)

	rootPath := "/plugin/" + info.ID

	r := &shared.Route{
		Method: http.MethodGet,
		Path:   "/index",
	}
	router.routes[r.Method][rootPath+r.Path] = m.index
	info.Routes = append(info.Routes, r)

	r = &shared.Route{
		Method: http.MethodPost,
		Path:   "/sum",
	}
	router.routes[r.Method][rootPath+r.Path] = m.sum
	info.Routes = append(info.Routes, r)

	r = &shared.Route{
		Method: http.MethodGet,
		Path:   "/sse",
		UseSSE: true,
	}
	router.sseRoutes[r.Method][rootPath+r.Path] = m.sse
	info.Routes = append(info.Routes, r)

	// To create a new route, copy a route like above, and make the map[string]handler if needed
}

// helper is a function to help with rendering the templ component & responding
func helper(ctx context.Context, comp templ.Component, status int, header http.Header) (*shared.Response, error) {
	resp := &bytes.Buffer{}
	err := comp.Render(ctx, resp)
	if err != nil {
		return nil, err
	}

	return &shared.Response{
		Status: status,
		Header: header,
		Body:   resp.String(),
	}, nil
}

// index "GET /index" return's the plugin's root page.
// A /index is the REQUIRED starting point of a plugin.
func (m *ModuleExample) index(ctx context.Context, req *http.Request) (*shared.Response, error) {
	// If logger is debug level or lower, print all request headers
	if m.logger.GetLevel() <= hclog.Debug {
		for k, v := range req.Header {
			m.logger.Debug("request header", k, v)
		}
	}

	// PTT passes user & project info in request headers
	username := req.Header.Get(shared.PTTUsername)
	userID := req.Header.Get(shared.PTTUserID)
	projectName := req.Header.Get(shared.PTTProjectName)
	projectID := req.Header.Get(shared.PTTProjectID)

	// Query PTT database store for previously stored value
	key := fmt.Sprintf("%s:%s", userID, projectID)

	// Send the "Get" request over the gRPC client
	resp, err := storeClient.Get(ctx, &proto.GetRequest{
		PluginId: info.ID,
		Key:      key,
	})
	if err != nil {
		err = fmt.Errorf("error sending Get request to store: %w", err)
		m.logger.Error(err.Error())
		return nil, err
	}

	// Convert data from JSON (stored as []byte) back to int
	val := new(data)
	if len(resp.Value) > 0 {
		err = json.Unmarshal(resp.Value, &val)
		if err != nil {
			err = fmt.Errorf("error unmarshaling data: %w", err)
			m.logger.Error(err.Error())
			return nil, err
		}
	} else {
		m.logger.Debug("Get response value was empty")
	}

	// Note that we can pass an http.Header that will be seen from the frontend client
	return helper(ctx, template.Example(req.Method, req.URL.String(),
		username, userID, projectName, projectID, val.Sum),
		http.StatusOK, http.Header{"Example": {"Hello World!"}})
}

type data struct {
	Sum int `json:"sum"`
}

// sum "POST /sum" example of an http.MethodPost request.
func (m *ModuleExample) sum(ctx context.Context, req *http.Request) (*shared.Response, error) {
	sum := 0
	numStr := req.PostFormValue("numbers")
	if numStr != "" {
		for _, s := range strings.Split(numStr, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				return helper(ctx, template.Error("'"+s+"' is not an integer"),
					http.StatusUnprocessableEntity, nil)
			}
			sum += n
		}
	}

	// We want to save this value to the PTT database store
	// Create key based on user & project
	userID := req.Header.Get(shared.PTTUserID)
	projectID := req.Header.Get(shared.PTTProjectID)
	key := fmt.Sprintf("%s:%s", userID, projectID)

	// Convert data we want to store into JSON (stored as []byte)
	val, err := json.Marshal(data{Sum: sum})
	if err != nil {
		err = fmt.Errorf("error marshaling data: %w", err)
		m.logger.Error(err.Error())
		return nil, err
	}

	// Send the "Set" request over the gRPC client
	_, err = storeClient.Set(ctx, &proto.SetRequest{
		PluginId: info.ID,
		Key:      key,
		Value:    val,
	})
	if err != nil {
		err = fmt.Errorf("error sending Set request to store: %w", err)
		m.logger.Error(err.Error())
		return nil, err
	}

	return helper(ctx, template.Numbers(strconv.Itoa(sum)),
		http.StatusOK, nil)
}

// sse "GET /sse" example of a sse request.
// Returns a Server-sent event rather than a templ component's HTML.
// Requires special handling on the frontend.
func (m *ModuleExample) sse(ctx context.Context, req *http.Request) (chan *shared.Response, error) {
	m.logger.Debug("SSE request received by plugin implementation", "method", req.Method, "URL", req.URL.String(),
		"protocol", req.Proto)

	ch := make(chan *shared.Response, 1)
	go func() {
		timer := time.NewTimer(4 * time.Second)
		defer timer.Stop()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				m.logger.Debug("closing plugin's SSE channel", "cause", "ctx.Done()")
				close(ch)
				return
			case <-ticker.C:
				resp := "time: " + time.Now().Format(time.RFC1123)

				ch <- &shared.Response{
					Status: http.StatusOK,
					Body:   resp,
				}

				m.logger.Debug("sent a response to plugin's SSE channel")
			case <-timer.C:
				m.logger.Debug("plugin's timer expired, sending stop response")
				ch <- &shared.Response{
					Status: http.StatusOK,
					Body:   "stop",
				}

				m.logger.Debug("closing plugin's SSE channel", "cause", "done working")
				close(ch)
				return
			}
		}
	}()
	return ch, nil
}
