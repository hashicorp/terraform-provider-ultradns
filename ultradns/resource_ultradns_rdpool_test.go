package ultradns

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"

	udnssdk "github.com/aliasgharmhowwala/ultradns-sdk-go"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockUltraDNSRecordRDPool struct {
	client *udnssdk.Client
}

func (m *mockUltraDNSRecordRDPool) Create(k udnssdk.RRSetKey, rrset udnssdk.RRSet) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecordRDPool) Select(k udnssdk.RRSetKey) ([]udnssdk.RRSet, error) {

	jsonData := []byte(`
	[{
		"ownerName": "test.provider.ultradns.net",
		"rrtype": "A (1)",
		"ttl": 3600,
		"rdata": [
			"10.0.0.1",
			"10.0.0.2",
			"10.0.0.3"
		],
		"profile": {
			"@context": "http://schemas.ultradns.com/RDPool.jsonschema",
			"order": "ROUND_ROBIN",
			"description": "testing"
		}
	}]
	`)

	rrsets := []udnssdk.RRSet{}
	err := json.Unmarshal(jsonData, &rrsets)
	if err != nil {
		log.Println(err)
	}

	return rrsets, nil

}

func (m *mockUltraDNSRecordRDPool) SelectWithOffset(k udnssdk.RRSetKey, offset int) ([]udnssdk.RRSet, udnssdk.ResultInfo, *http.Response, error) {
	return nil, udnssdk.ResultInfo{}, nil, nil
}

func (m *mockUltraDNSRecordRDPool) Update(udnssdk.RRSetKey, udnssdk.RRSet) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecordRDPool) Delete(k udnssdk.RRSetKey) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecordRDPool) SelectWithOffsetWithLimit(k udnssdk.RRSetKey, offset int, limit int) (rrsets []udnssdk.RRSet, ResultInfo udnssdk.ResultInfo, resp *http.Response, err error) {
	return []udnssdk.RRSet{}, udnssdk.ResultInfo{}, nil, nil
}

