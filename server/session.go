package server

import (
	"fmt"
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
	session.Values[sessionJoined] = user.Joined
	apollo.User = user
	return apollo.store.Save(apollo.Request, apollo.Writer, session)
}

// SetActiveOrganisation changes the "active organisation". This might influence some other organisation-dependent requests
// such as permission checks.
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
	return apollo.store.Save(apollo.Request, apollo.Writer, session)
}

// Utility function that retrieves a full core.User object from the current session, if one exists.
// If there is no active session, this will return core.ErrUnauthenticated
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
	session.Values[sessionUserName] = ""
	session.Values[sessionUserID] = 0
	session.Values[sessionOrganisationName] = ""
	session.Values[sessionOrganisationID] = 0
	session.Values[sessionOrganisationParent] = 0
	session.Values[sessionEmail] = ""
	return session.Store().Save(apollo.Request, apollo.Writer, session)
}
