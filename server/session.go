package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prior-it/apollo/core"
)

type contextKey uint

const cookieUser = "apollo-user"

const (
	sessionLoggedIn = "apollo-logged-in"
	sessionIsAdmin  = "apollo-user-admin"
	sessionUserName = "apollo-user-name"
	sessionEmail    = "apollo-user-email"
	sessionJoined   = "apollo-user-joined"
	sessionUserID   = "apollo-user-id"
)

// Login will log in with the specified user.
func (apollo *Apollo) Login(user *core.User) error {
	if user == nil {
		panic("you cannot log in with a nil user")
	}
	if apollo.store == nil {
		panic("you need to specify a session store before logging in")
	}
	session := Session(apollo.Context())
	if apollo.IsDebug {
		session.Options.Secure = true
		session.Options.HttpOnly = true
		session.Options.SameSite = http.SameSiteStrictMode
	} else {
		session.Options.Secure = false
		session.Options.HttpOnly = false
		session.Options.SameSite = http.SameSiteLaxMode
	}
	session.Values[sessionLoggedIn] = true
	session.Values[sessionIsAdmin] = user.Admin
	session.Values[sessionUserName] = user.Name
	session.Values[sessionEmail] = user.Email.String()
	session.Values[sessionUserID] = user.ID
	session.Values[sessionJoined] = user.Joined
	apollo.User = user
	return apollo.store.Save(apollo.Request, apollo.Writer, session)
}

// Utility function that retrieves a full core.User object from the current session, if one exists.
// If there is no active session, this will return core.ErrUnauthenticated
func (apollo *Apollo) retrieveUser() (*core.User, error) {
	session := Session(apollo.Context())

	loggedIn, ok := session.Values[sessionLoggedIn].(bool)

	// No user data in session
	if !ok || !loggedIn {
		return nil, core.ErrUnauthenticated
	}

	ID, ok := session.Values[sessionUserID].(core.UserID)
	if !ok {
		return nil, fmt.Errorf(
			"invalid user id stored in session: %v",
			session.Values[sessionUserID],
		)
	}

	IsAdmin, ok := session.Values[sessionIsAdmin].(bool)
	if session.Values[sessionIsAdmin] == nil {
		IsAdmin = false
	} else if !ok {
		return nil, fmt.Errorf(
			"invalid user is admin stored in session: %v",
			session.Values[sessionIsAdmin],
		)
	}

	Name, ok := session.Values[sessionUserName].(string)
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
	Email, err := core.NewEmailAddress(emailStr)
	if err != nil {
		return nil, fmt.Errorf("session e-mail address invalid: %w", err)
	}

	Joined, ok := session.Values[sessionJoined].(time.Time)
	if !ok {
		return nil, fmt.Errorf(
			"invalid joined time stored in session: %v",
			session.Values[sessionJoined],
		)
	}

	return &core.User{
		ID:     ID,
		Name:   Name,
		Email:  Email,
		Admin:  IsAdmin,
		Joined: Joined,
	}, nil
}

// Logout will log the current user out.
func (apollo *Apollo) Logout() error {
	session := Session(apollo.Context())
	session.Values[sessionLoggedIn] = false
	session.Values[sessionIsAdmin] = false
	session.Values[sessionUserName] = ""
	session.Values[sessionUserID] = 0
	session.Values[sessionEmail] = ""
	return session.Store().Save(apollo.Request, apollo.Writer, session)
}
