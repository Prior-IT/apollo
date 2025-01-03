package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-chi/httplog/v2"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/prior-it/apollo/config"
	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/permissions"
)

const CsrfName = "_csrf_token"

var (
	ErrCSRFNotFound = errors.New("CSRF token not found")
	ErrCSRFFail     = errors.New("CSRF token does not match")
)

type Apollo struct {
	Writer       http.ResponseWriter
	Request      *http.Request
	logger       *slog.Logger
	layout       templ.Component
	User         *core.User
	Cfg          *config.Config
	Organisation *core.Organisation
	permissions  permissions.Service
	store        sessions.Store
	ctx          context.Context
	decoder      *schema.Decoder
}

// Populate populates the Apollo object with fields that need to be retrieved after initialisation.
// E.g. fields that are stored in the active session.
func (apollo *Apollo) populate() {
	apollo.ctx = apollo.Request.Context()

	if apollo.store != nil {
		apollo.populateUser()
		apollo.populateOrganisation()
	} else {
		slog.Warn("No session store provided, it will not be possible to log in")
	}

	if apollo.permissions != nil {
		err := permissions.RegisterApolloPermissions(apollo.permissions)
		if err != nil {
			slog.Error("Could not register Apollo permissions", "error", err)
		}
	} else {
		slog.Warn("No permissions service provided, all permission checks will fail")
	}
}

func (apollo *Apollo) populateUser() {
	user, err := apollo.retrieveUser()
	if errors.Is(err, core.ErrUnauthenticated) {
		apollo.User = nil
		apollo.LogField("active_user_id", slog.AnyValue(nil))
	} else if err != nil {
		slog.Error("Could not retrieve user object from session", "error", err)
		err = apollo.Logout()
		if err != nil {
			slog.Error("Could not log out of the invalid session", "error", err)
		}
	} else {
		apollo.User = user
		apollo.LogField("active_user_id", slog.AnyValue(apollo.User.ID))
	}
}

func (apollo *Apollo) populateOrganisation() {
	organisation, err := apollo.retrieveOrganisation()
	if errors.Is(err, core.ErrUnauthenticated) || errors.Is(err, core.ErrNoActiveOrganisation) {
		apollo.Organisation = nil
		apollo.LogField("active_organisation_id", slog.AnyValue(nil))
	} else if err != nil {
		slog.Error("Could not retrieve organisation object from session", "error", err)
		err = apollo.Logout()
		if err != nil {
			slog.Error("Could not log out of the invalid session", "error", err)
		}
	} else {
		apollo.Organisation = organisation
		apollo.LogField("active_organisation_id", slog.AnyValue(apollo.Organisation.ID))
	}
}

func InjectApollo[state any](apollo *Apollo, _ state) (context.Context, error) {
	ctx := apollo.Context()
	ctx = context.WithValue(ctx, ctxApollo, apollo)
	return ctx, nil
}

func Ap(ctx context.Context) *Apollo {
	apollo, ok := ctx.Value(ctxApollo).(*Apollo)
	if !ok {
		panic("You need to inject the apollo object in the context before you can retrieve it")
	}
	return apollo
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
//
// # Example
//
// apollo.LogField("user_id", slog.IntValue(user.id)
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
	return apollo.ctx
}

