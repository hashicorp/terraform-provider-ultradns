package ultradns

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	udnssdk "github.com/ultradns/ultradns-sdk-go"
)

func setResourceRecordHTTPProbe() (resourceRecord *schema.Resource) {
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
				Type:     schema.TypeSet,
				Set:      schema.HashString,
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
			"http_probe": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     schemaHTTPProbes(),
			},
			// Computed
			"http_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func schemaHTTPProbes() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"transaction": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:     schema.TypeString,
							Required: true,
						},
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"transmitted_data": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"follow_redirects": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"limit": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      hashLimits,
							Elem:     resourceProbeLimits(),
						},
					},
				},
			},
			"total_limits": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"warning": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"critical": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"fail": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func compareResourcesDataProbeHTTP(t *testing.T, actual *schema.ResourceData, expected *schema.ResourceData) {
	assert.Equal(t, expected.Get("name"), actual.Get("name"), true)
	assert.Equal(t, expected.Get("zone"), actual.Get("zone"), true)
	assert.Equal(t, expected.Get("threshold"), actual.Get("threshold"), true)
	assert.Equal(t, expected.Get("interval"), actual.Get("interval"), true)
	assert.Equal(t, expected.Get("pool_record"), actual.Get("pool_record"), true)
	assert.Equal(t, expected.Get("id"), actual.Get("id"), true)
	assert.Equal(t, expected.Get("agents.2144410488"), actual.Get("agents.2144410488"), true)
	assert.Equal(t, expected.Get("agents.4091180299"), actual.Get("agents.4091180299"), true)
	assert.Equal(t, expected.Get("http_probe.0.total_limits"), actual.Get("http_probe.0.total_limits"), true)
	assert.Equal(t, expected.Get("http_probe.0.transaction.0.follow_redirects"), actual.Get("http_probe.0.transaction.0.follow_redirects"), true)
	assert.Equal(t, expected.Get("http_probe.0.transaction.0.limit.1959786783"), actual.Get("http_probe.0.transaction.0.limit.1959786783"), true)

}

func TestMakeHTTPProbeResource(t *testing.T) {
	resourceRecordObj := setResourceRecordHTTPProbe()
	resourceData := resourceRecordObj.TestResourceData()
	expectedData := []byte(`{"id":"0608485359E134B2","type": "HTTP","poolRecord":"10.2.1.1","interval":"ONE_MINUTE","agents":["AMSTERDAM","DALLAS"],"threshold":2,"details":{"transactions":[{"method":"POST","url":"http://www.google.com/","transmittedData":"{}","limits":{"connect":{"warning":10,"critical":11,"fail":12}},"followRedirects":true}],"totalLimits":{"warning":13,"critical":14,"fail":15}}}`)
	expectedResource := probeResource{}

	err := json.Unmarshal(expectedData, &expectedResource)

	if err != nil {
		log.Println(err)
	}
	expectedResource.Name = "test.provider.ultradns.net"
	expectedResource.Zone = "test.provider.ultradns.net"
	HttpDetail := &udnssdk.ProbeDetailsDTO{
		Detail: udnssdk.HTTPProbeDetailsDTO{
			Transactions: []udnssdk.Transaction{{
				Method:          "POST",
				URL:             "http://www.google.com/",
				TransmittedData: "{}",
				FollowRedirects: true,
				Limits: map[string]udnssdk.ProbeDetailsLimitDTO{
					"connect": udnssdk.ProbeDetailsLimitDTO{
						Warning:  10,
						Critical: 11,
						Fail:     12,
					},
				},
			}},
			TotalLimits: &udnssdk.ProbeDetailsLimitDTO{
				Warning:  13,
				Critical: 14,
				Fail:     15,
			},
		},
	}

	expectedResource.Details = HttpDetail

	HttpProbeDTO := make([]map[string]interface{}, 1)
	HttpProbe := []byte(`
	{
		"transaction": [{
				"method": "POST",
				"url": "http://www.google.com/",
				"transmitted_data": "{}",
				"limit": [{
							"name" : "connect",
							"warning": 10,
							"critical": 11,
							"fail": 12
						}],
				"follow_redirects": true
		}],

		"total_limits": [{
			"warning": 13,
			"critical": 14,
			"fail": 15
		}]
	}`)

	err = json.Unmarshal(HttpProbe, &HttpProbeDTO[0])

	if err != nil {
		log.Println(err)
	}

	resourceData.Set("name", "test.provider.ultradns.net")
	resourceData.Set("zone", "test.provider.ultradns.net")
	resourceData.SetId("0608485359E134B2")
	resourceData.Set("agents", []string{"DALLAS", "AMSTERDAM"})
	resourceData.Set("interval", "ONE_MINUTE")
	resourceData.Set("pool_record", "10.2.1.1")
	resourceData.Set("threshold", 2)
	err = resourceData.Set("http_probe", HttpProbeDTO)
	log.Infof("Err : %+v", err)
	res, _ := makeHTTPProbeResource(resourceData)
	assert.Equal(t, expectedResource, res, true)

}

