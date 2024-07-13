package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
)

type Apollo struct {
	Writer  http.ResponseWriter
	Request *http.Request
	logger  *slog.Logger
	layout  templ.Component
	IsDebug bool
	UseSSL  bool
}

func (apollo *Apollo) StatusCode(code int) {
	apollo.Writer.WriteHeader(code)
}

// Log the specified error message. args is a list of structured fields to add to the error message.
// The arguments should alternate between a field's name (string) and its value (any).
// This behaves the same as [log/slog.Error]
//
// # Example
//
//	server.Error("Something went wrong", "error", err, "user", user)
func (apollo *Apollo) Error(msg string, args ...any) {
	apollo.logger.Error(msg, args...)
}

// Log the specified debug message. args is a list of structured fields to add to the message.
// The arguments should alternate between a field's name (string) and its value (any).
// This behaves the same as [log/slog.Debug]
//
// # Example
//
//	server.Debug("New user registered", "user", user, "id", id)
func (apollo *Apollo) Debug(msg string, args ...any) {
	apollo.logger.Debug(msg, args...)
}

// LogField will add the specified field and its value to the current request's span
func (apollo *Apollo) LogString(field string, value string) {
	apollo.LogField(field, slog.StringValue(value))
}

// LogField will add the specified field and its value to the current request's span
func (apollo *Apollo) LogField(field string, value slog.Value) {
	httplog.LogEntrySetField(apollo.Context(), field, value)
}

// Context returns the request's context.
//
// The returned context is always non-nil; it defaults to the
// background context.
//
// The context is canceled when the
// client's connection closes, the request is canceled (with HTTP/2),
// or when the ServeHTTP method returns.
func (apollo *Apollo) Context() context.Context {
	return apollo.Request.Context()
}

// Host specifies the host on which the URL is sought.
// For HTTP/1 (per RFC 7230, section 5.4), this is either the value of the "Host" header or the host name
// given in the URL itself. For HTTP/2, it is the value of the ":authority" pseudo-header field.
// It may be of the form "host:port". For international domain names, Host may be in Punycode or Unicode form. Use
// golang.org/x/net/idna to convert it to either format if needed.
// To prevent DNS rebinding attacks, server Handlers should validate that the Host header has a value for which the
// Handler considers itself authoritative. The included ServeMux supports patterns registered to particular host
// names and thus protects its registered Handlers.
func (apollo *Apollo) Host() string {
	return apollo.Request.Host
}

// Path returns the path of the request.
func (apollo *Apollo) Path() string {
	return apollo.Request.URL.Path
}

// PathValue returns the value for the named path wildcard in the router pattern
// that matched the request.
// It returns the empty string if the request was not matched against a pattern
// or there is no such wildcard in the pattern.
func (apollo *Apollo) PathValue(key string) string {
	return apollo.Request.PathValue(key)
}

// ParseBody parses the request body into an interface using the form decoder.
// @TODO: Allow other decoders as well, it should be possible to get the correct
// decoder from the request headers.
//
// # Example:
//
//	var data SomeStruct
//	if err := server.ParseBody(&data); err != nil {
//		return fmt.Errorf("cannot parse body: %w", err)
//	}
func (apollo *Apollo) ParseBody(v interface{}) error {
	return render.DecodeForm(apollo.Request.Body, v)
}

// GetQuery returns the first value associated with the given query parameter in the request url.
// If there are no values set for the query param, this returns the empty string.
// This silently discards malformed value pairs. To check query errors use [Request().ParseQuery].
func (apollo *Apollo) GetQuery(param string) string {
	return apollo.Request.URL.Query().Get(param)
}

// GetHeader returns the first value associated with the given header in the request.
// If there are no values set for the header, this returns the empty string.
// Both the header and its value are case-insensitive.
func (apollo *Apollo) GetHeader(header string) string {
	return apollo.Request.Header.Get(header)
}

// AddHeader adds the header, value pair to the response header. It appends to any existing values associated with key.
// Both the header and its value are case-insensitive.
func (apollo *Apollo) AddHeader(header string, value string) {
	apollo.Writer.Header().Add(header, value)
}

// Protocol returns the currently used protocol (either "http://" or "https://")
func (apollo *Apollo) Protocol() string {
	if apollo.UseSSL {
		return "https://"
	}
	return "http://"
}

// CreateURL will return the url for the given endpoint. If you need to include the current protocol as well, use [CreateProtocolURL] instead.
func (apollo *Apollo) CreateURL(endpoint string) string {
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	return fmt.Sprintf("%s%s", apollo.Request.Host, endpoint)
}

// CreateProtocolURL will return the full url for the given endpoint, including its protocol. If you don't want the current protocol to be included, use [CreateURL] instead.
func (apollo *Apollo) CreateProtocolURL(endpoint string) string {
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	return fmt.Sprintf("%s%s%s", apollo.Protocol(), apollo.Request.Host, endpoint)
}

// Redirect will return a response that redirects the user to the specified url.
// If HTMX is available, this will redirect using HTMX.
func (apollo *Apollo) Redirect(url string) {
	if apollo.GetHeader("HX-Request") == "true" {
		apollo.AddHeader("HX-Redirect", url)
		apollo.StatusCode(http.StatusOK)
	} else {
		apollo.AddHeader("Location", url)
		apollo.StatusCode(http.StatusSeeOther)
	}
}

// RenderComponent renders the specified component in the response body.
// You can render multiple components and they will all be returned by the response,
// this can be used to perform out-of-band swaps with HTMX, for example.
//
// # Example:
//
//	if err := server.RenderComponent(components.OOBNotification(data)); err != nil {
//		return err
//	}
//	return server.RenderComponent(components.ActualResponse(otherData))
func (apollo *Apollo) RenderComponent(
	component templ.Component,
) error {
	ctx := apollo.Context()
	return component.Render(ctx, apollo.Writer)
}

// RenderPage renders the specified page in the response body.
// If the request was made using HTMX, it will simply return the pages contents.
// If the request was made without HTMX (e.g. by refreshing), it will return the page
// surrounded with the current default lay-out. You can change the default lay-out by
// calling `SetDefaultLayout` on the router during configuration.
// RenderPage can be combined with RenderComponent to perform out-of-band swaps in a
// single response.
//
// # Example:
//
//	if err := server.RenderComponent(components.OOBNotification(data)); err != nil {
//		return err
//	}
//	return server.RenderPage(pages.ActualResponse(otherData))
func (apollo *Apollo) RenderPage(
	page templ.Component,
) error {
	ctx := apollo.Context()
	comp := page
	if apollo.GetHeader("hx-request") != "true" {
		ctx = templ.WithChildren(ctx, page)
		comp = apollo.layout
	}
	return comp.Render(ctx, apollo.Writer)
}
