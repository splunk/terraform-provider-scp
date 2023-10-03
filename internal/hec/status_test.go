package hec_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/hec"
	"github.com/splunk/terraform-provider-scp/internal/utils"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	mockHecName = "mock-hec-token"
	mockStack   = "mock-stack"
)

var (
	mockDefaultIndex      = "mock-default-index"
	mockDefaultSource     = "mock-default-source"
	mockDefaultSourceType = "mock-default-source-type"
	mockDisabled          = false
	mockToken             = "mock-token"
	mockUseAck            = false
	mockAllowedIndexes    = []string{"main", "summary"}
	mockDeploymentId      = "mock-id"
	mockStatusRunning     = "running"
	mockStatusSucceeded   = "completed"
	mockStatusFailed      = "failed"
	mockStatusNew         = "new"

	mockUnupdated               = "some-other-value"
	mockUnupdatedAllowedIndexes = []string{"main", "index1"}
	mockUnupdatedBool           = true

	successRespSucceeded = &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(bytes.NewReader(deploymentInfoSucceededBody)),
	}

	successRespFailed = &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(bytes.NewReader(deploymentInfoFailedBody)),
	}

	successRespNew = &http.Response{
		StatusCode: http.StatusAccepted,
		Status:     "202-Accepted",
		Body:       io.NopCloser(bytes.NewReader(deploymentInfoNewBody)),
	}

	successRespRunning = &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(bytes.NewReader(deploymentInfoRunningBody)),
	}

	acceptedResp = &http.Response{
		StatusCode: http.StatusAccepted,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}

	successRespOk = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}

	notFoundResp = &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}

	badReqResp = &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}

	rateLimitResp = &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}

	failedDependencyResp = &http.Response{
		StatusCode: http.StatusFailedDependency,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}

	deploymentInfoRunningBody, _   = json.Marshal(&v2.DeploymentInfo{Id: mockDeploymentId, Status: &mockStatusRunning})
	deploymentInfoSucceededBody, _ = json.Marshal(&v2.DeploymentInfo{Id: mockDeploymentId, Status: &mockStatusSucceeded})
	deploymentInfoFailedBody, _    = json.Marshal(&v2.DeploymentInfo{Id: mockDeploymentId, Status: &mockStatusFailed})
	deploymentInfoNewBody, _       = json.Marshal(&v2.DeploymentInfo{Id: mockDeploymentId, Status: &mockStatusNew})
)