func TestMakeHTTPProbeDetails(t *testing.T) {
	HttpProbeDTO := make([]map[string]interface{}, 1)
	HttpProbe := []byte(`
	{
		"transaction": [{
				"method": "POST",
				"url": "http://www.google.com/",
				"transmitted_data": "{}",
				"limit": [{
							"name" : "connect",
							"warning": 10,
							"critical": 11,
							"fail": 12
						}],
				"follow_redirects": true
		}],

		"total_limits": [{
				"warning": 13,
				"critical": 14,
				"fail": 15
		}]
	}`)

	HttpDetail := &udnssdk.ProbeDetailsDTO{
		Detail: udnssdk.HTTPProbeDetailsDTO{
			Transactions: []udnssdk.Transaction{{
				Method:          "POST",
				URL:             "http://www.google.com/",
				TransmittedData: "{}",
				FollowRedirects: true,
				Limits: map[string]udnssdk.ProbeDetailsLimitDTO{
					"connect": udnssdk.ProbeDetailsLimitDTO{
						Warning:  10,
						Critical: 11,
						Fail:     12,
					},
				},
			}},
			TotalLimits: &udnssdk.ProbeDetailsLimitDTO{
				Warning:  13,
				Critical: 14,
				Fail:     15,
			},
		},
	}
	json.Unmarshal(HttpProbe, &HttpProbeDTO[0])
	resourceRecordObj := setResourceRecordHTTPProbe()
	resourceData := resourceRecordObj.TestResourceData()
	resourceData.Set("http_probe", HttpProbeDTO)
	res := makeHTTPProbeDetails(resourceData.Get("http_probe").([]interface{})[0])
	assert.Equal(t, HttpDetail.Detail, res.Detail, true)
}

