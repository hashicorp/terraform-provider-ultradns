package ultradns

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	udnssdk "github.com/ultradns/ultradns-sdk-go"
)

type mockUltraDNSRecord struct {
	client *udnssdk.Client
}

func (m *mockUltraDNSRecord) Create(k udnssdk.RRSetKey, rrset udnssdk.RRSet) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecord) Select(k udnssdk.RRSetKey) ([]udnssdk.RRSet, error) {

	return []udnssdk.RRSet{{
		OwnerName: "test-ultradns-provider.com.",
		RRType:    "A",
		RData:     []string{"10.0.0.1"},
		TTL:       3600,
	}}, nil

}

func (m *mockUltraDNSRecord) SelectWithOffset(k udnssdk.RRSetKey, offset int) ([]udnssdk.RRSet, udnssdk.ResultInfo, *http.Response, error) {
	return nil, udnssdk.ResultInfo{}, nil, nil
}

func (m *mockUltraDNSRecord) Update(udnssdk.RRSetKey, udnssdk.RRSet) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecord) Delete(k udnssdk.RRSetKey) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecord) SelectWithOffsetWithLimit(k udnssdk.RRSetKey, offset int, limit int) (rrsets []udnssdk.RRSet, ResultInfo udnssdk.ResultInfo, resp *http.Response, err error) {
	return []udnssdk.RRSet{{
		OwnerName: "test-ultradns-provider.com.",
		RRType:    "A",
		RData:     []string{"10.5.0.1"},
		TTL:       86400,
	}}, udnssdk.ResultInfo{}, nil, nil
}

