package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/acs/v2/mocks"
	client "github.com/splunk/terraform-provider-scp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	mockUsername          = "mock-username"
	mockPassword          = "mock-password"
	mockToken             = "mock-token"
	mockStack             = "mock-stack"
	mockTokenID           = "mock-token-id"
	mockServer            = "https://mock.admin.splunk.com"
	mockVersion           = "1.0.0"
	mockSplunkbaseSession = "xyz"
	mockSplunkLoginToken  = "mock-splunk-login-token" //nolint:gosec
)

func TestGetClient(t *testing.T) {
	asserttion := assert.New(t)

	t.Run("test basic get client", func(_ *testing.T) {
		acsClient, err := client.GetClient(mockServer, mockToken, mockVersion, mockSplunkbaseSession, mockSplunkLoginToken)
		asserttion.NoError(err)
		asserttion.NotNil(acsClient)
	})
}

func TestCommonRequestEditors(t *testing.T) {
	assertion := assert.New(t)

	t.Run("test bearer auth request editors", func(_ *testing.T) {
		reqEditorFn := client.CommonRequestEditors(mockToken, mockVersion, mockSplunkbaseSession, mockSplunkLoginToken)
		assertion.NotNil(reqEditorFn)
		assertion.Equal(len(reqEditorFn), 4)
	})
}

func TestAddBearerAuth(t *testing.T) {
	assertion := assert.New(t)

	t.Run("test valid add basic auth", func(_ *testing.T) {
		err := addBearerAuthTestCase(mockToken)
		assertion.NoError(err)
	})

	t.Run("test empty token returns error", func(_ *testing.T) {
		err := addBearerAuthTestCase("")
		assertion.ErrorContainsf(err, err.Error(), "provide a valid token")
	})
}

func addBearerAuthTestCase(token string) error {
	req, err := http.NewRequest(http.MethodGet, "some-url", nil)
	if err != nil {
		return err
	}
	setToken := token
	middlewareFunc := client.AddBearerAuth(token)
	if err := middlewareFunc(nil, req); err != nil {
		return err
	}

	setTokenValue := "Bearer " + setToken
	if receivedToken := req.Header.Get("Authorization"); receivedToken == "" {
		return fmt.Errorf("no auth headers set")
	} else if receivedToken != setTokenValue {
		return fmt.Errorf("unexpected token received, expected (%s), received (%s)", setToken, receivedToken)
	}
	return nil
}

func TestAddUserAgent(t *testing.T) {
	assertion := assert.New(t)
	req, err := http.NewRequest(http.MethodGet, "some-url", nil)
	assertion.NoError(err)
	err = client.AddUserAgent(req, mockVersion)
	assertion.NoError(err)
	expectedUserAgent := fmt.Sprintf("ACS-terraform-%s", mockVersion)
	assertion.Equal(expectedUserAgent, req.Header.Get("User-Agent"))
}

func TestGetClientBasicAuth(t *testing.T) {
	assertion := assert.New(t)

	t.Run("test basic get client", func(_ *testing.T) {
		acsClient, err := client.GetClientBasicAuth(mockServer, mockUsername, mockPassword, mockVersion)
		assertion.NoError(err)
		assertion.NotNil(acsClient)
	})
}

func TestCommonRequestEditorsBasicAuth(t *testing.T) {
	assertion := assert.New(t)

	t.Run("test basic auth request editors", func(_ *testing.T) {
		reqEditorFn := client.CommonRequestEditorsBasicAuth(mockUsername, mockPassword, mockVersion)
		assertion.NotNil(reqEditorFn)
		assertion.Equal(len(reqEditorFn), 2)
	})

	t.Run("test basic auth request editors", func(_ *testing.T) {
		reqEditorFn := client.CommonRequestEditorsBasicAuth(mockUsername, "", mockVersion)
		assertion.NotNil(reqEditorFn)
		assertion.Equal(len(reqEditorFn), 2)
	})
}

func TestAddBasicAuth(t *testing.T) {
	assertion := assert.New(t)

	t.Run("test valid add basic auth", func(_ *testing.T) {
		err := addBasicAuthTestCase(mockUsername, mockPassword)
		assertion.NoError(err)
	})

	t.Run("test empty username returns error", func(_ *testing.T) {
		err := addBasicAuthTestCase("", mockPassword)
		assertion.ErrorContainsf(err, err.Error(), "provide a valid username")
	})

	t.Run("test empty password returns error", func(_ *testing.T) {
		err := addBasicAuthTestCase(mockUsername, "")
		assertion.ErrorContainsf(err, err.Error(), "provide a valid password")
	})
}

func addBasicAuthTestCase(username string, password string) error {
	req, err := http.NewRequest(http.MethodGet, "some-url", nil)
	if err != nil {
		return err
	}
	setUsername := username
	setPassword := password
	middlewareFunc := client.AddBasicAuth(username, password)
	if err := middlewareFunc(nil, req); err != nil {
		return err
	}
	if receivedUsername, receivedPassword, ok := req.BasicAuth(); !ok {
		return fmt.Errorf("no basic auth headers set")
	} else if receivedUsername != setUsername || receivedPassword != setPassword {
		return fmt.Errorf("unexpected (username, password) received, expected (%s, %s), received (%s, %s)", setUsername, setPassword, receivedUsername, receivedPassword)
	}
	return nil
}

