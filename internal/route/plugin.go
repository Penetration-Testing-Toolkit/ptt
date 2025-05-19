package route

import (
	"bytes"
	"fmt"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/plugin"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/templates"
	"github.com/Penetration-Testing-Toolkit/ptt/shared"
	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

// RegisterPluginRoutes registers a plugin's shared.Module routes as an echo.Group.
// It registers these routes based on shared.ModuleInfo's routes.
func RegisterPluginRoutes(logger hclog.Logger, e *echo.Echo, g *app.Global, plug *plugin.ModulePlugin) error {
	info := plug.Info()
	// Create group based on the plugin's ID
	group := e.Group(plug.RootURL().String())

	for _, r := range info.Routes {
		// Setup echo.HandlerFunc
		handler := func(c echo.Context) error {
			if !r.UseSSE {
				// Regular (non-SSE) HTTP request
				// Proxy request to plugin's handler
				resp, err := plug.Module().Handle(c.Request().Context(), c.Request())
				if err != nil {
					return err
				}

				// Replace the existing response status & headers with plugin's response
				registerHelper(c, resp)

				return Layout(c, g, templates.PluginContent(resp.Body)).Render(c.Request().Context(), c.Response())
			} else {
				// Handle SSE request (https://echo.labstack.com/docs/cookbook/sse)
				logger.Info("frontend SSE client connected", "IP", c.RealIP(), "Path", c.Request().URL.Path)

				// Setup headers for SSE
				h := c.Response().Header()
				h.Set("Content-Type", "text/event-stream")
				h.Set("Cache-Control", "no-cache")
				h.Set("Connection", "keep-alive")

				// Proxy request to plugin's SSE handler
				ch, err := plug.Module().HandleSSE(c.Request().Context(), c.Request())
				if err != nil {
					return err
				}

				for {
					select {
					case <-c.Request().Context().Done():
						logger.Info("frontend SSE client disconnected", "cause", "c.Request().Context().Done()", "IP", c.RealIP(), "Path", c.Request().URL.Path)
						return nil
					case resp, ok := <-ch:
						if !ok {
							logger.Debug("SSE channel was closed")
							return nil
						}

						logger.Debug("SSE response came through server's channel", "status", resp.Status, "resp", resp.Body)

						// Replace the existing response status & headers with plugin's response
						registerHelper(c, resp)

						// Convert response's body into an SSE & write to echo.Context response
						event := sseEvent{Data: []byte(resp.Body)}
						if err := event.marshalTo(c.Response()); err != nil {
							return err
						}

						// Sent response up to the frontend client
						c.Response().Flush()
						logger.Debug("SSE response flushed")
					}
				}
			}
		}

		// Dynamically map route's method to the correct echo route type
		methods := map[string]func(string, echo.HandlerFunc, ...echo.MiddlewareFunc) *echo.Route{
			http.MethodGet:     group.GET,
			http.MethodHead:    group.HEAD,
			http.MethodPost:    group.POST,
			http.MethodPut:     group.PUT,
			http.MethodDelete:  group.DELETE,
			http.MethodConnect: group.CONNECT,
			http.MethodOptions: group.OPTIONS,
			http.MethodTrace:   group.TRACE,
			http.MethodPatch:   group.PATCH,
		}
		if echoMethod, exists := methods[r.Method]; exists {
			echoMethod(r.Path, handler)
		} else {
			logger.Error("invalid HTTP method", "method", r.Method)
		}
	}

	return nil
}

// registerHelper overwrites a shared.Response status & headers into the echo.Context response.
func registerHelper(c echo.Context, resp *shared.Response) {
	c.Response().Status = resp.Status
	for k, v := range resp.Header {
		c.Response().Header().Del(k)
		for _, i := range v {
			c.Response().Header().Add(k, i)
		}
	}
}

// sseEvent represents Server-Sent Event.
// SSE explanation: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#event_stream_format
// See: https://echo.labstack.com/docs/cookbook/sse
type sseEvent struct {
	// ID is used to set the EventSource object's last event ID value.
	ID []byte
	// Data field is for the message. When the EventSource receives multiple consecutive lines
	// that begin with data:, it concatenates them, inserting a newline character between each one.
	// Trailing newlines are removed.
	Data []byte
	// Event is a string identifying the type of event described. If this is specified, an event
	// will be dispatched on the browser to the listener for the specified event name; the website
	// source code should use addEventListener() to listen for named events. The onmessage handler
	// is called if no event name is specified for a message.
	Event []byte
	// Retry is the reconnection time. If the connection to the server is lost, the browser will
	// wait for the specified time before attempting to reconnect. This must be an integer, specifying
	// the reconnection time in milliseconds. If a non-integer value is specified, the field is ignored.
	Retry []byte
	// Comment line can be used to prevent connections from timing out; a server can send a comment
	// periodically to keep the connection alive.
	Comment []byte
}

// marshalTo marshals sseEvent to given io.Writer.
// See: https://echo.labstack.com/docs/cookbook/sse
func (ev *sseEvent) marshalTo(w io.Writer) error {
	// Marshalling part is taken from: https://github.com/r3labs/sse/blob/c6d5381ee3ca63828b321c16baa008fd6c0b4564/http.go#L16
	if len(ev.Data) == 0 && len(ev.Comment) == 0 {
		return nil
	}

	if len(ev.Data) > 0 {
		if _, err := fmt.Fprintf(w, "id: %s\n", ev.ID); err != nil {
			return err
		}

		sd := bytes.Split(ev.Data, []byte("\n"))
		for i := range sd {
			if _, err := fmt.Fprintf(w, "data: %s\n", sd[i]); err != nil {
				return err
			}
		}

		if len(ev.Event) > 0 {
			if _, err := fmt.Fprintf(w, "event: %s\n", ev.Event); err != nil {
				return err
			}
		}

		if len(ev.Retry) > 0 {
			if _, err := fmt.Fprintf(w, "retry: %s\n", ev.Retry); err != nil {
				return err
			}
		}
	}

	if len(ev.Comment) > 0 {
		if _, err := fmt.Fprintf(w, ": %s\n", ev.Comment); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(w, "\n"); err != nil {
		return err
	}

	return nil
}
