package splunkbaseapps

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/acs/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func generateResponse(statusCode int) *http.Response {
	return &http.Response{
		Status:     http.StatusText(statusCode),
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader("{}")), // Non-nil body
	}
}

func TestHandlesSuccessfulAppCreationResponse(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	params := v2.InstallAppVictoriaParams{}
	body := io.NopCloser(strings.NewReader("{}"))

	client.On("InstallAppVictoriaWithBody", ctx, stack, &params, "application/x-www-form-urlencoded", body).Return(generateResponse(http.StatusAccepted), nil).Once()

	state, status, err := AppStatusCreate(ctx, client, stack, params, body)()
	assert.NoError(t, err)
	assert.Equal(t, "Accepted", status)
	assert.NotNil(t, state)
}

func TestHandlesConflictDuringAppCreationResponse(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	params := v2.InstallAppVictoriaParams{}
	body := io.NopCloser(strings.NewReader("{}"))

	client.On("InstallAppVictoriaWithBody", ctx, stack, &params, "application/x-www-form-urlencoded", body).Return(generateResponse(http.StatusConflict), errors.New("conflict error")).Once()

	state, status, err := AppStatusCreate(ctx, client, stack, params, body)()
	assert.Error(t, err)
	assert.Equal(t, "Conflict", status)
	assert.Nil(t, state)
}

func TestHandlesNilResponseDuringAppCreationResponse(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	params := v2.InstallAppVictoriaParams{}
	body := io.NopCloser(strings.NewReader("{}"))

	client.On("InstallAppVictoriaWithBody", ctx, stack, &params, "application/x-www-form-urlencoded", body).Return(nil, errors.New("network error")).Once()

	var (
		state  interface{}
		status string
		err    error
	)

	assert.NotPanics(t, func() {
		state, status, err = AppStatusCreate(ctx, client, stack, params, body)()
	})
	assert.Error(t, err)
	assert.Empty(t, status)
	assert.Nil(t, state)
}

func TestHandlesSuccessfulAppReadResponse(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	appName := v2.AppName("mock-app")

	client.On("DescribeAppVictoria", ctx, stack, appName).Return(generateResponse(http.StatusOK), nil).Once()
	state, status, err := AppStatusRead(ctx, client, stack, appName)()
	assert.NoError(t, err)
	assert.Equal(t, "OK", status)
	assert.NotNil(t, state)
}

func TestHandlesErrorDuringAppReadResponse(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	appName := v2.AppName("mock-app")

	client.On("DescribeAppVictoria", ctx, stack, appName).Return(nil, errors.New("network error")).Once()

	_, _, err := AppStatusRead(ctx, client, stack, appName)()
	assert.Error(t, err)
}

func TestHandlesSuccessfulAppDeletionResponse(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	appName := v2.AppName("mock-app")

	client.On("UninstallAppVictoria", ctx, stack, appName, &v2.UninstallAppVictoriaParams{}).Return(generateResponse(http.StatusNotFound), nil).Once()

	state, status, err := AppStatusDelete(ctx, client, stack, appName, v2.UninstallAppVictoriaParams{})()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusText(http.StatusNotFound), status)
	assert.NotNil(t, state)
}

func TestHandlesFailedAppDeletionResponse(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	appName := v2.AppName("mock-app")

	client.On("UninstallAppVictoria", ctx, stack, appName, &v2.UninstallAppVictoriaParams{}).Return(generateResponse(http.StatusInternalServerError), errors.New("server error")).Once()
	state, status, err := AppStatusDelete(ctx, client, stack, appName, v2.UninstallAppVictoriaParams{})()
	assert.Error(t, err)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), status)
	assert.Nil(t, state)
}

func TestHandlesNilResponseDuringAppDeletionResponse(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	appName := v2.AppName("mock-app")

	client.On("UninstallAppVictoria", ctx, stack, appName, &v2.UninstallAppVictoriaParams{}).Return(nil, errors.New("network error")).Once()

	var (
		state  interface{}
		status string
		err    error
	)

	assert.NotPanics(t, func() {
		state, status, err = AppStatusDelete(ctx, client, stack, appName, v2.UninstallAppVictoriaParams{})()
	})
	assert.Error(t, err)
	assert.Empty(t, status)
	assert.Nil(t, state)
}

func TestHandlesSuccessfulAppUpdateResponse(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	appName := v2.AppName("mock-app")
	params := v2.PatchAppVictoriaParams{}
	body := io.NopCloser(strings.NewReader("{}"))

	client.On("PatchAppVictoriaWithBody", ctx, stack, appName, &params, "application/x-www-form-urlencoded", body).Return(generateResponse(http.StatusOK), nil).Once()
	state, status, err := AppStatusUpdate(ctx, client, stack, appName, params, body)()
	assert.NoError(t, err)
	assert.Equal(t, "OK", status)
	assert.NotNil(t, state)
}

func TestHandlesErrorDuringAppUpdateResponse(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	appName := v2.AppName("mock-app")
	params := v2.PatchAppVictoriaParams{}
	body := io.NopCloser(strings.NewReader("{}"))

	client.On("PatchAppVictoriaWithBody", ctx, stack, appName, &params, "application/x-www-form-urlencoded", body).Return(nil, errors.New("update error")).Once()

	_, _, err := AppStatusUpdate(ctx, client, stack, appName, params, body)()
	assert.Error(t, err)
}
