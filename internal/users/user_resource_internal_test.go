package users

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Test_parseUserRequest_ForceChangePass(t *testing.T) {
	tests := []struct {
		name    string
		raw     map[string]interface{}
		wantVal bool
	}{
		{
			name:    "explicit true",
			raw:     map[string]interface{}{schemaKeyName: "tester", schemaKeyForceChangePass: true},
			wantVal: true,
		},
		{
			name:    "explicit false",
			raw:     map[string]interface{}{schemaKeyName: "tester", schemaKeyForceChangePass: false},
			wantVal: false,
		},
		{
			name:    "unset falls back to schema default",
			raw:     map[string]interface{}{schemaKeyName: "tester"},
			wantVal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, userResourceSchema(), tt.raw)

			userInfo := parseUserRequest(d)

			if userInfo.ForceChangePass == nil {
				t.Fatalf("expected ForceChangePass to be set, got nil")
			}
			if *userInfo.ForceChangePass != tt.wantVal {
				t.Errorf("ForceChangePass = %v, want %v", *userInfo.ForceChangePass, tt.wantVal)
			}
		})
	}
}
