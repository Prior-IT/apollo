package oauth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"github.com/prior-it/apollo/login"
)

func getEntraIDUser(
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
		return nil, fmt.Errorf("failed to create EntraId user data request: %w", err)
	}
	userReq.Header.Set("Accept", "application/json")
	userReq.Header.Set(
		"Authorization",
		fmt.Sprintf("%s %s", data.TokenType, data.AccessToken),
	)

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve token data for EntraId: %w",
			err,
		)
	}
	defer userResp.Body.Close()

	respData := make(map[string]interface{})
	err = render.DecodeJSON(userResp.Body, &respData)
	if err != nil {
		return nil, fmt.Errorf("cannot parse entraId user response: %w", err)
	}
	slog.Debug("Received entraId user data", "data", respData)

	id, ok := respData["id"]
	if !ok || id == nil {
		return nil, fmt.Errorf("cannot parse entraId user data: %q", respData)
	}
	name := respData["displayName"]
	if name == nil {
		name = ""
	}
	email := respData["mail"]
	if email == nil {
		email = ""
	}
	return &login.UserData{
		Name:       name.(string),
		Email:      email.(string),
		Provider:   "entraid",
		ProviderID: id.(string),
	}, nil
}
