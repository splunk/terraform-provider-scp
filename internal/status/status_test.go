package status_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	idx "github.com/splunk/terraform-provider-scp/internal/indexes"
	"github.com/splunk/terraform-provider-scp/internal/status"
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

func Test_IsStatusCodeRetryable(t *testing.T) {
	assert := assert.New(t)

	/* test general retryable error status code returns true */
	t.Run("verify returns true on general retryable codes", func(t *testing.T) {
		for generalCode, _ := range idx.GeneralRetryableStatusCodes {
			t.Run(fmt.Sprintf("with general code %d", generalCode), func(t *testing.T) {
				assert.True(status.IsStatusCodeExpected(generalCode, idx.TargetStatusResourceChange, idx.PendingStatusCRUD))

			})
		}
	})

	/* test target status codes input return true when present */
	t.Run("verify returns true on target status resource change code", func(t *testing.T) {
		assert.True(status.IsStatusCodeExpected(acceptedResp.StatusCode, idx.TargetStatusResourceChange, idx.PendingStatusCRUD))
	})

	t.Run("verify returns true on target status resource exists code", func(t *testing.T) {
		assert.True(status.IsStatusCodeExpected(successRespOk.StatusCode, idx.TargetStatusResourceExists, idx.PendingStatusVerifyCreated))
	})

	t.Run("verify returns true on target status resource deleted code", func(t *testing.T) {
		assert.True(status.IsStatusCodeExpected(notFoundResp.StatusCode, idx.TargetStatusResourceDeleted, idx.PendingStatusVerifyDeleted))
	})

	/* test return false when status code absent in both */
	t.Run("verify returns false on non target and non retryable code", func(t *testing.T) {
		assert.False(status.IsStatusCodeExpected(badReqResp.StatusCode, idx.TargetStatusResourceDeleted, idx.PendingStatusVerifyDeleted))
	})
}

func Test_ProcessResponse(t *testing.T) {
	assert := assert.New(t)

	/* test nil resp, bad req status text, and error for bad request */
	t.Run("verify returns correct output on bad requests", func(t *testing.T) {
		for _, targetCodes := range [][]string{idx.TargetStatusResourceChange, idx.TargetStatusResourceExists, idx.TargetStatusResourceDeleted} {
			t.Run(fmt.Sprintf("with target codes %v", targetCodes), func(t *testing.T) {
				resp, statusText, err := status.ProcessResponse(badReqResp, targetCodes, idx.PendingStatusCRUD)
				assert.Nil(resp)
				assert.Equal(http.StatusText(badReqResp.StatusCode), statusText)
				assert.Error(err)
			})
		}
	})

	/* test nil resp returns error */
	t.Run("verify returns correct output on accepted resource change", func(t *testing.T) {
		resp, statusText, err := status.ProcessResponse(nil, idx.TargetStatusResourceChange, idx.PendingStatusCRUD)
		assert.Nil(resp)
		assert.Error(err)
		assert.Equal(statusText, "")
	})

	/* test non-nil resp, correct status text, and nil error for expected resp */
	t.Run("verify returns correct output on accepted resource change", func(t *testing.T) {
		resp, statusText, err := status.ProcessResponse(acceptedResp, idx.TargetStatusResourceChange, idx.PendingStatusCRUD)
		assert.NotNil(resp)
		assert.Equal(http.StatusText(acceptedResp.StatusCode), statusText)
		assert.NoError(err)
	})

	t.Run("verify returns correct output on success resource exists after create", func(t *testing.T) {
		resp, statusText, err := status.ProcessResponse(successRespOk, idx.TargetStatusResourceExists, idx.PendingStatusVerifyCreated)
		assert.NotNil(resp)
		assert.Equal(http.StatusText(successRespOk.StatusCode), statusText)
		assert.NoError(err)
	})

	t.Run("verify returns correct output on resource does not exist after delete", func(t *testing.T) {
		resp, statusText, err := status.ProcessResponse(notFoundResp, idx.TargetStatusResourceDeleted, idx.PendingStatusVerifyDeleted)
		assert.NotNil(resp)
		assert.Equal(http.StatusText(notFoundResp.StatusCode), statusText)
		assert.NoError(err)
	})
}