func TestPopulateResourceDataFromHTTPProbe(t *testing.T) {

	resourceRecordObj := setResourceRecordHTTPProbe()
	resourceData := resourceRecordObj.TestResourceData()
	expectedResourceRecordObj := setResourceRecordHTTPProbe()
	expectedResourceData := expectedResourceRecordObj.TestResourceData()
	expectedData := []byte(`{"id":"0608485359E134B2","type": "HTTP","poolRecord":"10.2.1.1","interval":"ONE_MINUTE","agents":["AMSTERDAM","DALLAS"],"threshold":2,"details":{"transactions":[{"method":"POST","url":"http://www.google.com/","transmittedData":"{}","limits":{"connect":{"warning":10,"critical":11,"fail":12}},"followRedirects":true}],"totalLimits":{"warning":13,"critical":14,"fail":15}}}`)
	expectedResource := udnssdk.ProbeInfoDTO{}
	err := json.Unmarshal(expectedData, &expectedResource)

	if err != nil {
		log.Println(err)
	}

	HttpProbeDTO := make([]map[string]interface{}, 1)
	HttpProbe := []byte(`
	{
		"transaction": [{
				"method": "POST",
				"url": "http://www.google.com/",
				"transmitted_data": "{}",
				"limit": [{
							"name" : "connect",
							"warning": 10,
							"critical": 11,
							"fail": 12
						}],
				"follow_redirects": true
		}],

		"total_limits": [{
				"warning": 13,
				"critical": 14,
				"fail": 15
		}]
	}`)

	json.Unmarshal(HttpProbe, &HttpProbeDTO[0])

	expectedResourceData.Set("name", "test.provider.ultradns.net")
	expectedResourceData.Set("zone", "test.provider.ultradns.net")
	expectedResourceData.SetId("test.provider.ultradns.net:test.provider.ultradns.net:0608485359E134B2")
	expectedResourceData.Set("agents", []string{"DALLAS", "AMSTERDAM"})
	expectedResourceData.Set("interval", "ONE_MINUTE")
	expectedResourceData.Set("threshold", 2)
	expectedResourceData.Set("pool_record", "10.2.1.1")
	expectedResourceData.Set("http_probe", HttpProbeDTO)

	resourceData.Set("name", "test.provider.ultradns.net")
	resourceData.Set("zone", "test.provider.ultradns.net")
	resourceData.SetId("test.provider.ultradns.net:test.provider.ultradns.net:0608485359E134B2")
	resourceData.Set("http_probe", HttpProbeDTO)

	populateResourceDataFromHTTPProbe(expectedResource, resourceData)
	compareResourcesDataProbeHTTP(t, resourceData, expectedResourceData)
}

func TestRresourceUltradnsProbePingImport(t *testing.T) {
	resourceRecordObj := setResourceRecordHTTPProbe()
	d := resourceRecordObj.TestResourceData()
	d.SetId("test:test.provider.ultradns.net:0608485259D5AC79")
	newRecordData, _ := resourceUltradnsProbeHTTPImport(d, udnssdk.Client{})
	assert.Equal(t, newRecordData[0].Get("name"), "test", true)
	assert.Equal(t, newRecordData[0].Get("zone"), "test.provider.ultradns.net", true)
}

func TestResourceUltradnsProbeHTTPImportFailCase(t *testing.T) {
	resourceRecordObj := setResourceRecordPingProbe()
	d := resourceRecordObj.TestResourceData()
	d.SetId("test.test.provider.ultradns.net.0608485259D5AC79")
	_, err := resourceUltradnsProbeHTTPImport(d, udnssdk.Client{})
	log.Errorf("ERROR: %+v", err)
	assert.NotNil(t, err, true)

	// Case1 when only one delimiter are there
	d.SetId("test:test.provider.ultradns.net.0608485259D5AC79")
	_, err = resourceUltradnsProbeHTTPImport(d, udnssdk.Client{})
	log.Errorf("ERROR: %+v", err)
	assert.NotNil(t, err, true)

}

