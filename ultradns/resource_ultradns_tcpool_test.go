package ultradns

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	udnssdk "github.com/ultradns/ultradns-sdk-go"
)

type mockUltraDNSRecordTCPool struct {
	client *udnssdk.Client
}

func (m *mockUltraDNSRecordTCPool) Create(k udnssdk.RRSetKey, rrset udnssdk.RRSet) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecordTCPool) Select(k udnssdk.RRSetKey) ([]udnssdk.RRSet, error) {

	rrsets := []udnssdk.RRSet{{
		OwnerName: "test.provider.ultradns.net",
		RRType:    "A",
		RData:     []string{"10.1.1.2"},
		TTL:       300,
		Profile: udnssdk.RawProfile{
			"@context":    "http://schemas.ultradns.com/TCPool.jsonschema",
			"description": "testing",
			"runProbes":   false,
			"actOnProbes": false,
			"status":      "OK",
			"backupRecord": map[string]interface{}{
				"rdata":            "10.6.1.4",
				"failoverDelay":    30,
				"availableToServe": true,
			},
			"rdataInfo": []interface{}{map[string]interface{}{
				"state":         "ACTIVE",
				"runProbes":     true,
				"priority":      1,
				"failoverDelay": 30,
				"threshold":     1,
				"weight":        2,
			}},
			"maxToLB": 2,
		},
	}}
	return rrsets, nil

}

func (m *mockUltraDNSRecordTCPool) SelectWithOffset(k udnssdk.RRSetKey, offset int) ([]udnssdk.RRSet, udnssdk.ResultInfo, *http.Response, error) {
	return nil, udnssdk.ResultInfo{}, nil, nil
}

func (m *mockUltraDNSRecordTCPool) Update(udnssdk.RRSetKey, udnssdk.RRSet) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecordTCPool) Delete(k udnssdk.RRSetKey) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecordTCPool) SelectWithOffsetWithLimit(k udnssdk.RRSetKey, offset int, limit int) (rrsets []udnssdk.RRSet, ResultInfo udnssdk.ResultInfo, resp *http.Response, err error) {
	return []udnssdk.RRSet{}, udnssdk.ResultInfo{}, nil, nil
}

func setResourceRecordTCPool() (resourceRecord *schema.Resource) {
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
			"description": {
				Type:     schema.TypeString,
				Required: true,
				// 0-255 char
			},
			"rdata": {
				Type:     schema.TypeSet,
				Set:      hashRdatas,
				Required: true,
				// Valid: len(rdataInfo) == len(rdata)
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required
						"host": {
							Type:     schema.TypeString,
							Required: true,
						},
						// Optional
						"failover_delay": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
							// Valid: 0-30
							// Units: Minutes
						},
						"priority": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						"run_probes": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"state": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "NORMAL",
						},
						"threshold": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						"weight": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  2,
							// Valid: i%2 == 0 && 2 <= i <= 100
						},
					},
				},
			},
			// Optional
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3600,
			},
			"run_probes": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"act_on_probes": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"max_to_lb": {
				Type:     schema.TypeInt,
				Optional: true,
				// Valid: 0 <= i <= len(rdata)
			},
			"backup_record_rdata": {
				Type:     schema.TypeString,
				Optional: true,
				// Valid: IPv4 address or CNAME
			},
			"backup_record_failover_delay": {
				Type:     schema.TypeInt,
				Optional: true,
				// Valid: 0-30
				// Units: Minutes
			},
			// Computed
			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func compareTCPoolResources(t *testing.T, expected *schema.ResourceData, actual *schema.ResourceData) {
	assert.Equal(t, expected.Get("name"), actual.Get("name"), true)
	assert.Equal(t, expected.Get("zone"), actual.Get("zone"), true)
	assert.Equal(t, expected.Get("hostname"), actual.Get("hostname"), true)
	assert.Equal(t, expected.Get("type"), actual.Get("type"), true)
	assert.Equal(t, expected.Get("description"), actual.Get("description"), true)
	assert.Equal(t, expected.Get("backup_record_rdata"), actual.Get("backup_record_rdata"), true)
	assert.Equal(t, expected.Get("backup_record_failover_delay"), actual.Get("backup_record_failover_delay"), true)
	assert.Equal(t, expected.Get("max_to_lb"), actual.Get("max_to_lb"), true)
	assert.Equal(t, expected.Get("run_probes"), actual.Get("run_probes"), true)
	assert.Equal(t, expected.Get("ttl"), actual.Get("ttl"), true)
	assert.Equal(t, expected.Get("act_on_probes"), actual.Get("act_on_probes"), true)
	assert.Equal(t, expected.Get("rdata.0"), actual.Get("rdata.0"), true)

}

