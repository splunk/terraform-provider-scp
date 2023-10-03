package roles_test

import (
	"fmt"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/roles"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	mockCumulativeRTSrchJobsQuota = 100
	mockCumulativeSrchJobsQuota   = 200
	mockDefaultApp                = "mock_default_app"
	mockImportedRoles             = []string{"user", "power"}
	mockRoleCapabilities          = []string{"mock_capability_1"}
	mockRtSrchJobsQuota           = 300
	mockSrchDiskQuota             = 400
	mockSrchFilter                = ""
	mockSrchIndexesAllowed        = []string{"index-1", "index-2"}
	mockSrchIndexesDefault        = []string{"index-1", "index-2"}
	mockSrchJobsQuota             = 500
	mockSrchTimeEarliest          = 1000
	mockSrchTimeWin               = 2000
	mockUpdatedIntValue           = -1
	mockUpdatedStringValue        = "-1"
	mockUpdatedListValue          = []string{"-1", "-2"}
)

func Test_VerifyRoleUpdate(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		expectedResult bool
		patchRequest   *v2.PatchRoleInfoJSONRequestBody
		roleResponse   *v2.RolesResponse
	}{
		// Test Case 0: Expected true for no fields to update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				CumulativeRTSrchJobsQuota: nil,
				CumulativeSrchJobsQuota:   nil,
				DefaultApp:                nil,
				ImportedRoles:             nil,
				RolesInfo: v2.RolesInfo{
					Capabilities:       nil,
					RtSrchJobsQuota:    nil,
					SrchDiskQuota:      nil,
					SrchFilter:         nil,
					SrchIndexesAllowed: nil,
					SrchIndexesDefault: nil,
					SrchJobsQuota:      nil,
					SrchTimeEarliest:   nil,
					SrchTimeWin:        nil,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 1: Expected true for all fields update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				ImportedRoles:             &mockImportedRoles,
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 2: Expected true for single field update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 3: Expected true for single field update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				CumulativeSrchJobsQuota: &mockCumulativeSrchJobsQuota,
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 4: Expected true for single field update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				DefaultApp: &mockDefaultApp,
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 5: Expected true for single field update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				ImportedRoles: &mockImportedRoles,
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 6: Expected true for single field update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					RtSrchJobsQuota: &mockRtSrchJobsQuota,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 7: Expected true for single field update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					SrchJobsQuota: &mockSrchJobsQuota,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 8: Expected true for single field update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					SrchDiskQuota: &mockSrchDiskQuota,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 9: Expected true for single field update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					SrchFilter: &mockSrchFilter,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 10: Expected true for single field update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					SrchJobsQuota: &mockSrchJobsQuota,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 11: Expected true for single field update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					SrchTimeEarliest: &mockSrchTimeEarliest,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 12: Expected true for single field update
		{
			true,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					SrchTimeWin: &mockSrchTimeWin,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 13: Expected false for uncompleted field update
		{
			false,
			&v2.PatchRoleInfoJSONRequestBody{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockUpdatedIntValue, // different from input
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 14: Expected false for uncompleted field update
		{
			false,
			&v2.PatchRoleInfoJSONRequestBody{
				CumulativeSrchJobsQuota: &mockCumulativeSrchJobsQuota,
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockUpdatedIntValue, // different from input
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 15: Expected false for uncompleted field update
		{
			false,
			&v2.PatchRoleInfoJSONRequestBody{
				DefaultApp: &mockDefaultApp,
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockUpdatedStringValue, // different from input
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		//Test Case 16: Expected false for uncompleted field update
		{
			false,
			&v2.PatchRoleInfoJSONRequestBody{
				ImportedRoles: &mockImportedRoles,
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockUpdatedListValue, // different from input
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 17: Expected false for uncompleted field update
		{
			false,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					RtSrchJobsQuota: &mockRtSrchJobsQuota,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockUpdatedIntValue, // different from input
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 18: Expected false for uncompleted field update
		{
			false,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					SrchDiskQuota: &mockSrchDiskQuota,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockUpdatedIntValue, // different from input
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 19: Expected false for uncompleted field update
		{
			false,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					SrchFilter: &mockSrchFilter,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockUpdatedStringValue, // different from input
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 20: Expected false for uncompleted field update
		{
			false,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					SrchJobsQuota: &mockSrchJobsQuota,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockUpdatedIntValue, // different from input
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 21: Expected false for uncompleted field update
		{
			false,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					SrchTimeEarliest: &mockSrchTimeEarliest,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockUpdatedIntValue, // different from input
					SrchTimeWin:        &mockSrchTimeWin,
				},
			},
		},
		// Test Case 22: Expected false for uncompleted field update
		{
			false,
			&v2.PatchRoleInfoJSONRequestBody{
				RolesInfo: v2.RolesInfo{
					SrchTimeWin: &mockSrchTimeWin,
				},
			},
			&v2.RolesResponse{
				CumulativeRTSrchJobsQuota: &mockCumulativeRTSrchJobsQuota,
				CumulativeSrchJobsQuota:   &mockCumulativeSrchJobsQuota,
				DefaultApp:                &mockDefaultApp,
				Imported: &v2.ImportedRolesInfo{
					Roles: &mockImportedRoles,
				},
				RolesInfo: v2.RolesInfo{
					Capabilities:       &mockRoleCapabilities,
					RtSrchJobsQuota:    &mockRtSrchJobsQuota,
					SrchDiskQuota:      &mockSrchDiskQuota,
					SrchFilter:         &mockSrchFilter,
					SrchIndexesAllowed: &mockSrchIndexesAllowed,
					SrchIndexesDefault: &mockSrchIndexesDefault,
					SrchJobsQuota:      &mockSrchJobsQuota,
					SrchTimeEarliest:   &mockSrchTimeEarliest,
					SrchTimeWin:        &mockUpdatedIntValue, // different from input
				},
			},
		},
	}
	for i, test := range cases {
		test := test // Capture
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			result := roles.VerifyRoleUpdate(*test.patchRequest, *test.roleResponse)
			compare := assert.Equal(test.expectedResult, result)
			if compare == false {
				print()
			}
		})
	}
}
