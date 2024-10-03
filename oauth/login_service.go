package oauth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/render"
	"github.com/prior-it/apollo/config"
	"github.com/prior-it/apollo/login"
)

func NewLoginService(
	providers map[string]config.OauthProviderConfig,
) *LoginService {
	return &LoginService{providers}
}

// Postgres implementation of the core LoginService interface.
type LoginService struct {
	providers map[string]config.OauthProviderConfig
}

// Force struct to implement the core interface
var _ login.Service = &LoginService{}

type accessToken struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func (s *LoginService) GetLoginRedirectURL(provider string, callbackURL string) (string, error) {
	config, exists := s.providers[provider]

	if !exists {
		return "", fmt.Errorf("unknown provider: %s", provider)
	}

	data := url.Values{}
	data.Set("client_id", config.ID)
	data.Set("redirect_uri", callbackURL)
	data.Set("scope", strings.Join(config.Scope, ","))
	data.Set("response_type", "code")
	url := fmt.Sprintf(
		"%s?%s",
		config.AuthURL,
		data.Encode(),
	)

	return url, nil
}

func (s *LoginService) LoginCallback(
	ctx context.Context,
	provider string,
	code string,
	redirectURL string,
) (*login.UserData, error) {
	slog.Debug("Login callback received", "provider", provider, "code", code)

	config, exists := s.providers[provider]

	if !exists {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	if len(code) == 0 {
		return nil, errors.New("expected to receive a code")
	}

	reqData := url.Values{}
	reqData.Set("client_id", config.ID)
	reqData.Set("client_secret", config.Secret)
	reqData.Set("code", code)
	reqData.Set("grant_type", "authorization_code")
	reqData.Set("redirect_uri", redirectURL)

	bodyReader := strings.NewReader(reqData.Encode())
	tokenReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		config.TokenURL,
		bodyReader,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	tokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Add("Accept", "application/json")

	tokenResp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve token data for provider %q: %w",
			provider,
			err,
		)
	}
	defer tokenResp.Body.Close()

	tokenData := accessToken{}
	err = render.DecodeJSON(tokenResp.Body, &tokenData)
	if err != nil {
		return nil, fmt.Errorf("cannot decode response for provider %q: %w", provider, err)
	}

	slog.Debug("Token request complete", "response", tokenData)

	var userData *login.UserData

	switch provider {
	case "github":
		userData, err = getGithubUser(ctx, tokenData, config.UserURL)
		if err != nil {
			return nil, err
		}
	case "entraid":
		userData, err = getEntraIDUser(ctx, tokenData, config.UserURL)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf(
			"retrieving user data is currently not supported for provider %q",
			provider,
		)
	}

	slog.Debug("User data retrieved", "data", userData)

	return userData, nil
}

// @TEMP: Zoho login Rust code
//
//         let jwt = token.id_token.expect("Zoho login should always send an id_token");
//         let payload_start = jwt
//             .find('.')
//             .context("Cannot find payload start in the Zoho id_token")?;
//         let jwt_end: String = jwt.chars().skip(payload_start + 1).collect();
//         let payload_len = jwt_end
//             .find('.')
//             .context("Cannot find payload end in the Zoho id_token")?;
//         let mut payload: String = jwt.chars().skip(payload_start + 1).take(payload_len).collect();
//         trace!(jwt, payload, "Zoho id_token payload extracted");
//         let padding_length = 4 - payload.len() % 4;
//         payload.push_str("=".repeat(padding_length).as_str());
//         let binary_token = base64::engine::general_purpose::URL_SAFE
//             .decode(payload)
//             .context("Cannot base64-decode the Zoho id_token")?;
//         let token = String::from_utf8(binary_token).context("Cannot decode the binary Zoho id_token")?;
//         trace!(token, "zoho id_token found");
//         let claims: ZohoIdClaims =
//             serde_json::from_str(token.as_str()).context("Cannot parse Zoho id_token")?;
//         trace!(?claims, "zoho id_token parsed successfully");
//         UserData {
//             id: claims.sub,
//             name: None,
//             email: Some(claims.email),
//         }
