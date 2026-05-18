package client

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/appinspect"
	"github.com/splunk/terraform-provider-scp/internal/utils"
)

const TokenType = "ephemeral"
const SplunkbaseSessionURL = "https://splunkbase.splunk.com/api/account:login"
const SplunkLoginURL = "https://api.splunk.com/2.0/rest/login/splunk"

type ACSProvider struct {
	Client           *v2.ClientInterface
	Stack            v2.Stack
	AppInspectClient *appinspect.ClientInterface
	TargetClients    map[v2.Stack]v2.ClientInterface

	mu                sync.RWMutex
	server            string
	version           string
	username          string
	password          string
	authToken         string
	splunkbaseSession string
	splunkLoginToken  string
}

type LoginResult struct {
	User     string `json:"user"`
	Audience string `json:"audience"`
	// nolint
	Id        string `json:"id"`
	Token     string `json:"token"`
	Status    string `json:"status"`
	ExpiresOn string `json:"expiresOn"`
	NotBefore string `json:"notBefore"`
}

type TargetFields struct {
	Client v2.ClientInterface
	Stack  v2.Stack
}

type errInvalidAuth struct {
	field string
}

func (e errInvalidAuth) Error() string {
	return fmt.Sprintf("provide a valid %s", e.field)
}

// GetClient retrieves client with bearer authentication
func GetClient(server string, token string, version string, splunkbaseSession string, splunkLoginToken string) (v2.ClientInterface, error) {
	acsClient, err := v2.NewClient(server)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the client: %w", err)
	}
	acsClient.RequestEditors = CommonRequestEditors(token, version, splunkbaseSession, splunkLoginToken)
	return acsClient, nil
}

func (p *ACSProvider) Configure(server, version, username, password, authToken, splunkbaseSession, splunkLoginToken string) {
	p.server = server
	p.version = version
	p.username = username
	p.password = password
	p.authToken = authToken
	p.splunkbaseSession = splunkbaseSession
	p.splunkLoginToken = splunkLoginToken
	if p.TargetClients == nil {
		p.TargetClients = map[v2.Stack]v2.ClientInterface{}
	}
}

func (p *ACSProvider) ClientForTarget(ctx context.Context, stack v2.Stack) (v2.ClientInterface, error) {
	p.mu.RLock()
	value, ok := p.TargetClients[stack]
	p.mu.RUnlock()

	if ok {
		return value, nil
	}

	if p.username == "" || p.password == "" {
		return nil, fmt.Errorf("targeted app installs require provider username and password to generate per-instance tokens")
	}

	tmpClient, err := GetClientBasicAuth(p.server, p.username, p.password, p.version)
	if err != nil {
		return nil, err
	}

	token, err := GenerateToken(ctx, tmpClient, p.username, string(stack))
	if err != nil {
		return nil, err
	}

	client, err := GetClient(p.server, token, p.version, p.splunkbaseSession, p.splunkLoginToken)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	if p.TargetClients == nil {
		p.TargetClients = map[v2.Stack]v2.ClientInterface{}
	}
	p.TargetClients[stack] = client
	p.mu.Unlock()

	return client, err
}

func GetTargetInstall(ctx context.Context, target string, stack v2.Stack, acsProvider *ACSProvider) (TargetFields, error) {
	targetStack, err := utils.TargetStackName(target, stack)
	if err != nil {
		return TargetFields{}, err
	}

	targetClient, err := acsProvider.ClientForTarget(ctx, targetStack)
	if err != nil {
		return TargetFields{}, err
	}

	return TargetFields{
		Client: targetClient,
		Stack:  targetStack,
	}, nil
}

func CommonRequestEditors(token string, version string, splunkbaseSession string, splunkLoginToken string) []v2.RequestEditorFn {
	addUserAgent := func(_ context.Context, req *http.Request) error {
		return AddUserAgent(req, version)
	}
	return []v2.RequestEditorFn{AddBearerAuth(token), addUserAgent, AddXSplunkbaseAuthorizationHeader(splunkbaseSession), AddSplunkLoginToken(splunkLoginToken)}
}

func AddBearerAuth(token string) v2.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
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
func GetClientBasicAuth(server string, username string, password string, version string) (*v2.Client, error) {
	acsClient, err := v2.NewClient(server)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the client: %w", err)
	}
	acsClient.RequestEditors = CommonRequestEditorsBasicAuth(username, password, version)
	return acsClient, nil
}

func CommonRequestEditorsBasicAuth(username string, password string, version string) []v2.RequestEditorFn {
	addUserAgent := func(_ context.Context, req *http.Request) error {
		return AddUserAgent(req, version)
	}
	return []v2.RequestEditorFn{AddBasicAuth(username, password), addUserAgent}
}

func AddBasicAuth(username string, password string) v2.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
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

func AddXSplunkbaseAuthorizationHeader(splunkbaseSession string) v2.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("X-Splunkbase-Authorization", splunkbaseSession)
		return nil
	}
}
func AddSplunkLoginToken(splunkLoginToken string) v2.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("X-Splunk-Authorization", splunkLoginToken)
		return nil
	}
}

// GetSplunkLoginTokenWithClient retrieves a Splunk login token with provided HTTP client and URL
func GetSplunkLoginTokenWithClient(username, password string, httpClient *http.Client, url string) (string, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.SetBasicAuth(username, password)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			tflog.Error(context.Background(), fmt.Sprintf("Error closing response body: %v", err))
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}
	var responseData struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return "", fmt.Errorf("error parsing JSON: %w", err)
	}
	return responseData.Data.Token, nil
}

// GetSplunkLoginToken retrieves a Splunk login token using default client and URL
func GetSplunkLoginToken(username, password string) (string, error) {
	return GetSplunkLoginTokenWithClient(username, password, &http.Client{}, SplunkLoginURL)
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error closing response body: %v", err))
		}
	}(resp.Body)
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

// GetSplunkbaseSessionWithClient retrieves a Splunkbase session with provided HTTP client and URL
func GetSplunkbaseSessionWithClient(ctx context.Context, username, password string, httpClient *http.Client, url string) (string, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	tflog.Info(ctx, "Getting Splunkbase session")
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	err := writer.WriteField("username", username)
	if err != nil {
		return "", fmt.Errorf("error writing field: %w", err)
	}
	err = writer.WriteField("password", password)
	if err != nil {
		return "", fmt.Errorf("error writing field: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("error closing writer: %w", err)
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error closing response body: %v", err))
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		if len(body) == 0 {
			return "", fmt.Errorf("splunkbase login failed with status %d", res.StatusCode)
		}
		return "", fmt.Errorf("splunkbase login failed with status %d: %s", res.StatusCode, string(body))
	}
	if len(body) == 0 {
		return "", fmt.Errorf("splunkbase login failed")
	}
	type LoginResponse struct {
		ID string `xml:"id"`
	}

	var loginResponse LoginResponse
	err = xml.Unmarshal(body, &loginResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling XML: %w", err)
	}
	if loginResponse.ID == "" {
		return "", fmt.Errorf("splunkbase login response did not include a session id")
	}
	return loginResponse.ID, nil
}

// GetSplunkbaseSession retrieves a Splunkbase session using default client and URL
func GetSplunkbaseSession(ctx context.Context, username, password string) (string, error) {
	return GetSplunkbaseSessionWithClient(ctx, username, password, &http.Client{}, SplunkbaseSessionURL)
}
