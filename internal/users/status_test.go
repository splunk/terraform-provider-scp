package users_test

import (
	"fmt"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/users"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	mockDefaultApp      = "mock_default_app"
	mockEmail           = "mock_email"
	mockForceChangePass = true
	mockFullName        = "mock_full_name"
	mockOldPassword     = "mock_old_password"
	mockPassword        = "mock_password"
	mockRoles           = []string{"user", "power"}
	mockName            = "mock_name"
)

func Test_VerifyUserUpdate(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		expectedResult bool
		patchRequest   *v2.PatchUserJSONRequestBody
		userInfo       *v2.UsersResponse
	}{
		// Test Case 0: Expected true for no fields to update
		{
			true,
			&v2.PatchUserJSONRequestBody{
				DefaultApp:      nil,
				Email:           nil,
				ForceChangePass: nil,
				FullName:        nil,
				OldPassword:     nil,
				Password:        nil,
				Roles:           nil,
			},
			&v2.UsersResponse{
				DefaultApp: mockDefaultApp,
				Email:      mockEmail,
				FullName:   mockFullName,
				Roles:      mockRoles,
				Name:       mockName,
			},
		},
		// Test Case 1: Expected true for all fields update
		{
			true,
			&v2.PatchUserJSONRequestBody{
				DefaultApp:      &mockDefaultApp,
				Email:           &mockEmail,
				ForceChangePass: &mockForceChangePass,
				FullName:        &mockFullName,
				OldPassword:     &mockOldPassword,
				Password:        &mockPassword,
				Roles:           &mockRoles,
			},
			&v2.UsersResponse{
				DefaultApp: mockDefaultApp,
				Email:      mockEmail,
				FullName:   mockFullName,
				Roles:      mockRoles,
				Name:       mockName,
			},
		},
		// Test Case 2: Expected true for single field update
		{
			true,
			&v2.PatchUserJSONRequestBody{
				Email: &mockEmail,
			},
			&v2.UsersResponse{
				DefaultApp: mockDefaultApp,
				Email:      mockEmail,
				FullName:   mockFullName,
				Roles:      mockRoles,
				Name:       mockName,
			},
		},
		// Test Case 3: Expected true for single field update
		{
			true,
			&v2.PatchUserJSONRequestBody{
				DefaultApp: &mockDefaultApp,
			},
			&v2.UsersResponse{
				DefaultApp: mockDefaultApp,
				Email:      mockEmail,
				FullName:   mockFullName,
				Roles:      mockRoles,
				Name:       mockName,
			},
		},
		// Test Case 4: Expected true for single field update
		{
			true,
			&v2.PatchUserJSONRequestBody{
				Roles: &mockRoles,
			},
			&v2.UsersResponse{
				DefaultApp: mockDefaultApp,
				Email:      mockEmail,
				FullName:   mockFullName,
				Roles:      mockRoles,
				Name:       mockName,
			},
		},
		// Test Case 5: Expected true for single field update
		{
			true,
			&v2.PatchUserJSONRequestBody{
				FullName: &mockFullName,
			},
			&v2.UsersResponse{
				DefaultApp: mockDefaultApp,
				Email:      mockEmail,
				FullName:   mockFullName,
				Roles:      mockRoles,
				Name:       mockName,
			},
		},
		// Test Case 6: Expected false for uncompleted field update
		{
			false,
			&v2.PatchUserJSONRequestBody{
				Email: &mockEmail,
			},
			&v2.UsersResponse{
				DefaultApp: mockDefaultApp,
				Email:      "",
				FullName:   mockFullName,
				Roles:      mockRoles,
				Name:       mockName,
			},
		},
		// Test Case 7: Expected false for uncompleted field update
		{
			false,
			&v2.PatchUserJSONRequestBody{
				DefaultApp: &mockDefaultApp,
			},
			&v2.UsersResponse{
				DefaultApp: "",
				Email:      mockEmail,
				FullName:   mockFullName,
				Roles:      mockRoles,
				Name:       mockName,
			},
		},
		// Test Case 8: Expected false for uncompleted field update
		{
			false,
			&v2.PatchUserJSONRequestBody{
				Roles: &mockRoles,
			},
			&v2.UsersResponse{
				DefaultApp: mockDefaultApp,
				Email:      mockEmail,
				FullName:   mockFullName,
				Roles:      nil,
				Name:       mockName,
			},
		},
		// Test Case 9: Expected false for uncompleted field update
		{
			false,
			&v2.PatchUserJSONRequestBody{
				FullName: &mockFullName,
			},
			&v2.UsersResponse{
				DefaultApp: mockDefaultApp,
				Email:      mockEmail,
				FullName:   "",
				Roles:      mockRoles,
				Name:       mockName,
			},
		},
	}
	for i, test := range cases {
		test := test // Capture
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			result := users.VerifyUserUpdate(*test.patchRequest, *test.userInfo)
			assert.Equal(test.expectedResult, result)
		})
	}
}
