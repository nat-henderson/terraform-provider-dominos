package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"net/http"
	"strings"
	"time"
)

func dataSourceMenuItem() *schema.Resource {
	return &schema.Resource{
		Read: resourceMenuItemRead,
		Schema: map[string]*schema.Schema{
			"store_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"query_string": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"matches": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"code": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"price_cents": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceMenuItemRead(d *schema.ResourceData, m interface{}) error {
	var client = &http.Client{Timeout: 10 * time.Second}
	store_id := d.Get("store_id").(string)
	menuitems, err := getAllMenuItems(fmt.Sprintf("https://order.dominos.com/power/store/%s/menu?lang=en&structured=true", store_id), client)
	if err != nil {
		return err
	}
	menu := make([]map[string]interface{}, 0, len(menuitems))
	queries := d.Get("query_string").([]interface{})
Menu:
	for i := range menuitems {
		for j := range queries {
			if !strings.Contains(strings.ToLower(menuitems[i].Name), strings.ToLower(queries[j].(string))) {
				continue Menu
			}
		}
		menu = append(menu, map[string]interface{}{"name": menuitems[i].Name, "code": menuitems[i].Code, "price_cents": menuitems[i].PriceCents})
	}
	if err := d.Set("matches", menu); err != nil {
		return err
	}
	log.Printf("len menu: %d", len(menu))
	d.SetId(store_id)
	return nil
}
