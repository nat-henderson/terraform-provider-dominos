package main

import (
	"fmt"
    "encoding/json"
    "log"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAddress() *schema.Resource {
	return &schema.Resource{
		Create: resourceAddressCreate,
		Read:   resourceAddressRead,
		Delete: resourceAddressDelete,
		Schema: map[string]*schema.Schema{
			"street": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
                ForceNew: true,
			},
			"city": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
                ForceNew: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
                ForceNew: true,
			},
			"zip": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
                ForceNew: true,
			},
			"url_object": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
            "api_object": &schema.Schema{
                Type: schema.TypeString,
                Computed: true,
            },
		},
	}
}

func resourceAddressCreate(d *schema.ResourceData, m interface{}) error {
	d.SetId("address")
	urlobj := map[string]string{
		"line1": d.Get("street").(string),
		"line2": fmt.Sprintf("%s, %s %s", d.Get("city").(string), d.Get("state").(string), d.Get("zip").(string)),
	}
    apiobj := map[string]string{
        "Street": d.Get("street").(string),
        "City": d.Get("city").(string),
        "Region": d.Get("state").(string),
        "PostalCode": d.Get("zip").(string),
        "Type": "House",
    }
    url_json, err := json.Marshal(urlobj)
    if err != nil {
        return err
    }
    url_json_string := string(url_json)
    log.Printf("[DEBUG] url json: %#v to %s", urlobj, url_json_string)
    if err := d.Set("url_object", url_json_string); err != nil {
        return err
    }
    api_json, err := json.Marshal(apiobj)
    if err != nil {
        return err
    }
    api_json_string := string(api_json)
    if err := d.Set("api_object", api_json_string); err != nil {
        return err
    }
	return nil
}

func resourceAddressRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceAddressDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
