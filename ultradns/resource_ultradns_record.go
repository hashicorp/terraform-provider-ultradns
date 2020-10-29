package ultradns

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	log "github.com/sirupsen/logrus"
	udnssdk "github.com/ultradns/ultradns-sdk-go"
)

func newRRSetResource(d *schema.ResourceData) (rRSetResource, error) {
	log.Infof("Schema  = %+v", d)
	r := rRSetResource{}

	// TODO: return error if required attributes aren't ok

	if attr, ok := d.GetOk("name"); ok {
		r.OwnerName = attr.(string)
	}

	if attr, ok := d.GetOk("type"); ok {
		r.RRType = attr.(string)
	}

	if attr, ok := d.GetOk("zone"); ok {
		r.Zone = attr.(string)
	}

	if attr, ok := d.GetOk("rdata"); ok {
		rdata := attr.(*schema.Set).List()
		r.RData = make([]string, len(rdata))
		for i, j := range rdata {
			r.RData[i] = j.(string)
		}
	}

	if attr, ok := d.GetOk("ttl"); ok {
		r.TTL, _ = strconv.Atoi(attr.(string))
	}

	return r, nil
}

func populateResourceDataFromRRSet(r []udnssdk.RRSet, d *schema.ResourceData) error {
	for _, rrset := range r {
		zone := d.Get("zone")
		typ := d.Get("type")
		log.Infof("RRTYPE: %v", rrset.RRType)
		if (typ != (strings.Split(rrset.RRType, " "))[0]) && (typ != "TXT") {
			continue
		}

		//setting type
		d.Set("type", typ)

		log.Infof("typ = %s %s %s", typ, zone, strconv.Itoa(rrset.TTL))
		// ttl
		d.Set("ttl", strconv.Itoa(rrset.TTL))
		// rdata
		rdata := rrset.RData

		// UltraDNS API returns answers double-encoded like JSON, so we must decode. This is their bug.
		if typ == "TXT" {
			rdata = make([]string, len(rrset.RData))
			for i := range rrset.RData {
				var s string
				err := json.Unmarshal([]byte(rrset.RData[i]), &s)
				if err != nil {
					log.Printf("[INFO] TXT answer parse error: %+v", err)
					s = rrset.RData[i]
				}
				rdata[i] = s

			}
		}

		err := d.Set("rdata", makeSetFromStrings(rdata))
		if err != nil {
			return fmt.Errorf("ultradns_record.rdata set failed: %#v", err)
		}
		// hostname
		if rrset.OwnerName == "" {
			d.Set("hostname", zone)
		} else {
			if strings.HasSuffix(rrset.OwnerName, ".") {
				d.Set("hostname", rrset.OwnerName)
			} else {
				d.Set("hostname", fmt.Sprintf("%s.%s", rrset.OwnerName, zone))
			}
		}
		break
	}
	return nil
}

func resourceUltradnsRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceUltraDNSRecordCreate,
		Read:   resourceUltraDNSRecordRead,
		Update: resourceUltraDNSRecordUpdate,
		Delete: resourceUltraDNSRecordDelete,

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

		Importer: &schema.ResourceImporter{
			State: resourceUltradnsRecordImport,
		},
	}
}

// CRUD Operations

func resourceUltraDNSRecordCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newRRSetResource(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ultradns_record create: %+v", r)
	_, err = client.RRSets.Create(r.RRSetKey(), r.RRSet())
	if err != nil {
		return fmt.Errorf("create failed: %v", err)
	}
	d.SetId(r.ID())
	log.Printf("[INFO] ultradns_record.id: %v", d.Id())

	return resourceUltraDNSRecordRead(d, meta)
}

func resourceUltraDNSRecordRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newRRSetResource(d)
	if err != nil {
		return err
	}

	rrsets, err := client.RRSets.Select(r.RRSetKey())
	log.Infof("RRSET INFO: %v", rrsets)
	if err != nil {
		uderr, ok := err.(*udnssdk.ErrorResponseList)
		if ok {
			for _, r := range uderr.Responses {
				// 70002 means Records Not Found
				if r.ErrorCode == 70002 {
					d.SetId("")
					return nil
				}
				return fmt.Errorf("not found: %v", err)
			}
		}
		return fmt.Errorf("not found: %v", err)
	}
	rec := rrsets
	return populateResourceDataFromRRSet(rec, d)
}

func resourceUltraDNSRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newRRSetResource(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ultradns_record update: %+v", r)
	_, err = client.RRSets.Update(r.RRSetKey(), r.RRSet())
	if err != nil {
		return fmt.Errorf("update failed: %v", err)
	}

	return resourceUltraDNSRecordRead(d, meta)
}

func resourceUltraDNSRecordDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newRRSetResource(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ultradns_record delete: %+v", r)
	_, err = client.RRSets.Delete(r.RRSetKey())
	if err != nil {
		return fmt.Errorf("delete failed: %v", err)
	}

	return nil
}

// State Function to seperate id into appropriate name and zone
func resourceUltradnsRecordImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return setResourceAndParseId(d)
}
