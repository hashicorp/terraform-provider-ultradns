package ultradns

import (
	"encoding/json"
	"fmt"
	"testing"

	udnssdk "github.com/aliasgharmhowwala/ultradns-sdk-go"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestProbeRRSetKeyCreate(t *testing.T) {
	newProbeResource := probeResource{
		Name: "test.provider.ultradns.com",
		Zone: "test.provider.ultradns.com",
		ID:   "0608485259D5AC79",
	}

	actualRRSetKey := newProbeResource.RRSetKey()

	expectedRRSetKey := udnssdk.RRSetKey{
		Name: "test.provider.ultradns.com",
		Zone: "test.provider.ultradns.com",
		Type: "A",
	}
	assert.Equal(t, expectedRRSetKey, actualRRSetKey, true)

}

func TestProbeInfoDTO(t *testing.T) {
	expectedData := []byte(`{"id":"0608485259D5AC79","type":"PING","interval":"ONE_MINUTE","agents":["DALLAS","AMSTERDAM"],"threshold":2,"details":{"packets":15,"packetSize":56,"limit":{"lossPercent":{"warning":1,"critical":2,"fail":3},"total":{"warning":2,"critical":3,"fail":4}}}}`)
	expectedResource := probeResource{}
	DetailsDTO := &udnssdk.ProbeDetailsDTO{
		Detail: udnssdk.PingProbeDetailsDTO{
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
		},
	}

	err := json.Unmarshal(expectedData, &expectedResource)
	expectedResource.Details = DetailsDTO

	actualData := []byte(`{"id":"0608485259D5AC79","type":"PING","interval":"ONE_MINUTE","agents":["DALLAS","AMSTERDAM"],"threshold":2,"details":{"packets":15,"packetSize":56,"limit":{"lossPercent":{"warning":1,"critical":2,"fail":3},"total":{"warning":2,"critical":3,"fail":4}}}}`)
	actualResource := udnssdk.ProbeInfoDTO{}
	actualDataDetail := &udnssdk.ProbeDetailsDTO{
		Detail: udnssdk.PingProbeDetailsDTO{
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
		},
	}

	err = json.Unmarshal(actualData, &actualResource)

	actualResource.Details = actualDataDetail

	if err != nil {
		log.Println(err)
	}

	if err != nil {
		log.Println(err)
	}
	expectedResource.Name = "test.provider.ultradns.net"
	expectedResource.Zone = "test.provider.ultradns.net"

	expectedDTO := expectedResource.ProbeInfoDTO()

	assert.Equal(t, expectedDTO, actualResource, true)
}

func TestMapFromLimit(t *testing.T) {

	actualDTO := udnssdk.ProbeDetailsLimitDTO{
		Warning:  1,
		Critical: 2,
		Fail:     3,
	}

	expectedDTO := map[string]interface{}{
		"name":     "lossPercent",
		"warning":  1,
		"critical": 2,
		"fail":     3,
	}

	actualResource := mapFromLimit("lossPercent", actualDTO)
	assert.Equal(t, expectedDTO, actualResource, true)
}

func TestHashLimits(t *testing.T) {
	actualInterface := map[string]interface{}{
		"name": "10.0.1.1",
	}

	expected := 658287524
	actualValue := hashLimits(actualInterface)
	assert.Equal(t, expected, actualValue, true)

}

func TestMakeProbeDetailsLimit(t *testing.T) {
	expectedDTO := &udnssdk.ProbeDetailsLimitDTO{
		Warning:  1,
		Critical: 2,
		Fail:     3,
	}

	actualDTO := map[string]interface{}{
		"name":     "lossPercent",
		"warning":  1,
		"critical": 2,
		"fail":     3,
	}

	actualData := makeProbeDetailsLimit(actualDTO)
	assert.Equal(t, expectedDTO, actualData, true)
}

func TestHashRdatas(t *testing.T) {
	actualInterface := map[string]interface{}{
		"host": "10.0.1.1",
	}
	expected := 658287524
	actualValue := hashRdatas(actualInterface)
	assert.Equal(t, expected, actualValue, true)
}

func testAccRdpoolCheckDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*udnssdk.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ultradns_rdpool" {
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

func testAccTcpoolCheckDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*udnssdk.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ultradns_tcpool" {
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

func testAccCheckUltradnsRecordExists(n string, record *udnssdk.RRSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*udnssdk.Client)
		k := udnssdk.RRSetKey{
			Zone: rs.Primary.Attributes["zone"],
			Name: rs.Primary.Attributes["name"],
			Type: rs.Primary.Attributes["type"],
		}

		foundRecord, err := client.RRSets.Select(k)

		if err != nil {
			return err
		}

		if foundRecord[0].OwnerName != rs.Primary.Attributes["hostname"] {
			return fmt.Errorf("Record not found: %+v,\n %+v\n", foundRecord, rs.Primary.Attributes)
		}

		*record = foundRecord[0]

		return nil
	}
}
