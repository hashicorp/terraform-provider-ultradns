package ultradns

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func setSchemaProvider() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ULTRADNS_USERNAME", nil),
				Description: "UltraDNS Username.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ULTRADNS_PASSWORD", nil),
				Description: "UltraDNS User Password",
			},
			"baseurl": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ULTRADNS_BASEURL", nil),
				Description: "UltraDNS Base URL",
			},
		},
	}
}

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"ultradns": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func TestProviderConfigure(t *testing.T) {

	resourceRecordObj := setSchemaProvider()
	resourceData := resourceRecordObj.TestResourceData()

	resourceData.Set("username", "abcd")
	resourceData.Set("password", "abcd")
	resourceData.Set("baseurl", "abcd")

	config := Config{
		Username: "abcd",
		Password: "abcd",
		BaseURL:  "abcd",
	}

	expected, _ := config.Client()

	actualData, _ := providerConfigure(resourceData)
	log.Infof("%#v", actualData)
	assert.Equal(t, expected, actualData, true)

}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ULTRADNS_USERNAME"); v == "" {
		t.Fatal("ULTRADNS_USERNAME must be set for acceptance tests")
	}

	if v := os.Getenv("ULTRADNS_PASSWORD"); v == "" {
		t.Fatal("ULTRADNS_PASSWORD must be set for acceptance tests")
	}

	if v := os.Getenv("ULTRADNS_DOMAIN"); v == "" {
		t.Fatal("ULTRADNS_DOMAIN must be set for acceptance tests. The domain is used to create and destroy record against.")
	}

	if v := os.Getenv("ULTRADNS_BASEURL"); v == "" {
		t.Fatal("ULTRADNS_BASEURL must be set for acceptance tests")
	}
}
