package roles_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/acs/v2/mocks"
	"github.com/splunk/terraform-provider-scp/internal/roles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	mockAck                   = "Y"
	unexpectedStatusCodes     = []int{400, 401, 403, 404, 409, 501, 503}
	unexpectedStatusCodesPoll = []int{400, 401, 403, 409, 501, 500, 503}
)

const (
	mockRoleName = "mock-user"
	mockStack    = "mock-stack"
)

func Test_WaitRoleCreate(t *testing.T) {
	client := &mocks.ClientInterface{}
	mockCreateParam := v2.CreateRoleParams{
		FederatedSearchManageAck: (*v2.FederatedSearchManage)(&mockAck),
	}

	mockCreateBody := v2.CreateRoleJSONRequestBody{
		Name: mockRoleName,
	}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("CreateRole", mock.Anything, v2.Stack(mockStack), &mockCreateParam, mockCreateBody).Return(nil, errors.New("some error")).Once()
		err := roles.WaitRoleCreate(context.TODO(), client, mockStack, mockCreateParam, mockCreateBody)
		assert.Error(t, err)
	})

	t.Run("with http response 200", func(t *testing.T) {
		client.On("CreateRole", mock.Anything, v2.Stack(mockStack), &mockCreateParam, mockCreateBody).Return(generateResponse(200), nil).Once()
		err := roles.WaitRoleCreate(context.TODO(), client, mockStack, mockCreateParam, mockCreateBody)
		assert.NoError(t, err)
	})

	t.Run("with unexpected http responses", func(t *testing.T) {
		for _, unexpectedStatusCode := range unexpectedStatusCodes {
			t.Run(fmt.Sprintf("with unexpected status %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("CreateRole", mock.Anything, v2.Stack(mockStack), &mockCreateParam, mockCreateBody).Return(generateResponse(unexpectedStatusCode), nil).Once()
				err := roles.WaitRoleCreate(context.TODO(), client, mockStack, mockCreateParam, mockCreateBody)
				assert.Error(t, err)
			})
		}
	})
}

func generateResponse(code int) *http.Response {
	var b []byte
	if code == http.StatusOK {

		role := v2.RolesResponse{
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
		}

		b, _ = json.Marshal(&role)
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

func Test_WaitRoleRead(t *testing.T) {
	client := &mocks.ClientInterface{}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("DescribeRole", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName)).Return(nil, errors.New("some error")).Once()
		user, err := roles.WaitRoleRead(context.TODO(), client, mockStack, mockRoleName)
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("with http 200 response", func(t *testing.T) {
		client.On("DescribeRole", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName)).Return(generateResponse(http.StatusOK), nil).Once()
		user, err := roles.WaitRoleRead(context.TODO(), client, mockStack, mockRoleName)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotNil(t, user.Name)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range unexpectedStatusCodes {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DescribeRole", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName)).Return(generateResponse(unexpectedStatusCode), nil).Once()
				index, err := roles.WaitRoleRead(context.TODO(), client, mockStack, mockRoleName)
				assert.Error(t, err)
				assert.Nil(t, index)
			})
		}
	})
}

func Test_WaitRoleUpdate(t *testing.T) {
	client := &mocks.ClientInterface{}
	mockUpdateParam := v2.PatchRoleInfoParams{
		FederatedSearchManageAck: (*v2.FederatedSearchManage)(&mockAck),
	}
	mockUpdateBody := v2.PatchRoleInfoJSONRequestBody{
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
	}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("PatchRoleInfo", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName), &mockUpdateParam, mockUpdateBody).Return(nil, errors.New("some error")).Once()
		err := roles.WaitRoleUpdate(context.TODO(), client, mockStack, mockUpdateParam, mockUpdateBody, mockRoleName)
		assert.Error(t, err)
	})

	t.Run("with http response 200", func(t *testing.T) {
		client.On("PatchRoleInfo", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName), &mockUpdateParam, mockUpdateBody).Return(generateResponse(http.StatusOK), nil).Once()
		err := roles.WaitRoleUpdate(context.TODO(), client, mockStack, mockUpdateParam, mockUpdateBody, mockRoleName)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range unexpectedStatusCodes {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("PatchRoleInfo", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName), &mockUpdateParam, mockUpdateBody).Return(generateResponse(unexpectedStatusCode), nil).Once()
				err := roles.WaitRoleUpdate(context.TODO(), client, mockStack, mockUpdateParam, mockUpdateBody, mockRoleName)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitVerifyRoleUpdate(t *testing.T) {
	mockUpdateBody := v2.PatchRoleInfoJSONRequestBody{
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
	}

	t.Run("with some client interface error", func(t *testing.T) {
		client := &mocks.ClientInterface{}
		client.On("DescribeRole", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName)).Return(nil, errors.New("some error")).Once()
		err := roles.WaitVerifyRoleUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockRoleName)
		assert.Error(t, err)
	})

	t.Run("with http 200 response", func(t *testing.T) {
		client := &mocks.ClientInterface{}
		client.On("DescribeRole", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName)).Return(generateResponse(http.StatusOK), nil).Once()
		err := roles.WaitVerifyRoleUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockRoleName)
		assert.NoError(t, err)
	})

	t.Run("with non updated response first", func(t *testing.T) {
		client := &mocks.ClientInterface{}
		client.On("DescribeRole", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName)).Return(generateResponse(http.StatusOK), nil).Once()
		err := roles.WaitVerifyRoleUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockRoleName)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		client := &mocks.ClientInterface{}
		for _, unexpectedStatusCode := range unexpectedStatusCodesPoll {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DescribeRole", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName)).Return(generateResponse(unexpectedStatusCode), nil).Once()
				err := roles.WaitVerifyRoleUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockRoleName)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitRoleDelete(t *testing.T) {
	client := &mocks.ClientInterface{}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("DeleteRole", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName)).Return(nil, errors.New("some error")).Once()
		err := roles.WaitRoleDelete(context.TODO(), client, mockStack, mockRoleName)
		assert.Error(t, err)
	})

	t.Run("with http response 202", func(t *testing.T) {
		client.On("DeleteRole", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName)).Return(generateResponse(http.StatusOK), nil).Once()
		err := roles.WaitRoleDelete(context.TODO(), client, mockStack, mockRoleName)
		assert.NoError(t, err)
	})

	t.Run("with retry on rate limit", func(t *testing.T) {
		client.On("DeleteRole", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName)).Return(generateResponse(http.StatusTooManyRequests), nil).Once()
		client.On("DeleteRole", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName)).Return(generateResponse(http.StatusOK), nil).Once()
		err := roles.WaitRoleDelete(context.TODO(), client, mockStack, mockRoleName)
		assert.NoError(t, err)
	})

	t.Run("with unexpected error resp", func(t *testing.T) {
		for _, unexpectedStatusCode := range unexpectedStatusCodes {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DeleteRole", mock.Anything, v2.Stack(mockStack), v2.RoleName(mockRoleName)).Return(generateResponse(unexpectedStatusCode), nil).Once()
				err := roles.WaitRoleDelete(context.TODO(), client, mockStack, mockRoleName)
				assert.Error(t, err)
			})
		}
	})
}
