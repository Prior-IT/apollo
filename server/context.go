package server

import (
	"context"

	"github.com/gorilla/sessions"
	"github.com/prior-it/apollo/core"
)

const (
	ctxLoggedIn contextKey = iota
	ctxUserID
	ctxUserName
	ctxOrganisationID
	ctxOrganisationName
	ctxSession
	ctxDebug
	ctxIsAdmin
)

func IsLoggedIn(ctx context.Context) bool {
	loggedIn, ok := ctx.Value(ctxLoggedIn).(bool)
	return ok && loggedIn
}

func IsAdmin(ctx context.Context) bool {
	isAdmin, ok := ctx.Value(ctxIsAdmin).(bool)
	return ok && isAdmin
}

func UserID(ctx context.Context) core.UserID {
	return ctx.Value(ctxUserID).(core.UserID)
}

func UserName(ctx context.Context) string {
	return ctx.Value(ctxUserName).(string)
}

func OrganisationID(ctx context.Context) core.OrganisationID {
	return ctx.Value(ctxOrganisationID).(core.OrganisationID)
}

func OrganisationName(ctx context.Context) string {
	return ctx.Value(ctxOrganisationName).(string)
}

// Session provides access to the current user's session.
// Applications can use this to attach or retrieve custom data from this session.
// Make sure to prefix all custom keys with "app-" so they won't interfere with the Apollo session context.
func Session(ctx context.Context) *sessions.Session {
	return ctx.Value(ctxSession).(*sessions.Session)
}

func Debug(ctx context.Context) bool {
	return ctx.Value(ctxDebug).(bool)
}