// RebuildContext rebuilds the apollo context based on the current session data.
func (apollo *Apollo) rebuildContext() {
	ctx := apollo.ctx
	session := apollo.Session()
	ctx = buildSessionContext(ctx, session)
	apollo.ctx = ctx
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

// Path returns the full path of the request.
func (apollo *Apollo) Path() string {
	return apollo.Request.URL.Path
}

// GetPath returns the value for the named path wildcard in the router pattern
// that matched the request.
// It returns the empty string if the request was not matched against a pattern
// or there is no such wildcard in the pattern.
//
// E.g.: A route defined as `/users/{id}` can call `GetPath("id")` to return the
// value for "id" in the current path.
func (apollo *Apollo) GetPath(key string) string {
	return apollo.Request.PathValue(key)
}

// ParseBody parses the request body into an interface using the form decoder.
// @TODO: Allow other decoders as well, it should be possible to get the correct
// decoder from the request headers.
// @TODO: This decoder has an "IgnoreUnknownKeys" option, do we want to expose that as well?
//
// # Example:
//
//	var data SomeStruct
//	if err := server.ParseBody(&data); err != nil {
//		return fmt.Errorf("cannot parse body: %w", err)
//	}
func (apollo *Apollo) ParseBody(v interface{}) error {
	if apollo.decoder == nil {
		apollo.decoder = schema.NewDecoder()
	}
	err := apollo.Request.ParseForm()
	if err != nil {
		return fmt.Errorf("cannot parse form: %w", err)
	}
	return apollo.decoder.Decode(v, apollo.Request.PostForm)
}

// FormValue returns the first value for the named component of the POST, PUT, or PATCH request body.
// URL query parameters are ignored. This ignores any errors, if the key is not present, FormValue returns the empty string.
func (apollo *Apollo) FormValue(key string) string {
	return apollo.Request.PostFormValue(key)
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
	if apollo.Cfg.App.SSL {
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

// RequiresLogin will return core.ErrUnauthenticated if there is no user logged in and nil otherwise.
func (apollo *Apollo) RequiresLogin() error {
	if apollo.User == nil {
		return core.ErrUnauthenticated
	}
	return nil
}

// Requires will return core.ErrForbidden if the current user does not have the specified permission and nil otherwise.
// If no user is logged in at all, this will return core.ErrUnauthenticated.
func (apollo *Apollo) Requires(permission permissions.Permission) error {
	if err := apollo.RequiresLogin(); err != nil {
		return err
	}
	if !apollo.Has(permission) {
		return core.ErrForbidden
	}
	return nil
}

// RequiresStrict will return core.ErrForbidden if the current user does not have the specified permission and nil otherwise.
// If no user is logged in at all, this will return core.ErrUnauthenticated.
// This will use the strict permission check, which ignores global permissions and only checks the active organisation.
func (apollo *Apollo) RequiresStrict(permission permissions.Permission) error {
	if err := apollo.RequiresLogin(); err != nil {
		return err
	}
	if !apollo.HasStrict(permission) {
		return core.ErrForbidden
	}
	return nil
}

// RequiresInOrganisation will return core.ErrForbidden if the current user does not have the specified permission, either globally, or within
// the lineage of the specified organisation and nil otherwise.
// If no user is logged in at all, this will return core.ErrUnauthenticated.
func (apollo *Apollo) RequiresInOrganisation(
	permission permissions.Permission,
	organisation core.OrganisationID,
) error {
	if err := apollo.RequiresLogin(); err != nil {
		return err
	}
	if !apollo.HasInOrganisation(permission, organisation) {
		return core.ErrForbidden
	}
	return nil
}

// RequiresInOrganisationStrict will return core.ErrForbidden if the current user does not have the specified permission and nil otherwise.
// If no user is logged in at all, this will return core.ErrUnauthenticated.
// This will use the strict permission check, which ignores global permissions and only checks the specified organisation, not its lineage.
func (apollo *Apollo) RequiresInOrganisationStrict(
	permission permissions.Permission,
	organisation core.OrganisationID,
) error {
	if err := apollo.RequiresLogin(); err != nil {
		return err
	}
	if !apollo.HasInOrganisationStrict(permission, organisation) {
		return core.ErrForbidden
	}
	return nil
}

// Has returns a boolean indicating whether or not the currently logged in user has the specified permission in any
// of their permission groups or not. If no user is logged in, this will return false.
// If there is an active organisation set, this will recursively check the permissions in that organisation's lineage.
func (apollo *Apollo) Has(permission permissions.Permission) bool {
	if apollo.permissions == nil {
		slog.Warn(
			"Trying to use permission system while Apollo does not have access to a permissions.Service!",
		)
		return false
	}
	if apollo.User == nil {
		slog.Warn(
			"Trying to use permission system while no user is logged in!",
		)
		return false
	}
	if apollo.User.Admin {
		return true
	}
	ok, err := apollo.permissions.HasAny(apollo.Context(), apollo.User.ID, permission)
	if err != nil {
		slog.Error("Error while checking global permissions", "error", err)
		return false
	}
	if !ok && apollo.Organisation != nil {
		// User does not have the global permission -> check active organisation
		return apollo.HasInOrganisation(permission, apollo.Organisation.ID)
	}
	return ok
}

// HasStrict returns a boolean indicating whether or not the currently logged in user has the specified permission in
// the currently active organisation (and only there, global permissions and parent organisations are ignored).
// If no user is logged in or they haven't chosen an active organisation yet, this will always return false.
func (apollo *Apollo) HasStrict(permission permissions.Permission) bool {
	if apollo.Organisation == nil {
		return false
	}

	return apollo.HasInOrganisationStrict(permission, apollo.Organisation.ID)
}

// HasInOrganisation returns a boolean indicating whether or not the currently logged in user has the specified permission within the
// specified organisation or not. If no user is logged in, this will always return false.
// This will recursively check the permissions in the specified organisation's lineage.
func (apollo *Apollo) HasInOrganisation(
	permission permissions.Permission,
	organisation core.OrganisationID,
) bool {
	if apollo.permissions == nil {
		slog.Warn(
			"Trying to use permission system while Apollo does not have access to a permissions.Service!",
		)
		return false
	}
	if apollo.User == nil {
		slog.Warn(
			"Trying to use permission system while no user is logged in!",
		)
		return false
	}
	if apollo.User.Admin {
		return true
	}
	ok, err := apollo.permissions.HasAnyForOrgTree(
		apollo.Context(),
		apollo.User.ID,
		organisation,
		permission,
	)
	if err != nil {
		slog.Error("Error while checking organisation permissions", "error", err,
			"organisation_id",
			organisation,
		)
		return false
	}
	return ok
}

// HasInOrganisationStrict returns a boolean indicating whether or not the currently logged in user has the specified permission within the
// specified organisation (and only there, global permissions and parent organisations are ignored). If no user is logged in, this will always return false.
func (apollo *Apollo) HasInOrganisationStrict(
	permission permissions.Permission,
	organisation core.OrganisationID,
) bool {
	if apollo.permissions == nil {
		slog.Warn(
			"Trying to use permission system while Apollo does not have access to a permissions.Service!",
		)
		return false
	}
	if apollo.User == nil {
		slog.Warn(
			"Trying to use permission system while no user is logged in!",
		)
		return false
	}
	if apollo.User.Admin {
		return true
	}
	ok, err := apollo.permissions.HasAnyForOrg(
		apollo.Context(),
		apollo.User.ID,
		organisation,
		permission,
	)
	if err != nil {
		slog.Error(
			"Error while checking organisation permissions",
			"error",
			err,
			"organisation_id",
			organisation,
		)
		return false
	}
	return ok
}

// CheckCSRF will check if a CSRF token was added to the requests form body and if that token matches the
// token specified in the CSRF cookie. If either of these are false, this will return an error. If the correct CSRF
// token was specified, this will return nil.
// Note that you can only check the CSRF token once, the token will be discarded if the check passes.
// Also note that currently only POST requests using form values are supported.
// GET requests cannot use CSRF since that's less safe.
// JSON API requests can also not use CSRF at this moment, but that might change in the future using a CSRF header, if
// that need would ever arise.
func (apollo *Apollo) CheckCSRF() error {
	sessionToken, ok := oldCSRF(apollo.Context())

	// No (valid) token in session
	if !ok || len(sessionToken) == 0 {
		return ErrCSRFNotFound
	}

	token := apollo.FormValue(CsrfName)

	if token != sessionToken {
		return ErrCSRFFail
	}

	// Remove the csrf token field from the form data
	apollo.Request.PostForm.Del(CsrfName)

	return nil
}

// Redirect will return a response that redirects the user to the specified url.
// If HTMX is available, this will redirect using HTMX.
func (apollo *Apollo) Redirect(url string) {
	apollo.RedirectWithCode(url, http.StatusSeeOther)
}

// Redirect will return a response that redirects the user to the specified url with the specified status code.
// If HTMX is available, this will redirect using HTMX with status code 200 (so that HTMX would be triggered).
func (apollo *Apollo) RedirectWithCode(url string, statusCode int) {
	if apollo.GetHeader("HX-Request") == "true" {
		apollo.AddHeader("HX-Redirect", url)
		apollo.StatusCode(http.StatusOK)
	} else {
		apollo.AddHeader("Location", url)
		apollo.StatusCode(statusCode)
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

// RenderOptions specifies options you can use when rendering a page with the Apollo object.
type RenderOptions struct {
	// If CustomLayout is set, the page will render with the specified layout instead of the default one
	CustomLayout templ.Component
}

// RenderPage renders the specified page in the response body.
// If the request was made using HTMX, it will simply return the pages contents.
// If the request was made without HTMX (e.g. by refreshing), it will return the page
// surrounded with the current default lay-out, unless a custom layout was specified in the render options.
// You can change the default lay-out by calling `SetDefaultLayout` on the router during configuration.
// RenderPage can be combined with RenderComponent to perform out-of-band swaps in a
// single response.
//
// # Example:
//
//	if err := server.RenderComponent(components.OOBNotification(data)); err != nil {
//		return err
//	}
//	return server.RenderPage(pages.ActualResponse(otherData), nil)
func (apollo *Apollo) RenderPage(
	page templ.Component,
	renderOptions *RenderOptions,
) error {
	if renderOptions == nil {
		renderOptions = &RenderOptions{} // leave everything default
	}
	ctx := apollo.Context()
	comp := page
	// If we're requesting with htmx, do not rerender the layout
	if apollo.GetHeader("hx-request") != "true" {
		ctx = templ.WithChildren(ctx, page)
		if renderOptions.CustomLayout != nil {
			comp = renderOptions.CustomLayout
		} else {
			comp = apollo.layout
		}
	}
	return comp.Render(ctx, apollo.Writer)
}
