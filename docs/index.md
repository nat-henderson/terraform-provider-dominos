# Provider Purpose
The dominos provider exists to ensure that while your cloud infrastructure is spinning up, you can have a hot pizza delivered to you.  This paradigm-shifting expansion of Terraform's "resource" model into the physical world was inspired in part by the realization that Google has a REST API for Interconnects, e.g. for people with hard-hats laying digging up the ground, laying fiber.  If you can use Terraform to summon folks with shovels to drop a fiber line, why shouldn't you be able
to summon a teenager with a pizza?

# Provider Overview

The Dominos Pizza provider is made up primarily of data sources.  The only thing you can truly `Create` with this provider is, of course, an order from Dominos.

Further, since Hashicorp is not extremely interested in integrating `terraform-provider-dominos` into Terraform, you will have to install the provider manually.  Follow instructions in README.md.

# Using the Provider

If you are a true Dominos afficionado, you may already know the four-digit store ID of the store closest to you, the correct json-format for your address, the six-to-ten-digit code for the item you want to order.  If you are one of those people, you can feel free to construct a `dominos_order` resource from scratch.

For the rest of us, I recommend one of each of the data sources.  They feed into each other in an obvious way.

## Provider Configuration

If you plan to place an order, you need to set the following fields in the `provider "dominos" {}` block:
* `email_address`
* `first_name`
* `last_name`
* `phone_number`

The credit card fields are optional - if you do not configure a credit card, you will be paying by cash when the delivery driver arrives.

The `credit_card` block requires the following fields:
* `number`: just an integer, the whole credit card number.  The API accepts Visa, Amex, and Mastercard.
* `cvv`: also an integer
* `zip`: integer!
* `date`: The experiation date, as a string, with a slash between the month and year, e.g. `03/19`.

If you don't plan to place an order, you don't need to fill this out.

## Data Sources
### `dominos_address`

This data source takes in your address and writes it back out in the two different JSON formats that the API expects.  Configure it with `street`, `city`, `state`, and `zip`, and use `url_object` and `api_object` in other data sources where required.

### `dominos_store`

This data source takes in the `url_object` of your address, and returns the `store_id`, and, in case it's useful to you somehow, the `delivery_minutes`, an integer showing the estimated minutes until your pizza will be delivered.

### `dominos_menu_item`

This data source takes in the `store_id` and a list of strings (as `query_string`), and outputs the menu items in `matches`.  Each item in `matches` has three attributes: `name`, `code`, and `price_cents`.  The name is human-readable, but not useful for ordering.  The `price_cents` is also only informational.  `code` is the value that will be useful in a `dominos_order`.

Each string in `query_string` must literally match the `name` of the menu item for the menu item to appear in `matches`.

### `dominos_menu`

If you would prefer to do your own filtering, you can get access to every item on the dominos menu in your area using this data source.  This data source takes in `store_id` and provides `menu`, a list of all (186, at my dominos) `name`/`code`/`price_cents` blocks.

For the love of all that's holy, do not accidentally feed this data source directly into the `dominos_order`.  This will be expensive and probably pretty annoying to the Dominos store, which will be serving you 1 of each 2-liter bottle of sode, 1 of each 20oz bottle, at least 4 different kinds of salad, probably like 6 different kinds of chicken wings, and I think 12 of each kind of pizza?  (Small, medium, large) x (Hand Tossed, Pan, Stuffed Crust, Gluten Free)?  Oh plus breads.  There's breads on the menu, I found that out while trawling through API responses.  I wonder who eats those.  Are they good?  Let me know!

## Resources

### `dominos_order`

This is it!  This will order you your pizzas!  Configure it with:
* `address_api_object`, from your `dominos_address` data source.
* `item_codes`, a list of strings, from your `dominos_menu_item` or `dominos_menu` data source.
* `store_id`, the ID from your `dominos_store` data source.

As far as I know there is no way to cancel a dominos order programmatically, so if you made a mistake, you'll have to call the store.  You should receive an email confirmation almost instantly, and that email will have the store's phone number in it.

# Using the Dominos Provider

It's pretty simple.

## Sample Configuration

```
provider "dominos" {
  first_name = "My"
  last_name = "Name"
  email_address = "my@name.com"
  phone_number = "15555555555"

  credit_card {
    number = 123456789101112
    cvv = 1314
    date = "15/16"
    zip = 18192
  }
}

data "dominos_address" "addr" {
  street = "123 Main St"
  city = "Anytown"
  state = "WA"
  zip = "02122"
}

data "dominos_store" "store" {
  address_url_object = "${data.dominos_address.addr.url_object}"
}

data "dominos_menu_item" "item" {
  store_id = "${data.dominos_store.store.store_id}"
  query_string = ["philly", "medium"]
}

resource "dominos_order" "order" {
  address_api_object = "${data.dominos_address.addr.api_object}"
  item_codes = ["${data.dominos_menu_item.matches.0.code}"]
  store_id = "${data.dominos_store.store.store_id}"
}
```

Now I don't know what you're going to get since I don't know what a medium philly is in your area, but in my area it gets you a 12" hand-tossed philly cheesesteak pizza, and it's pretty good.  It's all right.  Regular dominos.

# Warnings and Caveats

1)  The author(s) of this software are not in any sense associated with Domino's Pizza.  It was an idea a bunch of us had while working on the Google provider, but this software isn't associated with Google, either.  For further details you can read LICENSE.md.
1)  If your cloud infrastructure is kubernetes-based or otherwise slow to spin up, your pizza might arrive before your changes finish applying.  This will be very embarrassing, and potentially distracting.  Use caution.
1)  This is not a joke provider.  Or, it kind of is a joke, but even though it's a joke it will still order you a pizza.  You are going to get a pizza.  You should be careful with this provider, if you don't want a pizza.
1)  Even if you do want a pizza, you should probably be careful with this provider.  In testing, I once nearly ordered every item on the Domino's menu, which would probably have been expensive and embarrassing.
1)  You do have to put your actual credit card information into this provider, because you will, again, be purchasing and receiving a pizza.
1)  Although all your credit card information is marked `Sensitive` in schema, that's the only protection they've got.  If your state storage isn't secure, maybe don't use this provider.  Or use a virtual card number, or something.  Be smart.  Again, real credit card, real money, real pizza.
1)  I cannot emphasize enough how much you are actually going to be ordering a pizza.  Please do not be surprised when you receive a pizza and a corresponding charge to your credit card.
1)  As far as I know, there is no programmatic way to `destroy` an existing pizza.  `terraform destroy` is implemented on the client side, by consuming the pizza.
1)  The dominos API supports an astonishing amount of customization of your items.  This is where "none pizza with left beef" comes from.  You can't do any of that with this provider.  Order off the menu!
1)  Dominos probably exists outside the US, but I have no idea what will happen if you try to order a pizza outside the US.
1)  This provider auto-accepts Dominos' canonicalization of your address.  If you live someplace the post office doesn't know about, you might have trouble.
