package privateapps_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/acs/v2/mocks"
	privateapps "github.com/splunk/terraform-provider-scp/internal/private_apps"
	"github.com/stretchr/testify/assert"
)

func generateResponse(statusCode int) *http.Response {
	return &http.Response{
		Status:     http.StatusText(statusCode),
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader("{}")), // Add a non-nil Body
	}
}

func TestHandlesSuccessfulAppCreation(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	params := v2.InstallAppVictoriaParams{}
	body := io.NopCloser(strings.NewReader("{}"))

	client.On("InstallAppVictoriaWithBody", ctx, stack, &params, "application/x-www-form-urlencoded", body).Return(generateResponse(http.StatusAccepted), nil).Once()

	state, status, err := privateapps.AppStatusCreate(ctx, client, stack, params, body)()
	assert.NoError(t, err)
	assert.Equal(t, "Accepted", status)
	assert.NotNil(t, state)
}

func TestHandlesConflictDuringAppCreation(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	params := v2.InstallAppVictoriaParams{}
	body := io.NopCloser(strings.NewReader("{}"))

	client.On("InstallAppVictoriaWithBody", ctx, stack, &params, "application/x-www-form-urlencoded", body).Return(generateResponse(http.StatusConflict), nil).Once()

	state, status, err := privateapps.AppStatusCreate(ctx, client, stack, params, body)()
	assert.NoError(t, err)
	assert.Equal(t, "Conflict", status)
	assert.Nil(t, state)
}

func TestHandlesNilResponseDuringAppCreation(t *testing.T) {
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
		state, status, err = privateapps.AppStatusCreate(ctx, client, stack, params, body)()
	})
	assert.Error(t, err)
	assert.Empty(t, status)
	assert.Nil(t, state)
}

func TestHandlesErrorDuringAppRead(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	appName := v2.AppName("mock-app")

	client.On("DescribeAppVictoria", ctx, stack, appName).Return(nil, errors.New("network error")).Once()

	_, _, err := privateapps.AppStatusRead(ctx, client, stack, appName)()
	assert.Error(t, err)
}

func TestHandlesSuccessfulAppDeletion(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	appName := v2.AppName("mock-app")
	params := v2.UninstallAppVictoriaParams{}

	client.On("UninstallAppVictoria", ctx, stack, appName, &params).Return(generateResponse(http.StatusNotFound), nil).Once()

	state, status, err := privateapps.AppStatusDelete(ctx, client, stack, appName, params)()
	assert.NoError(t, err)
	assert.Equal(t, "Not Found", status)
	assert.NotNil(t, state)
}

func TestHandlesNilResponseDuringAppDeletion(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	appName := v2.AppName("mock-app")
	params := v2.UninstallAppVictoriaParams{}

	client.On("UninstallAppVictoria", ctx, stack, appName, &params).Return(nil, errors.New("network error")).Once()

	var (
		state  interface{}
		status string
		err    error
	)

	assert.NotPanics(t, func() {
		state, status, err = privateapps.AppStatusDelete(ctx, client, stack, appName, params)()
	})
	assert.Error(t, err)
	assert.Empty(t, status)
	assert.Nil(t, state)
}

func TestHandlesErrorDuringAppUpdate(t *testing.T) {
	client := &mocks.ClientInterface{}
	ctx := context.TODO()
	stack := v2.Stack("mock-stack")
	appName := v2.AppName("mock-app")
	params := v2.PatchAppVictoriaParams{}
	body := io.NopCloser(strings.NewReader("{}"))

	client.On("PatchAppVictoriaWithBody", ctx, stack, appName, &params, "application/x-www-form-urlencoded", body).Return(nil, errors.New("update error")).Once()

	_, _, err := privateapps.AppStatusUpdate(ctx, client, stack, appName, params, body)()
	assert.Error(t, err)
}
