package ultradns

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	log "github.com/sirupsen/logrus"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terra-farm/udnssdk"
)

func newRRSetResource(d *schema.ResourceData) (rRSetResource, error) {
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

func populateResourceDataFromRRSet(r udnssdk.RRSet, d *schema.ResourceData) error {
	zone := d.Get("zone")
	typ := d.Get("type")
	if typ == "" {
		typ = (strings.Split(r.RRType," "))[0]
		d.Set("type",typ)
	}
	log.Infof("type = %s %s %s",typ,zone,strconv.Itoa(r.TTL))
	// ttl
	d.Set("ttl", strconv.Itoa(r.TTL))
	// rdata
	rdata := r.RData

	// UltraDNS API returns answers double-encoded like JSON, so we must decode. This is their bug.
	if typ == "TXT" {
		rdata = make([]string, len(r.RData))
		for i := range r.RData {
			var s string
			err := json.Unmarshal([]byte(r.RData[i]), &s)
			if err != nil {
				log.Printf("[INFO] TXT answer parse error: %+v", err)
				s = r.RData[i]
			}
			rdata[i] = s

		}
	}

	err := d.Set("rdata", makeSetFromStrings(rdata))
	if err != nil {
		return fmt.Errorf("ultradns_record.rdata set failed: %#v", err)
	}
	// hostname
	if r.OwnerName == "" {
		d.Set("hostname", zone)
	} else {
		if strings.HasSuffix(r.OwnerName, ".") {
			d.Set("hostname", r.OwnerName)
		} else {
			d.Set("hostname", fmt.Sprintf("%s.%s", r.OwnerName, zone))
		}
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

                Importer: &schema.ResourceImporter{
                        State: resourceUltradnsRecordImport ,
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
	log.Infof("RRSET INFO: %v",rrsets)
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
	rec := rrsets[0]
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

// Conversion helper functions

func parse(id string) (name, zone string) {
	var id_parts = strings.Split(id, ".")
	for x := len(id_parts) - 1; x >= 0; x-- {
		var n = strings.Join(id_parts[0:x], ".")
		var z = strings.Join(id_parts[x:], ".")
		if len(n) < len(z) {
			break
		}
		if strings.HasSuffix(n, z) {
			name = n
			zone = z
			return
		}
	}
	return
}

func resourceUltradnsRecordImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Infof("id= %s",d.Id())
//	name,zone := parse(d.Id())
	newId := strings.TrimSuffix(d.Id(),".")
	attributes := strings.SplitN(newId, ".", 2)
	if len(attributes) > 1{
		d.Set("zone",attributes[1])
	        d.Set("name",attributes[0])
		log.Infof("name = %s   %s",attributes[0],attributes[1])
	}else{
		d.Set("zone",attributes)
		d.Set("name",attributes)
	}
	return []*schema.ResourceData{d}, nil
}
