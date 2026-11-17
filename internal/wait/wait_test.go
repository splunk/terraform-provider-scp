package wait_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/splunk/terraform-provider-scp/internal/wait"
	"github.com/stretchr/testify/require"
)

func TestWaitForCreated(t *testing.T) {
	t.Run("retries pending statuses and returns readable state", func(t *testing.T) {
		codes := []int{http.StatusNotFound, http.StatusFailedDependency, http.StatusTooManyRequests, http.StatusOK}
		expected := &struct{ Name string }{Name: "created"}
		result, err := wait.WaitForCreated(context.Background(), func(context.Context) resource.StateRefreshFunc {
			return func() (interface{}, string, error) {
				code := codes[0]
				codes = codes[1:]
				return expected, http.StatusText(code), nil
			}
		})
		require.NoError(t, err)
		require.Same(t, expected, result)
		require.Empty(t, codes, "must wait through every pending status")
	})

	t.Run("bounds active reads and honors cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		result, err := wait.WaitForCreated(ctx, func(readCtx context.Context) resource.StateRefreshFunc {
			deadline, ok := readCtx.Deadline()
			require.True(t, ok, "read must have a visibility deadline")
			require.WithinDuration(t, time.Now().Add(wait.CreateVisibilityTimeout), deadline, time.Second)
			return func() (interface{}, string, error) {
				cancel()
				<-readCtx.Done()
				return nil, "", readCtx.Err()
			}
		})
		require.ErrorIs(t, err, context.Canceled)
		require.Nil(t, result)
	})

	t.Run("preserves earlier deadline and returns successful state", func(t *testing.T) {
		deadline := time.Now().Add(time.Minute)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()
		var requestCtx context.Context
		expected := &struct{ Name string }{Name: "created"}

		result, err := wait.WaitForCreated(ctx, func(readCtx context.Context) resource.StateRefreshFunc {
			requestCtx = readCtx
			actualDeadline, ok := readCtx.Deadline()
			require.True(t, ok)
			require.Equal(t, deadline, actualDeadline)
			return func() (interface{}, string, error) {
				return expected, http.StatusText(http.StatusOK), nil
			}
		})
		require.NoError(t, err)
		require.Same(t, expected, result)
		require.ErrorIs(t, requestCtx.Err(), context.Canceled, "visibility context must be released")
		require.NoError(t, ctx.Err(), "parent context must remain usable")
	})

	t.Run("honors cancellation before polling", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result, err := wait.WaitForCreated(ctx, func(readCtx context.Context) resource.StateRefreshFunc {
			return func() (interface{}, string, error) {
				return nil, "", readCtx.Err()
			}
		})
		require.ErrorIs(t, err, context.Canceled)
		require.Nil(t, result)
	})
}
