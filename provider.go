package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"strconv"
	"strings"
)

// Provider exports the actual provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"email_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"first_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"last_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"phone_number": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"credit_card": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number": &schema.Schema{
							Type:      schema.TypeInt,
							Required:  true,
							Sensitive: true,
						},
						"cvv": &schema.Schema{
							Type:      schema.TypeInt,
							Required:  true,
							Sensitive: true,
						},
						"date": &schema.Schema{
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						"zip": &schema.Schema{
							Type:      schema.TypeInt,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"dominos_order": resourceOrder(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"dominos_address":   dataSourceAddress(),
			"dominos_store":     dataSourceStore(),
			"dominos_menu":      dataSourceMenu(),
			"dominos_menu_item": dataSourceMenuItem(),
		},
		ConfigureFunc: providerConfigure,
	}
}

type Config struct {
	FirstName        string
	LastName         string
	EmailAddr        string
	PhoneNumber      string
	CreditCardNumber int64
	Cvv              int64
	ExprDate         string
	Zip              int64
	CardType         string
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		FirstName:   d.Get("first_name").(string),
		LastName:    d.Get("last_name").(string),
		EmailAddr:   d.Get("email_address").(string),
		PhoneNumber: d.Get("phone_number").(string),
	}

	if _, ok := d.GetOk("credit_card"); ok {
		config.CreditCardNumber = int64(d.Get("credit_card.0.number").(int))
		config.Cvv = int64(d.Get("credit_card.0.cvv").(int))
		config.ExprDate = d.Get("credit_card.0.date").(string)
		config.Zip = int64(d.Get("credit_card.0.zip").(int))
		n := strconv.Itoa(int(config.CreditCardNumber))

		if strings.HasPrefix(n, "3") {
			config.CardType = "AMEX"
		} else if strings.HasPrefix(n, "4") {
			config.CardType = "VISA"
		} else if strings.HasPrefix(n, "5") {
			config.CardType = "MASTERCARD"
		} else if strings.HasPrefix(n, "6") {
			config.CardType = "DISCOVER"
		}
	}
	return &config, nil
}