func genHecResp(code int) *http.Response {
	var b []byte
	if code == http.StatusOK {
		hecSpec := v2.HecSpec{
			AllowedIndexes:    &mockAllowedIndexes,
			DefaultIndex:      &mockDefaultIndex,
			DefaultSource:     &mockDefaultSource,
			DefaultSourcetype: nil,
			Disabled:          &mockDisabled,
			Name:              mockHecName,
			Token:             &mockToken,
			UseAck:            &mockUseAck,
		}
		hec := hec.HecBody{HttpEventCollector: &v2.HecInfo{
			Spec:  &hecSpec,
			Token: nil,
		}}

		b, _ = json.Marshal(&hec)
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

func genDeploymentInfoResp(code int, status string) *http.Response {
	var b []byte
	if code == http.StatusOK {
		deploymentInfo := v2.DeploymentInfo{
			Id:     mockDeploymentId,
			Status: &status,
		}

		b, _ = json.Marshal(&deploymentInfo)
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

func Test_VerifyHecUpdate(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		expectedResult bool
		patchRequest   *v2.PatchHECJSONRequestBody
		hecResponse    *v2.HecSpec
	}{
		// Test Case 0: Expected true for no fields to update
		{
			true,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes:    nil,
				DefaultIndex:      nil,
				DefaultSource:     nil,
				DefaultSourcetype: nil,
				Disabled:          nil,
				UseAck:            nil,
			},
			&v2.HecSpec{
				AllowedIndexes:    &mockAllowedIndexes,
				DefaultIndex:      &mockDefaultIndex,
				DefaultSource:     &mockDefaultSource,
				DefaultSourcetype: nil,
				Disabled:          &mockDisabled,
				Name:              mockHecName,
				UseAck:            &mockUseAck,
			},
		},
		// Test Case 1: Tests complete update for single field update
		{
			true,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes: &mockAllowedIndexes,
			},
			&v2.HecSpec{
				AllowedIndexes:    &mockAllowedIndexes,
				DefaultIndex:      &mockDefaultIndex,
				DefaultSource:     &mockDefaultSource,
				DefaultSourcetype: nil,
				Disabled:          &mockDisabled,
				Name:              mockHecName,
				UseAck:            &mockUseAck,
			},
		},
		// Test Case 2: Tests complete update for all fields updated
		{
			true,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes:    &mockAllowedIndexes,
				DefaultIndex:      &mockDefaultIndex,
				DefaultSource:     &mockDefaultSource,
				DefaultSourcetype: &mockDefaultSourceType,
				Disabled:          &mockDisabled,
				Name:              mockHecName,
				UseAck:            &mockUseAck,
			},
			&v2.HecSpec{
				AllowedIndexes:    &mockAllowedIndexes,
				DefaultIndex:      &mockDefaultIndex,
				DefaultSource:     &mockDefaultSource,
				DefaultSourcetype: &mockDefaultSourceType,
				Disabled:          &mockDisabled,
				Name:              mockHecName,
				UseAck:            &mockUseAck,
			},
		},
		// Test Case 3: Tests incomplete update (defaultSource not updated)
		{
			false,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  &mockDefaultSource,
			},
			&v2.HecSpec{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  &mockUnupdated,
				Disabled:       &mockDisabled,
				UseAck:         &mockUseAck,
			},
		},
		// Test Case 4: Tests incomplete update (nil defaultSource)
		{
			false,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  &mockDefaultSource,
			},
			&v2.HecSpec{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  nil,
				Disabled:       &mockDisabled,
				UseAck:         &mockUseAck,
			},
		},
		// Test Case 5: Tests incomplete update (defaultIndex not updated)
		{
			false,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes: nil,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  nil,
			},
			&v2.HecSpec{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockUnupdated,
				DefaultSource:  nil,
				Disabled:       &mockDisabled,
				UseAck:         &mockUseAck,
			},
		},
		// Test Case 6: Tests incomplete update (defaultIndex nil)
		{
			false,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes: nil,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  nil,
			},
			&v2.HecSpec{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   nil,
				DefaultSource:  nil,
				Disabled:       &mockDisabled,
				UseAck:         &mockUseAck,
			},
		},
		// Test Case 7: Tests incomplete update (AllowedIndexes not updated)
		{
			false,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  nil,
			},
			&v2.HecSpec{
				AllowedIndexes: &mockUnupdatedAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  nil,
				Disabled:       &mockDisabled,
				UseAck:         &mockUseAck,
			},
		},
		// Test Case 8: Tests incomplete update (AllowedIndexes nil)
		{
			false,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  nil,
			},
			&v2.HecSpec{
				AllowedIndexes: nil,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  nil,
				Disabled:       &mockDisabled,
				UseAck:         &mockUseAck,
			},
		},
		// Test Case 9: Tests incomplete update (disabled not updated)
		{
			false,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				Disabled:       &mockDisabled,
			},
			&v2.HecSpec{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  nil,
				Disabled:       &mockUnupdatedBool,
				Token:          &mockToken,
				UseAck:         &mockUseAck,
			},
		},
		// Test Case 10: Tests incomplete update (disabled nil)
		{
			false,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				Disabled:       &mockDisabled,
			},
			&v2.HecSpec{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				Disabled:       nil,
				Token:          &mockToken,
				UseAck:         &mockUseAck,
			},
		},
		// Test Case 11: Tests incomplete update (useAck not updated)
		{
			false,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				UseAck:         &mockUseAck,
			},
			&v2.HecSpec{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  nil,
				Disabled:       &mockUnupdatedBool,
				UseAck:         &mockUnupdatedBool,
			},
		},
		// Test Case 12: Tests incomplete update (useAck nil)
		{
			false,
			&v2.PatchHECJSONRequestBody{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				UseAck:         &mockUseAck,
			},
			&v2.HecSpec{
				AllowedIndexes: &mockAllowedIndexes,
				DefaultIndex:   &mockDefaultIndex,
				DefaultSource:  nil,
				Disabled:       &mockUnupdatedBool,
				Token:          &mockToken,
				UseAck:         nil,
			},
		},
		// Test Case 13: Tests incomplete update (defaultSourcetype not updated)
		{
			false,
			&v2.PatchHECJSONRequestBody{
				DefaultSourcetype: &mockDefaultSourceType,
			},
			&v2.HecSpec{
				DefaultSourcetype: &mockUnupdated,
				AllowedIndexes:    &mockAllowedIndexes,
				DefaultIndex:      &mockDefaultIndex,
				DefaultSource:     nil,
				Disabled:          &mockDisabled,
				Token:             &mockToken,
				UseAck:            nil,
			},
		},
		// Test Case 14: Tests incomplete update (defaultSourcetype nil)
		{
			false,
			&v2.PatchHECJSONRequestBody{
				DefaultSourcetype: &mockDefaultSourceType,
			},
			&v2.HecSpec{
				DefaultSourcetype: nil,
				AllowedIndexes:    &mockAllowedIndexes,
				DefaultIndex:      &mockDefaultIndex,
				DefaultSource:     nil,
				Disabled:          &mockDisabled,
				Token:             &mockToken,
				UseAck:            nil,
			},
		},
	}
	for i, test := range cases {
		test := test // Capture
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			result := hec.VerifyHecUpdate(*test.patchRequest, *test.hecResponse)
			assert.Equal(result, test.expectedResult)
		})
	}
}

func Test_TestIsSliceEqual(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		expectedResult bool
		first          *[]string
		second         *[]string
	}{
		// Test Case 0: Expected true for empty
		{
			true,
			&[]string{},
			&[]string{},
		},
		// Test Case 1: Expected true for same array
		{
			true,
			&[]string{"a", "b"},
			&[]string{"a", "b"},
		},
		// Test Case 2: Expected false for different length
		{
			false,
			&[]string{"a"},
			&[]string{"b", "b"},
		},
		// Test Case 3: Expected true for different order
		{
			true,
			&[]string{"a", "b"},
			&[]string{"b", "a"},
		},
		// Test Case 4: Expected false for different entries
		{
			false,
			&[]string{"a", "b"},
			&[]string{"a", "a"},
		},
	}
	for i, test := range cases {
		test := test // Capture
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			result := utils.IsSliceEqual(test.first, test.second)
			assert.Equal(result, test.expectedResult)
		})
	}
}
