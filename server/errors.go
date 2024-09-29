package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/prior-it/apollo/core"
)

func DefaultErrorHandler(apollo *Apollo, err error) {
	apollo.Error("Server error", "error", err)
	code, msg := func() (int, string) {
		switch {
		case errors.Is(err, core.ErrUnauthenticated):
			return http.StatusUnauthorized, "unauthorized"
		case errors.Is(err, core.ErrForbidden):
			return http.StatusForbidden, "forbidden"
		case errors.Is(err, core.ErrConflict):
			return http.StatusConflict, "conflict"
		case errors.Is(err, core.ErrNotFound):
			return http.StatusNotFound, "not found"
		}
		return http.StatusInternalServerError, "internal server error"
	}()
	apollo.Writer.WriteHeader(code)
	render.PlainText(apollo.Writer, apollo.Request, msg)
}
