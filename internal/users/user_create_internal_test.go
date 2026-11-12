package users

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/acs/v2/mocks"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResourceUserCreateWaitsUntilReadable(t *testing.T) {
	data := schema.TestResourceDataRaw(t, userResourceSchema(), map[string]interface{}{schemaKeyName: "test-user", schemaKeyPassword: "example-password"})
	mockClient := &mocks.ClientInterface{}
	var acsClient v2.ClientInterface = mockClient
	provider := &client.ACSProvider{Client: &acsClient, Stack: "test-stack"}
	user := &v2.UsersResponse{
		Name:       "test-user",
		DefaultApp: "search",
		Email:      "test@example.invalid",
		FullName:   "Test User",
		Roles:      []string{"user"},
	}
	mockClient.On("CreateUser", mock.Anything, v2.Stack("test-stack"), mock.Anything, mock.Anything).Return(userCreateTestResponse(t, http.StatusOK, nil), nil).Once()
	mockClient.On("DescribeUser", mock.Anything, v2.Stack("test-stack"), v2.UserName("test-user")).Return(userCreateTestResponse(t, http.StatusNotFound, nil), nil).Once()
	mockClient.On("DescribeUser", mock.Anything, v2.Stack("test-stack"), v2.UserName("test-user")).Return(userCreateTestResponse(t, http.StatusOK, user), nil).Once()
	diagnostics := resourceUserCreate(context.Background(), data, provider)
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
	require.Equal(t, "test-user", data.Id())
	require.Equal(t, "test-user", data.Get(schemaKeyName))
	require.Equal(t, user.DefaultApp, data.Get(schemaKeyDefaultApp))
	require.Equal(t, user.Email, data.Get(schemaKeyEmail))
	require.Equal(t, user.FullName, data.Get(schemaKeyFullName))
	require.Equal(t, []interface{}{"user"}, data.Get(schemaKeyRoles).(*schema.Set).List())
	mockClient.AssertExpectations(t)
}

func TestResourceUserCreatePreservesIDOnReadFailure(t *testing.T) {
	for _, scenario := range []string{"forbidden", "canceled"} {
		t.Run(scenario, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, userResourceSchema(), map[string]interface{}{schemaKeyName: "test-user", schemaKeyPassword: "example-password"})
			mockClient := &mocks.ClientInterface{}
			var acsClient v2.ClientInterface = mockClient
			provider := &client.ACSProvider{Client: &acsClient, Stack: "test-stack"}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mockClient.On("CreateUser", mock.Anything, v2.Stack("test-stack"), mock.Anything, mock.Anything).Return(userCreateTestResponse(t, http.StatusOK, nil), nil).Once()
			read := mockClient.On("DescribeUser", mock.Anything, v2.Stack("test-stack"), v2.UserName("test-user"))
			if scenario == "canceled" {
				read.Run(func(args mock.Arguments) {
					if _, ok := args.Get(0).(context.Context).Deadline(); !ok {
						t.Error("read must receive a visibility deadline")
					}
					cancel()
				}).Return(nil, context.Canceled).Once()
			} else {
				read.Return(userCreateTestResponse(t, http.StatusForbidden, nil), nil).Once()
			}

			diagnostics := resourceUserCreate(ctx, data, provider)
			require.True(t, diagnostics.HasError(), "%v", diagnostics)
			require.Equal(t, "test-user", data.Id())
			mockClient.AssertExpectations(t)
		})
	}
}

func userCreateTestResponse(t *testing.T, code int, body interface{}) *http.Response {
	t.Helper()
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(code)
	if body != nil {
		require.NoError(t, json.NewEncoder(recorder).Encode(body))
	}
	return recorder.Result()
}

func TestUserCreateReadPendingStatuses(t *testing.T) {
	for _, code := range []int{http.StatusNotFound, http.StatusFailedDependency, http.StatusTooManyRequests} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			mockClient := &mocks.ClientInterface{}
			recorder := httptest.NewRecorder()
			recorder.WriteHeader(code)
			mockClient.On("DescribeUser", mock.Anything, v2.Stack("test-stack"), v2.UserName("test")).Return(recorder.Result(), nil).Once()
			result, state, err := userStatusRead(context.Background(), mockClient, "test-stack", "test", true)()
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, http.StatusText(code), state)
		})
	}
}
