package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"net/http"
	"time"
)

func dataSourceTracking() *schema.Resource {
	return &schema.Resource{
		Read: resourceTrackingRead,
		Schema: map[string]*schema.Schema{
			"order_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"store_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceTrackingRead(d *schema.ResourceData, m interface{}) error {
	var client = &http.Client{Timeout: 10 * time.Second}
	order_id := d.Get("order_id").(string)
	store_id := d.Get("store_id").(string)
	_, err := getTrackingApiObject(fmt.Sprintf("https://trkweb.dominos.com/orderstorage/GetTrackerData?StoreID=%s&OrderKey=%s", store_id, order_id), client)
	if err != nil {
		return err
	}
	return nil
}

func getTrackingApiObject(url string, client *http.Client) (map[string]interface{}, error) {
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	resp := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&resp)
	log.Printf("Tracking response: %#v", resp)
	return resp, err
}
