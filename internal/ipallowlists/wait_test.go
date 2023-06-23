package ipallowlists_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/acs/v2/mocks"
	"github.com/splunk/terraform-provider-scp/internal/ipallowlists"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	mockStack   = "mock-stack"
	mockFeature = "s2s"
	mockSubnets = []string{"1.1.1.1/32", "1.1.1.2/32"}
)

var (
	clientErrorCodes = []int{400, 401, 403, 404, 409}
	serverErrorCodes = []int{501, 500, 503}
)

func Test_WaitIPAllowlistCreate(t *testing.T) {
	client := &mocks.ClientInterface{}

	mockCreateBody := v2.AddSubnetsJSONRequestBody{
		Subnets: &mockSubnets,
	}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("AddSubnets", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature), mockCreateBody).Return(nil, errors.New("some error")).Once()
		err := ipallowlists.WaitIPAllowlistCreate(context.TODO(), client, v2.Stack(mockStack), v2.Feature(mockFeature), mockSubnets)
		assert.Error(t, err)
	})

	t.Run("with http response 200", func(t *testing.T) {
		client.On("AddSubnets", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature), mockCreateBody).Return(successRespOk, nil).Once()
		err := ipallowlists.WaitIPAllowlistCreate(context.TODO(), client, v2.Stack(mockStack), v2.Feature(mockFeature), mockSubnets)
		assert.NoError(t, err)
	})

	t.Run("with retryable response 429", func(t *testing.T) {
		client.On("AddSubnets", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature), mockCreateBody).Return(rateLimitResp, nil).Once()
		client.On("AddSubnets", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature), mockCreateBody).Return(successRespOk, nil).Once()
		err := ipallowlists.WaitIPAllowlistCreate(context.TODO(), client, v2.Stack(mockStack), v2.Feature(mockFeature), mockSubnets)
		assert.NoError(t, err)
	})

	t.Run("with unexpected http responses", func(t *testing.T) {
		for _, statusCode := range append(clientErrorCodes, serverErrorCodes...) {
			t.Run(fmt.Sprintf("with unexpected status %v", statusCode), func(t *testing.T) {
				client.On("AddSubnets", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature), mockCreateBody).Return(getIPAllowlistResponse(statusCode), nil).Once()
				err := ipallowlists.WaitIPAllowlistCreate(context.TODO(), client, v2.Stack(mockStack), v2.Feature(mockFeature), mockSubnets)
				assert.Error(t, err)
			})
		}
	})
}

func Test_WaitIPAllowlistRead(t *testing.T) {
	client := &mocks.ClientInterface{}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("DescribeAllowlist", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature)).Return(nil, errors.New("some error")).Once()
		subnets, err := ipallowlists.WaitIPAllowlistRead(context.TODO(), client, v2.Stack(mockStack), mockFeature)
		assert.Error(t, err)
		assert.Nil(t, subnets)
	})

	t.Run("with http 200 response", func(t *testing.T) {
		client.On("DescribeAllowlist", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature)).Return(getIPAllowlistResponse(200), nil).Once()
		subnets, err := ipallowlists.WaitIPAllowlistRead(context.TODO(), client, v2.Stack(mockStack), mockFeature)
		assert.NoError(t, err)
		assert.NotNil(t, subnets)
		assert.ElementsMatch(t, mockSubnets, subnets)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, statusCode := range append(clientErrorCodes, serverErrorCodes...) {
			t.Run(fmt.Sprintf("with unexpected response %v", statusCode), func(t *testing.T) {
				client.On("DescribeAllowlist", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature)).Return(getIPAllowlistResponse(statusCode), nil).Once()
				subnets, err := ipallowlists.WaitIPAllowlistRead(context.TODO(), client, v2.Stack(mockStack), mockFeature)
				assert.Error(t, err)
				assert.Nil(t, subnets)
			})
		}
	})
}

func Test_WaitIPAllowlistDelete(t *testing.T) {
	client := &mocks.ClientInterface{}

	mockDeleteBody := v2.DeleteSubnetsJSONRequestBody{
		Subnets: &mockSubnets,
	}

	t.Run("with some client interface error", func(t *testing.T) {
		client.On("DeleteSubnets", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature), mockDeleteBody).Return(nil, errors.New("some error")).Once()
		err := ipallowlists.WaitIPAllowlistDelete(context.TODO(), client, v2.Stack(mockStack), v2.Feature(mockFeature), mockSubnets)
		assert.Error(t, err)
	})

	t.Run("with http response 200", func(t *testing.T) {
		client.On("DeleteSubnets", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature), mockDeleteBody).Return(successRespOk, nil).Once()
		err := ipallowlists.WaitIPAllowlistDelete(context.TODO(), client, v2.Stack(mockStack), v2.Feature(mockFeature), mockSubnets)
		assert.NoError(t, err)
	})

	t.Run("with retryable response 429", func(t *testing.T) {
		client.On("DeleteSubnets", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature), mockDeleteBody).Return(rateLimitResp, nil).Once()
		client.On("DeleteSubnets", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature), mockDeleteBody).Return(successRespOk, nil).Once()
		err := ipallowlists.WaitIPAllowlistDelete(context.TODO(), client, v2.Stack(mockStack), v2.Feature(mockFeature), mockSubnets)
		assert.NoError(t, err)
	})

	t.Run("with unexpected http responses", func(t *testing.T) {
		for _, statusCode := range append(clientErrorCodes, serverErrorCodes...) {
			t.Run(fmt.Sprintf("with unexpected status %v", statusCode), func(t *testing.T) {
				client.On("DeleteSubnets", mock.Anything, v2.Stack(mockStack), v2.Feature(mockFeature), mockDeleteBody).Return(getIPAllowlistResponse(statusCode), nil).Once()
				err := ipallowlists.WaitIPAllowlistDelete(context.TODO(), client, v2.Stack(mockStack), v2.Feature(mockFeature), mockSubnets)
				assert.Error(t, err)
			})
		}
	})
}
