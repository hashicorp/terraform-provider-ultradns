package ultradns

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	_ "reflect"
	"testing"

	udnssdk "github.com/aliasgharmhowwala/ultradns-sdk-go"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockUltraDNSRecordDirPool struct {
	client *udnssdk.Client
}

func (m *mockUltraDNSRecordDirPool) Create(k udnssdk.RRSetKey, rrset udnssdk.RRSet) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecordDirPool) Select(k udnssdk.RRSetKey) ([]udnssdk.RRSet, error) {

	rrsets := []udnssdk.RRSet{{
		OwnerName: "test.provider.ultradns.net",
		RRType:    "A",
		RData:     []string{"10.1.1.2"},
		TTL:       0,
		Profile: udnssdk.RawProfile{
			"@context":        "http://schemas.ultradns.com/DirPool.jsonschema",
			"conflictResolve": "GEO",
			"description":     "testing",
			"noResponse": map[string]interface{}{
				"allNonConfigured": bool(false),
				"geoInfo": map[string]interface{}{
					"isAccountLevel": bool(false),
					"codes":          []interface{}{"EUR"},
					"name":           "america",
				},
				"ipInfo": map[string]interface{}{
					"ips": []interface{}{map[string]interface{}{
						"address": "200.20.0.1",
					}},
					"name": "nrIpInfo",
				},
			},
			"rdataInfo": []interface{}{map[string]interface{}{
				"ttl": 300,
				"geoInfo": map[string]interface{}{
					"codes": []interface{}{"US"},
					"name":  "North America",
				},
				"ipInfo": map[string]interface{}{
					"ips": []interface{}{map[string]interface{}{
						"address": "200.212.1.1",
					}},
					"name": "rdataIpInfo",
				}},
			}},
	}}
	return rrsets, nil

}

func (m *mockUltraDNSRecordDirPool) SelectWithOffset(k udnssdk.RRSetKey, offset int) ([]udnssdk.RRSet, udnssdk.ResultInfo, *http.Response, error) {
	return nil, udnssdk.ResultInfo{}, nil, nil
}

func (m *mockUltraDNSRecordDirPool) Update(udnssdk.RRSetKey, udnssdk.RRSet) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecordDirPool) Delete(k udnssdk.RRSetKey) (*http.Response, error) {
	return nil, nil
}

func (m *mockUltraDNSRecordDirPool) SelectWithOffsetWithLimit(k udnssdk.RRSetKey, offset int, limit int) (rrsets []udnssdk.RRSet, ResultInfo udnssdk.ResultInfo, resp *http.Response, err error) {
	return []udnssdk.RRSet{}, udnssdk.ResultInfo{}, nil, nil
}

