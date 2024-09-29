package server

import (
	"context"
)

func RequireLogin[state any](apollo *Apollo, _ state) (context.Context, error) {
	return apollo.Context(), apollo.RequiresLogin()
}
