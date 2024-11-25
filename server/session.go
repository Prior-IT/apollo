package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/prior-it/apollo/config"
	"github.com/prior-it/apollo/core"
)

const (
	cookieUser = "apollo-user"
	cookieCSRF = "apollo-csrf"
)

const (
	// User cookie
	sessionLoggedIn           = "apollo-logged-in"
	sessionIsAdmin            = "apollo-user-admin"
	sessionUserName           = "apollo-user-name"
	sessionEmail              = "apollo-user-email"
	sessionLanguage           = "apollo-user-lang"
	sessionJoined             = "apollo-user-joined"
	sessionUserID             = "apollo-user-id"
	sessionOrganisationID     = "apollo-organisation-id"
	sessionOrganisationName   = "apollo-organisation-name"
	sessionOrganisationParent = "apollo-organisation-parent"

	// CSRF cookie
	sessionCSRFToken = "token"
)

func (apollo *Apollo) Session() *sessions.Session {
	session := Session(apollo.Context())
	return configureCookie(apollo.Cfg, session)
}

func configureCookie(cfg *config.Config, session *sessions.Session) *sessions.Session {
	if cfg.IsTest() {
		session.Options.Secure = false
		session.Options.HttpOnly = false
		session.Options.SameSite = http.SameSiteNoneMode
	} else if cfg.App.Debug {
		session.Options.Secure = true
		session.Options.HttpOnly = true
		session.Options.SameSite = http.SameSiteNoneMode
	} else { // production
		session.Options.Secure = true
		session.Options.HttpOnly = true
		session.Options.SameSite = http.SameSiteLaxMode
	}
	return session
}

func buildSessionContext(ctx context.Context, session *sessions.Session) context.Context {
	loggedIn, ok := session.Values[sessionLoggedIn].(bool)
	ctx = context.WithValue(ctx, ctxLoggedIn, ok && loggedIn)

	isAdmin, ok := session.Values[sessionIsAdmin].(bool)
	ctx = context.WithValue(ctx, ctxIsAdmin, ok && isAdmin)

	userName, ok := session.Values[sessionUserName].(string)
	if ok {
		ctx = context.WithValue(ctx, ctxUserName, userName)
	}

	userID, ok := session.Values[sessionUserID].(core.UserID)
	if ok {
		ctx = context.WithValue(ctx, ctxUserID, userID)
	}

	organisationID, ok := session.Values[sessionOrganisationID].(core.OrganisationID)
	if ok {
		ctx = context.WithValue(ctx, ctxOrganisationID, organisationID)
	}

	organisationName, ok := session.Values[sessionOrganisationName].(string)
	if ok {
		ctx = context.WithValue(ctx, ctxOrganisationName, organisationName)
	}

	organisationParent, ok := session.Values[sessionOrganisationParent].(core.OrganisationID)
	if ok {
		ctx = context.WithValue(ctx, ctxOrganisationParent, organisationParent)
	}

	return ctx
}

// Login will log in with the specified user.
func (apollo *Apollo) Login(user *core.User) error {
	if user == nil {
		panic("you cannot log in with a nil user")
	}
	if apollo.store == nil {
		panic("you need to specify a session store before logging in")
	}
	session := apollo.Session()
	session.Values[sessionLoggedIn] = true
	session.Values[sessionIsAdmin] = user.Admin
	session.Values[sessionUserName] = user.Name
	session.Values[sessionEmail] = user.Email.String()
	session.Values[sessionUserID] = user.ID
	session.Values[sessionLanguage] = user.Lang
	session.Values[sessionJoined] = user.Joined
	apollo.User = user
	err := apollo.store.Save(apollo.Request, apollo.Writer, session)
	if err != nil {
		return err
	}
	apollo.LogField("active_user_id", slog.AnyValue(apollo.User.ID))
	apollo.rebuildContext()
	return nil
}

// SetActiveOrganisation changes the "active organisation". This might influence some other organisation-dependent requests
// such as permission checks.
// Beware: this might update apollo.Context(), reretrieve that if you were already using it.
func (apollo *Apollo) SetActiveOrganisation(organisation *core.Organisation) error {
	if apollo.store == nil {
		panic("you need to specify a session store before setting an active organisation")
	}
	session := apollo.Session()
	if organisation != nil {
		session.Values[sessionOrganisationID] = organisation.ID
		session.Values[sessionOrganisationName] = organisation.Name
		if organisation.ParentID != nil {
			session.Values[sessionOrganisationParent] = *organisation.ParentID
		} else {
			session.Values[sessionOrganisationParent] = nil
		}
	} else {
		session.Values[sessionOrganisationID] = nil
		session.Values[sessionOrganisationName] = nil
		session.Values[sessionOrganisationParent] = nil
	}
	apollo.Organisation = organisation
	err := apollo.store.Save(apollo.Request, apollo.Writer, session)
	if err != nil {
		return err
	}
	apollo.LogField("active_organisation_id", slog.AnyValue(apollo.Organisation.ID))
	apollo.rebuildContext()
	return nil
}

