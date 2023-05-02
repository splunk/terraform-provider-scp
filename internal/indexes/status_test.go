package indexes_test

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

var rateLimitResp = &http.Response{
	StatusCode: http.StatusTooManyRequests,
	Body:       io.NopCloser(bytes.NewReader(nil)),
}

func uint64Ptr(value int64) *uint64 {
	uint64Val := uint64(value)
	return &uint64Val
}

func genIndexResp(code int) *http.Response {

	var b []byte
	if code == http.StatusOK {
		index := v2.IndexResponse{
			Datatype:                    mockDatatype,
			MaxDataSizeMB:               uint64(mockMaxDataSizeMB),
			Name:                        mockIndexName,
			SearchableDays:              uint64(mockSearchableDays),
			SelfStorageBucketPath:       &mockSelfStorageBucketPath,
			SplunkArchivalRetentionDays: nil,
			TotalEventCount:             nil,
			TotalRawSizeMB:              nil,
		}

		b, _ = json.Marshal(&index)
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

func Test_VerifyIndexUpdate(t *testing.T) {
	assert := assert.New(t)

	altMockBucketPath := "s3://some_alt_bucket_path"

	cases := []struct {
		expectedResult bool
		patchRequest   *v2.PatchIndexInfoJSONRequestBody
		indexResponse  *v2.IndexResponse
	}{
		// Test Case 0: Expected true for no fields to update
		{
			true,
			&v2.PatchIndexInfoJSONRequestBody{
				MaxDataSizeMB:               nil,
				SearchableDays:              nil,
				SelfStorageBucketPath:       nil,
				SplunkArchivalRetentionDays: nil,
			},
			&v2.IndexResponse{
				Datatype:                    mockDatatype,
				MaxDataSizeMB:               uint64(mockMaxDataSizeMB),
				Name:                        mockIndexName,
				SearchableDays:              uint64(mockSearchableDays),
				SelfStorageBucketPath:       &mockSelfStorageBucketPath,
				SplunkArchivalRetentionDays: nil,
				TotalEventCount:             nil,
				TotalRawSizeMB:              nil,
			},
		},
		// Test Case 1: Tests complete update for single field update
		{
			true,
			&v2.PatchIndexInfoJSONRequestBody{
				MaxDataSizeMB:               &mockMaxDataSizeMB,
				SearchableDays:              nil,
				SelfStorageBucketPath:       nil,
				SplunkArchivalRetentionDays: nil,
			},
			&v2.IndexResponse{
				Datatype:                    mockDatatype,
				MaxDataSizeMB:               uint64(mockMaxDataSizeMB),
				Name:                        mockIndexName,
				SearchableDays:              uint64(mockSearchableDays),
				SelfStorageBucketPath:       &mockSelfStorageBucketPath,
				SplunkArchivalRetentionDays: nil,
				TotalEventCount:             nil,
				TotalRawSizeMB:              nil,
			},
		},
		// Test Case 2: Tests complete update for all fields updated except searchableDays
		{
			true,
			&v2.PatchIndexInfoJSONRequestBody{
				MaxDataSizeMB:               &mockMaxDataSizeMB,
				SearchableDays:              nil,
				SelfStorageBucketPath:       &mockSelfStorageBucketPath,
				SplunkArchivalRetentionDays: &mockSplunkArchivalRetentionDays,
			},
			&v2.IndexResponse{
				Datatype:                    mockDatatype,
				MaxDataSizeMB:               uint64(mockMaxDataSizeMB),
				Name:                        mockIndexName,
				SearchableDays:              uint64(mockSearchableDays),
				SelfStorageBucketPath:       &mockSelfStorageBucketPath,
				SplunkArchivalRetentionDays: uint64Ptr(mockSplunkArchivalRetentionDays),
				TotalEventCount:             nil,
				TotalRawSizeMB:              nil,
			},
		},
		// Test Case 3: Tests complete update for all fields updated except SplunkArchivalRetentionDays
		{
			true,
			&v2.PatchIndexInfoJSONRequestBody{
				MaxDataSizeMB:               &mockMaxDataSizeMB,
				SearchableDays:              &mockSearchableDays,
				SelfStorageBucketPath:       &mockSelfStorageBucketPath,
				SplunkArchivalRetentionDays: nil,
			},
			&v2.IndexResponse{
				Datatype:                    mockDatatype,
				MaxDataSizeMB:               uint64(mockMaxDataSizeMB),
				Name:                        mockIndexName,
				SearchableDays:              uint64(mockSearchableDays),
				SelfStorageBucketPath:       &mockSelfStorageBucketPath,
				SplunkArchivalRetentionDays: nil,
				TotalEventCount:             nil,
				TotalRawSizeMB:              nil,
			},
		},
		// Test Case 4: Tests incomplete update (nil selfstorage bucket path)
		{
			false,
			&v2.PatchIndexInfoJSONRequestBody{
				MaxDataSizeMB:               &mockMaxDataSizeMB,
				SearchableDays:              &mockSearchableDays,
				SelfStorageBucketPath:       &mockSelfStorageBucketPath,
				SplunkArchivalRetentionDays: nil,
			},
			&v2.IndexResponse{
				Datatype:                    mockDatatype,
				MaxDataSizeMB:               uint64(mockMaxDataSizeMB),
				Name:                        mockIndexName,
				SearchableDays:              uint64(mockSearchableDays),
				SelfStorageBucketPath:       nil,
				SplunkArchivalRetentionDays: nil,
				TotalEventCount:             nil,
				TotalRawSizeMB:              nil,
			},
		},
		// Test Case 5: Tests incomplete update (MaxDataSizeMB not updated)
		{
			false,
			&v2.PatchIndexInfoJSONRequestBody{
				MaxDataSizeMB:               &mockMaxDataSizeMB,
				SearchableDays:              nil,
				SelfStorageBucketPath:       nil,
				SplunkArchivalRetentionDays: nil,
			},
			&v2.IndexResponse{
				Datatype:                    mockDatatype,
				MaxDataSizeMB:               uint64(512),
				Name:                        mockIndexName,
				SearchableDays:              uint64(mockSearchableDays),
				SelfStorageBucketPath:       nil,
				SplunkArchivalRetentionDays: nil,
				TotalEventCount:             nil,
				TotalRawSizeMB:              nil,
			},
		},
		// Test Case 6: Tests incomplete update (searchableDays not updated)
		{
			false,
			&v2.PatchIndexInfoJSONRequestBody{
				MaxDataSizeMB:               &mockMaxDataSizeMB,
				SearchableDays:              &mockSearchableDays,
				SelfStorageBucketPath:       &mockSelfStorageBucketPath,
				SplunkArchivalRetentionDays: nil,
			},
			&v2.IndexResponse{
				Datatype:                    mockDatatype,
				MaxDataSizeMB:               uint64(mockMaxDataSizeMB),
				Name:                        mockIndexName,
				SearchableDays:              uint64(90),
				SelfStorageBucketPath:       nil,
				SplunkArchivalRetentionDays: nil,
				TotalEventCount:             nil,
				TotalRawSizeMB:              nil,
			},
		},
		// Test Case 7: Tests incomplete update (SelfStorageBucketPath not updated)
		{
			false,
			&v2.PatchIndexInfoJSONRequestBody{
				MaxDataSizeMB:               &mockMaxDataSizeMB,
				SearchableDays:              &mockSearchableDays,
				SelfStorageBucketPath:       &mockSelfStorageBucketPath,
				SplunkArchivalRetentionDays: nil,
			},
			&v2.IndexResponse{
				Datatype:                    mockDatatype,
				MaxDataSizeMB:               uint64(mockMaxDataSizeMB),
				Name:                        mockIndexName,
				SearchableDays:              uint64(mockSearchableDays),
				SelfStorageBucketPath:       &altMockBucketPath,
				SplunkArchivalRetentionDays: nil,
				TotalEventCount:             nil,
				TotalRawSizeMB:              nil,
			},
		},
		// Test Case 7: Tests incomplete update (SplunkArchivalRetentionDays not updated)
		{
			false,
			&v2.PatchIndexInfoJSONRequestBody{
				MaxDataSizeMB:               &mockMaxDataSizeMB,
				SearchableDays:              &mockSearchableDays,
				SelfStorageBucketPath:       nil,
				SplunkArchivalRetentionDays: &mockSplunkArchivalRetentionDays,
			},
			&v2.IndexResponse{
				Datatype:                    mockDatatype,
				MaxDataSizeMB:               uint64(mockMaxDataSizeMB),
				Name:                        mockIndexName,
				SearchableDays:              uint64(mockSearchableDays),
				SelfStorageBucketPath:       &mockSelfStorageBucketPath,
				SplunkArchivalRetentionDays: uint64Ptr(400),
				TotalEventCount:             nil,
				TotalRawSizeMB:              nil,
			},
		},
	}
	for i, test := range cases {
		test := test // Capture
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			result := idx.VerifyIndexUpdate(*test.patchRequest, *test.indexResponse)
			assert.Equal(result, test.expectedResult)
		})
	}
}