func TestGenerateToken(t *testing.T) {
	mockClient := &mocks.ClientInterface{}
	assertion := assert.New(t)
	tokenType := client.TokenType

	mockCreateBody := v2.CreateTokenJSONRequestBody{
		User:     mockUsername,
		Audience: mockUsername,
		Type:     &tokenType,
	}

	t.Run("with some client interface error", func(_ *testing.T) {
		mockClient.On("CreateToken", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(nil, errors.New("some error")).Once()
		token, err := client.GenerateToken(context.TODO(), mockClient, mockUsername, mockStack)
		assertion.Error(err)
		assertion.Equal(token, "")
	})

	t.Run("with some unmarshal error", func(_ *testing.T) {
		mockClient.On("CreateToken", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(genInvalidTokenResp(200), errors.New("some error")).Once()
		token, err := client.GenerateToken(context.TODO(), mockClient, mockUsername, mockStack)
		assertion.ErrorContainsf(err, err.Error(), "unmarshal error")
		assertion.Equal(token, "")
	})

	t.Run("with valid params and http response 200", func(_ *testing.T) {
		mockClient.On("CreateToken", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(genTokenResp(200), nil).Once()
		token, err := client.GenerateToken(context.TODO(), mockClient, mockUsername, mockStack)
		assertion.NoError(err)
		assertion.Equal(token, mockToken)
	})

	// http unexpected status codes
	t.Run("with unexpected http responses", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{101, 400, 401, 403, 404, 409, 500, 501, 503} {
			t.Run(fmt.Sprintf("with unexpected status %v", unexpectedStatusCode), func(_ *testing.T) {
				mockClient.On("CreateToken", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(genTokenResp(unexpectedStatusCode), nil).Once()
				token, err := client.GenerateToken(context.TODO(), mockClient, mockUsername, mockStack)
				assertion.Error(err)
				assertion.Equal(token, "")
			})
		}
	})
}

func genTokenResp(code int) *http.Response {
	var b []byte
	token := mockToken
	if code == http.StatusOK {
		tokenInfo := v2.TokenInfo{
			Id:    mockTokenID,
			Token: &token,
		}

		b, _ = json.Marshal(&tokenInfo)
	} else {
		b, _ = json.Marshal(&v2.Error{
			Code:    http.StatusText(code),
			Message: http.StatusText(code),
		})
	}
	recorder := httptest.NewRecorder()
	recorder.Header().Add("Content-Type", "json")
	recorder.WriteHeader(code)
	if b != nil {
		_, _ = recorder.Write(b)
	}
	return recorder.Result()
}

func genInvalidTokenResp(code int) *http.Response {
	var b []byte
	if code == http.StatusOK {
		tokenInfo := v2.HecSpec{}

		b, _ = json.Marshal(&tokenInfo)
	} else {
		b, _ = json.Marshal(&v2.Error{
			Code:    http.StatusText(code),
			Message: http.StatusText(code),
		})
	}
	recorder := httptest.NewRecorder()
	recorder.Header().Add("Content-Type", "json")
	recorder.WriteHeader(code)
	if b != nil {
		_, _ = recorder.Write(b)
	}
	return recorder.Result()
}

func TestAddSplunkbaseSession(t *testing.T) {
	assertion := assert.New(t)

	t.Run("test valid splunkbase session header", func(_ *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "some-url", nil)
		assertion.NoError(err)

		middlewareFunc := client.AddXSplunkbaseAuthorizationHeader(mockSplunkbaseSession)
		err = middlewareFunc(nil, req)
		assertion.NoError(err)
		assertion.Equal(mockSplunkbaseSession, req.Header.Get("X-Splunkbase-Authorization"))
	})

	t.Run("test empty splunkbase session header", func(_ *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "some-url", nil)
		assertion.NoError(err)

		middlewareFunc := client.AddXSplunkbaseAuthorizationHeader("")
		err = middlewareFunc(nil, req)
		assertion.NoError(err)
		assertion.Equal("", req.Header.Get("X-Splunkbase-Authorization"))
	})
}

func TestAddSplunkLoginToken(t *testing.T) {
	assertion := assert.New(t)

	t.Run("test valid splunk login token header", func(_ *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "some-url", nil)
		assertion.NoError(err)

		middlewareFunc := client.AddSplunkLoginToken(mockSplunkLoginToken)
		err = middlewareFunc(nil, req)
		assertion.NoError(err)
		assertion.Equal(mockSplunkLoginToken, req.Header.Get("X-Splunk-Authorization"))
	})

	t.Run("test empty splunk login token header", func(_ *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "some-url", nil)
		assertion.NoError(err)

		middlewareFunc := client.AddSplunkLoginToken("")
		err = middlewareFunc(nil, req)
		assertion.NoError(err)
		assertion.Equal("", req.Header.Get("X-Splunk-Authorization"))
	})
}

