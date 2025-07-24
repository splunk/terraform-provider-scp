package indexes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/acs/v2/mocks"
	idx "github.com/splunk/terraform-provider-scp/internal/indexes"
	"github.com/splunk/terraform-provider-scp/internal/wait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	mockIndexName = "mock-index"
	mockStack     = "mock-stack"
)

var (
	mockDatatype                          = "mock-data-type"
	mockMaxDataSizeMB               int64 = 1024
	mockSearchableDays              int64 = 10
	mockSelfStorageBucketPath             = "s3://some_bucket_path"
	mockSplunkArchivalRetentionDays int64 = 1099
)

func Test_WaitIndexCreate(t *testing.T) {
	client := &mocks.ClientInterface{}

	mockCreateBody := v2.CreateIndexJSONRequestBody{
		Datatype:       &mockDatatype,
		Name:           mockIndexName,
		MaxDataSizeMB:  &mockMaxDataSizeMB,
		SearchableDays: &mockSearchableDays,
	}

	t.Run("with some client interface error", func(_ *testing.T) {
		client.On("CreateIndex", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(nil, errors.New("some error")).Once()
		err := idx.WaitIndexCreate(context.TODO(), client, mockStack, mockCreateBody)
		assert.Error(t, err)
	})

	t.Run("with http response 202", func(t *testing.T) {
		client.On("CreateIndex", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(acceptedResp, nil).Once()
		err := idx.WaitIndexCreate(context.TODO(), client, mockStack, mockCreateBody)
		assert.NoError(t, err)
	})

	t.Run("with retryable response 429", func(t *testing.T) {
		client.On("CreateIndex", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(rateLimitResp, nil).Once()
		client.On("CreateIndex", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(acceptedResp, nil).Once()
		err := idx.WaitIndexCreate(context.TODO(), client, mockStack, mockCreateBody)
		assert.NoError(t, err)
	})

	t.Run("with unexpected http responses", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 404, 409, 501} {
			t.Run(fmt.Sprintf("with unexpected status %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("CreateIndex", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(genIndexResp(unexpectedStatusCode), nil).Once()
				err := idx.WaitIndexCreate(context.TODO(), client, mockStack, mockCreateBody)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitIndexPoll(t *testing.T) {
	client := &mocks.ClientInterface{}

	t.Run("with some client interface error", func(_ *testing.T) {
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(nil, errors.New("some error")).Once()
		err := idx.WaitIndexPoll(context.TODO(), client, mockStack, mockIndexName, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
		assert.Error(t, err)
	})

	/* Test Poll to Verify Creation */
	t.Run("with http response 200", func(t *testing.T) {
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(successRespOk, nil).Once()
		err := idx.WaitIndexPoll(context.TODO(), client, mockStack, mockIndexName, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
		assert.NoError(t, err)
	})

	t.Run("with retryable response 404 verify create", func(t *testing.T) {
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(rateLimitResp, nil).Once()
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(successRespOk, nil).Once()
		err := idx.WaitIndexPoll(context.TODO(), client, mockStack, mockIndexName, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 409, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(genIndexResp(unexpectedStatusCode), nil).Once()
				err := idx.WaitIndexPoll(context.TODO(), client, mockStack, mockIndexName, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
				assert.Error(t, err)
			})
		}
	})

	/* Test Poll to Verify Deletion */
	t.Run("with expected http response 404", func(t *testing.T) {
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(notFoundResp, nil).Once()
		err := idx.WaitIndexPoll(context.TODO(), client, mockStack, mockIndexName, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
		assert.NoError(t, err)
	})

	t.Run("with retryable response 200", func(t *testing.T) {
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(successRespOk, nil).Once()
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(notFoundResp, nil).Once()
		err := idx.WaitIndexPoll(context.TODO(), client, mockStack, mockIndexName, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 409, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(genIndexResp(unexpectedStatusCode), nil).Once()
				err := idx.WaitIndexPoll(context.TODO(), client, mockStack, mockIndexName, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitIndexRead(t *testing.T) {
	client := &mocks.ClientInterface{}

	t.Run("with some client interface error", func(_ *testing.T) {
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(nil, errors.New("some error")).Once()
		index, err := idx.WaitIndexRead(context.TODO(), client, mockStack, mockIndexName)
		assert.Error(t, err)
		assert.Nil(t, index)
	})

	t.Run("with http 200 response", func(t *testing.T) {
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(genIndexResp(200), nil).Once()
		index, err := idx.WaitIndexRead(context.TODO(), client, mockStack, mockIndexName)
		assert.NoError(t, err)
		assert.NotNil(t, index)
		assert.NotNil(t, index.Name)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 404, 409, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(genIndexResp(unexpectedStatusCode), nil).Once()
				index, err := idx.WaitIndexRead(context.TODO(), client, mockStack, mockIndexName)
				assert.Error(t, err)
				assert.Nil(t, index)
			})
		}
	})
}

func Test_WaitIndexUpdate(t *testing.T) {
	client := &mocks.ClientInterface{}

	mockUpdateBody := v2.PatchIndexInfoJSONRequestBody{
		MaxDataSizeMB:  &mockMaxDataSizeMB,
		SearchableDays: &mockSearchableDays,
	}

	t.Run("with some client interface error", func(_ *testing.T) {
		client.On("PatchIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName), mockUpdateBody).Return(nil, errors.New("some error")).Once()
		err := idx.WaitIndexUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockIndexName)
		assert.Error(t, err)
	})

	t.Run("with http response 202", func(t *testing.T) {
		client.On("PatchIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName), mockUpdateBody).Return(acceptedResp, nil).Once()
		err := idx.WaitIndexUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockIndexName)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 404, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("PatchIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName), mockUpdateBody).Return(genIndexResp(unexpectedStatusCode), nil).Once()
				err := idx.WaitIndexUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockIndexName)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitVerifyIndexUpdate(t *testing.T) {
	client := &mocks.ClientInterface{}

	mockUpdateBody := v2.PatchIndexInfoJSONRequestBody{
		MaxDataSizeMB:               &mockMaxDataSizeMB,
		SearchableDays:              &mockSearchableDays,
		SelfStorageBucketPath:       &mockSelfStorageBucketPath,
		SplunkArchivalRetentionDays: nil,
	}

	t.Run("with some client interface error", func(_ *testing.T) {
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(nil, errors.New("some error")).Once()
		err := idx.WaitVerifyIndexUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockIndexName)
		assert.Error(t, err)
	})

	t.Run("with http 200 response", func(t *testing.T) {
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(genIndexResp(200), nil).Once()
		err := idx.WaitVerifyIndexUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockIndexName)
		assert.NoError(t, err)
	})

	tmpMockSearchableDays := int64(50)
	mockIndexNotUpdated, _ := json.Marshal(v2.IndexResponse{
		Datatype: mockDatatype,
		// nolint
		MaxDataSizeMB: uint64(mockMaxDataSizeMB),
		Name:          mockIndexName,
		// nolint
		SearchableDays:              uint64(tmpMockSearchableDays),
		SelfStorageBucketPath:       &mockSelfStorageBucketPath,
		SplunkArchivalRetentionDays: nil,
		TotalEventCount:             nil,
		TotalRawSizeMB:              nil,
	})

	mockResp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Body:       io.NopCloser(bytes.NewBufferString(string(mockIndexNotUpdated))),
	}

	t.Run("with non updated response first", func(t *testing.T) {
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(mockResp, nil).Once()
		client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(genIndexResp(200), nil).Once()
		err := idx.WaitVerifyIndexUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockIndexName)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 404, 409, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("GetIndexInfo", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName)).Return(genIndexResp(unexpectedStatusCode), nil).Once()
				err := idx.WaitVerifyIndexUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockIndexName)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitIndexDelete(t *testing.T) {
	client := &mocks.ClientInterface{}

	t.Run("with some client interface error", func(_ *testing.T) {
		client.On("DeleteIndex", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName), v2.DeleteIndexJSONRequestBody{}).Return(nil, errors.New("some error")).Once()
		err := idx.WaitIndexDelete(context.TODO(), client, mockStack, mockIndexName)
		assert.Error(t, err)
	})

	t.Run("with http response 202", func(t *testing.T) {
		client.On("DeleteIndex", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName),
			v2.DeleteIndexJSONRequestBody{}).Return(acceptedResp, nil).Once()
		err := idx.WaitIndexDelete(context.TODO(), client, mockStack, mockIndexName)
		assert.NoError(t, err)
	})

	t.Run("with retry on rate limit", func(t *testing.T) {
		client.On("DeleteIndex", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName),
			v2.DeleteIndexJSONRequestBody{}).Return(rateLimitResp, nil).Once()
		client.On("DeleteIndex", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName),
			v2.DeleteIndexJSONRequestBody{}).Return(acceptedResp, nil).Once()
		err := idx.WaitIndexDelete(context.TODO(), client, mockStack, mockIndexName)
		assert.NoError(t, err)
	})

	t.Run("with unexpected error resp", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 404, 409, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DeleteIndex", mock.Anything, v2.Stack(mockStack), v2.Index(mockIndexName),
					v2.DeleteIndexJSONRequestBody{}).Return(badReqResp, nil).Once()
				err := idx.WaitIndexDelete(context.TODO(), client, mockStack, mockIndexName)
				assert.Error(t, err)
			})
		}
	})
}
