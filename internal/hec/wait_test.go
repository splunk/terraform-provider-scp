package hec_test

import (
	"context"
	"errors"
	"fmt"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/acs/v2/mocks"
	hec "github.com/splunk/terraform-provider-scp/internal/hec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_WaitHecCreate(t *testing.T) {
	client := &mocks.ClientInterface{}

	mockCreateBody := v2.CreateHECJSONRequestBody{
		Name: mockHecName,
	}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("CreateHEC", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(nil, errors.New("some error")).Once()
		err := hec.WaitHecCreate(context.TODO(), client, mockStack, mockCreateBody)
		assert.Error(t, err)
	})

	t.Run("with http response 202", func(t *testing.T) {
		client.On("CreateHEC", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(acceptedResp, nil).Once()
		err := hec.WaitHecCreate(context.TODO(), client, mockStack, mockCreateBody)
		assert.NoError(t, err)
	})

	t.Run("with retryable response 429", func(t *testing.T) {
		client.On("CreateHEC", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(rateLimitResp, nil).Once()
		client.On("CreateHEC", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(acceptedResp, nil).Once()
		err := hec.WaitHecCreate(context.TODO(), client, mockStack, mockCreateBody)
		assert.NoError(t, err)
	})

	t.Run("with unexpected http responses", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 404, 409, 501} {
			t.Run(fmt.Sprintf("with unexpected status %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("CreateHEC", mock.Anything, v2.Stack(mockStack), mockCreateBody).Return(genHecResp(unexpectedStatusCode), nil).Once()
				err := hec.WaitHecCreate(context.TODO(), client, mockStack, mockCreateBody)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitHecPoll(t *testing.T) {
	client := &mocks.ClientInterface{}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(nil, errors.New("some error")).Once()
		err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, hec.TargetStatusResourceExists, hec.PendingStatusVerifyCreated)
		assert.Error(t, err)
	})

	/* Test Poll to Verify Creation */
	t.Run("with http response 200", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(successRespOk, nil).Once()
		err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, hec.TargetStatusResourceExists, hec.PendingStatusVerifyCreated)
		assert.NoError(t, err)
	})

	t.Run("with retryable response 404 verify create", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(rateLimitResp, nil).Once()
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(successRespOk, nil).Once()
		err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, hec.TargetStatusResourceExists, hec.PendingStatusVerifyCreated)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 409, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(genHecResp(unexpectedStatusCode), nil).Once()
				err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, hec.TargetStatusResourceExists, hec.PendingStatusVerifyCreated)
				assert.Error(t, err)
			})
		}
	})

	/* Test Poll to Verify Deletion */
	t.Run("with expected http response 404", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(notFoundResp, nil).Once()
		err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, hec.TargetStatusResourceDeleted, hec.PendingStatusVerifyDeleted)
		assert.NoError(t, err)
	})

	t.Run("with retryable response 200", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(successRespOk, nil).Once()
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(notFoundResp, nil).Once()
		err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, hec.TargetStatusResourceDeleted, hec.PendingStatusVerifyDeleted)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 409, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(genHecResp(unexpectedStatusCode), nil).Once()
				err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, hec.TargetStatusResourceDeleted, hec.PendingStatusVerifyDeleted)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitHecRead(t *testing.T) {
	client := &mocks.ClientInterface{}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(nil, errors.New("some error")).Once()
		Hec, err := hec.WaitHecRead(context.TODO(), client, mockStack, mockHecName)
		assert.Error(t, err)
		assert.Nil(t, Hec)
	})

	t.Run("with http 200 response", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(genHecResp(200), nil).Once()
		Hec, err := hec.WaitHecRead(context.TODO(), client, mockStack, mockHecName)
		assert.NoError(t, err)
		assert.NotNil(t, Hec)
		assert.NotNil(t, Hec.Name)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 404, 409, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(genHecResp(unexpectedStatusCode), nil).Once()
				Hec, err := hec.WaitHecRead(context.TODO(), client, mockStack, mockHecName)
				assert.Error(t, err)
				assert.Nil(t, Hec)
			})
		}
	})
}
