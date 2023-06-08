package hec_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/acs/v2/mocks"
	"github.com/splunk/terraform-provider-scp/internal/hec"
	"github.com/splunk/terraform-provider-scp/internal/wait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
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
		err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
		assert.Error(t, err)
	})

	/* Test Poll to Verify Creation */
	t.Run("with http response 200", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(successRespOk, nil).Once()
		err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
		assert.NoError(t, err)
	})

	t.Run("with retryable response 404 verify create", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(rateLimitResp, nil).Once()
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(successRespOk, nil).Once()
		err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 409, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(genHecResp(unexpectedStatusCode), nil).Once()
				err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
				assert.Error(t, err)
			})
		}
	})

	/* Test Poll to Verify Deletion */
	t.Run("with expected http response 404", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(notFoundResp, nil).Once()
		err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
		assert.NoError(t, err)
	})

	t.Run("with retryable response 200", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(successRespOk, nil).Once()
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(notFoundResp, nil).Once()
		err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 409, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(genHecResp(unexpectedStatusCode), nil).Once()
				err := hec.WaitHecPoll(context.TODO(), client, mockStack, mockHecName, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
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

func Test_WaitHecUpdate(t *testing.T) {
	client := &mocks.ClientInterface{}

	mockUpdateBody := v2.PatchHECJSONRequestBody{
		DefaultSource: &mockDefaultSource,
		DefaultIndex:  &mockDefaultIndex,
	}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("PatchHEC", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName), mockUpdateBody).Return(nil, errors.New("some error")).Once()
		err := hec.WaitHecUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockHecName)
		assert.Error(t, err)
	})

	t.Run("with http response 202", func(t *testing.T) {
		client.On("PatchHEC", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName), mockUpdateBody).Return(acceptedResp, nil).Once()
		err := hec.WaitHecUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockHecName)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 404, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("PatchHEC", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName), mockUpdateBody).Return(genHecResp(unexpectedStatusCode), nil).Once()
				err := hec.WaitHecUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockHecName)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitVerifyHecUpdate(t *testing.T) {
	client := &mocks.ClientInterface{}

	mockUpdateBody := v2.PatchHECJSONRequestBody{
		AllowedIndexes:    &mockAllowedIndexes,
		DefaultIndex:      &mockDefaultIndex,
		DefaultSource:     &mockDefaultSource,
		DefaultSourcetype: nil,
		Disabled:          &mockDisabled,
		UseAck:            &mockUseAck,
	}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(nil, errors.New("some error")).Once()
		err := hec.WaitVerifyHecUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockHecName)
		assert.Error(t, err)
	})

	t.Run("with http 200 response", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(genHecResp(200), nil).Once()
		err := hec.WaitVerifyHecUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockHecName)
		assert.NoError(t, err)
	})

	var tmpDefaultSource = "tmp-default-source"
	mockHecNotUpdated, _ := json.Marshal(hec.HecBody{HttpEventCollector: &v2.HecInfo{
		Spec: &v2.HecSpec{
			DefaultIndex:      &mockDefaultIndex,
			DefaultSourcetype: &mockDefaultSourceType,
			DefaultSource:     &tmpDefaultSource,
			UseAck:            &mockUseAck,
			Disabled:          &mockDisabled,
			AllowedIndexes:    &mockAllowedIndexes,
		}}})

	mockResp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Body:       io.NopCloser(bytes.NewBufferString(string(mockHecNotUpdated))),
	}

	t.Run("with non updated response first", func(t *testing.T) {
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(mockResp, nil).Once()
		client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(genHecResp(200), nil).Once()
		err := hec.WaitVerifyHecUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockHecName)
		assert.NoError(t, err)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 404, 409, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DescribeHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName)).Return(genHecResp(unexpectedStatusCode), nil).Once()
				err := hec.WaitVerifyHecUpdate(context.TODO(), client, mockStack, mockUpdateBody, mockHecName)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitHecDelete(t *testing.T) {
	client := &mocks.ClientInterface{}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("DeleteHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName), v2.DeleteHecJSONRequestBody{}).Return(nil, errors.New("some error")).Once()
		err := hec.WaitHecDelete(context.TODO(), client, mockStack, mockHecName)
		assert.Error(t, err)
	})

	t.Run("with http response 202", func(t *testing.T) {
		client.On("DeleteHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName),
			v2.DeleteHecJSONRequestBody{}).Return(acceptedResp, nil).Once()
		err := hec.WaitHecDelete(context.TODO(), client, mockStack, mockHecName)
		assert.NoError(t, err)
	})

	t.Run("with retry on rate limit", func(t *testing.T) {
		client.On("DeleteHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName),
			v2.DeleteHecJSONRequestBody{}).Return(rateLimitResp, nil).Once()
		client.On("DeleteHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName),
			v2.DeleteHecJSONRequestBody{}).Return(acceptedResp, nil).Once()
		err := hec.WaitHecDelete(context.TODO(), client, mockStack, mockHecName)
		assert.NoError(t, err)
	})

	t.Run("with unexpected error resp", func(t *testing.T) {
		for _, unexpectedStatusCode := range []int{400, 401, 403, 404, 409, 501, 500, 503} {
			t.Run(fmt.Sprintf("with unexpected response %v", unexpectedStatusCode), func(t *testing.T) {
				client.On("DeleteHec", mock.Anything, v2.Stack(mockStack), v2.Hec(mockHecName),
					v2.DeleteHecJSONRequestBody{}).Return(badReqResp, nil).Once()
				err := hec.WaitHecDelete(context.TODO(), client, mockStack, mockHecName)
				assert.Error(t, err)
			})
		}
	})
}
