package oauth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"github.com/prior-it/apollo/login"
)

func getGithubUser(
	ctx context.Context,
	data accessToken,
	url string,
) (*login.UserData, error) {
	userReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create github user data request: %w", err)
	}
	userReq.Header.Set("User-Agent", "Prior-IT Login")
	userReq.Header.Set("Accept", "application/json")
	userReq.Header.Set(
		"Authorization",
		fmt.Sprintf("%s %s", data.TokenType, data.AccessToken),
	)

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve token data for GitHub: %w",
			err,
		)
	}
	defer userResp.Body.Close()

	respData := make(map[string]interface{})
	err = render.DecodeJSON(userResp.Body, &respData)
	if err != nil {
		return nil, fmt.Errorf("cannot parse github user response: %w", err)
	}
	slog.Debug("Received github user data", "data", respData)

	id, ok := respData["id"]
	if !ok || id == nil {
		return nil, fmt.Errorf("cannot parse github user data: %q", respData)
	}
	switch id.(type) {
	// Why do these IDs get detected as f64?
	case float64:
		id = fmt.Sprintf("%.0f", id)
	}

	name := respData["name"]
	if name == nil {
		name = ""
	}
	email := respData["email"]
	if email == nil {
		email = ""
	}
	return &login.UserData{
		Name:       name.(string),
		Email:      email.(string),
		Provider:   "github",
		ProviderID: id.(string),
	}, nil
}