func setResourceRecord() (resourceRecord *schema.Resource) {
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
			"type": {
				Type:     schema.TypeString,
				Required: true,
				//ForceNew: true,
			},
			"rdata": {
				Type:     schema.TypeSet,
				Set:      schema.HashString,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// Optional
			"ttl": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "3600",
			},
			// Computed
			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func compareResourceData(t *testing.T, expected *schema.ResourceData, actual *schema.ResourceData) {

	assert.Equal(t, expected.Get("zone"), actual.Get("zone"), true)
	assert.Equal(t, expected.Get("ttl"), actual.Get("ttl"), true)
	assert.Equal(t, expected.Get("type"), actual.Get("type"), true)
	assert.Equal(t, expected.Get("rdata.654229907"), actual.Get("rdata.654229907"), true)

}

func TestUltradnsNewRRSetResourceRecord(t *testing.T) {
	resourceRecordObj := setResourceRecord()
	resourceData := resourceRecordObj.TestResourceData()
	resourceData.Set("name", "test.provider.ultradns.net")
	resourceData.Set("ttl", "3600")
	resourceData.Set("type", "A")
	resourceData.Set("rdata", []string{"10.0.0.1"})
	resourceData.Set("zone", "test.provider.ultradns.net")

	expected := rRSetResource{
		Zone:      "test.provider.ultradns.net",
		OwnerName: "test.provider.ultradns.net",
		TTL:       3600,
		RRType:    "A",
		RData:     []string{"10.0.0.1"},
	}

	res, _ := newRRSetResource(resourceData)
	assert.Equal(t, reflect.DeepEqual(expected, res), true)
}

func TestPopulateResourceDataFromRRSet(t *testing.T) {
	resourceRecordObj := setResourceRecord()
	expectedResourceRecordObj := setResourceRecord()
	expectedData := expectedResourceRecordObj.TestResourceData()

	expectedData.Set("ttl", "3600")
	expectedData.Set("type", "A")
	expectedData.Set("rdata", []string{"10.0.0.1"})
	expectedData.Set("zone", "test.provider.ultradns.net")

	d := resourceRecordObj.TestResourceData()
	d.Set("zone", "test.provider.ultradns.net")

	rRSetCase1 := []udnssdk.RRSet{{
		OwnerName: "test.",
		TTL:       3600,
		RRType:    "A",
		RData:     []string{"10.0.0.1"},
	}}

	rRSetCase2 := []udnssdk.RRSet{{
		OwnerName: "",
		TTL:       3600,
		RRType:    "A",
		RData:     []string{"10.0.0.1"},
	}}

	rRSetCase3 := []udnssdk.RRSet{{
		OwnerName: "test.provider.ultradns.net",
		TTL:       3600,
		RRType:    "A",
		RData:     []string{"10.0.0.1"},
	}}

	rRSetCase4 := []udnssdk.RRSet{{
		OwnerName: "test.provider.ultradns.net",
		TTL:       3600,
		RRType:    "TXT",
		RData:     []string{"Text one Test com"},
	}}

	//Case 1 when the owner name has suffix dot
	log.Infof("Case 1 when the owner name has suffix dot")
	expectedData.Set("hostname", "test.")
	populateResourceDataFromRRSet(rRSetCase1, d)
	compareResourceData(t, expectedData, d)

	//Case 2 When the owner name is empty
	log.Infof("Case 2 When the owner name is empty")
	expectedData.Set("name", "")
	populateResourceDataFromRRSet(rRSetCase2, d)
	compareResourceData(t, expectedData, d)

	//Case 3 when we provide normal owner name
	log.Infof("Case 3 when we provide normal owner name")
	expectedData.Set("name", "test.provider.ultradns.net")
	populateResourceDataFromRRSet(rRSetCase3, d)
	compareResourceData(t, expectedData, d)

	//Case 4 When we are sending txt record
	log.Infof("Case 4 when we provide normal owner name")
	expectedData.Set("name", "test.provider.ultradns.net")
	d.Set("type", "TXT")
	expectedData.Set("type", "TXT")
	expectedData.Set("rdata", []string{"Text one Test com"})
	populateResourceDataFromRRSet(rRSetCase4, d)
	compareResourceData(t, expectedData, d)
}

func TestResourceUltradnsRecordCreate(t *testing.T) {
	expectedResourceRecordObj := setResourceRecord()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecord()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecord{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("ttl", "3600")
	actualData.Set("rdata", []string{"10.0.0.1"})
	actualData.Set("type", "A")
	actualData.Set("zone", "test.provider.ultradns.net")

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("ttl", "3600")
	expectedData.Set("rdata", []string{"10.0.0.1"})
	expectedData.Set("type", "A")
	expectedData.Set("zone", "test.provider.ultradns.net")

	resourceUltraDNSRecordCreate(actualData, client)
	compareResourceData(t, expectedData, actualData)

}

func TestResourceUltradnsRecordUpdate(t *testing.T) {
	expectedResourceRecordObj := setResourceRecord()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecord()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecord{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("ttl", "3600")
	actualData.Set("type", "A")
	actualData.Set("rdata", []string{"10.0.0.1"})
	actualData.Set("zone", "test.provider.ultradns.net")

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("ttl", "3600")
	expectedData.Set("type", "A")
	expectedData.Set("rdata", []string{"10.0.0.1"})
	expectedData.Set("zone", "test.provider.ultradns.net")

	resourceUltraDNSRecordUpdate(actualData, client)
	compareResourceData(t, expectedData, actualData)

}

func TestResourceUltradnsRecordDelete(t *testing.T) {
	expectedResourceRecordObj := setResourceRecord()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecord()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecord{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("ttl", "3600")
	actualData.Set("type", "A")
	actualData.Set("rdata", []string{"10.0.0.1"})
	actualData.Set("zone", "test.provider.ultradns.net")

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("ttl", "3600")
	expectedData.Set("rdata", []string{"10.0.0.1"})
	expectedData.Set("type", "A")
	expectedData.Set("zone", "test.provider.ultradns.net")

	resourceUltraDNSRecordDelete(actualData, client)
	compareResourceData(t, expectedData, actualData)

}

//Testcase to check proper split of iD into appropriate fields
func TestResourceUltradnsRecordImport(t *testing.T) {
	mocked := mockUltraDNSRecord{}
	client := &udnssdk.Client{
		RRSets: &mocked,
	}
	resourceRecordObj := setResourceRecord()
	d := resourceRecordObj.TestResourceData()
	d.SetId("test:test.provider.ultradns.net")
	newRecordData, _ := resourceUltradnsRecordImport(d, client)
	assert.Equal(t, newRecordData[0].Get("name"), "test", true)
	assert.Equal(t, newRecordData[0].Get("zone"), "test.provider.ultradns.net", true)

	//Case 1 when using type for same recordname
	d.SetId("test:test.provider.ultradns.net:MX")
	newRecordData1, _ := resourceUltradnsRecordImport(d, client)
	assert.Equal(t, newRecordData1[0].Get("name"), "test", true)
	assert.Equal(t, newRecordData1[0].Get("zone"), "test.provider.ultradns.net", true)
	assert.Equal(t, newRecordData1[0].Get("type"), "MX", true)

}

//Testcase to check fail case
func TestResourceUltradnsRecordImportFailCase(t *testing.T) {
	mocked := mockUltraDNSRecord{}
	client := &udnssdk.Client{
		RRSets: &mocked,
	}
	resourceRecordObj := setResourceRecord()
	d := resourceRecordObj.TestResourceData()
	d.SetId("testabc.test.provider.ultradns.net")
	_, err := resourceUltradnsRecordImport(d, client)
	log.Errorf("Error: %+v", err)
	assert.NotNil(t, err, true)

}

func TestAccUltradnsRecord(t *testing.T) {
	var record udnssdk.RRSet
	domain, _ := os.LookupEnv("ULTRADNS_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRecordCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testCfgRecordMinimal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_record.it", &record),
					resource.TestCheckResourceAttr("ultradns_record.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_record.it", "name", "test-record"),
					resource.TestCheckResourceAttr("ultradns_record.it", "rdata.3994963683", "10.5.0.1"),
				),
			},
			{
				Config: fmt.Sprintf(testCfgRecordMinimal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_record.it", &record),
					resource.TestCheckResourceAttr("ultradns_record.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_record.it", "name", "test-record"),
					resource.TestCheckResourceAttr("ultradns_record.it", "rdata.3994963683", "10.5.0.1"),
				),
			},
			{
				Config: fmt.Sprintf(testCfgRecordUpdated, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_record.it", &record),
					resource.TestCheckResourceAttr("ultradns_record.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_record.it", "name", "test-record"),
					resource.TestCheckResourceAttr("ultradns_record.it", "rdata.1998004057", "10.5.0.2"),
				),
			},

			{
				ResourceName:      "ultradns_record.it",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUltradnsRecordTXT(t *testing.T) {
	var record udnssdk.RRSet
	domain, _ := os.LookupEnv("ULTRADNS_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRecordCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testCfgRecordTXTMinimal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_record.it", &record),
					resource.TestCheckResourceAttr("ultradns_record.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_record.it", "name", "test-record-txt"),
					resource.TestCheckResourceAttr("ultradns_record.it", "rdata.1447448707", "simple answer"),
					resource.TestCheckResourceAttr("ultradns_record.it", "rdata.3337444205", "backslash answer \\"),
					resource.TestCheckResourceAttr("ultradns_record.it", "rdata.3135730072", "quote answer \""),
					resource.TestCheckResourceAttr("ultradns_record.it", "rdata.126343430", "complex answer \\ \""),
				),
			},
		},
	})
}

func testAccRecordCheckDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*udnssdk.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ultradns_record" {
			continue
		}

		k := udnssdk.RRSetKey{
			Zone: rs.Primary.Attributes["zone"],
			Name: rs.Primary.Attributes["name"],
			Type: rs.Primary.Attributes["type"],
		}

		_, err := client.RRSets.Select(k)

		if err == nil {
			return fmt.Errorf("Record still exists")
		}
	}

	return nil
}

const testCfgRecordMinimal = `
resource "ultradns_record" "it" {
  zone = "%s"
  name  = "test-record"
  rdata = ["10.5.0.1"]
  type  = "A"
  ttl   = 3600
}
`

const testCfgRecordUpdated = `
resource "ultradns_record" "it" {
  zone = "%s"
  name  = "test-record"
  rdata = ["10.5.0.2"]
  type  = "A"
  ttl   = 3600
}
`

const testCfgRecordTXTMinimal = `
resource "ultradns_record" "it" {
  zone = "%s"
  name  = "test-record-txt"
  rdata = [
    "simple answer",
    "backslash answer \\",
    "quote answer \"",
    "complex answer \\ \"",
  ]
  type  = "TXT"
  ttl   = 3600
}
`
