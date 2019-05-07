Terraform Provider for Dominos Pizza
==================

# [Documentation](https://ndmckinley.github.io/terraform-provider-dominos/)

# Quickstart

Download `bin/terraform-provider-dominos` and place it on your machine at `~/.terraform.d/plugins/terraform-provider-dominos`.  Make sure to `chmod +x` it.  This is the normal way to install third-party providers - follow instructions at [Installing 3rd Party Plugins](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins) if you have trouble.

```sh
mkdir ~/.terraform.d/plugins && \
wget https://github.com/ndmckinley/terraform-provider-dominos/raw/master/bin/terraform-provider-dominos -O ~/.terraform.d/plugins/terraform-provider-dominos && \
chmod +x ~/.terraform.d/plugins/terraform-provider-dominos
```

Then write your config.  Here's a sample config - a variation on this worked for me last night.

```hcl
provider "dominos" {
  first_name    = "My"
  last_name     = "Name"
  email_address = "my@name.com"
  phone_number  = "15555555555"

  credit_card {
    number = 123456789101112
    cvv    = 1314
    date   = "15/16"
    zip    = 18192
  }
}

data "dominos_address" "addr" {
  street = "123 Main St"
  city   = "Anytown"
  state  = "WA"
  zip    = "02122"
}

data "dominos_store" "store" {
  address_url_object = "${data.dominos_address.addr.url_object}"
}

data "dominos_menu_item" "item" {
  store_id     = "${data.dominos_store.store.store_id}"
  query_string = ["philly", "medium"]
}

resource "dominos_order" "order" {
  address_api_object = "${data.dominos_address.addr.api_object}"
  item_codes         = ["${data.dominos_menu_item.item.matches.0.code}"]
  store_id           = "${data.dominos_store.store.store_id}"
}
```


`terraform init` as usual and `plan`!  `apply` when ready - but use caution, since this is going to charge you money.

Please view the docs [here](https://ndmckinley.github.io/terraform-provider-dominos/) for more information past the quickstart, as well as some caveats it's probably worth being aware of.
