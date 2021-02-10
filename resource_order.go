package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

func resourceOrder() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrderCreate,
		Read:   resourceOrderRead,
		Delete: resourceOrderDelete,
		Schema: map[string]*schema.Schema{
			"address_api_object": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"item_codes": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ForceNew: true,
			},
			"store_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceOrderCreate(d *schema.ResourceData, m interface{}) error {
	var client = &http.Client{Timeout: 10 * time.Second}
	store_id := d.Get("store_id").(string)
	address_obj := make(map[string]interface{})
	err := json.NewDecoder(strings.NewReader(d.Get("address_api_object").(string))).Decode(&address_obj)
	if err != nil {
		return err
	}
	order_data := map[string]interface{}{
		"Address":               address_obj,
		"Coupons":               []string{},
		"CustomerID":            "",
		"Extension":             "",
		"OrderChannel":          "OLO",
		"OrderID":               "",
		"NoCombine":             true,
		"OrderMethod":           "Web",
		"OrderTaker":            nil,
		"Payments":              []map[string]interface{}{map[string]interface{}{"Type": "Cash"}},
		"Products":              []map[string]interface{}{},
		"Market":                "",
		"Currency":              "",
		"ServiceMethod":         "Delivery",
		"Tags":                  map[string]string{},
		"Version":               "1.0",
		"SourceOrganizationURI": "order.dominos.com",
		"LanguageCode":          "en",
		"Partners":              map[string]string{},
		"NewUser":               true,
		"metaData":              map[string]string{},
		"Amounts":               map[string]string{},
		"BusinessDate":          "",
		"EstimatedWaitMinutes":  "",
		"PriceOrderTime":        "",
		"AmountBreakdown":       map[string]string{},
		"StoreID":               store_id,
	}

	menuapi, err := getMenuApiObject(fmt.Sprintf("https://order.dominos.com/power/store/%s/menu?lang=en&structured=true", store_id), client)
	if err != nil {
		return err
	}
	codes := d.Get("item_codes").([]interface{})
	for i := range codes {
		variant := menuapi["Variants"].(map[string]interface{})[codes[i].(string)].(map[string]interface{})
		variant["ID"] = 1
		variant["isNew"] = true
		variant["Qty"] = 1
		variant["AutoRemove"] = false
		order_data["Products"] = append(order_data["Products"].([]map[string]interface{}), variant)
	}
	log.Printf("order data: %#v", order_data)
	config := m.(*Config)
	order_data["Email"] = config.EmailAddr
	order_data["FirstName"] = config.FirstName
	order_data["LastName"] = config.LastName
	order_data["Phone"] = config.PhoneNumber
	val_bytes, err := json.Marshal(map[string]interface{}{"Order": order_data})
	if err != nil {
		return err
	}
	val_req, err := http.NewRequest("POST", "https://order.dominos.com/power/price-order", strings.NewReader(string(val_bytes)))
	if err != nil {
		return err
	}
	val_req.Header.Set("Referer", "https://order.dominos.com/en/pages/order/")
	val_req.Header.Set("Content-Type", "application/json")
	dumpreq, err := httputil.DumpRequest(val_req, true)
	if err != nil {
		return err
	}
	log.Printf("http request:  %#v", string(dumpreq))
	val_rsp, err := client.Do(val_req)
	if err != nil {
		return err
	}
	dumprsp, err := httputil.DumpResponse(val_rsp, true)
	if err != nil {
		return err
	}
	log.Printf("http response: %#v", string(dumprsp))
	validate_response_obj := make(map[string]interface{})
	err = json.NewDecoder(val_rsp.Body).Decode(&validate_response_obj)
	if validate_response_obj["Status"].(float64) == -1 {
		return fmt.Errorf("The Dominos API didn't like this request: %#v", validate_response_obj["StatusItems"])
	}
	for k, v := range validate_response_obj["Order"].(map[string]interface{}) {
		if list, ok := v.([]interface{}); !ok || len(list) > 0 {
			order_data[k] = v
		}
	}

	if config.CreditCardNumber != 0 {
		order_data["Payments"] = []map[string]interface{}{map[string]interface{}{
			"Type":         "CreditCard",
			"Expiration":   config.ExprDate,
			"Amount":       order_data["Amounts"].(map[string]interface{})["Customer"],
			"CardType":     config.CardType,
			"Number":       config.CreditCardNumber,
			"SecurityCode": config.Cvv,
			"PostalCode":   config.Zip,
		}}
	}

	order_bytes, err := json.Marshal(map[string]interface{}{"Order": order_data})
	if err != nil {
		return err
	}
	order_req, err := http.NewRequest("POST", "https://order.dominos.com/power/place-order", strings.NewReader(string(order_bytes)))
	if err != nil {
		return err
	}
	order_req.Header.Set("Referer", "https://order.dominos.com/en/pages/order/")
	order_req.Header.Set("Content-Type", "application/json")

	dump_order_req, err := httputil.DumpRequest(order_req, true)
	if err != nil {
		return err
	}
	log.Printf("http request:  %#v", string(dump_order_req))
	order_rsp, err := client.Do(order_req)
	if err != nil {
		return err
	}
	dump_order_rsp, err := httputil.DumpResponse(order_rsp, true)
	if err != nil {
		return err
	}
	log.Printf("http response: %#v", string(dump_order_rsp))
	order_response_obj := make(map[string]interface{})
	err = json.NewDecoder(order_rsp.Body).Decode(&order_response_obj)
	if err != nil {
		return err
	}
	orderid := order_response_obj["Order"].(map[string]interface{})["StoreOrderID"]
	d.SetId(orderid.(string))
	return resourceOrderRead(d, m)
}

