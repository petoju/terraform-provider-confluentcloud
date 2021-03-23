package ccloud

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"confluentcloud": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func TestProvider_login(t *testing.T) {
	ctx := context.Background()
	for _, name := range []string{"CONFLUENT_CLOUD_USERNAME", "CONFLUENT_CLOUD_PASSWORD"} {
		if v := os.Getenv(name); v == "" {
			t.Fatal("CONFLUENT_CLOUD_USERNAME and CONFLUENT_CLOUD_PASSWORD must be set for acceptance tests")
		}
	}

	err := testAccProvider.Configure(ctx, terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}
