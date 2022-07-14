package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func dataSourceMenu() *schema.Resource {
	return &schema.Resource{
		Read: resourceMenuRead,
		Schema: map[string]*schema.Schema{
			"store_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"menu": &schema.Schema{
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

func resourceMenuRead(d *schema.ResourceData, m interface{}) error {
	var client = &http.Client{Timeout: 10 * time.Second}
	store_id := d.Get("store_id").(string)
	menuitems, err := getAllMenuItems(fmt.Sprintf("https://order.dominos.com/power/store/%s/menu?lang=en&structured=true", store_id), client)
	if err != nil {
		return err
	}
	menu := make([]map[string]interface{}, 0, len(menuitems))
	for i := range menuitems {
		menu = append(menu, map[string]interface{}{"name": menuitems[i].Name, "code": menuitems[i].Code, "price_cents": menuitems[i].PriceCents})
	}
	if err := d.Set("menu", menu); err != nil {
		return err
	}
	log.Printf("len menu: %d", len(menu))
	d.SetId(store_id)
	return nil
}

type MenuItem struct {
	Code       string
	Name       string
	PriceCents int64
}

func getMenuApiObject(url string, client *http.Client) (map[string]interface{}, error) {
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	resp := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func getAllMenuItems(url string, client *http.Client) ([]MenuItem, error) {
	resp, err := getMenuApiObject(url, client)
	if err != nil {
		return nil, err
	}
	products := resp["Variants"].(map[string]interface{})
	all_products := make([]MenuItem, 0, len(products))
	log.Printf("len products: %d", len(products))
	for name, d := range products {
		dict := d.(map[string]interface{})
		price := dict["Price"].(string)
		price = strings.Replace(price, ".", "", 1)
		price_cents, err := strconv.ParseInt(price, 10, 64)
		if err != nil {
			continue
		}
		all_products = append(all_products, MenuItem{
			Code:       name,
			Name:       dict["Name"].(string),
			PriceCents: price_cents,
		})
	}
	sort.Slice(all_products, func(i, j int) bool {
		return all_products[i].Code < all_products[j].Code
	})
	// for each entry in Products, make a MenuItem struct and return it.
	return all_products, nil
}