func getOrderStatus(client *http.Client, phonenumber, orderid string) (string, error) {
	order, err := getOrder(client, phonenumber, orderid)
	log.Printf("[DEBUG] order: %#v", order)
	return order.OrderStatus, err
}

type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    Body     `xml:"Body"`
}

type Body struct {
	XMLName         xml.Name        `xml:"Body"`
	TrackerResponse TrackerResponse `xml:"GetTrackerDataResponse"`
}

type TrackerResponse struct {
	XMLName       xml.Name      `xml:"GetTrackerDataResponse"`
	OrderStatuses OrderStatuses `xml:"OrderStatuses"`
}

type OrderStatuses struct {
	XMLName     xml.Name      `xml:"OrderStatuses"`
	OrderStatus []OrderStatus `xml:"OrderStatus"`
}

type OrderStatus struct {
	XMLName          xml.Name `xml:"OrderStatus"`
	OrderStatus      string   `xml:"OrderStatus"`
	OrderDescription string   `xml:"OrderDescription"`
	OrderID          string   `xml:"OrderID"`
	StartTime        string   `xml:"StartTime"`
	ParsedStartTime  time.Time
}

func getOrder(client *http.Client, phonenumber, orderid string) (*OrderStatus, error) {
	// hm.  easy to fetch, hard to process.
	// grab out "soap:Envelope.soap:Body.GetTrackerDataResponse.OrderStatuses.[OrderStatus where OrderID = orderid]
	r, err := client.Get(fmt.Sprintf("https://trkweb.dominos.com/orderstorage/GetTrackerData?Phone=%s", phonenumber))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	resp := Envelope{}
	b, err := ioutil.ReadAll(r.Body)
	log.Printf("[DEBUG] Response from tracker: %s", b)
	if err != nil {
		return nil, err
	}
	err = xml.Unmarshal(b, &resp)
	order := resp.Body.TrackerResponse.OrderStatuses.OrderStatus
	for i := range order {
		if order[i].OrderID == orderid {
			t, err := time.Parse("2006-01-02T15:04:05", order[i].StartTime)
			if err != nil {
				return nil, err
			}
			order[i].ParsedStartTime = t
			return &order[i], nil
		}
	}
	return nil, fmt.Errorf("Could not find order ID %s in tracker.", orderid)
}

func resourceOrderRead(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)
	client := &http.Client{Timeout: 10 * time.Second}
	var status string
	var err error
	for ; err == nil && status != "Complete"; status, err = getOrderStatus(client, config.PhoneNumber, d.Id()) {
		time.Sleep(10 * time.Second)
		log.Printf("[DEBUG] checking tracker...")
	}
	if err != nil {
		return err
	}
	order, err := getOrder(client, config.PhoneNumber, d.Id())
	if err != nil {
		return err
	}
	if time.Now().Sub(order.ParsedStartTime).Hours() > 12 {
		d.SetId("")
	}
	return nil
}

func resourceOrderDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}
