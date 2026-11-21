# Tiendanube / Nuvemshop Development Tools

Command-line program to assist with the development and maintenance of Tiendanube and Nuvemshop apps and stores.

## Installation

Download the version for your platform on the [releases page](https://github.com/ScreenStaring/tiendanube-dev-tools/releases).
Windows, macOS/OS X, and GNU/Linux are supported.

### Using [Homebrew](https://brew.sh/)

```
brew tap screenstaring/tap
brew install tiendanube-dev-tools
```

## Usage

The CLI interface uses the executable `tndt`:

    NAME:
       tndt - Tiendanube / Nuvemshop Development Tools

    USAGE:
       tndt [global options] [command [command options]]

    VERSION:
       0.0.1

    COMMANDS:
       products    Do things with products
       locations   Do things with locations
       metafields     Metafield utilities
       orders, order  Information about orders
       scripts        Script utilities
       store       Information about the given store
       webhooks    Webhook utilities
       help, h     Shows a list of commands or help for one command

    GLOBAL OPTIONS:
       --help, -h     show help
       --version, -v  print the version

## Credentials & Configuration

To make calls to the Tiendanube / Nuvemshop REST API, you need your store ID and access token.

```
tndt webhooks ls --store 123456 --access-token YOUR_ACCESS_TOKEN
```

### Environment Variables

You can configure options via environment variables:

- `TNDT_STORE_ID` or `TIENDANUBE_STORE_ID`
- `TNDT_ACCESS_TOKEN` or `TIENDANUBE_ACCESS_TOKEN`
- `TNDT_USER_AGENT` - custom User-Agent header for API requests (can also be set with `--ua` or `--user-agent`)
- `TNDT_PRODUCT_FIELDS` - default fields for the `products` command's `--fields` flag
- `TNDT_MAX_RETRY_ATTEMPTS` (default: `10`)
- `TNDT_READONLY` - set to `1` or `true` to block state-mutating requests (POST, PUT, DELETE)

### Dynamic Access Token Command

The `--access-token` option supports prefixing with `<` to execute a shell command dynamically:

```
tndt webhooks ls --store 123456 --access-token '<get-token.sh'
```

The command is executed and passed the store ID as an argument.

## Commands

### Locations

    NAME:
       tndt locations - Do things with locations

    USAGE:
       tndt locations [command [command options]]

    COMMANDS:
       ls, l  List the shop's locations or the locations given by the specified IDs

    OPTIONS:
       --help, -h  show help

#### Listing Locations

List all locations for the store:

```
tndt locations ls
```

List specific locations by ID:

```
tndt locations ls 1001 1002
```

Output as JSONL (one JSON object per line):

```
tndt locations ls -j
```

### Metafields

Tiendanube / Nuvemshop supports namespaced key-value metadata attached to store entities (`products`, `product_variants`, `categories`, `orders`, `customers`, and `pages`).

    NAME:
       tndt metafields - Metafield utilities

    USAGE:
       tndt metafields [command [command options]]

    COMMANDS:
       delete      Delete one or more metafields
       product     List metafields for the given product(s)
       variant     List metafields for the given variant(s)
       customer    List metafields for the given customer
       collection  List metafields for the given collection(s)
       orders      List metafields for the given order(s)
       page        List metafields for the given page(s)

    OPTIONS:
       --help, -h  show help

#### Listing Metafields

List metafields for a product by ID or variant SKU:

```
tndt metafields product 12345678
tndt metafields product sku:REM-RED-S
```

List metafields for a variant by ID or SKU:

```
tndt metafields variant 98765432
tndt metafields variant sku:REM-RED-S
```

List metafields for a customer, collection/category, order, or page:

```
tndt metafields customer 5555
tndt metafields collection 12849
tndt metafields orders 7777
tndt metafields page 3333
```

Filter by namespace and/or key:

```
tndt metafields product -n details -k material 12345678
```

Reverse results order with `-r` / `--reverse`:

```
tndt metafields product -r 12345678
```

Output in JSONL format with `-j` / `--jsonl`:

```
tndt metafields product -j 12345678
```

#### Deleting Metafields

Delete metafields by ID:

```
tndt metafields delete 13226 13227
```

Delete via stdin:

```
tndt metafields delete < metafield_ids.txt
```

### Orders

Information about orders

    NAME:
       tndt orders - Information about orders

    USAGE:
       tndt orders [command [command options]]

    COMMANDS:
       ls  List the shop's orders or the orders matching the given IDs, 'sku:VALUE' and/or 'name:VALUE' arguments

#### Listing Orders

`tndt orders ls` lists open orders by default. Use `-s`/`--status` to filter by status (e.g. `open`, `closed`, `cancelled`, `any`).
The maximum number of orders returned is configured with `-l`/`--limit` (defaults to 10).

List open orders:

```
tndt orders ls
```

Filter by status:

```
tndt orders ls -s closed
tndt orders ls -s any
```

Orders can also be looked up by their IDs, SKU, or name/number:

```
tndt orders ls 871254203
tndt orders ls sku:SHIRT-RED
tndt orders ls name:#101
```

The `sku:` and `name:` prefixes are matched case-insensitively.

Output in JSONL format with `-j` / `--jsonl`:

```
tndt orders ls -j
```

### Products

    NAME:
       tndt products - Do things with products

    USAGE:
       tndt products [command [command options]]

    COMMANDS:
       ls      List some of a shop's products or the products matching the given IDs and/or 'sku:VALUE' arguments
       delete  Delete products by ID

    OPTIONS:
       --help, -h  show help

#### Listing Products

List the first 10 products:

```
tndt products ls
```

Specify a limit:

```
tndt products ls -l 25
```

Filter by IDs and/or variant SKUs:

```
tndt products ls 12345678 sku:REM-RED-S 87654321
```

Filter by status (`active` / `published` or `draft` / `archived` / `unpublished`):

```
tndt products ls -s active
```

Limit displayed fields with `-f` / `--fields`:

```
tndt products ls -f id,name,published,canonical_url
```

Output as JSONL (one JSON object per line):

```
tndt products ls -j
```

#### Deleting Products

Delete one or more products by ID:

```
tndt products delete 12345678 87654321
```

Delete products from stdin (one ID per line):

```
tndt products delete < ids.txt
```

### Scripts

Script utilities

    NAME:
       tndt scripts - Script utilities

    USAGE:
       tndt scripts [command [command options]]

    COMMANDS:
       create  Associate a script with the store
       update  Update an existing script association
       delete  Delete a script by ID
       ls      List scripts for the given store or scripts matching the given IDs or URLs

#### Creating Scripts

Associate a script with a store by its ID:

```
tndt scripts create 12345
```

Pass query parameters using `-q` / `--query`:

```
tndt scripts create -q key1=val1 -q key2=val2 12345
```

#### Updating Scripts

Update an existing script association by its ID:

```
tndt scripts update -q key1=val1 -q key2=val2 12345
```

#### Deleting Scripts

Delete a script by its ID:

```
tndt scripts delete 12345
```

#### Listing Scripts

List scripts for the store:

```
tndt scripts ls
```

List scripts by ID or script URL:

```
tndt scripts ls 12345
tndt scripts ls https://example.com/app.js
```

Output in JSONL format with `-j` / `--jsonl`:

```
tndt scripts ls -j
```

### Store

    NAME:
       tndt store - Information about the given store

    USAGE:
       tndt store [command [command options]]

    COMMANDS:
       info  Information about the store

    OPTIONS:
       --help, -h  show help

#### Store Information

Display details about the authenticated store:

```
tndt store info
```

### Webhooks

    NAME:
       tndt webhooks - Webhook utilities

    USAGE:
       tndt webhooks [command [command options]]

    COMMANDS:
       create  Create a webhook for the given shop
       delete  Delete the given webhook
       update  Update the given webhook
       ls      List the shop's webhooks

    OPTIONS:
       --help, -h  show help

#### Listing Webhooks

List all webhooks:

```
tndt webhooks ls
```

List specific webhooks by ID:

```
tndt webhooks ls 12345 67890
```

Filter by event or URL / address:

```
tndt webhooks ls -e order/created
tndt webhooks ls -a https://myapp.com/webhook
```

Output in JSONL format:

```
tndt webhooks ls -j
```

#### Creating Webhooks

Create a webhook with an event and target address:

```
tndt webhooks create -e product/created -a https://myapp.com/product_hook
```

Avoid duplicates using `-1` / `--one`:

```
tndt webhooks create -1 -e product/created -a https://myapp.com/product_hook
```

#### Updating Webhooks

Update an existing webhook's endpoint URL or event:

```
tndt webhooks update 8901 -a https://myapp.com/new_url
```

#### Deleting Webhooks

Delete by ID:

```
tndt webhooks delete 8901
```

Delete by event name:

```
tndt webhooks delete product/created
```

Delete multiple IDs or events in one call:

```
tndt webhooks delete 8901 order/paid 8902
```

Delete all webhooks for the store:

```
tndt webhooks delete --all
```

## Development

### Running Tests

Run the test suite with:

```
go test ./...
```

Tests that do not require network access run automatically.

### Integration Tests

Tests that hit the live Tiendanube / Nuvemshop API (such as webhook CRUD tests) are skipped by default. To run them, set the following environment variables:

- `TNDT_STORE_ID` or `TIENDANUBE_STORE_ID` - The store ID
- `TNDT_ACCESS_TOKEN` or `TIENDANUBE_ACCESS_TOKEN` - The API access token

To run all tests including live API integration tests:

```
TNDT_STORE_ID=123456 TNDT_ACCESS_TOKEN=your_token go test -v ./...
```

Note: Integration tests create and delete real resources on the configured store, so always run them against a dedicated test or development store.

## See Also

- [tiendanube_url](https://github.com/ScreenStaring/tiendanube_url/) - Ruby gem to Build URLs to Tiendanube and Nuvemshop web pages
- [Shopify Dev Tools](https://github.com/ScreenStaring/shopify-dev-tools/) - Command-line program to assist with the development and/or maintenance of Shopify apps and stores

## License

Released under the MIT License.

---

Made by [ScreenStaring](https://screenstaring.com)
