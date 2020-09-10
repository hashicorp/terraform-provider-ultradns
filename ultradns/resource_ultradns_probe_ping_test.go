package ultradns

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	udnssdk "github.com/aliasgharmhowwala/ultradns-sdk-go"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockUltraDNSRecordProbePing struct {
	client *udnssdk.Client
}

func (m *mockUltraDNSRecordProbePing) Create(k udnssdk.RRSetKey, rrset udnssdk.RRSet) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecordProbePing) Select(k udnssdk.RRSetKey) ([]udnssdk.RRSet, error) {

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
                        "@context": "http://schemas.ultradns.com/ProbePing.jsonschema",
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

func (m *mockUltraDNSRecordProbePing) SelectWithOffset(k udnssdk.RRSetKey, offset int) ([]udnssdk.RRSet, udnssdk.ResultInfo, *http.Response, error) {
	return nil, udnssdk.ResultInfo{}, nil, nil
}

func (m *mockUltraDNSRecordProbePing) Update(udnssdk.RRSetKey, udnssdk.RRSet) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecordProbePing) Delete(k udnssdk.RRSetKey) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecordProbePing) SelectWithOffsetWithLimit(k udnssdk.RRSetKey, offset int, limit int) (rrsets []udnssdk.RRSet, ResultInfo udnssdk.ResultInfo, resp *http.Response, err error) {
	return []udnssdk.RRSet{}, udnssdk.ResultInfo{}, nil, nil
}

func setResourceRecordPingProbe() (resourceRecord *schema.Resource) {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			// Key
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
			"pool_record": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			// Required
			"agents": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"threshold": {
				Type:     schema.TypeInt,
				Required: true,
			},
			// Optional
			"interval": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "FIVE_MINUTES",
			},
			"ping_probe": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     schemaPingProbe(),
			},
			// Computed
			"ping_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func compareResourcesProbePing(t *testing.T, actual probeResource, expected probeResource, expectedDetails udnssdk.PingProbeDetailsDTO) {
	assert.Equal(t, expected.Name, actual.Name, true)
	assert.Equal(t, expected.Zone, actual.Zone, true)
	assert.Equal(t, expected.ID, actual.ID, true)
	assert.Equal(t, expectedDetails, actual.Details.Detail, true)

}

func compareResourcesDataProbePing(t *testing.T, actual *schema.ResourceData, expected *schema.ResourceData) {
	assert.Equal(t, expected.Get("name"), actual.Get("name"), true)
	assert.Equal(t, expected.Get("zone"), actual.Get("zone"), true)
	assert.Equal(t, expected.Get("threshold"), actual.Get("threshold"), true)
	assert.Equal(t, expected.Get("interval"), actual.Get("interval"), true)
	assert.Equal(t, expected.Get("pool_record"), actual.Get("pool_record"), true)
	assert.Equal(t, expected.Get("id"), actual.Get("id"), true)
	assert.Equal(t, expected.Get("agents"), actual.Get("agents"), true)
	assert.Equal(t, expected.Get("ping_probe.limit"), actual.Get("ping_probe.limit"), true)
	assert.Equal(t, expected.Get("ping_probe.packets"), actual.Get("ping_probe.packets"), true)
	assert.Equal(t, expected.Get("ping_probe.packet_size"), actual.Get("ping_probe.packet_size"), true)

}

