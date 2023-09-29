package users_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/acs/v2/mocks"
	"github.com/splunk/terraform-provider-scp/internal/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	mockAck                   = "Y"
	mockCapabilities          = []string{"mock-capability-1"}
	mockDefaultAppSource      = "mock-default-app-source"
	unexpectedStatusCodes     = []int{400, 401, 403, 404, 409, 501, 503}
	unexpectedStatusCodesPoll = []int{400, 401, 403, 409, 501, 500, 503}
)

const (
	mockUserName = "mock-user"
	mockStack    = "mock-stack"
)

func Test_WaitUserCreate(t *testing.T) {
	client := &mocks.ClientInterface{}
	mockCreateParam := v2.CreateUserParams{
		FederatedSearchManageAck: (*v2.FederatedSearchManage)(&mockAck),
	}

	mockCreateBody := v2.CreateUserJSONRequestBody{
		Name: mockUserName,
	}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("CreateUser", mock.Anything, v2.Stack(mockStack), &mockCreateParam, mockCreateBody).Return(nil, errors.New("some error")).Once()
		err := users.WaitUserCreate(context.TODO(), client, mockStack, mockCreateParam, mockCreateBody)
		assert.Error(t, err)
	})

	t.Run("with http response 200", func(t *testing.T) {
		client.On("CreateUser", mock.Anything, v2.Stack(mockStack), &mockCreateParam, mockCreateBody).Return(generateResponse(200), nil).Once()
		err := users.WaitUserCreate(context.TODO(), client, mockStack, mockCreateParam, mockCreateBody)
		assert.NoError(t, err)
	})

	t.Run("with unexpected http responses", func(t *testing.T) {
		for _, unexpectedStatusCode := range unexpectedStatusCodes {
			t.Run(fmt.Sprintf("with unexpected status %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("CreateUser", mock.Anything, v2.Stack(mockStack), &mockCreateParam, mockCreateBody).Return(generateResponse(unexpectedStatusCode), nil).Once()
				err := users.WaitUserCreate(context.TODO(), client, mockStack, mockCreateParam, mockCreateBody)
				assert.Error(t, err)
			})
		}
	})
}

func generateResponse(code int) *http.Response {
	var b []byte
	if code == http.StatusOK {

		user := v2.UsersResponse{
			Capabilities:     mockCapabilities,
			DefaultApp:       mockDefaultApp,
			DefaultAppSource: mockDefaultAppSource,
			Email:            mockEmail,
			FullName:         mockFullName,
			Name:             mockUserName,
			Roles:            mockRoles,
		}

		b, _ = json.Marshal(&user)
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

func Test_WaitUserRead(t *testing.T) {
	client := &mocks.ClientInterface{}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("DescribeUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName)).Return(nil, errors.New("some error")).Once()
		user, err := users.WaitUserRead(context.TODO(), client, mockStack, mockUserName)
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("with http 200 response", func(t *testing.T) {
		client.On("DescribeUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName)).Return(generateResponse(http.StatusOK), nil).Once()
		user, err := users.WaitUserRead(context.TODO(), client, mockStack, mockUserName)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotNil(t, user.Name)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range unexpectedStatusCodes {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DescribeUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName)).Return(generateResponse(unexpectedStatusCode), nil).Once()
				index, err := users.WaitUserRead(context.TODO(), client, mockStack, mockUserName)
				assert.Error(t, err)
				assert.Nil(t, index)
			})
		}
	})
}

func Test_WaitUserUpdate(t *testing.T) {
	client := &mocks.ClientInterface{}
	mockUpdateParam := v2.PatchUserParams{
		FederatedSearchManageAck: (*v2.FederatedSearchManage)(&mockAck),
	}
	mockUpdateBody := v2.PatchUserJSONRequestBody{
		DefaultApp:      &mockDefaultApp,
		Email:           &mockEmail,
		ForceChangePass: &mockForceChangePass,
		FullName:        &mockFullName,
		OldPassword:     &mockPassword,
		Roles:           &mockRoles,
	}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("PatchUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName), &mockUpdateParam, mockUpdateBody).Return(nil, errors.New("some error")).Once()
		err := users.WaitUserUpdate(context.TODO(), client, mockStack, mockUpdateParam, mockUpdateBody, mockUserName)
		assert.Error(t, err)
	})

	t.Run("with http response 200", func(t *testing.T) {
		client.On("PatchUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName), &mockUpdateParam, mockUpdateBody).Return(generateResponse(http.StatusOK), nil).Once()
		err := users.WaitUserUpdate(context.TODO(), client, mockStack, mockUpdateParam, mockUpdateBody, mockUserName)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range unexpectedStatusCodes {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("PatchUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName), &mockUpdateParam, mockUpdateBody).Return(generateResponse(unexpectedStatusCode), nil).Once()
				err := users.WaitUserUpdate(context.TODO(), client, mockStack, mockUpdateParam, mockUpdateBody, mockUserName)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitVerifyUserUpdate(t *testing.T) {
	client := &mocks.ClientInterface{}
	mockUpdateBody := v2.PatchUserJSONRequestBody{
		DefaultApp:      &mockDefaultApp,
		Email:           &mockEmail,
		ForceChangePass: &mockForceChangePass,
		FullName:        &mockFullName,
		OldPassword:     &mockPassword,
		Roles:           &mockRoles,
	}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("DescribeUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName)).Return(nil, errors.New("some error")).Once()
		err := users.WaitVerifyUserUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockUserName)
		assert.Error(t, err)
	})

	t.Run("with http 200 response", func(t *testing.T) {
		client.On("DescribeUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName)).Return(generateResponse(http.StatusOK), nil).Once()
		err := users.WaitVerifyUserUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockUserName)
		assert.NoError(t, err)
	})

	t.Run("with non updated response first", func(t *testing.T) {
		client.On("DescribeUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName)).Return(generateResponse(http.StatusOK), nil).Once()
		err := users.WaitVerifyUserUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockUserName)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range unexpectedStatusCodesPoll {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DescribeUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName)).Return(generateResponse(unexpectedStatusCode), nil).Once()
				err := users.WaitVerifyUserUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockUserName)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitUserDelete(t *testing.T) {
	client := &mocks.ClientInterface{}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("DeleteUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName)).Return(nil, errors.New("some error")).Once()
		err := users.WaitUserDelete(context.TODO(), client, mockStack, mockUserName)
		assert.Error(t, err)
	})

	t.Run("with http response 202", func(t *testing.T) {
		client.On("DeleteUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName)).Return(generateResponse(http.StatusOK), nil).Once()
		err := users.WaitUserDelete(context.TODO(), client, mockStack, mockUserName)
		assert.NoError(t, err)
	})

	t.Run("with retry on rate limit", func(t *testing.T) {
		client.On("DeleteUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName)).Return(generateResponse(http.StatusTooManyRequests), nil).Once()
		client.On("DeleteUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName)).Return(generateResponse(http.StatusOK), nil).Once()
		err := users.WaitUserDelete(context.TODO(), client, mockStack, mockUserName)
		assert.NoError(t, err)
	})

	t.Run("with unexpected error resp", func(t *testing.T) {
		for _, unexpectedStatusCode := range unexpectedStatusCodes {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DeleteUser", mock.Anything, v2.Stack(mockStack), v2.UserName(mockUserName)).Return(generateResponse(unexpectedStatusCode), nil).Once()
				err := users.WaitUserDelete(context.TODO(), client, mockStack, mockUserName)
				assert.Error(t, err)
			})
		}
	})
}
