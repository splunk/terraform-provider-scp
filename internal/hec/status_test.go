package hec_test

import (
	"bytes"
	"encoding/json"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"io"
	"net/http"
	"net/http/httptest"
)

const (
	mockHecName = "mock-hec-token"
	mockStack   = "mock-stack"
)

var (
	mockDefaultIndex               = "mock-default-index"
	mockDefaultSource              = "mock-default-source"
	mockDefaultSourceType          = "mock-default-source-type"
	mockDisabled                   = false
	mockDefaultHost                = "mock-default-host"
	mockToken                      = "mock-token"
	mockUseAck                     = false
	mockAllowedIndexes    []string = nil
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

var rateLimitResp = &http.Response{
	StatusCode: http.StatusTooManyRequests,
	Body:       io.NopCloser(bytes.NewReader(nil)),
}

func genHecResp(code int) *http.Response {
	var b []byte
	if code == http.StatusOK {
		hec := v2.HecSpec{
			AllowedIndexes:    &mockAllowedIndexes,
			DefaultHost:       &mockDefaultHost,
			DefaultIndex:      &mockDefaultIndex,
			DefaultSource:     &mockDefaultSource,
			DefaultSourcetype: nil,
			Disabled:          &mockDisabled,
			Name:              mockHecName,
			Token:             &mockToken,
			UseAck:            &mockUseAck,
		}

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
