package acctest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/provider"
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