// Utility function that retrieves a full core.User object from the current session, if one exists.
// If there is no active session, this will return core.ErrUnauthenticated
//
//nolint:cyclop
func (apollo *Apollo) retrieveUser() (*core.User, error) {
	session := apollo.Session()

	loggedIn, ok := session.Values[sessionLoggedIn].(bool)

	// No user data in session
	if !ok || !loggedIn {
		return nil, core.ErrUnauthenticated
	}

	id, ok := session.Values[sessionUserID].(core.UserID)
	if !ok {
		return nil, fmt.Errorf(
			"invalid user id stored in session: %v",
			session.Values[sessionUserID],
		)
	}

	isAdmin, ok := session.Values[sessionIsAdmin].(bool)
	if session.Values[sessionIsAdmin] == nil {
		isAdmin = false
	} else if !ok {
		return nil, fmt.Errorf(
			"invalid user is admin stored in session: %v",
			session.Values[sessionIsAdmin],
		)
	}

	name, ok := session.Values[sessionUserName].(string)
	if !ok {
		return nil, fmt.Errorf(
			"invalid user name stored in session: %v",
			session.Values[sessionUserName],
		)
	}

	emailStr, ok := session.Values[sessionEmail].(string)
	if !ok {
		return nil, fmt.Errorf(
			"invalid email string stored in session: %v",
			session.Values[sessionEmail],
		)
	}
	email, err := core.NewEmailAddress(emailStr)
	if err != nil {
		return nil, fmt.Errorf("session e-mail address invalid: %w", err)
	}

	lang, ok := session.Values[sessionLanguage].(string)
	if !ok {
		return nil, fmt.Errorf(
			"invalid language stored in session: %v",
			session.Values[sessionLanguage],
		)
	}

	joined, ok := session.Values[sessionJoined].(time.Time)
	if !ok {
		return nil, fmt.Errorf(
			"invalid joined time stored in session: %v",
			session.Values[sessionJoined],
		)
	}

	return &core.User{
		ID:     id,
		Name:   name,
		Email:  email,
		Admin:  isAdmin,
		Lang:   lang,
		Joined: joined,
	}, nil
}

// Utility function that retrieves a full core.Organisation object from the current session, if one exists.
// If there is no active session, this will return core.ErrUnauthenticated
// If the active session does not have an "active organisation", this will return core.ErrNoActiveOrganisation
func (apollo *Apollo) retrieveOrganisation() (*core.Organisation, error) {
	session := apollo.Session()

	loggedIn, ok := session.Values[sessionLoggedIn].(bool)

	// No user data in session
	if !ok || !loggedIn {
		return nil, core.ErrUnauthenticated
	}

	if session.Values[sessionOrganisationID] == nil || session.Values[sessionOrganisationID] == 0 ||
		session.Values[sessionOrganisationName] == nil {
		return nil, core.ErrNoActiveOrganisation
	}

	id, ok := session.Values[sessionOrganisationID].(core.OrganisationID)
	if !ok {
		return nil, fmt.Errorf(
			"invalid organisation id stored in session: %v",
			session.Values[sessionOrganisationID],
		)
	}

	name, ok := session.Values[sessionOrganisationName].(string)
	if !ok {
		return nil, fmt.Errorf(
			"invalid organisation name stored in session: %v",
			session.Values[sessionOrganisationName],
		)
	}

	var parentID *core.OrganisationID
	if session.Values[sessionOrganisationParent] != nil {
		pID, ok := session.Values[sessionOrganisationParent].(core.OrganisationID)
		if !ok {
			return nil, fmt.Errorf(
				"invalid organisation parent id stored in session: %v",
				session.Values[sessionOrganisationParent],
			)
		}
		parentID = &pID
	}

	return &core.Organisation{
		ID:       id,
		Name:     name,
		ParentID: parentID,
	}, nil
}

// Logout will log the current user out.
func (apollo *Apollo) Logout() error {
	session := apollo.Session()
	session.Values[sessionLoggedIn] = false
	session.Values[sessionIsAdmin] = false
	session.Values[sessionUserName] = nil
	session.Values[sessionLanguage] = nil
	session.Values[sessionUserID] = nil
	session.Values[sessionOrganisationName] = nil
	session.Values[sessionOrganisationID] = nil
	session.Values[sessionOrganisationParent] = nil
	session.Values[sessionEmail] = nil
	return session.Store().Save(apollo.Request, apollo.Writer, session)
}
