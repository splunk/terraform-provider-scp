package roles

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

func TestResourceRoleCreateWaitsUntilReadable(t *testing.T) {
	data := schema.TestResourceDataRaw(t, roleResourceSchema(), map[string]interface{}{schemaKeyName: "test-role"})
	mockClient := &mocks.ClientInterface{}
	var acsClient v2.ClientInterface = mockClient
	provider := &client.ACSProvider{Client: &acsClient, Stack: "test-stack"}
	defaultApp := "search"
	quota := 17
	importedRoles := []string{"user"}
	role := &v2.RolesResponse{
		DefaultApp:              &defaultApp,
		CumulativeSrchJobsQuota: &quota,
		Imported:                &v2.ImportedRolesInfo{Roles: &importedRoles},
	}
	mockClient.On("CreateRole", mock.Anything, v2.Stack("test-stack"), mock.Anything, mock.Anything).Return(roleCreateTestResponse(t, http.StatusOK, nil), nil).Once()
	mockClient.On("DescribeRole", mock.Anything, v2.Stack("test-stack"), v2.RoleName("test-role")).Return(roleCreateTestResponse(t, http.StatusNotFound, nil), nil).Once()
	mockClient.On("DescribeRole", mock.Anything, v2.Stack("test-stack"), v2.RoleName("test-role")).Return(roleCreateTestResponse(t, http.StatusOK, role), nil).Once()
	diagnostics := resourceRoleCreate(context.Background(), data, provider)
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
	require.Equal(t, "test-role", data.Id())
	require.Equal(t, "test-role", data.Get(schemaKeyName))
	require.Equal(t, defaultApp, data.Get(schemaKeyDefaultApp))
	require.Equal(t, quota, data.Get(schemaKeyCumulativeSrchJobsQuota))
	require.Equal(t, []interface{}{"user"}, data.Get(schemaKeyImportedRoles).(*schema.Set).List())
	mockClient.AssertExpectations(t)
}

func TestResourceRoleCreatePreservesIDOnReadFailure(t *testing.T) {
	for _, scenario := range []string{"forbidden", "canceled"} {
		t.Run(scenario, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, roleResourceSchema(), map[string]interface{}{schemaKeyName: "test-role"})
			mockClient := &mocks.ClientInterface{}
			var acsClient v2.ClientInterface = mockClient
			provider := &client.ACSProvider{Client: &acsClient, Stack: "test-stack"}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mockClient.On("CreateRole", mock.Anything, v2.Stack("test-stack"), mock.Anything, mock.Anything).Return(roleCreateTestResponse(t, http.StatusOK, nil), nil).Once()
			read := mockClient.On("DescribeRole", mock.Anything, v2.Stack("test-stack"), v2.RoleName("test-role"))
			if scenario == "canceled" {
				read.Run(func(args mock.Arguments) {
					if _, ok := args.Get(0).(context.Context).Deadline(); !ok {
						t.Error("read must receive a visibility deadline")
					}
					cancel()
				}).Return(nil, context.Canceled).Once()
			} else {
				read.Return(roleCreateTestResponse(t, http.StatusForbidden, nil), nil).Once()
			}

			diagnostics := resourceRoleCreate(ctx, data, provider)
			require.True(t, diagnostics.HasError(), "%v", diagnostics)
			require.Equal(t, "test-role", data.Id())
			mockClient.AssertExpectations(t)
		})
	}
}

func roleCreateTestResponse(t *testing.T, code int, body interface{}) *http.Response {
	t.Helper()
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(code)
	if body != nil {
		require.NoError(t, json.NewEncoder(recorder).Encode(body))
	}
	return recorder.Result()
}

func TestRoleCreateReadPendingStatuses(t *testing.T) {
	for _, code := range []int{http.StatusNotFound, http.StatusFailedDependency, http.StatusTooManyRequests} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			mockClient := &mocks.ClientInterface{}
			recorder := httptest.NewRecorder()
			recorder.WriteHeader(code)
			mockClient.On("DescribeRole", mock.Anything, v2.Stack("test-stack"), v2.RoleName("test")).Return(recorder.Result(), nil).Once()
			result, state, err := roleStatusRead(context.Background(), mockClient, "test-stack", "test", true)()
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, http.StatusText(code), state)
		})
	}
}