func Test_ProcessResponse(t *testing.T) {
	assert := assert.New(t)

	/* test nil resp, bad req status text, and error for bad request */
	t.Run("verify returns correct output on bad requests", func(t *testing.T) {
		for _, targetCodes := range [][]string{idx.TargetStatusResourceChange, idx.TargetStatusResourceExists, idx.TargetStatusResourceDeleted} {
			t.Run(fmt.Sprintf("with target codes %v", targetCodes), func(t *testing.T) {
				resp, statusText, err := idx.ProcessResponse(badReqResp, targetCodes, idx.PendingStatusCRUD)
				assert.Nil(resp)
				assert.Equal(http.StatusText(badReqResp.StatusCode), statusText)
				assert.Error(err)
			})
		}
	})

	/* test nil resp returns error */
	t.Run("verify returns correct output on accepted resource change", func(t *testing.T) {
		resp, statusText, err := idx.ProcessResponse(nil, idx.TargetStatusResourceChange, idx.PendingStatusCRUD)
		assert.Nil(resp)
		assert.Error(err)
		assert.Equal(statusText, "")
	})

	/* test non-nil resp, correct status text, and nil error for expected resp */
	t.Run("verify returns correct output on accepted resource change", func(t *testing.T) {
		resp, statusText, err := idx.ProcessResponse(acceptedResp, idx.TargetStatusResourceChange, idx.PendingStatusCRUD)
		assert.NotNil(resp)
		assert.Equal(http.StatusText(acceptedResp.StatusCode), statusText)
		assert.NoError(err)
	})

	t.Run("verify returns correct output on success resource exists after create", func(t *testing.T) {
		resp, statusText, err := idx.ProcessResponse(successRespOk, idx.TargetStatusResourceExists, idx.PendingStatusVerifyCreated)
		assert.NotNil(resp)
		assert.Equal(http.StatusText(successRespOk.StatusCode), statusText)
		assert.NoError(err)
	})

	t.Run("verify returns correct output on resource does not exist after delete", func(t *testing.T) {
		resp, statusText, err := idx.ProcessResponse(notFoundResp, idx.TargetStatusResourceDeleted, idx.PendingStatusVerifyDeleted)
		assert.NotNil(resp)
		assert.Equal(http.StatusText(notFoundResp.StatusCode), statusText)
		assert.NoError(err)
	})
}