func TestMakePingProbeResource(t *testing.T) {
	resourceRecordObj := setResourceRecordPingProbe()
	resourceData := resourceRecordObj.TestResourceData()
	expectedData := []byte(`{"id":"0608485259D5AC79","type":"PING","interval":"ONE_MINUTE","agents":["DALLAS","AMSTERDAM"],"threshold":2,"details":{"packets":15,"packetSize":56,"limit":{"lossPercent":{"warning":1,"critical":2,"fail":3},"total":{"warning":2,"critical":3,"fail":4}}}}`)
	expectedResource := probeResource{}
	DetailsDTO := udnssdk.PingProbeDetailsDTO{
		Packets:    15,
		PacketSize: 56,
		Limits: map[string]udnssdk.ProbeDetailsLimitDTO{
			"lossPercent": udnssdk.ProbeDetailsLimitDTO{
				Warning:  1,
				Critical: 2,
				Fail:     3,
			},
			"total": udnssdk.ProbeDetailsLimitDTO{
				Warning:  2,
				Critical: 3,
				Fail:     4,
			},
		},
	}

	err := json.Unmarshal(expectedData, &expectedResource)

	if err != nil {
		log.Println(err)
	}
	expectedResource.Name = "test.provider.ultradns.net"
	expectedResource.Zone = "test.provider.ultradns.net"
	pingProbeDTO := make([]map[string]interface{}, 1)
	pingProbe := []byte(`
		{
		"packets": 15,
		"packet_size": 56,
		"limit": [
					{
						"name": "lossPercent",
						"warning": 1,
						"critical": 2,
						"fail": 3
					},
					{
						"name": "total",
						"warning": 2,
						"critical": 3,
						"fail": 4
					}
				]
		}
	`)

	json.Unmarshal(pingProbe, &pingProbeDTO[0])
	resourceData.Set("name", "test.provider.ultradns.net")
	resourceData.Set("zone", "test.provider.ultradns.net")
	resourceData.SetId("0608485259D5AC79")
	resourceData.Set("agents", []string{"DALLAS", "AMSTERDAM"})
	resourceData.Set("interval", "ONE_MINUTE")
	resourceData.Set("threshold", 2)
	resourceData.Set("ping_probe", pingProbeDTO)
	res, _ := makePingProbeResource(resourceData)
	compareResourcesProbePing(t, res, expectedResource, DetailsDTO)

}

func TestMakePingProbeDetails(t *testing.T) {
	pingProbeDTO := make([]map[string]interface{}, 1)
	pingProbe := []byte(`
		{
		"packets": 15,
		"packet_size": 56,
		"limit": [
				{
					"name": "lossPercent",
					"warning": 1,
					"critical": 2,
					"fail": 3
				},
				{
					"name": "total",
					"warning": 2,
					"critical": 3,
					"fail": 4
				}
			]
		}
	`)

	expectedData := &udnssdk.ProbeDetailsDTO{
		Detail: udnssdk.PingProbeDetailsDTO{
			Packets:    15,
			PacketSize: 56,
			Limits: map[string]udnssdk.ProbeDetailsLimitDTO{
				"lossPercent": udnssdk.ProbeDetailsLimitDTO{
					Warning: 1,

					Critical: 2,
					Fail:     3,
				},
				"total": udnssdk.ProbeDetailsLimitDTO{
					Warning:  2,
					Critical: 3,
					Fail:     4,
				},
			},
		},
	}

	json.Unmarshal(pingProbe, &pingProbeDTO[0])
	resourceRecordObj := setResourceRecordPingProbe()
	resourceData := resourceRecordObj.TestResourceData()
	resourceData.Set("ping_probe", pingProbeDTO)
	res := makePingProbeDetails(resourceData.Get("ping_probe").([]interface{})[0])
	assert.Equal(t, expectedData, res, true)
}