func setResourceRecordRDPool() (resourceRecord *schema.Resource) {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			// Required
			"zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rdata": {
				Type:     schema.TypeSet,
				Set:      schema.HashString,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// Optional
			"order": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ROUND_ROBIN",
				ValidateFunc: validation.StringInSlice([]string{
					"ROUND_ROBIN",
					"FIXED",
					"RANDOM",
				}, false),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3600,
			},
			// Computed
			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func compareResourceDataRDPool(t *testing.T, expected *schema.ResourceData, actual *schema.ResourceData) {
	log.Infof("RData values : %+v", expected.Get("rdata"))
	assert.Equal(t, expected.Get("zone"), actual.Get("zone"), true)
	assert.Equal(t, expected.Get("ttl"), actual.Get("ttl"), true)
	assert.Equal(t, expected.Get("type"), actual.Get("type"), true)
	assert.Equal(t, expected.Get("name"), actual.Get("name"), true)
	assert.Equal(t, expected.Get("order"), actual.Get("order"), true)
	assert.Equal(t, expected.Get("description"), actual.Get("description"), true)
	assert.Equal(t, expected.Get("rdata.654229907"), actual.Get("rdata.654229907"), true)
	assert.Equal(t, expected.Get("rdata.3220672553"), actual.Get("rdata.3220672553"), true)
	assert.Equal(t, expected.Get("rdata.3371212991"), actual.Get("rdata.3371212991"), true)

}

func TestUltradnsNewRRSetResourceRecordRDPool(t *testing.T) {
	resourceRecordObj := setResourceRecordRDPool()
	resourceData := resourceRecordObj.TestResourceData()
	resourceData.Set("name", "test.provider.ultradns.net")
	resourceData.Set("ttl", 3600)
	resourceData.Set("rdata", []string{"10.0.0.2", "10.0.0.1", "10.0.0.3"})
	resourceData.Set("zone", "test.provider.ultradns.net")
	resourceData.Set("order", "ROUND_ROBIN")
	resourceData.Set("description", "testing")

	rdpoolProfile := udnssdk.RDPoolProfile{
		Context:     "http://schemas.ultradns.com/RDPool.jsonschema",
		Order:       "ROUND_ROBIN",
		Description: "testing",
	}

	expected := rRSetResource{
		Zone:      "test.provider.ultradns.net",
		OwnerName: "test.provider.ultradns.net",
		TTL:       3600,
		RRType:    "A",
		RData:     []string{"10.0.0.2", "10.0.0.3", "10.0.0.1"},
		Profile:   rdpoolProfile.RawProfile(),
	}

	res, _ := newRRSetResourceFromRdpool(resourceData)
	log.Infof("Actual: %+v", res)
	log.Infof("Expected : %+v", expected)
	assert.Equal(t, reflect.DeepEqual(expected, res), true)
}

func TestResourceUltradnsRDPoolCreate(t *testing.T) {
	expectedResourceRecordObj := setResourceRecordRDPool()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecordRDPool()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecordRDPool{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("ttl", 3600)
	actualData.Set("rdata", []string{"10.0.0.2", "10.0.0.1", "10.0.0.3"})
	actualData.Set("zone", "test.provider.ultradns.net")
	actualData.Set("order", "ROUND_ROBIN")
	actualData.Set("description", "testing")

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("ttl", 3600)
	expectedData.Set("rdata", []string{"10.0.0.2", "10.0.0.1", "10.0.0.3"})
	expectedData.Set("zone", "test.provider.ultradns.net")
	expectedData.Set("order", "ROUND_ROBIN")
	expectedData.Set("description", "testing")

	resourceUltradnsRdpoolCreate(actualData, client)
	compareResourceDataRDPool(t, expectedData, actualData)

}

func TestResourceUltradnsRDPoolRead(t *testing.T) {
        expectedResourceRecordObj := setResourceRecordRDPool()
        expectedData := expectedResourceRecordObj.TestResourceData()
        resourceRecordObject := setResourceRecordRDPool()
        actualData := resourceRecordObject.TestResourceData()
        mocked := mockUltraDNSRecordRDPool{}

        client := &udnssdk.Client{
                RRSets: &mocked,
        }

        actualData.Set("name", "test.provider.ultradns.net")
        actualData.Set("zone", "test.provider.ultradns.net")

        expectedData.Set("name", "test.provider.ultradns.net")
        expectedData.Set("ttl", 3600)
        expectedData.Set("rdata", []string{"10.0.0.2", "10.0.0.1", "10.0.0.3"})
        expectedData.Set("zone", "test.provider.ultradns.net")
        expectedData.Set("order", "ROUND_ROBIN")
        expectedData.Set("description", "testing")

        resourceUltradnsRdpoolRead(actualData, client)
        compareResourceDataRDPool(t, expectedData, actualData)

}


func TestResourceUltradnsRDPoolUpdate(t *testing.T) {
	expectedResourceRecordObj := setResourceRecordRDPool()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecordRDPool()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecordRDPool{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("ttl", 3600)
	actualData.Set("rdata", []string{"10.0.0.2", "10.0.0.1", "10.0.0.3"})
	actualData.Set("zone", "test.provider.ultradns.net")
	actualData.Set("order", "ROUND_ROBIN")
	actualData.Set("description", "testing")

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("ttl", 3600)
	expectedData.Set("rdata", []string{"10.0.0.2", "10.0.0.1", "10.0.0.3"})
	expectedData.Set("zone", "test.provider.ultradns.net")
	expectedData.Set("order", "ROUND_ROBIN")
	expectedData.Set("description", "testing")

	resourceUltradnsRdpoolUpdate(actualData, client)
	compareResourceDataRDPool(t, expectedData, actualData)

}

func TestResourceUltradnsRDPoolDelete(t *testing.T) {
	expectedResourceRecordObj := setResourceRecordRDPool()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecordRDPool()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecordRDPool{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("ttl", 3600)
	actualData.Set("rdata", []string{"10.0.0.2", "10.0.0.1", "10.0.0.3"})
	actualData.Set("zone", "test.provider.ultradns.net")
	actualData.Set("order", "ROUND_ROBIN")
	actualData.Set("description", "testing")

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("ttl", 3600)
	expectedData.Set("rdata", []string{"10.0.0.2", "10.0.0.1", "10.0.0.3"})
	expectedData.Set("zone", "test.provider.ultradns.net")
	expectedData.Set("order", "ROUND_ROBIN")
	expectedData.Set("description", "testing")

	resourceUltradnsRdpoolDelete(actualData, client)
	compareResourceDataRDPool(t, expectedData, actualData)

}

func TestResourceUltradnsRDPoolImport(t *testing.T) {
	mocked := mockUltraDNSRecordRDPool{}
	client := &udnssdk.Client{
		RRSets: &mocked,
	}
	resourceRecordObj := setResourceRecordRDPool()
	d := resourceRecordObj.TestResourceData()
	d.SetId("test:test.provider.ultradns.net")
	newRecordData, _ := resourceUltradnsRdpoolImport(d, client)
	assert.Equal(t, newRecordData[0].Get("name"), "test", true)
	assert.Equal(t, newRecordData[0].Get("zone"), "test.provider.ultradns.net", true)

}

//Testcase to check fail case
func TestResourceUltradnsRDPoolImportFailCase(t *testing.T) {
        mocked := mockUltraDNSRecordRDPool{}
        client := &udnssdk.Client{
                RRSets: &mocked,
        }
        resourceRecordObj := setResourceRecordRDPool()
        d := resourceRecordObj.TestResourceData()
        d.SetId("testabc.test.provider.ultradns.net")
        _,err := resourceUltradnsRecordImport(d, client)
        log.Errorf("Error: %+v",err)
        assert.NotNil(t,err,true)

}



func TestAccUltradnsRdpool(t *testing.T) {
	var record udnssdk.RRSet
	domain, _ := os.LookupEnv("ULTRADNS_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRdpoolCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testCfgRdpoolMinimal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_rdpool.it", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "name", "test-rdpool-minimal"),
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "ttl", "300"),

					// hashRdatas(): 10.6.0.1 -> 2847814707
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "rdata.2847814707", "10.6.0.1"),
					// Defaults
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "description", "Minimal RD Pool"),
					// Generated
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "id", fmt.Sprintf("test-rdpool-minimal:%s", domain)),
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "hostname", fmt.Sprintf("test-rdpool-minimal.%s.", domain)),
				),
			},

			{
				ResourceName:      "ultradns_rdpool.it",
				ImportState:       true,
				ImportStateVerify: true,
			},

			{
				Config: fmt.Sprintf(testCfgRdpoolMaximal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_rdpool.it", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "name", "test-rdpool-maximal"),
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "ttl", "300"),
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "description", "traffic controller pool with all settings tuned"),

					// hashRdatas(): 10.6.1.1 -> 2826722820
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "rdata.2826722820", "10.6.1.1"),

					// hashRdatas(): 10.6.1.2 -> 829755326
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "rdata.829755326", "10.6.1.2"),

					// Generated
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "id", fmt.Sprintf("test-rdpool-maximal:%s", domain)),
					resource.TestCheckResourceAttr("ultradns_rdpool.it", "hostname", fmt.Sprintf("test-rdpool-maximal.%s.", domain)),
				),
			},

			{
				ResourceName:      "ultradns_rdpool.it",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testCfgRdpoolMinimal = `
resource "ultradns_rdpool" "it" {
  zone        = "%s"
  name        = "test-rdpool-minimal"
  ttl         = 300
  description = "Minimal RD Pool"
  rdata       = ["10.6.0.1"]
}
`

const testCfgRdpoolMaximal = `
resource "ultradns_rdpool" "it" {
  zone        = "%s"
  name        = "test-rdpool-maximal"
  order       = "ROUND_ROBIN"
  ttl         = 300
  description = "traffic controller pool with all settings tuned"
  rdata       = ["10.6.1.1","10.6.1.2"]
}
`

