package acctest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/provider"
	"github.com/splunk/terraform-provider-scp/internal/utils"
	"github.com/splunk/terraform-provider-scp/version"
)

var Provider *schema.Provider

// ProviderFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var ProviderFactories = map[string]func() (*schema.Provider, error){
	"scp": func() (*schema.Provider, error) {
		return provider.New(version.ProviderVersion)(), nil
	},
}

func init() {
	var err error
	Provider = provider.New(version.ProviderVersion)()

	if err != nil {
		panic(err)
	}
}

// PreCheck is run prior to any test case execution, add code here to run before any test execution
// For example, assertions about the appropriate environment
func PreCheck(t *testing.T) {
	variables := []string{
		"ACS_SERVER",
		"STACK_TOKEN",
		"SPLUNK_STACK",
		"STACK_USERNAME",
		"STACK_PASSWORD",
	}

	for _, variable := range variables {
		value := os.Getenv(variable)
		if value == "" {
			t.Fatalf("`%s` must be set for acceptance tests!", variable)
		}
	}
}

// PreCheckSplunkbaseApps is run prior to splunkbase apps test case execution as an additional check
// It ensures that the environment variables needed for splunkbase authentication are set
func PreCheckSplunkbaseApps(t *testing.T) {
	appsVariables := []string{
		"SPLUNK_USERNAME",
		"SPLUNK_PASSWORD",
	}

	for _, variable := range appsVariables {
		value := os.Getenv(variable)
		if value == "" {
			t.Fatalf("`%s` must be set for splunkbase apps acceptance tests!", variable)
		}
	}
}

func describeAppResource(id string) (*http.Response, error) {
	providerNew := Provider
	diags := providerNew.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if diags != nil {
		return nil, fmt.Errorf("%+v", diags)
	}

	acsProvider := providerNew.Meta().(*client.ACSProvider).Client
	acsClient := *acsProvider
	stack := providerNew.Meta().(*client.ACSProvider).Stack

	resp, err := acsClient.DescribeAppVictoria(context.TODO(), stack, v2.AppName(id))
	if err != nil {
		return nil, fmt.Errorf("error describing app resource: %s", err)
	}

	return resp, nil
}

func CheckAppResourceDeleted(name string, id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if ok {
			return fmt.Errorf("resource still in state: %s", name)
		}

		resp, err := describeAppResource(id)
		if err != nil {
			return fmt.Errorf("error while fetching app resource: %e", err)
		}
		defer resp.Body.Close()

		statusCode := resp.StatusCode
		if statusCode != http.StatusNotFound {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("error reading response body: %e", err)
			}
			return fmt.Errorf("expected %d, got %d, %s", http.StatusNotFound, statusCode, string(body))
		}

		return nil
	}
}

func CheckAppResourceCreated(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not in state: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is not set")
		}

		resp, err := describeAppResource(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error while fetching app resource: %e", err)
		}
		defer resp.Body.Close()

		statusCode := resp.StatusCode
		if statusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("error reading response body: %e", err)
			}
			return fmt.Errorf("expected %d, got %d, %s", http.StatusOK, statusCode, string(body))
		}

		return nil
	}
}

func CheckAppResourceUpdated(name string, updatedFields map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		time.Sleep(25 * time.Second)
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not in state: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is not set")
		}

		resp, err := describeAppResource(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error while fetching app resource: %e", err)
		}
		defer resp.Body.Close()

		statusCode := resp.StatusCode
		if statusCode != http.StatusOK {
			return fmt.Errorf("expected %d, got %d, %s", http.StatusOK, statusCode, resp.Body)
		}

		result := make(map[string]interface{})
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return fmt.Errorf("error decoding response body: %e", err)
		}

		for fieldName, expectedValue := range updatedFields {
			if actualValue, ok := result[fieldName]; ok {
				if expectedValue != actualValue {
					return fmt.Errorf("field %s differs %s %s", fieldName, expectedValue, actualValue)
				}
			}
		}

		return nil
	}
}

// PreCheckTAI validates that the TAI-specific environment variables are set.
// TAI_TARGET_1 is required (at least one search head target).
// TAI_TARGET_2 is optional and enables multi-target test scenarios.
func PreCheckTAI(t *testing.T) {
	target1 := os.Getenv("TAI_TARGET_1")
	if target1 == "" {
		t.Fatalf("`TAI_TARGET_1` must be set for TAI acceptance tests!")
	}
}

