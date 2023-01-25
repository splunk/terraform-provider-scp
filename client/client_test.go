package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/acs/v2/mocks"
	client "github.com/splunk/terraform-provider-scp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	mockUsername = "mock-username"
	mockPassword = "mock-password"
	mockToken    = "mock-token"
	mockStack    = "mock-stack"
	mockTokenId  = "mock-token-id"
	mockServer   = "https://mock.admin.splunk.com"
	mockVersion  = "1.0.0"
)

func TestGetClient(t *testing.T) {
	assert := assert.New(t)

	t.Run(fmt.Sprint("test basic get client"), func(t *testing.T) {
		client, err := client.GetClient(mockServer, mockToken, mockVersion)
		assert.NoError(err)
		assert.NotNil(client)
	})
}

func TestCommonRequestEditors(t *testing.T) {
	assert := assert.New(t)

	t.Run(fmt.Sprint("test bearer auth request editors"), func(t *testing.T) {
		reqEditorFn := client.CommonRequestEditors(mockToken, mockVersion)
		assert.NotNil(reqEditorFn)
		assert.Equal(len(reqEditorFn), 2)
	})
}

func TestAddBearerAuth(t *testing.T) {
	assert := assert.New(t)

	t.Run(fmt.Sprint("test valid add basic auth"), func(t *testing.T) {
		err := addBearerAuthTestCase(mockToken)
		assert.NoError(err)
	})

	t.Run(fmt.Sprint("test empty token returns error"), func(t *testing.T) {
		err := addBearerAuthTestCase("")
		assert.ErrorContainsf(err, err.Error(), "provide a valid token")
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
	assert := assert.New(t)
	req, err := http.NewRequest(http.MethodGet, "some-url", nil)
	assert.NoError(err)
	client.AddUserAgent(req, mockVersion)
	expectedUserAgent := fmt.Sprintf("ACS-terraform-%s", mockVersion)
	assert.Equal(expectedUserAgent, req.Header.Get("User-Agent"))
}

func TestGetClientBasicAuth(t *testing.T) {
	assert := assert.New(t)

	t.Run(fmt.Sprint("test basic get client"), func(t *testing.T) {
		client, err := client.GetClientBasicAuth(mockServer, mockUsername, mockPassword, mockVersion)
		assert.NoError(err)
		assert.NotNil(client)
	})
}

func TestCommonRequestEditorsBasicAuth(t *testing.T) {
	assert := assert.New(t)

	t.Run(fmt.Sprint("test basic auth request editors"), func(t *testing.T) {
		reqEditorFn := client.CommonRequestEditorsBasicAuth(mockUsername, mockPassword, mockVersion)
		assert.NotNil(reqEditorFn)
		assert.Equal(len(reqEditorFn), 2)
	})

	t.Run(fmt.Sprint("test basic auth request editors"), func(t *testing.T) {
		reqEditorFn := client.CommonRequestEditorsBasicAuth(mockUsername, "", mockVersion)
		assert.NotNil(reqEditorFn)
		assert.Equal(len(reqEditorFn), 2)
	})
}

func TestAddBasicAuth(t *testing.T) {
	assert := assert.New(t)

	t.Run(fmt.Sprint("test valid add basic auth"), func(t *testing.T) {
		err := addBasicAuthTestCase(mockUsername, mockPassword)
		assert.NoError(err)
	})

	t.Run(fmt.Sprint("test empty username returns error"), func(t *testing.T) {
		err := addBasicAuthTestCase("", mockPassword)
		assert.ErrorContainsf(err, err.Error(), "provide a valid username")
	})

	t.Run(fmt.Sprint("test empty password returns error"), func(t *testing.T) {
		err := addBasicAuthTestCase(mockUsername, "")
		assert.ErrorContainsf(err, err.Error(), "provide a valid password")
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
	assert := assert.New(t)
	tokenType := client.TokenType

	mockCreateBody := v2.CreateTokenJSONRequestBody{
		User:     mockUsername,
		Audience: mockUsername,
		Type:     &tokenType,
	}

	t.Run("with some client interface error", func(t *testing.T) {
		mockClient.On("CreateToken", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(nil, errors.New("some error")).Once()
		token, err := client.GenerateToken(context.TODO(), mockClient, mockUsername, mockStack)
		assert.Error(err)
		assert.Equal(token, "")
	})

	t.Run("with valid params and http response 200", func(t *testing.T) {
		mockClient.On("CreateToken", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(genTokenResp(200), nil).Once()
		token, err := client.GenerateToken(context.TODO(), mockClient, mockUsername, mockStack)
		assert.NoError(err)
		assert.Equal(token, mockToken)
	})

	// http unexpected status codes
	t.Run("with unexpected http responses", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{101, 400, 401, 403, 404, 409, 500, 501, 503} {
			t.Run(fmt.Sprintf("with unexpected status %v", unexpectedStatusCode), func(t *testing.T) {
				mockClient.On("CreateToken", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(genTokenResp(unexpectedStatusCode), nil).Once()
				token, err := client.GenerateToken(context.TODO(), mockClient, mockUsername, mockStack)
				assert.Error(err)
				assert.Equal(token, "")
			})
		}
	})
}

func genTokenResp(code int) *http.Response {
	var b []byte
	token := mockToken
	if code == http.StatusOK {
		tokenInfo := v2.TokenInfo{
			Id:    mockTokenId,
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
