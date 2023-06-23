package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/splunk/terraform-provider-scp/acs/v2"
	"io"
	"net/http"
)

const TokenType = "ephemeral"

type ACSProvider struct {
	Client *v2.ClientInterface
	Stack  v2.Stack
}

type LoginResult struct {
	User      string `json:"user"`
	Audience  string `json:"audience"`
	Id        string `json:"id"`
	Token     string `json:"token"`
	Status    string `json:"status"`
	ExpiresOn string `json:"expiresOn"`
	NotBefore string `json:"notBefore"`
}

type errInvalidAuth struct {
	field string
}

func (e errInvalidAuth) Error() string {
	return fmt.Sprintf("provide a valid %s", e.field)
}

// GetClient retrieves client with bearer authentication
func GetClient(server string, token string, version string) (v2.ClientInterface, error) {
	acsClient, err := v2.NewClient(server)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the client: %w", err)
	}
	acsClient.RequestEditors = CommonRequestEditors(token, version)
	return acsClient, nil
}

func CommonRequestEditors(token string, version string) []v2.RequestEditorFn {
	addUserAgent := func(ctx context.Context, req *http.Request) error {
		return AddUserAgent(req, version)
	}
	return []v2.RequestEditorFn{AddBearerAuth(token), addUserAgent}
}

func AddBearerAuth(token string) v2.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		if token == "" {
			return &errInvalidAuth{field: "token"}
		}
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}

func AddUserAgent(req *http.Request, version string) error {
	userAgent := fmt.Sprintf("ACS-terraform-%s", version)
	req.Header.Set("User-Agent", userAgent)
	return nil
}

// GetClientBasicAuth retrieves client with Basic authentication instead of bearer authentication to use to generate token
func GetClientBasicAuth(server string, username string, password string, version string) (v2.ClientInterface, error) {
	acsClient, err := v2.NewClient(server)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the client: %w", err)
	}
	acsClient.RequestEditors = CommonRequestEditorsBasicAuth(username, password, version)
	return acsClient, nil
}

func CommonRequestEditorsBasicAuth(username string, password string, version string) []v2.RequestEditorFn {
	addUserAgent := func(ctx context.Context, req *http.Request) error {
		return AddUserAgent(req, version)
	}
	return []v2.RequestEditorFn{AddBasicAuth(username, password), addUserAgent}
}

func AddBasicAuth(username string, password string) v2.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		if username == "" {
			return &errInvalidAuth{field: "username"}
		}
		if password == "" {
			return &errInvalidAuth{field: "password"}
		}
		req.SetBasicAuth(username, password)
		return nil
	}
}

// GenerateToken creates an ephemeral token to be used for ACS client
func GenerateToken(ctx context.Context, clientInterface v2.ClientInterface, user string, stack string) (string, error) {
	tflog.Info(ctx, fmt.Sprintf("Creating token on stack %s", stack))
	tokenType := TokenType

	tokenBody := v2.CreateTokenJSONRequestBody{
		User:     user,
		Audience: user,
		Type:     &tokenType,
	}
	resp, err := clientInterface.CreateToken(ctx, v2.Stack(stack), tokenBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)

	tflog.Info(ctx, fmt.Sprintf("Create token request ID %s", resp.Header.Get("X-REQUEST-ID")))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to create token: %v", errors.New(string(bodyBytes)))
	}

	var loginResult LoginResult
	if err = json.Unmarshal(bodyBytes, &loginResult); err != nil {
		return "", fmt.Errorf("unmarshal error: %v", err)
	}

	return loginResult.Token, nil
}