func CleanupTAIApp(t *testing.T, appID string) {
	t.Helper()

	providerNew := Provider
	diags := providerNew.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if diags != nil {
		return
	}
	acsProvider := providerNew.Meta().(*client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Best-effort delete from global stack
	acsClient.UninstallAppVictoria(context.TODO(), stack, v2.AppName(appID), &v2.UninstallAppVictoriaParams{}) //nolint:errcheck

	// Best-effort delete from each known TAI target
	localScope := "local"
	for _, target := range []string{os.Getenv("TAI_TARGET_1"), os.Getenv("TAI_TARGET_2")} {
		if target == "" {
			continue
		}
		targetStack, err := utils.TargetStackName(target, stack)
		if err != nil {
			continue
		}
		targetClient, err := acsProvider.ClientForTarget(context.TODO(), targetStack)
		if err != nil {
			continue
		}
		targetClient.UninstallAppVictoria(context.TODO(), targetStack, v2.AppName(appID), &v2.UninstallAppVictoriaParams{Scope: &localScope}) //nolint:errcheck
	}
	time.Sleep(5 * time.Second)
}

func GetTAITarget1() string {
	return os.Getenv("TAI_TARGET_1")
}

func GetTAITarget2() string {
	return os.Getenv("TAI_TARGET_2")
}

func HasMultipleTAITargets() bool {
	return os.Getenv("TAI_TARGET_2") != ""
}

func describeAppResourceOnTarget(id string, target string) (*http.Response, error) {
	providerNew := Provider
	diags := providerNew.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if diags != nil {
		return nil, fmt.Errorf("%+v", diags)
	}

	acsProvider := providerNew.Meta().(*client.ACSProvider)
	stack := acsProvider.Stack

	targetStack, err := utils.TargetStackName(target, stack)
	if err != nil {
		return nil, fmt.Errorf("error building target stack name: %s", err)
	}

	targetClient, err := acsProvider.ClientForTarget(context.TODO(), targetStack)
	if err != nil {
		return nil, fmt.Errorf("error getting client for target %s: %s", target, err)
	}

	resp, err := targetClient.DescribeAppVictoria(context.TODO(), targetStack, v2.AppName(id))
	if err != nil {
		return nil, fmt.Errorf("error describing app resource on target %s: %s", target, err)
	}

	return resp, nil
}

func CheckAppResourceCreatedOnTarget(name string, target string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not in state: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is not set")
		}

		// Verify remote state: app exists on target
		resp, err := describeAppResourceOnTarget(rs.Primary.ID, target)
		if err != nil {
			return fmt.Errorf("error while fetching app resource on target %s: %s", target, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("error reading response body: %s", err)
			}
			return fmt.Errorf("expected %d on target %s, got %d, %s", http.StatusOK, target, resp.StatusCode, string(body))
		}

		return nil
	}
}

// CheckAppResourceCreatedOnTargetWithRetry is like CheckAppResourceCreatedOnTarget
// but retries on 404 responses for up to the given timeout. This is useful when an
// app install is propagating asynchronously to a new target.
func CheckAppResourceCreatedOnTargetWithRetry(name string, target string, timeout time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not in state: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is not set")
		}

		return resource.RetryContext(context.Background(), timeout, func() *resource.RetryError {
			resp, err := describeAppResourceOnTarget(rs.Primary.ID, target)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("error while fetching app resource on target %s: %s", target, err))
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				return resource.RetryableError(fmt.Errorf("app %s not yet available on target %s (404)", rs.Primary.ID, target))
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return resource.NonRetryableError(fmt.Errorf("expected %d on target %s, got %d, %s", http.StatusOK, target, resp.StatusCode, string(body)))
			}

			return nil
		})
	}
}

func CheckAppResourceDeletedOnTarget(id string, target string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resp, err := describeAppResourceOnTarget(id, target)
		if err != nil {
			return fmt.Errorf("error while fetching app resource on target %s: %s", target, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("error reading response body: %s", err)
			}
			return fmt.Errorf("expected %d on target %s, got %d, %s", http.StatusNotFound, target, resp.StatusCode, string(body))
		}

		return nil
	}
}

func CheckTAITargetsInState(resourceName string, expectedTargets []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not in state: %s", resourceName)
		}

		targetsCount := rs.Primary.Attributes["targets.#"]
		expectedCount := fmt.Sprintf("%d", len(expectedTargets))
		if targetsCount != expectedCount {
			return fmt.Errorf("expected %s targets in state, got %s", expectedCount, targetsCount)
		}

		// Collect actual target values from state and compare (order-independent)
		count := len(expectedTargets)
		actualTargets := make([]string, 0, count)
		for key, value := range rs.Primary.Attributes {
			if len(key) > len("targets.") && key[:len("targets.")] == "targets." && key != "targets.#" {
				actualTargets = append(actualTargets, value)
			}
		}

		sort.Strings(actualTargets)
		sortedExpected := make([]string, len(expectedTargets))
		copy(sortedExpected, expectedTargets)
		sort.Strings(sortedExpected)

		if len(actualTargets) != len(sortedExpected) {
			return fmt.Errorf("expected targets %v, got %v", sortedExpected, actualTargets)
		}
		for i := range sortedExpected {
			if actualTargets[i] != sortedExpected[i] {
				return fmt.Errorf("expected targets %v, got %v", sortedExpected, actualTargets)
			}
		}

		return nil
	}
}

func CheckAppResourceUpdatedOnTarget(name string, target string, expectedFields map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not in state: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is not set")
		}

		resp, err := describeAppResourceOnTarget(rs.Primary.ID, target)
		if err != nil {
			return fmt.Errorf("error while fetching app resource on target %s: %s", target, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("expected %d on target %s, got %d, %s", http.StatusOK, target, resp.StatusCode, string(body))
		}

		result := make(map[string]interface{})
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("error decoding response body from target %s: %s", target, err)
		}

		for fieldName, expectedValue := range expectedFields {
			actualValue, ok := result[fieldName]
			if !ok {
				return fmt.Errorf("field %q not found in response from target %s", fieldName, target)
			}
			actualStr := fmt.Sprintf("%v", actualValue)
			if actualStr != expectedValue {
				return fmt.Errorf("on target %s, field %q: expected %q, got %q", target, fieldName, expectedValue, actualStr)
			}
		}

		return nil
	}
}