func TestNewRRSetResourceFromTCPool(t *testing.T) {
	resourceRecordObj := setResourceRecordTCPool()
	resourceData := resourceRecordObj.TestResourceData()

	expectedResource := rRSetResource{}
	var context udnssdk.ProfileSchema
	context = "http://schemas.ultradns.com/TCPool.jsonschema"
	expectedResource = rRSetResource{
		OwnerName: "test.provider.ultradns.net",
		RRType:    "A",
		RData:     []string{"10.1.1.2"},
		TTL:       300,
		Profile: udnssdk.RawProfile{
			"@context":    context,
			"description": "testing",
			"runProbes":   false,
			"actOnProbes": false,
			"backupRecord": map[string]interface{}{
				"rdata":         "10.6.1.4",
				"failoverDelay": 30,
			},
			"rdataInfo": []interface{}{map[string]interface{}{
				"state":         "ACTIVE",
				"runProbes":     true,
				"priority":      1,
				"failoverDelay": 30,
				"threshold":     1,
				"weight":        2,
			}},
			"maxToLB": 2,
		},
		Zone: "test.provider.ultradns.net",
	}

	rrsetDTO := make([]map[string]interface{}, 1)
	data := []byte(`
		[{
			"host": "10.1.1.2",
			"state": "ACTIVE",
			"run_probes": true,
			"priority": 1,
			"failover_delay": 30,
			"threshold": 1,
			"weight": 2
		}]
	`)

	err := json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	resourceData.Set("name", "test.provider.ultradns.net")
	resourceData.Set("rdata", rrsetDTO)
	resourceData.Set("backup_record_rdata", "10.6.1.4")
	resourceData.Set("backup_record_failover_delay", 30)
	resourceData.Set("max_to_lb", 2)
	resourceData.Set("run_probes", false)
	resourceData.Set("ttl", 300)
	resourceData.Set("act_on_probes", false)
	resourceData.Set("zone", "test.provider.ultradns.net")
	resourceData.Set("type", "A")
	resourceData.Set("description", "testing")

	rrset, _ := newRRSetResourceFromTcpool(resourceData)
	log.Infof("Expected %+v", expectedResource)
	log.Infof("actual  %+v", rrset)
	assert.Equal(t, expectedResource, rrset, true)
	log.Infof("resourceData RData: %+v, err %+v", resourceData.Get("rdata"), err)
}

func TestResourceUltradnsTCPoolRead(t *testing.T) {
	mocked := mockUltraDNSRecordTCPool{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}
	resourceRecordObj := setResourceRecordTCPool()
	resourceData := resourceRecordObj.TestResourceData()
	expectedResourceRecordObj := setResourceRecordTCPool()
	expectedResourceData := expectedResourceRecordObj.TestResourceData()
	rrsetDTO := make([]map[string]interface{}, 1)
	data := []byte(`
		[{
			"host": "10.1.1.2",
			"state": "ACTIVE",
			"run_probes": true,
			"priority": 1,
			"failover_delay": 30,
			"threshold": 1,
			"weight": 2
		}]
	`)

	err := json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	expectedResourceData.Set("name", "test.provider.ultradns.net")
	expectedResourceData.Set("rdata", rrsetDTO)
	expectedResourceData.Set("backup_record_rdata", "10.6.1.4")
	expectedResourceData.Set("backup_record_failover_delay", 30)
	expectedResourceData.Set("max_to_lb", 2)
	expectedResourceData.Set("run_probes", false)
	expectedResourceData.Set("ttl", 300)
	expectedResourceData.Set("act_on_probes", false)
	expectedResourceData.Set("zone", "test.provider.ultradns.net")
	expectedResourceData.Set("type", "A")
	expectedResourceData.Set("description", "testing")
	expectedResourceData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	resourceData.Set("name", "test.provider.ultradns.net")
	resourceData.Set("zone", "test.provider.ultradns.net")

	resourceUltradnsTcpoolRead(resourceData, client)
	compareTCPoolResources(t, expectedResourceData, resourceData)
}