func setResourceRecordDirPool() (resourceRecord *schema.Resource) {
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
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if len(value) > 255 {
						errors = append(errors, fmt.Errorf(
							"'description' too long, must be less than 255 characters"))
					}
					return
				},
			},
			"rdata": {
				// UltraDNS API does not respect rdata ordering
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
						"ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  3600,
						},

						"all_non_configured": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"geo_info": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"is_account_level": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"codes": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
									},
								},
							},
						},
						"ip_info": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"is_account_level": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"ips": {
										Type:     schema.TypeSet,
										Optional: true,
										Set:      hashIPInfoIPs,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"start": {
													Type:     schema.TypeString,
													Optional: true,
													// ConflictsWith: []string{"cidr", "address"},
												},
												"end": {
													Type:     schema.TypeString,
													Optional: true,
													// ConflictsWith: []string{"cidr", "address"},
												},
												"cidr": {
													Type:     schema.TypeString,
													Optional: true,
													// ConflictsWith: []string{"start", "end", "address"},
												},
												"address": {
													Type:     schema.TypeString,
													Optional: true,
													// ConflictsWith: []string{"start", "end", "cidr"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"conflict_resolve": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "GEO",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value != "GEO" && value != "IP" {
						errors = append(errors, fmt.Errorf(
							"only 'GEO', and 'IP' are supported values for 'conflict_resolve'"))
					}
					return
				},
			},
			"no_response": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"all_non_configured": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"geo_info": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"is_account_level": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"codes": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
									},
								},
							},
						},
						"ip_info": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"is_account_level": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"ips": {
										Type:     schema.TypeSet,
										Optional: true,
										Set:      hashIPInfoIPs,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"start": {
													Type:     schema.TypeString,
													Optional: true,
													// ConflictsWith: []string{"cidr", "address"},
												},
												"end": {
													Type:     schema.TypeString,
													Optional: true,
													// ConflictsWith: []string{"cidr", "address"},
												},
												"cidr": {
													Type:     schema.TypeString,
													Optional: true,
													// ConflictsWith: []string{"start", "end", "address"},
												},
												"address": {
													Type:     schema.TypeString,
													Optional: true,
													// ConflictsWith: []string{"start", "end", "cidr"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			// Computed
			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func compareDirPoolResources(t *testing.T, expected *schema.ResourceData, actual *schema.ResourceData) {
	assert.Equal(t, expected.Get("name"), actual.Get("name"), true)
	assert.Equal(t, expected.Get("zone"), actual.Get("zone"), true)
	assert.Equal(t, expected.Get("hostname"), actual.Get("hostname"), true)
	assert.Equal(t, expected.Get("rdata.2203440046.all_non_configured"), actual.Get("rdata.2203440046.all_non_configured"), true)
	assert.Equal(t, expected.Get("rdata.2203440046.host"), actual.Get("rdata.2203440046.host"), true)
	assert.Equal(t, expected.Get("rdata.2203440046.geo_info.0.codes.1954003872"), actual.Get("rdata.2203440046.geo_info.0.codes.1954003872"), true)
	assert.Equal(t, expected.Get("rdata.2203440046.ip_info.0.ips.393713059.address"), actual.Get("rdata.2203440046.ip_info.0.ips.393713059.address"), true)
	assert.Equal(t, expected.Get("type"), actual.Get("type"), true)
	assert.Equal(t, expected.Get("conflict_resolve"), actual.Get("conflict_resolve"), true)
	assert.Equal(t, expected.Get("description"), actual.Get("description"), true)
	assert.Equal(t, expected.Get("no_response.0.all_non_configured"), actual.Get("no_response.0.all_non_configured"), true)
	assert.Equal(t, expected.Get("no_response.0.geo_info.0.codes.3417903088"), actual.Get("no_response.0.geo_info.0.codes.3417903088"), true)
	assert.Equal(t, expected.Get("no_response.0.geo_info.0.is_account_level"), actual.Get("no_response.0.geo_info.0.is_account_level"), true)
	assert.Equal(t, expected.Get("no_response.0.ip_info.0.name"), actual.Get("no_response.0.ip_info.0.name"), true)
	assert.Equal(t, expected.Get("no_response.0.ip_info.0.ips.3187989519.address"), actual.Get("no_response.0.ip_info.0.ips.3187989519.address"))
}

func TestMakeDirpoolRRSetResource(t *testing.T) {
	resourceRecordObj := setResourceRecordDirPool()
	resourceData := resourceRecordObj.TestResourceData()

	expectedResource := rRSetResource{}
	var context udnssdk.ProfileSchema
	context = "http://schemas.ultradns.com/DirPool.jsonschema"
	expectedResource = rRSetResource{
		OwnerName: "test.provider.ultradns.net",
		RRType:    "A", RData: []string{"10.1.1.2"},
		TTL: 0,
		Profile: udnssdk.RawProfile{
			"@context":        context,
			"conflictResolve": "GEO",
			"description":     "testing",
			"noResponse": map[string]interface{}{
				"geoInfo": map[string]interface{}{
					"codes": []string{"EUR"},
					"name":  "america",
				},
				"ipInfo": map[string]interface{}{
					"ips": []interface{}{map[string]interface{}{
						"address": "200.20.0.1",
					}},
					"name": "nrIpInfo",
				},
			},
			"rdataInfo": []interface{}{map[string]interface{}{
				"geoInfo": map[string]interface{}{
					"codes": []string{"US"},
					"name":  "North America",
				},
				"ipInfo": map[string]interface{}{
					"ips": []interface{}{map[string]interface{}{
						"address": "200.212.1.1",
					}},
					"name": "rdataIpInfo",
				}},
			}},
		Zone: "test.provider.ultradns.net",
	}

	rrsetDTO := make([]map[string]interface{}, 1)
	noResponseDTO := make([]map[string]interface{}, 1)
	data := []byte(`
                [{
                        "host": "10.1.1.2",
                        "ttl": 300,
                        "geo_info": [{
                                "name": "North America",
                                "codes": [
                                        "US"
                                ]
                        }],
                        "ip_info": [{

                                "name": "rdataIpInfo",
                                "ips": [{
                                        "address": "200.212.1.1"
                                }]
                        }]
                }]
        `)

	noResponseData := []byte(`
                [{
                        "geo_info": [{
                                "name": "america",
                                "codes": [
                                        "EUR"
                                ]
                        }],
                        "ip_info": [{

                                "name": "nrIpInfo",
                                "ips": [{
                                        "address": "200.20.0.1"
                                }]
                        }]
                   }]
        `)

	err := json.Unmarshal(noResponseData, &noResponseDTO)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	resourceData.Set("name", "test.provider.ultradns.net")
	resourceData.Set("rdata", rrsetDTO)
	err = resourceData.Set("no_response", noResponseDTO)
	resourceData.Set("zone", "test.provider.ultradns.net")
	resourceData.Set("type", "A")
	resourceData.Set("conflict_resolve", "GEO")
	resourceData.Set("description", "testing")

	rrset, _ := makeDirpoolRRSetResource(resourceData)
	log.Infof("Expected %+v", expectedResource)
	log.Infof("actual  %+v", rrset)
	assert.Equal(t, expectedResource, rrset, true)
	log.Infof("resourceData RData: %+v, err %+v", resourceData.Get("rdata"), err)
}

func TestPopulateResourceFromDirpool(t *testing.T) {
	resourceRecordObj := setResourceRecordDirPool()
	resourceData := resourceRecordObj.TestResourceData()
	expectedResourceRecordObj := setResourceRecordDirPool()
	expectedResourceData := expectedResourceRecordObj.TestResourceData()
	var booleanValue bool
	booleanValue = false
	rrsetResource := &udnssdk.RRSet{
		OwnerName: "test.provider.ultradns.net",
		RRType:    "A",
		RData:     []string{"10.1.1.2"},
		TTL:       0,
		Profile: udnssdk.RawProfile{
			"@context":        "http://schemas.ultradns.com/DirPool.jsonschema",
			"conflictResolve": "GEO",
			"description":     "testing",
			"noResponse": map[string]interface{}{
				"allNonConfigured": booleanValue,
				"geoInfo": map[string]interface{}{
					"isAccountLevel": booleanValue,
					"codes":          []interface{}{"EUR"},
					"name":           "america",
				},
				"ipInfo": map[string]interface{}{
					"ips": []interface{}{map[string]interface{}{
						"address": "200.20.0.1",
					}},
					"name": "nrIpInfo",
				},
			},
			"rdataInfo": []interface{}{map[string]interface{}{
				"ttl": 300,
				"geoInfo": map[string]interface{}{
					"codes": []interface{}{"US"},
					"name":  "North America",
				},
				"ipInfo": map[string]interface{}{
					"ips": []interface{}{map[string]interface{}{
						"address": "200.212.1.1",
					}},
					"name": "rdataIpInfo",
				}},
			}},
	}

	rrsetDTO := make([]map[string]interface{}, 1)
	noResponseDTO := make([]map[string]interface{}, 1)
	data := []byte(`
                [{
                        "host": "10.1.1.2",
                        "ttl": 300,
                        "geo_info": [{
                                "name": "North America",
                                "codes": [
                                        "US"
                                ]
                        }],
                        "ip_info": [{

                                "name": "rdataIpInfo",
                                "ips": [{
                                        "address": "200.212.1.1"
                                }]
                        }]
                }]
        `)

	noResponseData := []byte(`
                [{
                        "geo_info": [{
                                "name": "america",
                                "codes": [
                                        "EUR"
                                ]
                        }],
                        "ip_info": [{

                                "name": "nrIpInfo",
                                "ips": [{
                                        "address": "200.20.0.1"
                                }]
                        }]
                   }]
        `)

	err := json.Unmarshal(noResponseData, &noResponseDTO)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	expectedResourceData.Set("name", "test.provider.ultradns.net")
	expectedResourceData.Set("rdata", rrsetDTO)
	err = expectedResourceData.Set("no_response", noResponseDTO)
	expectedResourceData.Set("zone", "test.provider.ultradns.net")
	expectedResourceData.Set("type", "A")
	expectedResourceData.Set("conflict_resolve", "GEO")
	expectedResourceData.Set("description", "testing")
	expectedResourceData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	resourceData.Set("name", "test.provider.ultradns.net")
	resourceData.Set("zone", "test.provider.ultradns.net")
	resourceData.Set("conflict_resolve", "GEO")

	populateResourceFromDirpool(resourceData, rrsetResource)
	compareDirPoolResources(t, expectedResourceData, resourceData)
	
	//case 1  when  ownername is empty
	rrsetResource.OwnerName = ""
	expectedResourceData.Set("hostname", "test.provider.ultradns.net")
	populateResourceFromDirpool(resourceData, rrsetResource)
	compareDirPoolResources(t, expectedResourceData, resourceData)

	//case 2 when ownername has prefix .
	rrsetResource.OwnerName = "abc."
	expectedResourceData.Set("hostname", "abc.")
	populateResourceFromDirpool(resourceData, rrsetResource)
	compareDirPoolResources(t, expectedResourceData, resourceData)
}

func TestResourceUltradnsDirPoolCreate(t *testing.T) {
	expectedResourceRecordObj := setResourceRecordDirPool()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecordDirPool()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecordDirPool{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	rrsetDTO := make([]map[string]interface{}, 1)
	noResponseDTO := make([]map[string]interface{}, 1)
	data := []byte(`
			[{
					"host": "10.1.1.2",
					"ttl": 300,
					"geo_info": [{
							"name": "North America",
							"codes": [
									"US"
							]
					}],
					"ip_info": [{

							"name": "rdataIpInfo",
							"ips": [{
									"address": "200.212.1.1"
							}]
					}]
			}]
	`)

	noResponseData := []byte(`
			[{
					"geo_info": [{
							"name": "america",
							"codes": [
									"EUR"
							]
					}],
					"ip_info": [{

							"name": "nrIpInfo",
							"ips": [{
									"address": "200.20.0.1"
							}]
					}]
			   }]
	`)

	err := json.Unmarshal(noResponseData, &noResponseDTO)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("rdata", rrsetDTO)
	err = expectedData.Set("no_response", noResponseDTO)
	expectedData.Set("zone", "test.provider.ultradns.net")
	expectedData.Set("type", "A")
	expectedData.Set("conflict_resolve", "GEO")
	expectedData.Set("description", "testing")
	expectedData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("rdata", rrsetDTO)
	err = actualData.Set("no_response", noResponseDTO)
	actualData.Set("zone", "test.provider.ultradns.net")
	actualData.Set("type", "A")
	actualData.Set("conflict_resolve", "GEO")
	actualData.Set("description", "testing")
	actualData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	resourceUltradnsDirpoolCreate(actualData, client)
	compareDirPoolResources(t, expectedData, actualData)

}

func TestResourceUltradnsDirPoolRead(t *testing.T) {
	expectedResourceRecordObj := setResourceRecordDirPool()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecordDirPool()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecordDirPool{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	rrsetDTO := make([]map[string]interface{}, 1)
	noResponseDTO := make([]map[string]interface{}, 1)
	data := []byte(`
			[{
					"host": "10.1.1.2",
					"ttl": 300,
					"geo_info": [{
							"name": "North America",
							"codes": [
									"US"
							]
					}],
					"ip_info": [{

							"name": "rdataIpInfo",
							"ips": [{
									"address": "200.212.1.1"
							}]
					}]
			}]
	`)

	noResponseData := []byte(`
			[{
					"geo_info": [{
							"name": "america",
							"codes": [
									"EUR"
							]
					}],
					"ip_info": [{

							"name": "nrIpInfo",
							"ips": [{
									"address": "200.20.0.1"
							}]
					}]
			   }]
	`)

	err := json.Unmarshal(noResponseData, &noResponseDTO)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("rdata", rrsetDTO)
	err = expectedData.Set("no_response", noResponseDTO)
	expectedData.Set("zone", "test.provider.ultradns.net")
	expectedData.Set("type", "A")
	expectedData.Set("conflict_resolve", "GEO")
	expectedData.Set("description", "testing")
	expectedData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("zone", "test.provider.ultradns.net")

	resourceUltradnsDirpoolRead(actualData, client)
	compareDirPoolResources(t, expectedData, actualData)

}

func TestResourceUltradnsDirPoolUpdate(t *testing.T) {
	expectedResourceRecordObj := setResourceRecordDirPool()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecordDirPool()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecordDirPool{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	rrsetDTO := make([]map[string]interface{}, 1)
	noResponseDTO := make([]map[string]interface{}, 1)
	data := []byte(`
			[{
					"host": "10.1.1.2",
					"ttl": 300,
					"geo_info": [{
							"name": "North America",
							"codes": [
									"US"
							]
					}],
					"ip_info": [{

							"name": "rdataIpInfo",
							"ips": [{
									"address": "200.212.1.1"
							}]
					}]
			}]
	`)

	noResponseData := []byte(`
			[{
					"geo_info": [{
							"name": "america",
							"codes": [
									"EUR"
							]
					}],
					"ip_info": [{

							"name": "nrIpInfo",
							"ips": [{
									"address": "200.20.0.1"
							}]
					}]
			   }]
	`)

	err := json.Unmarshal(noResponseData, &noResponseDTO)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("rdata", rrsetDTO)
	err = expectedData.Set("no_response", noResponseDTO)
	expectedData.Set("zone", "test.provider.ultradns.net")
	expectedData.Set("type", "A")
	expectedData.Set("conflict_resolve", "GEO")
	expectedData.Set("description", "testing")
	expectedData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("rdata", rrsetDTO)
	err = actualData.Set("no_response", noResponseDTO)
	actualData.Set("zone", "test.provider.ultradns.net")
	actualData.Set("type", "A")
	actualData.Set("conflict_resolve", "GEO")
	actualData.Set("description", "testing")
	actualData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	resourceUltradnsDirpoolUpdate(actualData, client)
	compareDirPoolResources(t, expectedData, actualData)

}

func TestResourceUltradnsDirPoolDelete(t *testing.T) {
	expectedResourceRecordObj := setResourceRecordDirPool()
	expectedData := expectedResourceRecordObj.TestResourceData()
	resourceRecordObject := setResourceRecordDirPool()
	actualData := resourceRecordObject.TestResourceData()
	mocked := mockUltraDNSRecordDirPool{}

	client := &udnssdk.Client{
		RRSets: &mocked,
	}

	rrsetDTO := make([]map[string]interface{}, 1)
	noResponseDTO := make([]map[string]interface{}, 1)
	data := []byte(`
			[{
					"host": "10.1.1.2",
					"ttl": 300,
					"geo_info": [{
							"name": "North America",
							"codes": [
									"US"
							]
					}],
					"ip_info": [{

							"name": "rdataIpInfo",
							"ips": [{
									"address": "200.212.1.1"
							}]
					}]
			}]
	`)

	noResponseData := []byte(`
			[{
					"geo_info": [{
							"name": "america",
							"codes": [
									"EUR"
							]
					}],
					"ip_info": [{

							"name": "nrIpInfo",
							"ips": [{
									"address": "200.20.0.1"
							}]
					}]
			   }]
	`)

	err := json.Unmarshal(noResponseData, &noResponseDTO)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	expectedData.Set("name", "test.provider.ultradns.net")
	expectedData.Set("rdata", rrsetDTO)
	err = expectedData.Set("no_response", noResponseDTO)
	expectedData.Set("zone", "test.provider.ultradns.net")
	expectedData.Set("type", "A")
	expectedData.Set("conflict_resolve", "GEO")
	expectedData.Set("description", "testing")
	expectedData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("rdata", rrsetDTO)
	err = actualData.Set("no_response", noResponseDTO)
	actualData.Set("zone", "test.provider.ultradns.net")
	actualData.Set("type", "A")
	actualData.Set("conflict_resolve", "GEO")
	actualData.Set("description", "testing")
	actualData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	resourceUltradnsDirpoolDelete(actualData, client)
	compareDirPoolResources(t, expectedData, actualData)

}

func TestResourceUltradnsDirPoolImport(t *testing.T) {
	mocked := mockUltraDNSRecordRDPool{}
	client := &udnssdk.Client{
		RRSets: &mocked,
	}
	resourceRecordObj := setResourceRecordDirPool()
	d := resourceRecordObj.TestResourceData()
	d.SetId("test:test.provider.ultradns.net")
	newRecordData, _ := resourceUltradnsDirpoolImport(d, client)
	assert.Equal(t, newRecordData[0].Get("name"), "test", true)
	assert.Equal(t, newRecordData[0].Get("zone"), "test.provider.ultradns.net", true)

}

//Testcase to check fail case
func TestResourceUltradnsDirPoolImportFailCase(t *testing.T) {
	mocked := mockUltraDNSRecordDirPool{}
	client := &udnssdk.Client{
		RRSets: &mocked,
	}
	resourceRecordObj := setResourceRecordDirPool()
	d := resourceRecordObj.TestResourceData()
	d.SetId("testabc.test.provider.ultradns.net")
	_, err := resourceUltradnsDirpoolImport(d, client)
	log.Errorf("Error: %+v", err)
	assert.NotNil(t, err, true)

}

func TestResourceUltradnsDirPoolFailCases(t *testing.T){

	resourceRecordObject := setResourceRecordDirPool()
	actualData := resourceRecordObject.TestResourceData()

	actualData.Set("name", "test.provider.ultradns.net")
	actualData.Set("zone", "test.provider.ultradns.net")
	actualData.Set("type", "A")
	actualData.Set("conflict_resolve", "abcdefghi")
	actualData.Set("description", "testing")
	actualData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")
	actualData.Set("description", "xWVQfy7AtNcCATHLSNppqs2SjImlnYOBi2UVp9X5XlEzoRCkmttmb2tD2JZI7AW4cySeS9aOvSFOj0oZM8m78cExZtnIO8dTeilKp6iObO1ipB2g4966c630QBxsHotCqEjrQ8Ky70vw3hd6mL16qe9nuHr8BDxJ4LYm5OyyiMT85NSuA0PykDl1hJhL5t6pCuPYqQQ8tXuLBqArJBZGuoPIPQHHLf33aSASRuVPkKZ8wqgJFLz4zgJ8mUEtIc9TmBRsddadsdadasdsdasdsDasdasdaDADADwadaDAWDASDADSDDADDWDAWDASDWADDADWADWADAWDW")

	mocked := mockUltraDNSRecordRDPool{}
	client := &udnssdk.Client{
		RRSets: &mocked,
	}
	err := resourceUltradnsDirpoolCreate(actualData, client)
	log.Infof("Error : %+v",err)
	assert.NotNil(t,err,true)

}


func TestMakeDirpoolRRSetResourceFailCases(t *testing.T) {

	resourceRecordObj := setResourceRecordDirPool()
	resourceData := resourceRecordObj.TestResourceData()

	rrsetDTO := make([]map[string]interface{}, 1)
	noResponseDTO := make([]map[string]interface{}, 1)

	data := []byte(`
		[{
				"host": "10.1.1.2",
				"ttl": 300,
				"geo_info": [{
						"name": "North America",
						"codes": [
								"US"
						]
				}],
				"ip_info": [{
						"name": "rdataIpInfo",
						"ips": [{
								"address": "200.212.1.1"
						}]
				}]
		}]
		`)

	//Case 1 when there is more than one no_response block
	noResponseData := []byte(`
		[{
		"geo_info": [{
			"name": "america",
			"codes": [
				"EUR"
			]
		}],
		"ip_info": [{

			"name": "nrIpInfo",
			"ips": [{
				"address": "200.20.0.1"
			}]
		}]
	},
	{
		"geo_info": [{
			"name": "america",
			"codes": [
				"EUR"
			]
		}],
		"ip_info": [{

			"name": "nrIpInfo",
			"ips": [{
				"address": "200.20.0.1"
			}]
		}]
	}]
		`)

	err := json.Unmarshal(noResponseData, &noResponseDTO)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	resourceData.Set("name", "test.provider.ultradns.net")
	resourceData.Set("rdata", rrsetDTO)
	resourceData.Set("no_response", noResponseDTO)
	resourceData.Set("zone", "test.provider.ultradns.net")
	resourceData.Set("type", "A")
	resourceData.Set("conflict_resolve", "GEO")
	resourceData.Set("description", "testing")

	_, err1 := makeDirpoolRRSetResource(resourceData)
	log.Errorf("Error %+v", err1)
	assert.NotNil(t, err1, true)

	// Case2 when there is error in no_response block
	noResponseData = []byte(`
		[{
				"geo_info": [{
					"name": "america",
					"codes": [
						"EUR"
					]
				},
				{
					"name": "america",
					"codes": [
						"EUR"
					]
				}],
				"ip_info": [{
						"name": "nrIpInfo",
						"ips": [{
								"address": "23"
					},
				{
						"name": "nrIpInfo",
						"ips": [{
								"address": "23"
					}]
				}]
		}]`)

	err = json.Unmarshal(noResponseData, &noResponseDTO)
	if err != nil {
		log.Println(err)
	}
	resourceData.Set("no_response", noResponseDTO)
	_, noResponseError := makeDirpoolRRSetResource(resourceData)
	log.Errorf("Error %+v", noResponseError)
	assert.NotNil(t, noResponseError, true)

}

func TestPopulateResourceFromDirpoolFailCase(t *testing.T) {

	resourceRecordObj := setResourceRecordDirPool()
	resourceData := resourceRecordObj.TestResourceData()
	expectedResourceRecordObj := setResourceRecordDirPool()
	expectedResourceData := expectedResourceRecordObj.TestResourceData()
	rrsetResource := &udnssdk.RRSet{
		OwnerName: "test.provider.ultradns.net",
		RRType:    "A",
		RData: []string{
			"10.1.1.2",
		},
		TTL: 0,
	}

	rrsetDTO := make([]map[string]interface{}, 1)
	noResponseDTO := make([]map[string]interface{}, 1)
	data := []byte(`
		[{
				"host": "10.1.1.2",
				"ttl": 300,
				"geo_info": [{
						"name": "North America",
						"codes": [
								"US"
						]
				}],
				"ip_info": [{
						"name": "rdataIpInfo",
						"ips": [{
								"address": "200.212.1.1"
						}]
				}]
		}]
				`)

	noResponseData := []byte(`
		[{
				"geo_info": [{
						"name": "america",
						"codes": [
								"EUR"
						]
				}],
				"ip_info": [{
						"name": "nrIpInfo",
						"ips": [{
								"address": "200.20.0.1"
						}]
				}]
		}]
				`)

	err := json.Unmarshal(noResponseData, &noResponseDTO)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(data, &rrsetDTO)
	if err != nil {
		log.Println(err)
	}

	expectedResourceData.Set("name", "test.provider.ultradns.net")
	expectedResourceData.Set("rdata", rrsetDTO)
	err = expectedResourceData.Set("no_response", noResponseDTO)
	expectedResourceData.Set("zone", "test.provider.ultradns.net")
	expectedResourceData.Set("type", "A")
	expectedResourceData.Set("conflict_resolve", "GEO")
	expectedResourceData.Set("description", "testing")
	expectedResourceData.Set("hostname", "test.provider.ultradns.net.test.provider.ultradns.net")

	resourceData.Set("name", "test.provider.ultradns.net")
	resourceData.Set("zone", "test.provider.ultradns.net")
	resourceData.Set("conflict_resolve", "GEO")

	//Case when no profile parameter is passed in DTO  rrsetResource
	popError := populateResourceFromDirpool(resourceData, rrsetResource)
	assert.NotNil(t, popError, true)

}

func TestAccUltradnsDirpool(t *testing.T) {
	var record udnssdk.RRSet
	domain, _ := os.LookupEnv("ULTRADNS_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccDirpoolCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testCfgDirpoolMinimal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_dirpool.it", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "name", "test-dirpool-minimal"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "type", "A"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "description", "Minimal directional pool"),
					// hashRdatas(): 10.1.0.1 -> 463398947
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "rdata.463398947.host", "10.1.0.1"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "rdata.463398947.all_non_configured", "true"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "rdata.463398947.ttl", "300"),
					// Generated
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "id", fmt.Sprintf("test-dirpool-minimal:%s", domain)),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "hostname", fmt.Sprintf("test-dirpool-minimal.%s.", domain)),
				),
			},
			{
				Config: fmt.Sprintf(testCfgDirpoolMaximal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltradnsRecordExists("ultradns_dirpool.it", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "name", "test-dirpool-maximal"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "type", "A"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "description", "Description of pool"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "conflict_resolve", "GEO"),

					// hashRdatas(): 10.1.1.1 -> 442270228
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "rdata.442270228.host", "10.1.1.1"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "rdata.442270228.all_non_configured", "true"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "rdata.442270228.ttl", "300"),

					// hashRdatas(): 10.1.1.2 -> 2203440046
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "rdata.2203440046.host", "10.1.1.2"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "rdata.2203440046.geo_info.0.name", "North America"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "rdata.2203440046.ttl", "300"),

					// hashRdatas(): 10.1.1.3 -> 4099072824
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "rdata.4099072824.host", "10.1.1.3"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "rdata.4099072824.ip_info.0.name", "some Ips"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "no_response.0.geo_info.0.name", "nrGeo"),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "no_response.0.ip_info.0.name", "nrIP"),
					// Generated
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "id", fmt.Sprintf("test-dirpool-maximal:%s", domain)),
					resource.TestCheckResourceAttr("ultradns_dirpool.it", "hostname", fmt.Sprintf("test-dirpool-maximal.%s.", domain)),
				),
			},

			{
				ResourceName:      "ultradns_dirpool.it",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDirpoolCheckDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*udnssdk.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ultradns_dirpool" {
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

const testCfgDirpoolMinimal = `
resource "ultradns_dirpool" "it" {
  zone        = "%s"
  name        = "test-dirpool-minimal"
  type        = "A"
  description = "Minimal directional pool"

  rdata {
        ttl         = 300
        host = "10.1.0.1"
        all_non_configured = true
  }
}
`

const testCfgDirpoolMaximal = `
resource "ultradns_dirpool" "it" {
  zone        = "%s"
  name        = "test-dirpool-maximal"
  type        = "A"
  description = "Description of pool"

  conflict_resolve = "GEO"

  rdata {
        ttl         = 300
        host               = "10.1.1.1"
        all_non_configured = true
  }

  rdata {
        host = "10.1.1.2"
        ttl         = 300

        geo_info {
          name = "North America"

          codes = [
        "US-OK",
        "US-DC",
        "US-MA",
          ]
        }

  }

  rdata {
        host = "10.1.1.3"

        ip_info {
          name = "some Ips"

          ips {
        start = "200.20.0.1"
        end   = "200.20.0.10"
          }

          ips {
        cidr = "20.20.20.0/24"
          }

          ips {
        address = "50.60.70.80"
          }
        }
  }

#   rdata {
#     host = "10.1.1.4"
#
#     geo_info {
#       name             = "accountGeoGroup"
#       is_account_level = true
#     }
#
#     ip_info {
#       name             = "accountIPGroup"
#       is_account_level = true
#     }
#   }

  no_response {
        geo_info {
          name = "nrGeo"

          codes = [
        "LU", "ES", "SJ", "JE", "NL", "HU", "FR", "GB", "IE", "MT", "IM", "BG", "LT", "MD", "GR", "IT", "BA", "PL", "BE", "UA", "BY", "IS", "PT", "LV", "FO", "AM", "CZ", "VA", "LI", "AT", "NO", "EE", "AX", "U5", "DE", "SE", "AL", "SM", "AD", "SI", "GG", "CH", "MK", "FI", "MC", "GI", "GE", "HR", "ME", "AZ", "RO", "SK", "DK", "RS"
          ]
        }

        ip_info {
          name = "nrIP"

          ips {
        address = "197.231.41.3"
          }
        }
  }
}
`

