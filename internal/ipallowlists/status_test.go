package ipallowlists_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/ipallowlists"
	"github.com/stretchr/testify/assert"
)

var badReqResp = &http.Response{
	StatusCode: http.StatusBadRequest,
	Body:       io.NopCloser(bytes.NewReader(nil)),
}

var successRespOk = &http.Response{
	StatusCode: http.StatusOK,
	Body:       io.NopCloser(bytes.NewReader(nil)),
}

var rateLimitResp = &http.Response{
	StatusCode: http.StatusTooManyRequests,
	Body:       io.NopCloser(bytes.NewReader(nil)),
}

func Test_ProcessResponse(t *testing.T) {
	assert := assert.New(t)

	/* test nil resp, bad req status text, and error for bad request */
	t.Run("verify returns correct output on bad requests", func(t *testing.T) {
		for _, targetCodes := range [][]string{ipallowlists.TargetStatusResourceChange, ipallowlists.TargetStatusResourceExists, ipallowlists.TargetStatusResourceDeleted} {
			t.Run(fmt.Sprintf("with target codes %v", targetCodes), func(t *testing.T) {
				resp, statusText, err := ipallowlists.ProcessResponse(badReqResp, targetCodes, ipallowlists.PendingStatusCRUD)
				assert.Nil(resp)
				assert.Equal(http.StatusText(badReqResp.StatusCode), statusText)
				assert.Error(err)
			})
		}
	})

	/* test nil resp returns error */
	t.Run("verify returns correct output on success resource change", func(t *testing.T) {
		resp, statusText, err := ipallowlists.ProcessResponse(nil, ipallowlists.TargetStatusResourceChange, ipallowlists.PendingStatusCRUD)
		assert.Nil(resp)
		assert.Error(err)
		assert.Equal(statusText, "")
	})

	t.Run("verify returns correct output on accepted resource change", func(t *testing.T) {
		resp, statusText, err := ipallowlists.ProcessResponse(successRespOk, ipallowlists.TargetStatusResourceChange, ipallowlists.PendingStatusCRUD)
		assert.NotNil(resp)
		assert.Equal(http.StatusText(successRespOk.StatusCode), statusText)
		assert.NoError(err)
	})

}

func getIPAllowlistResponse(statusCode int) *http.Response {

	var b []byte
	if statusCode == http.StatusOK {
		subnets := struct {
			Subnets *[]string
		}{
			Subnets: &mockSubnets,
		}

		b, _ = json.Marshal(&subnets)
	} else {
		b, _ = json.Marshal(&v2.Error{
			Code:    http.StatusText(statusCode),
			Message: http.StatusText(statusCode),
		})
	}
	recorder := httptest.NewRecorder()
	recorder.Header().Add("Content-Type", "json")
	recorder.WriteHeader(statusCode)
	if b != nil {
		_, _ = recorder.Write(b)
	}
	return recorder.Result()
}
