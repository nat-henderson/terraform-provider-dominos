package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
	"net/url"
	"time"
)

func dataSourceStore() *schema.Resource {
	return &schema.Resource{
		Read: resourceStoreRead,
		Schema: map[string]*schema.Schema{
			"address_url_object": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"store_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"delivery_minutes": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceStoreRead(d *schema.ResourceData, m interface{}) error {
	var client = &http.Client{Timeout: 10 * time.Second}
	address_url_obj := make(map[string]string)
	err := json.Unmarshal([]byte(d.Get("address_url_object").(string)), &address_url_obj)
	if err != nil {
		return err
	}
	line1 := url.QueryEscape(address_url_obj["line1"])
	line2 := url.QueryEscape(address_url_obj["line2"])
	stores, err := getStores(fmt.Sprintf("https://order.dominos.com/power/store-locator?s=%s&c=%s&s=Delivery", line1, line2), client)
	if err != nil {
		return err
	}
	if len(stores) == 0 {
		return fmt.Errorf("No stores near the address %#v", address_url_obj)
	}
	d.Set("store_id", stores[0].StoreID)
	d.Set("delivery_minutes", stores[0].ServiceMethodEstimatedWaitMinutes.Delivery.Min)
	d.SetId("store")
	return nil
}

type StoresResponse struct {
	Stores []Store
}

type Store struct {
	StoreID                           string
	ServiceMethodEstimatedWaitMinutes WaitMinutes
}

type WaitMinutes struct {
	Delivery DeliveryMinutes
}

type DeliveryMinutes struct {
	Min int
}

func getStores(url string, client *http.Client) ([]Store, error) {
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	resp := StoresResponse{}

	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	return resp.Stores, nil
}
