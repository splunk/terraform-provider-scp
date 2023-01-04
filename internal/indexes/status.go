package indexes

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	v2 "github.com/splunk/terraform-provider-splunkcloud/acs/v2"
	"io"
	"net/http"
)

var GeneralRetryableStatusCodes = map[int]string{
	http.StatusTooManyRequests: http.StatusText(429),
	http.StatusForbidden:       http.StatusText(403), //Todo remove once WAF changes are in
}

func StatusIndexCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, createIndexRequest v2.CreateIndexJSONRequestBody) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := acsClient.CreateIndex(ctx, stack, createIndexRequest)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		return ProcessResponse(resp, TargetStatusResourceChange, PendingStatusCRUD)
	}
}

func StatusIndex(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, indexName string, targetStatus []string, pendingStatus []string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.GetIndexInfo(ctx, stack, v2.Index(indexName))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return ProcessResponse(resp, targetStatus, pendingStatus)
	}
}

func StatusIndexWithIndexResponse(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, indexName string) resource.StateRefreshFunc {
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

func StatusIndexDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, indexName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DeleteIndex(ctx, stack, v2.Index(indexName), v2.DeleteIndexJSONRequestBody{})
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return ProcessResponse(resp, TargetStatusResourceChange, PendingStatusCRUD)
	}
}

func StatusIndexUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchIndexRequest v2.PatchIndexInfoJSONRequestBody, indexName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		resp, err := acsClient.PatchIndexInfo(ctx, stack, v2.Index(indexName), patchIndexRequest)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return ProcessResponse(resp, TargetStatusResourceChange, PendingStatusCRUD)
	}
}

func StatusIndexPollUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchIndexInfoJSONRequestBody, indexName string) resource.StateRefreshFunc {
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
			updateComplete = ValidateIndexUpdateComplete(patchRequest, index)
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

func ValidateIndexUpdateComplete(patchRequest v2.PatchIndexInfoJSONRequestBody, index v2.IndexResponse) bool {
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

func ProcessResponse(resp *http.Response, targetStateCodes []string, pendingStatusCodes []string) (interface{}, string, error) {
	if resp == nil {
		return nil, "", &resource.UnexpectedStateError{LastError: errors.New("nil response")}
	}
	statusCode := resp.StatusCode
	statusText := http.StatusText(statusCode)

	if !IsStatusCodeExpected(statusCode, targetStateCodes, pendingStatusCodes) {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, statusText, &resource.UnexpectedStateError{
			State:         statusText,
			ExpectedState: targetStateCodes,
			LastError:     errors.New(string(bodyBytes))}
	}
	return resp, statusText, nil
}

func IsStatusCodeExpected(statusCode int, targetStatusCodes []string, pendingStatusCodes []string) bool {
	isRetryableError := false
	isTargetStatus := false

	for _, code := range targetStatusCodes {
		if code == http.StatusText(statusCode) {
			isTargetStatus = true
		}
	}

	for _, code := range pendingStatusCodes {
		if code == http.StatusText(statusCode) {
			isRetryableError = true
		}
	}

	return isTargetStatus || isRetryableError
}