func TestAccUltradnsProbeHTTP(t *testing.T) {
	var record udnssdk.RRSet
	domain, _ := os.LookupEnv("ULTRADNS_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccTcpoolCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testCfgProbeHTTPMinimal, domain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_tcpool.test-probe-http-minimal", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "name", "test-probe-http-minimal"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "pool_record", "10.2.0.1"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "agents.4091180299", "DALLAS"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "agents.2144410488", "AMSTERDAM"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "interval", "ONE_MINUTE"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "threshold", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.method", "GET"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.url", "http://google.com/"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.#", "2"),

					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1959786783.name", "connect"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1959786783.warning", "20"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1959786783.critical", "20"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1959786783.fail", "20"),

					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1349952704.name", "run"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1349952704.warning", "60"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1349952704.critical", "60"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1349952704.fail", "60"),
				),
			},
			{
				Config: fmt.Sprintf(testCfgProbeHTTPMaximal, domain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_tcpool.test-probe-http-maximal", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "name", "test-probe-http-maximal"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "pool_record", "10.2.1.1"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "agents.4091180299", "DALLAS"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "agents.2144410488", "AMSTERDAM"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "interval", "ONE_MINUTE"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "threshold", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.method", "POST"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.url", "http://google.com/"),

					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.#", "4"),

					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1349952704.name", "run"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1349952704.warning", "1"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1349952704.critical", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1349952704.fail", "3"),

					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.2720402232.name", "avgConnect"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.2720402232.warning", "4"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.2720402232.critical", "5"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.2720402232.fail", "6"),

					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.896769211.name", "avgRun"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.896769211.warning", "7"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.896769211.critical", "8"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.896769211.fail", "9"),

					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1959786783.name", "connect"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1959786783.warning", "10"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1959786783.critical", "11"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.transaction.0.limit.1959786783.fail", "12"),

					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.total_limits.0.warning", "13"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.total_limits.0.critical", "14"),
					resource.TestCheckResourceAttr("ultradns_probe_http.it", "http_probe.0.total_limits.0.fail", "15"),
				),
			},

			{
				ResourceName:      "ultradns_probe_http.it",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testCfgProbeHTTPMinimal = `
resource "ultradns_tcpool" "test-probe-http-minimal" {
  zone = "%s"
  name = "test-probe-http-minimal"
  ttl         = 30
  description = "traffic controller pool with probes"
  run_probes    = true
  act_on_probes = true
  max_to_lb     = 2
  rdata {
	host = "10.2.0.1"
	state          = "NORMAL"
	run_probes     = true
	priority       = 1
	failover_delay = 0
	threshold      = 1
	weight         = 2
  }
  rdata {
	host = "10.2.0.2"
	state          = "NORMAL"
	run_probes     = true
	priority       = 2
	failover_delay = 0
	threshold      = 1
	weight         = 2
  }
  backup_record_rdata = "10.2.0.3"
}
resource "ultradns_probe_http" "it" {
  zone = "%s"
  name = "test-probe-http-minimal"
  pool_record = "10.2.0.1"
  agents = ["DALLAS", "AMSTERDAM"]
  interval  = "ONE_MINUTE"
  threshold = 2
  http_probe {
	transaction {
	  method = "GET"
	  url    = "http://google.com/"
	  limit {
	name     = "run"
	warning  = 60
	critical = 60
	fail     = 60
	  }
	  limit {
	name     = "connect"
	warning  = 20
	critical = 20
	fail     = 20
	  }
	}
  }
  depends_on = ["ultradns_tcpool.test-probe-http-minimal"]
}
`

const testCfgProbeHTTPMaximal = `
resource "ultradns_tcpool" "test-probe-http-maximal" {
  zone  = "%s"
  name  = "test-probe-http-maximal"
  ttl   = 30
  description = "traffic controller pool with probes"
  run_probes    = true
  act_on_probes = true
  max_to_lb     = 2
  rdata {
	host = "10.2.1.1"
	state          = "NORMAL"
	run_probes     = true
	priority       = 1
	failover_delay = 0
	threshold      = 1
	weight         = 2
  }
  rdata {
	host = "10.2.1.2"
	state          = "NORMAL"
	run_probes     = true
	priority       = 2
	failover_delay = 0
	threshold      = 1
	weight         = 2
  }
  backup_record_rdata = "10.2.1.3"
}
resource "ultradns_probe_http" "it" {
  zone = "%s"
  name = "test-probe-http-maximal"
  pool_record = "10.2.1.1"
  agents = ["DALLAS", "AMSTERDAM"]
  interval  = "ONE_MINUTE"
  threshold = 2
  http_probe {
	transaction {
	  method           = "POST"
	  url              = "http://google.com/"
	  transmitted_data = "{}"
	  follow_redirects = true
	  limit {
	name = "run"
	warning  = 1
	critical = 2
	fail     = 3
	  }
	  limit {
	name = "avgConnect"
	warning  = 4
	critical = 5
	fail     = 6
	  }
	  limit {
	name = "avgRun"
	warning  = 7
	critical = 8
	fail     = 9
	  }
	  limit {
	name = "connect"
	warning  = 10
	critical = 11
	fail     = 12
	  }
	}
	total_limits {
	  warning  = 13
	  critical = 14
	  fail     = 15
	}
  }
  depends_on = ["ultradns_tcpool.test-probe-http-maximal"]
}
`
