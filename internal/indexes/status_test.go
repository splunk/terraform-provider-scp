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
	// nolint
	uint64Val := uint64(value)
	return &uint64Val
}

func genIndexResp(code int) *http.Response {

	var b []byte
	if code == http.StatusOK {
		index := v2.IndexResponse{
			Datatype: mockDatatype,
			// nolint
			MaxDataSizeMB: uint64(mockMaxDataSizeMB),
			Name:          mockIndexName,
			// nolint
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
				Datatype: mockDatatype,
				// nolint
				MaxDataSizeMB: uint64(mockMaxDataSizeMB),
				Name:          mockIndexName,
				// nolint
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
				Datatype: mockDatatype,
				// nolint
				MaxDataSizeMB: uint64(mockMaxDataSizeMB),
				Name:          mockIndexName,
				// nolint
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
				Datatype: mockDatatype,
				// nolint
				MaxDataSizeMB: uint64(mockMaxDataSizeMB),
				Name:          mockIndexName,
				// nolint
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
				Datatype: mockDatatype,
				// nolint
				MaxDataSizeMB: uint64(mockMaxDataSizeMB),
				Name:          mockIndexName,
				// nolint
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
				Datatype: mockDatatype,
				// nolint
				MaxDataSizeMB: uint64(mockMaxDataSizeMB),
				Name:          mockIndexName,
				// nolint
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
				Datatype:      mockDatatype,
				MaxDataSizeMB: uint64(512),
				Name:          mockIndexName,
				// nolint
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
				Datatype: mockDatatype,
				// nolint
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
				Datatype: mockDatatype,
				// nolint
				MaxDataSizeMB: uint64(mockMaxDataSizeMB),
				Name:          mockIndexName,
				// nolint
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
				Datatype: mockDatatype,
				// nolint
				MaxDataSizeMB: uint64(mockMaxDataSizeMB),
				Name:          mockIndexName,
				// nolint
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
		t.Run(fmt.Sprintf("case %d", i), func(_ *testing.T) {
			result := idx.VerifyIndexUpdate(*test.patchRequest, *test.indexResponse)
			assert.Equal(result, test.expectedResult)
		})
	}
}