func TestGetSplunkLoginTokenWithClient(t *testing.T) {
	assertion := assert.New(t)

	t.Run("test successful token retrieval", func(_ *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method
			assertion.Equal(http.MethodGet, r.Method)

			// Verify basic auth is set
			username, password, ok := r.BasicAuth()
			assertion.True(ok)
			assertion.Equal(mockUsername, username)
			assertion.Equal(mockPassword, password)

			// Return mock token response
			w.WriteHeader(http.StatusOK)
			response := `{"data":{"token":"test-login-token-123"}}`
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		token, err := client.GetSplunkLoginTokenWithClient(mockUsername, mockPassword, &http.Client{}, server.URL)
		assertion.NoError(err)
		assertion.Equal("test-login-token-123", token)
	})

	t.Run("test with invalid JSON response", func(_ *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		_, err := client.GetSplunkLoginTokenWithClient(mockUsername, mockPassword, &http.Client{}, server.URL)
		assertion.Error(err)
		assertion.Contains(err.Error(), "error parsing JSON")
	})

	t.Run("test with HTTP error", func(_ *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		token, err := client.GetSplunkLoginTokenWithClient(mockUsername, mockPassword, &http.Client{}, server.URL)
		// Function doesn't check HTTP status, so it will try to parse the response
		// and likely fail with JSON parsing error or return empty token
		assertion.NotNil(err)
		assertion.Equal("", token)
	})

	t.Run("test with invalid URL", func(_ *testing.T) {
		_, err := client.GetSplunkLoginTokenWithClient(mockUsername, mockPassword, &http.Client{}, "://invalid-url")
		assertion.Error(err)
		assertion.Contains(err.Error(), "error creating request")
	})

	t.Run("test with nil HTTP client", func(_ *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response := `{"data":{"token":"test-token-with-nil-client"}}`
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		// Should not panic and should use default client
		token, err := client.GetSplunkLoginTokenWithClient(mockUsername, mockPassword, nil, server.URL)
		assertion.NoError(err)
		assertion.Equal("test-token-with-nil-client", token)
	})
}

func TestGetSplunkbaseSessionWithClient(t *testing.T) {
	assertion := assert.New(t)
	ctx := context.Background()

	t.Run("test successful session retrieval", func(_ *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method
			assertion.Equal(http.MethodPost, r.Method)

			// Verify content type is multipart
			assertion.Contains(r.Header.Get("Content-Type"), "multipart/form-data")

			// Parse form data
			err := r.ParseMultipartForm(32 << 20)
			assertion.NoError(err)
			assertion.Equal(mockUsername, r.FormValue("username"))
			assertion.Equal(mockPassword, r.FormValue("password"))

			// Return mock XML response
			w.WriteHeader(http.StatusOK)
			response := `<response><id>test-session-id-456</id></response>`
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		sessionID, err := client.GetSplunkbaseSessionWithClient(ctx, mockUsername, mockPassword, &http.Client{}, server.URL)
		assertion.NoError(err)
		assertion.Equal("test-session-id-456", sessionID)
	})

	t.Run("test with invalid XML response", func(_ *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid xml"))
		}))
		defer server.Close()

		_, err := client.GetSplunkbaseSessionWithClient(ctx, mockUsername, mockPassword, &http.Client{}, server.URL)
		assertion.Error(err)
		assertion.Contains(err.Error(), "error unmarshalling XML")
	})

	t.Run("test with HTTP error", func(_ *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		sessionID, err := client.GetSplunkbaseSessionWithClient(ctx, mockUsername, mockPassword, &http.Client{}, server.URL)
		assertion.NotNil(err)
		assertion.Equal("", sessionID)
		assertion.Contains(err.Error(), "splunkbase login failed with status 403")
	})

	t.Run("test with empty response body", func(_ *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		sessionID, err := client.GetSplunkbaseSessionWithClient(ctx, mockUsername, mockPassword, &http.Client{}, server.URL)
		assertion.Error(err)
		assertion.Equal("", sessionID)
		assertion.Contains(err.Error(), "splunkbase login failed")
	})

	t.Run("test with missing session id", func(_ *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<response></response>`))
		}))
		defer server.Close()

		sessionID, err := client.GetSplunkbaseSessionWithClient(ctx, mockUsername, mockPassword, &http.Client{}, server.URL)
		assertion.Error(err)
		assertion.Equal("", sessionID)
		assertion.Contains(err.Error(), "did not include a session id")
	})

	t.Run("test with invalid URL", func(_ *testing.T) {
		_, err := client.GetSplunkbaseSessionWithClient(ctx, mockUsername, mockPassword, &http.Client{}, "://invalid-url")
		assertion.Error(err)
		assertion.Contains(err.Error(), "error creating request")
	})

	t.Run("test with nil HTTP client", func(_ *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response := `<response><id>test-session-with-nil-client</id></response>`
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		// Should not panic and should use default client
		sessionID, err := client.GetSplunkbaseSessionWithClient(ctx, mockUsername, mockPassword, nil, server.URL)
		assertion.NoError(err)
		assertion.Equal("test-session-with-nil-client", sessionID)
	})
}
