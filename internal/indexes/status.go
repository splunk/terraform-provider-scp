package indexes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/status"
)

var GeneralRetryableStatusCodes = map[int]string{
	http.StatusTooManyRequests: http.StatusText(429),
}

// IndexStatusCreate returns StateRefreshFunc that makes POST request and checks if response is accepted
func IndexStatusCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, createIndexRequest v2.CreateIndexJSONRequestBody) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := acsClient.CreateIndex(ctx, stack, createIndexRequest)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		return status.ProcessResponse(resp, TargetStatusResourceChange, PendingStatusCRUD)
	}
}

// IndexStatusPoll returns StateRefreshFunc that makes GET request and checks if response is desired target (200 for create and 404 for delete)
func IndexStatusPoll(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, indexName string, targetStatus []string, pendingStatus []string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.GetIndexInfo(ctx, stack, v2.Index(indexName))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return status.ProcessResponse(resp, targetStatus, pendingStatus)
	}
}

// IndexStatusRead returns StateRefreshFunc that makes GET request, checks if request was successful, and returns index response
func IndexStatusRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, indexName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.GetIndexInfo(ctx, stack, v2.Index(indexName))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		if _, ok := GeneralRetryableStatusCodes[resp.StatusCode]; !ok && resp.StatusCode != http.StatusOK {
			return nil, http.StatusText(resp.StatusCode), &resource.UnexpectedStateError{
				State:         http.StatusText(resp.StatusCode),
				ExpectedState: TargetStatusResourceExists,
				LastError:     errors.New(string(bodyBytes)),
			}
		}

		var index v2.IndexResponse
		if resp.StatusCode == 200 {
			if err = json.Unmarshal(bodyBytes, &index); err != nil {
				return nil, "", &resource.UnexpectedStateError{LastError: err}
			}
		}
		status := http.StatusText(resp.StatusCode)
		return &index, status, nil
	}
}

// IndexStatusDelete returns StateRefreshFunc that makes DELETE request and checks if request was accepted
func IndexStatusDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, indexName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DeleteIndex(ctx, stack, v2.Index(indexName), v2.DeleteIndexJSONRequestBody{})
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return status.ProcessResponse(resp, TargetStatusResourceChange, PendingStatusCRUD)
	}
}

// IndexStatusUpdate returns StateRefreshFunc that makes PATCH request and checks if request was accepted
func IndexStatusUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchIndexRequest v2.PatchIndexInfoJSONRequestBody, indexName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		resp, err := acsClient.PatchIndexInfo(ctx, stack, v2.Index(indexName), patchIndexRequest)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return status.ProcessResponse(resp, TargetStatusResourceChange, PendingStatusCRUD)
	}
}

// IndexStatusVerifyUpdate returns a StateRefreshFunc that makes a GET request and checks to see if the index fields matches those in patch request
func IndexStatusVerifyUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchIndexInfoJSONRequestBody, indexName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.GetIndexInfo(ctx, stack, v2.Index(indexName))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		if _, ok := GeneralRetryableStatusCodes[resp.StatusCode]; !ok && resp.StatusCode != 200 {
			return nil, http.StatusText(resp.StatusCode), &resource.UnexpectedStateError{LastError: errors.New(string(bodyBytes))}
		}

		var index v2.IndexResponse
		updateComplete := false
		if resp.StatusCode == 200 {
			if err = json.Unmarshal(bodyBytes, &index); err != nil {
				return nil, "", &resource.UnexpectedStateError{LastError: err}
			}
			updateComplete = VerifyIndexUpdate(patchRequest, index)
		}

		var status string
		if updateComplete {
			status = "UPDATED"
			return &index, status, nil
		} else {
			status = http.StatusText(resp.StatusCode)
			return nil, status, nil
		}
	}
}

// VerifyIndexUpdate is a helper to verify that the fields in patch request match fields in the index response
func VerifyIndexUpdate(patchRequest v2.PatchIndexInfoJSONRequestBody, index v2.IndexResponse) bool {
	if patchRequest.MaxDataSizeMB != nil && uint64(*patchRequest.MaxDataSizeMB) != index.MaxDataSizeMB {
		return false
	}
	if patchRequest.SearchableDays != nil && uint64(*patchRequest.SearchableDays) != index.SearchableDays {
		return false
	}
	if patchRequest.SelfStorageBucketPath != nil {
		if index.SelfStorageBucketPath == nil || *patchRequest.SelfStorageBucketPath != *index.SelfStorageBucketPath {
			return false
		}
	}
	if patchRequest.SplunkArchivalRetentionDays != nil {
		if index.SplunkArchivalRetentionDays == nil || uint64(*patchRequest.SplunkArchivalRetentionDays) != *index.SplunkArchivalRetentionDays {
			return false
		}
	}

	return true
}