func TestResourceUltradnsTCPoolCreate(t *testing.T) {
	expectedResourceRecordObj := setResourceRecordTCPool()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecordTCPool()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecordTCPool{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	rrsetDTO := make([]map[string]interface{}, 1)
	data := []byte(`
			[{
					"host": "10.1.1.2",
					"state": "ACTIVE",
					"run_probes": true,
					"priority": 1,
					"failover_delay": 30,
					"threshold": 1,
					"weight": 2
			}]
	`)

	err := json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("rdata", rrsetDTO)
	actualData.Set("backup_record_rdata", "10.6.1.4")
	actualData.Set("backup_record_failover_delay", 30)
	actualData.Set("max_to_lb", 2)
	actualData.Set("run_probes", false)
	actualData.Set("ttl", 300)
	actualData.Set("act_on_probes", false)
	actualData.Set("zone", "test.provider.ultradns.net")
	actualData.Set("type", "A")
	actualData.Set("description", "testing")

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("rdata", rrsetDTO)
	expectedData.Set("backup_record_rdata", "10.6.1.4")
	expectedData.Set("backup_record_failover_delay", 30)
	expectedData.Set("max_to_lb", 2)
	expectedData.Set("run_probes", false)
	expectedData.Set("ttl", 300)
	expectedData.Set("act_on_probes", false)
	expectedData.Set("zone", "test.provider.ultradns.net")
	expectedData.Set("type", "A")
	expectedData.Set("description", "testing")
	expectedData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	resourceUltradnsTcpoolCreate(actualData, client)
	compareTCPoolResources(t, expectedData, actualData)

}

func TestResourceUltradnsTCPoolUpdate(t *testing.T) {
	expectedResourceRecordObj := setResourceRecordTCPool()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecordTCPool()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecordTCPool{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	rrsetDTO := make([]map[string]interface{}, 1)
	data := []byte(`
			[{
					"host": "10.1.1.2",
					"state": "ACTIVE",
					"run_probes": true,
					"priority": 1,
					"failover_delay": 30,
					"threshold": 1,
					"weight": 2
			}]
	`)

	err := json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("rdata", rrsetDTO)
	actualData.Set("backup_record_rdata", "10.6.1.4")
	actualData.Set("backup_record_failover_delay", 30)
	actualData.Set("max_to_lb", 2)
	actualData.Set("run_probes", false)
	actualData.Set("ttl", 300)
	actualData.Set("act_on_probes", false)
	actualData.Set("zone", "test.provider.ultradns.net")
	actualData.Set("type", "A")
	actualData.Set("description", "testing")

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("rdata", rrsetDTO)
	expectedData.Set("backup_record_rdata", "10.6.1.4")
	expectedData.Set("backup_record_failover_delay", 30)
	expectedData.Set("max_to_lb", 2)
	expectedData.Set("run_probes", false)
	expectedData.Set("ttl", 300)
	expectedData.Set("act_on_probes", false)
	expectedData.Set("zone", "test.provider.ultradns.net")
	expectedData.Set("type", "A")
	expectedData.Set("description", "testing")
	expectedData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	resourceUltradnsTcpoolUpdate(actualData, client)
	compareTCPoolResources(t, expectedData, actualData)

}

func TestResourceUltradnsTCPoolDelete(t *testing.T) {

	expectedResourceRecordObj := setResourceRecordTCPool()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecordTCPool()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecordTCPool{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	rrsetDTO := make([]map[string]interface{}, 1)
	data := []byte(`
			[{
					"host": "10.1.1.2",
					"state": "ACTIVE",
					"run_probes": true,
					"priority": 1,
					"failover_delay": 30,
					"threshold": 1,
					"weight": 2
			}]
	`)

	err := json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("rdata", rrsetDTO)
	actualData.Set("backup_record_rdata", "10.6.1.4")
	actualData.Set("backup_record_failover_delay", 30)
	actualData.Set("max_to_lb", 2)
	actualData.Set("run_probes", false)
	actualData.Set("ttl", 300)
	actualData.Set("act_on_probes", false)
	actualData.Set("zone", "test.provider.ultradns.net")
	actualData.Set("type", "A")
	actualData.Set("description", "testing")
	actualData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("rdata", rrsetDTO)
	expectedData.Set("backup_record_rdata", "10.6.1.4")
	expectedData.Set("backup_record_failover_delay", 30)
	expectedData.Set("max_to_lb", 2)
	expectedData.Set("run_probes", false)
	expectedData.Set("ttl", 300)
	expectedData.Set("act_on_probes", false)
	expectedData.Set("zone", "test.provider.ultradns.net")
	expectedData.Set("type", "A")
	expectedData.Set("description", "testing")
	expectedData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	resourceUltradnsTcpoolDelete(actualData, client)
	compareTCPoolResources(t, expectedData, actualData)

}

func TestResourceUltradnsTCPoolImport(t *testing.T) {
	mocked := mockUltraDNSRecordTCPool{}
	client := &udnssdk.Client{
		RRSets: &mocked,
	}
	resourceRecordObj := setResourceRecordTCPool()
	d := resourceRecordObj.TestResourceData()
	d.SetId("test:test.provider.ultradns.net:A")
	newRecordData, _ := resourceUltradnsTcpoolImport(d, client)
	assert.Equal(t, newRecordData[0].Get("name"), "test", true)
	assert.Equal(t, newRecordData[0].Get("zone"), "test.provider.ultradns.net", true)

}

//Testcase to check fail case
func TestResourceUltradnsTCPoolImportFailCase(t *testing.T) {
	mocked := mockUltraDNSRecordTCPool{}
	client := &udnssdk.Client{
		RRSets: &mocked,
	}
	resourceRecordObj := setResourceRecordTCPool()
	d := resourceRecordObj.TestResourceData()
	d.SetId("testabc.test.provider.ultradns.net")
	_, err := resourceUltradnsTcpoolImport(d, client)
	log.Errorf("Error: %+v", err)
	assert.NotNil(t, err, true)

}

func TestAccUltradnsTcpool(t *testing.T) {
	var record udnssdk.RRSet
	domain, _ := os.LookupEnv("ULTRADNS_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccTcpoolCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testCfgTcpoolMinimal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_tcpool.it", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "name", "test-tcpool-minimal"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "ttl", "300"),

					// hashRdatas(): 10.6.0.1 -> 2847814707
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2847814707.host", "10.6.0.1"),
					// Defaults
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "act_on_probes", "true"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "description", "Minimal TC Pool"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "max_to_lb", "0"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "run_probes", "true"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2847814707.failover_delay", "0"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2847814707.priority", "1"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2847814707.run_probes", "true"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2847814707.state", "NORMAL"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2847814707.threshold", "1"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2847814707.weight", "2"),
					// Generated
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "id", fmt.Sprintf("test-tcpool-minimal:%s:A", domain)),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "hostname", fmt.Sprintf("test-tcpool-minimal.%s.", domain)),
				),
			},
			{
				Config: fmt.Sprintf(testCfgTcpoolMaximal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_tcpool.it", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "name", "test-tcpool-maximal"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "ttl", "300"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "description", "traffic controller pool with all settings tuned"),

					resource.TestCheckResourceAttr("ultradns_tcpool.it", "act_on_probes", "false"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "max_to_lb", "2"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "run_probes", "false"),

					// hashRdatas(): 10.6.1.1 -> 2826722820
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2826722820.host", "10.6.1.1"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2826722820.failover_delay", "30"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2826722820.priority", "1"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2826722820.run_probes", "true"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2826722820.state", "ACTIVE"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2826722820.threshold", "1"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.2826722820.weight", "2"),

					// hashRdatas(): 10.6.1.2 -> 829755326
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.829755326.host", "10.6.1.2"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.829755326.failover_delay", "30"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.829755326.priority", "2"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.829755326.run_probes", "true"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.829755326.state", "INACTIVE"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.829755326.threshold", "1"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.829755326.weight", "4"),

					// hashRdatas(): 10.6.1.3 -> 1181892392
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.1181892392.host", "10.6.1.3"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.1181892392.failover_delay", "30"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.1181892392.priority", "3"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.1181892392.run_probes", "false"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.1181892392.state", "NORMAL"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.1181892392.threshold", "1"),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "rdata.1181892392.weight", "8"),
					// Generated
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "id", fmt.Sprintf("test-tcpool-maximal:%s:A", domain)),
					resource.TestCheckResourceAttr("ultradns_tcpool.it", "hostname", fmt.Sprintf("test-tcpool-maximal.%s.", domain)),
				),
			},

			{
				ResourceName:      "ultradns_tcpool.it",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testCfgTcpoolMinimal = `
resource "ultradns_tcpool" "it" {
  zone        = "%s"
  name        = "test-tcpool-minimal"
  ttl         = 300
  description = "Minimal TC Pool"

  rdata {
	host = "10.6.0.1"
  }
}
`

const testCfgTcpoolMaximal = `
resource "ultradns_tcpool" "it" {
  zone        = "%s"
  name        = "test-tcpool-maximal"
  ttl         = 300
  description = "traffic controller pool with all settings tuned"

  act_on_probes = false
  max_to_lb     = 2
  run_probes    = false

  rdata {
	host = "10.6.1.1"

	failover_delay = 30
	priority       = 1
	run_probes     = true
	state          = "ACTIVE"
	threshold      = 1
	weight         = 2
  }

  rdata {
	host = "10.6.1.2"

	failover_delay = 30
	priority       = 2
	run_probes     = true
	state          = "INACTIVE"
	threshold      = 1
	weight         = 4
  }

  rdata {
	host = "10.6.1.3"

	failover_delay = 30
	priority       = 3
	run_probes     = false
	state          = "NORMAL"
	threshold      = 1
	weight         = 8
  }

  backup_record_rdata          = "10.6.1.4"
  backup_record_failover_delay = 30
}
`
