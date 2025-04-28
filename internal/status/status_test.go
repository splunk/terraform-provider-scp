package status_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	idx "github.com/splunk/terraform-provider-scp/internal/indexes"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"github.com/splunk/terraform-provider-scp/internal/wait"
	"github.com/stretchr/testify/assert"
)

var acceptedResp = &http.Response{
	StatusCode: http.StatusAccepted,
	Body:       io.NopCloser(bytes.NewReader(nil)),
}

var successRespOk = &http.Response{
	StatusCode: http.StatusOK,
	Body:       io.NopCloser(bytes.NewReader(nil)),
}

var notFoundResp = &http.Response{
	StatusCode: http.StatusNotFound,
	Body:       io.NopCloser(bytes.NewReader(nil)),
}

var badReqResp = &http.Response{
	StatusCode: http.StatusBadRequest,
	Body:       io.NopCloser(bytes.NewReader(nil)),
}

func genErrResp(code int, message string) *http.Response {
	var b []byte

	b, _ = json.Marshal(&v2.Error{
		Code:    http.StatusText(code),
		Message: message,
	})

	recorder := httptest.NewRecorder()
	recorder.Header().Add("Content-Type", "json")
	recorder.WriteHeader(code)
	if b != nil {
		_, _ = recorder.Write(b)
	}
	return recorder.Result()
}

func Test_IsStatusCodeRetryable(t *testing.T) {
	assert := assert.New(t)

	/* test general retryable error status code returns true */
	t.Run("verify returns true on general retryable codes", func(t *testing.T) {
		for generalCode := range idx.GeneralRetryableStatusCodes {
			t.Run(fmt.Sprintf("with general code %d", generalCode), func(_ *testing.T) {
				assert.True(status.IsStatusCodeExpected(generalCode, wait.TargetStatusResourceChange, wait.PendingStatusCRUD))

			})
		}
	})

	/* test target status codes input return true when present */
	t.Run("verify returns true on target status resource change code", func(_ *testing.T) {
		assert.True(status.IsStatusCodeExpected(acceptedResp.StatusCode, wait.TargetStatusResourceChange, wait.PendingStatusCRUD))
	})

	t.Run("verify returns true on target status resource exists code", func(_ *testing.T) {
		assert.True(status.IsStatusCodeExpected(successRespOk.StatusCode, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated))
	})

	t.Run("verify returns true on target status resource deleted code", func(_ *testing.T) {
		assert.True(status.IsStatusCodeExpected(notFoundResp.StatusCode, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted))
	})

	/* test return false when status code absent in both */
	t.Run("verify returns false on non target and non retryable code", func(_ *testing.T) {
		assert.False(status.IsStatusCodeExpected(badReqResp.StatusCode, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted))
	})
}

func Test_ProcessResponse(t *testing.T) {
	assert := assert.New(t)

	/* test nil resp, bad req status text, and error for bad request */
	t.Run("verify returns correct output on bad requests", func(t *testing.T) {
		for _, targetCodes := range [][]string{wait.TargetStatusResourceChange, wait.TargetStatusResourceExists, wait.TargetStatusResourceDeleted} {
			t.Run(fmt.Sprintf("with target codes %v", targetCodes), func(_ *testing.T) {
				resp, statusText, err := status.ProcessResponse(badReqResp, targetCodes, wait.PendingStatusCRUD)
				assert.Nil(resp)
				assert.Equal(http.StatusText(badReqResp.StatusCode), statusText)
				assert.Error(err)
			})
		}
	})

	/* test nil resp returns error */
	t.Run("verify returns correct output on accepted resource change", func(_ *testing.T) {
		resp, statusText, err := status.ProcessResponse(nil, wait.TargetStatusResourceChange, wait.PendingStatusCRUD)
		assert.Nil(resp)
		assert.Error(err)
		assert.Equal(statusText, "")
	})

	/* test non-nil resp, correct status text, and nil error for expected resp */
	t.Run("verify returns correct output on accepted resource change", func(_ *testing.T) {
		resp, statusText, err := status.ProcessResponse(acceptedResp, wait.TargetStatusResourceChange, wait.PendingStatusCRUD)
		assert.NotNil(resp)
		assert.Equal(http.StatusText(acceptedResp.StatusCode), statusText)
		assert.NoError(err)
	})

	t.Run("verify returns correct output on success resource exists after create", func(_ *testing.T) {
		resp, statusText, err := status.ProcessResponse(successRespOk, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
		assert.NotNil(resp)
		assert.Equal(http.StatusText(successRespOk.StatusCode), statusText)
		assert.NoError(err)
	})

	t.Run("verify returns correct output on resource does not exist after delete", func(_ *testing.T) {
		resp, statusText, err := status.ProcessResponse(notFoundResp, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
		assert.NotNil(resp)
		assert.Equal(http.StatusText(notFoundResp.StatusCode), statusText)
		assert.NoError(err)
	})

	t.Run(fmt.Sprintf("%s results in error", status.ErrFailedDependency), func(_ *testing.T) {
		errResp := genErrResp(http.StatusFailedDependency, status.ErrFailedDependency)
		resp, statusText, err := status.ProcessResponse(errResp, wait.TargetStatusResourceChange, wait.PendingStatusCRUD)
		assert.Nil(resp)
		assert.Equal(http.StatusText(errResp.StatusCode), statusText)
		assert.Error(err)
	})

	t.Run(fmt.Sprintf("%s does not result in error", status.ErrDependencyIncomplete), func(_ *testing.T) {
		errResp := genErrResp(http.StatusFailedDependency, status.ErrDependencyIncomplete)
		resp, statusText, err := status.ProcessResponse(errResp, wait.TargetStatusResourceChange, wait.PendingStatusCRUD)
		assert.NotNil(resp)
		assert.Equal(http.StatusText(errResp.StatusCode), statusText)
		assert.NoError(err)
	})
}
