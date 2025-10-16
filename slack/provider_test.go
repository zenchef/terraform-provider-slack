package slack

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

type testUser struct {
	id    string
	name  string
	email string
}

var (
	testUserCreator = testUser{
		id:    "U01D6L97N0M",
		name:  "contact",
		email: "contact@pablovarela.co.uk",
	}

	testUser00 = testUser{
		id:    "U01D31S1GUE",
		name:  "contact_test-user-ter",
		email: "contact+test-user-terraform-provider-slack-00@pablovarela.co.uk",
	}

	testUser01 = testUser{
		id:    "U01DZK10L1W",
		name:  "contact_test-user-206",
		email: "contact+test-user-terraform-provider-slack-01@pablovarela.co.uk",
	}
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"slack": providerserver.NewProtocol6WithError(NewFrameworkProvider("test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("SLACK_TOKEN"); v == "" {
		t.Fatal("SLACK_TOKEN must be set for acceptance tests")
	}
}