func TestPopulateResourceDataFromPingProbe(t *testing.T) {

	resourceRecordObj := setResourceRecordPingProbe()
	resourceData := resourceRecordObj.TestResourceData()
	expectedResourceRecordObj := setResourceRecordPingProbe()
	expectedResourceData := expectedResourceRecordObj.TestResourceData()
	expectedData := []byte(`{"id":"0608485259D5AC79","type":"PING","interval":"ONE_MINUTE","agents":["DALLAS","AMSTERDAM"],"threshold":2,"details":{"packets":15,"packetSize":56,"limit":{"lossPercent":{"warning":1,"critical":2,"fail":3},"total":{"warning":2,"critical":3,"fail":4}}}}`)
	expectedResource := udnssdk.ProbeInfoDTO{}
	//	expectedDataDetail := &udnssdk.ProbeDetailsDTO{
	//	Detail: udnssdk.PingProbeDetailsDTO{
	//                        Packets:15,
	//                        PacketSize:56,
	//                        Limits: map[string]udnssdk.ProbeDetailsLimitDTO{
	//                                        "lossPercent": udnssdk.ProbeDetailsLimitDTO{
	//                                                Warning: 1,
	//                                                Critical: 2,
	//                                                Fail: 3,
	//                                        },
	//                                        "total": udnssdk.ProbeDetailsLimitDTO{
	//                                                Warning: 2,
	//                                                Critical: 3,
	//                                                Fail: 4,
	//                                        },
	//                        },
	//                },
	//        }
	//
	//
	err := json.Unmarshal(expectedData, &expectedResource)
	//	expectedResource.Details =  expectedDataDetail
	//
	//	if err != nil {
	//			log.Println(err)
	//	}
	//
	pingProbeDTO := make([]map[string]interface{}, 1)
	pingProbe := []byte(`
		{
		"packets": 15,
		"packet_size": 56,
		"limit": [
					{
						"name": "lossPercent",
						"warning": 1,
						"critical": 2,
						"fail": 3
					},
					{
						"name": "total",
						"warning": 2,
						"critical": 3,
						"fail": 4
					}
				]
		}
	`)

	json.Unmarshal(pingProbe, &pingProbeDTO[0])

	expectedResourceData.Set("name", "test.provider.ultradns.net")
	expectedResourceData.Set("zone", "test.provider.ultradns.net")
	expectedResourceData.SetId("0608485259D5AC79")
	expectedResourceData.Set("agents", []string{"DALLAS", "AMSTERDAM"})
	expectedResourceData.Set("interval", "ONE_MINUTE")
	expectedResourceData.Set("threshold", 2)
	expectedResourceData.Set("pool_record", "")
	expectedResourceData.Set("ping_probe", pingProbeDTO)

	resourceData.Set("name", "test.provider.ultradns.net")
	resourceData.Set("zone", "test.provider.ultradns.net")
	resourceData.SetId("0608485259D5AC79")
	resourceData.Set("ping_probe", pingProbeDTO)

	populateResourceDataFromPingProbe(expectedResource, resourceData)
	log.Infof("%+v", err)
	compareResourcesDataProbePing(t, resourceData, expectedResourceData)
}

func TestResourceUltradnsProbePingImport(t *testing.T) {
	resourceRecordObj := setResourceRecordPingProbe()
	d := resourceRecordObj.TestResourceData()
	d.SetId("test:test.provider.ultradns.net:0608485259D5AC79")
	var interfaceEmpty *udnssdk.Client
	newRecordData, _ := resourceUltradnsProbePingImport(d, interfaceEmpty)
	assert.Equal(t, newRecordData[0].Get("name"), "test", true)
	assert.Equal(t, newRecordData[0].Get("zone"), "test.provider.ultradns.net", true)
}

func TestResourceUltradnsProbePingImportFailCase(t *testing.T) {
	resourceRecordObj := setResourceRecordPingProbe()
	d := resourceRecordObj.TestResourceData()
	d.SetId("test.test.provider.ultradns.net.0608485259D5AC79")
	var interfaceEmpty *udnssdk.Client
	_, err := resourceUltradnsProbePingImport(d, interfaceEmpty)
	log.Errorf("ERROR: %+v", err)
	assert.NotNil(t, err, true)
}

func TestAccUltradnsProbePing(t *testing.T) {
	var record udnssdk.RRSet
	domain, _ := os.LookupEnv("ULTRADNS_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccTcpoolCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testCfgProbePingRecord, domain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_tcpool.test-probe-ping-record", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "name", "test-probe-ping-record"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "pool_record", "10.3.0.1"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "agents.0", "DALLAS"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "agents.1", "AMSTERDAM"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "interval", "ONE_MINUTE"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "threshold", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.packets", "15"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.packet_size", "56"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.#", "2"),

					// hashLimits(): lossPercent -> 3375621462
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3375621462.name", "lossPercent"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3375621462.warning", "1"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3375621462.critical", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3375621462.fail", "3"),

					// hashLimits(): total -> 3257917790
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3257917790.name", "total"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3257917790.warning", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3257917790.critical", "3"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3257917790.fail", "4"),
				),
			},
			{
				Config: fmt.Sprintf(testCfgProbePingPool, domain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_tcpool.test-probe-ping-pool", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "name", "test-probe-ping-pool"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "agents.0", "DALLAS"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "agents.1", "AMSTERDAM"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "interval", "ONE_MINUTE"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "threshold", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.packets", "15"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.packet_size", "56"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.#", "2"),

					// hashLimits(): lossPercent -> 3375621462
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3375621462.name", "lossPercent"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3375621462.warning", "1"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3375621462.critical", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3375621462.fail", "3"),

					// hashLimits(): total -> 3257917790
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3257917790.name", "total"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3257917790.warning", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3257917790.critical", "3"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.it", "ping_probe.0.limit.3257917790.fail", "4"),
				),
			},

			{
				ResourceName:      "ultradns_probe_ping.it",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testCfgProbePingRecord = `
resource "ultradns_tcpool" "test-probe-ping-record" {
  zone  = "%s"
  name  = "test-probe-ping-record"

  ttl   = 30
  description = "traffic controller pool with probes"

  run_probes    = true
  act_on_probes = true
  max_to_lb     = 2

  rdata {
    host = "10.3.0.1"

    state          = "NORMAL"
    run_probes     = true
    priority       = 1
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  rdata {
    host = "10.3.0.2"

    state          = "NORMAL"
    run_probes     = true
    priority       = 2
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  backup_record_rdata = "10.3.0.3"
}

resource "ultradns_probe_ping" "it" {
  zone  = "%s"
  name  = "test-probe-ping-record"

  pool_record = "10.3.0.1"

  agents = ["DALLAS", "AMSTERDAM"]

  interval  = "ONE_MINUTE"
  threshold = 2

  ping_probe {
    packets    = 15
    packet_size = 56

    limit {
      name     = "lossPercent"
      warning  = 1
      critical = 2
      fail     = 3
    }

    limit {
      name     = "total"
      warning  = 2
      critical = 3
      fail     = 4
    }
  }

  depends_on = ["ultradns_tcpool.test-probe-ping-record"]
}
`

const testCfgProbePingPool = `
resource "ultradns_tcpool" "test-probe-ping-pool" {
  zone  = "%s"
  name  = "test-probe-ping-pool"

  ttl   = 30
  description = "traffic controller pool with probes"

  run_probes    = true
  act_on_probes = true
  max_to_lb     = 2

  rdata {
    host = "10.3.0.1"

    state          = "NORMAL"
    run_probes     = true
    priority       = 1
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  rdata {
    host = "10.3.0.2"

    state          = "NORMAL"
    run_probes     = true
    priority       = 2
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  backup_record_rdata = "10.3.0.3"
}

resource "ultradns_probe_ping" "it" {
  zone  = "%s"
  name  = "test-probe-ping-pool"

  agents = ["DALLAS", "AMSTERDAM"]

  interval  = "ONE_MINUTE"
  threshold = 2

  ping_probe {
    packets    = 15
    packet_size = 56

    limit {
      name     = "lossPercent"
      warning  = 1
      critical = 2
      fail     = 3
    }

    limit {
      name     = "total"
      warning  = 2
      critical = 3
      fail     = 4
    }
  }

  depends_on = ["ultradns_tcpool.test-probe-ping-pool"]
}
`
