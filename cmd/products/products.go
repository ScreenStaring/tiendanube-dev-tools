package products

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/cheynewallace/tabby"
	"github.com/urfave/cli/v3"

	"github.com/ScreenStaring/tiendanube-dev-tools/api"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd"
)

var Cmd cli.Command

// PrintJSONL outputs the products one JSON object per line.
func PrintJSONL(products []api.Product) error {
	for _, product := range products {
		line, err := json.Marshal(product)
		if err != nil {
			return fmt.Errorf("failed to marshal product: %w", err)
		}

		fmt.Println(string(line))
	}
	return nil
}

func displayFieldName(name string) string {
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := name[i-1]
			if prev >= 'a' && prev <= 'z' {
				result.WriteByte(' ')
			} else if prev >= 'A' && prev <= 'Z' && i+1 < len(name) && name[i+1] >= 'a' && name[i+1] <= 'z' {
				result.WriteByte(' ')
			}
		}
		result.WriteRune(r)
	}
	return result.String()
}

func isFieldToPrint(field string, selectedFields []string) bool {
	for _, f := range selectedFields {
		if f == field {
			return true
		}
	}
	return false
}

func printOptions(options []api.ProductOption) {
	t := tabby.New()
	t.AddHeader("Name", "Values")

	for _, opt := range options {
		t.AddLine(opt.Name, strings.Join(opt.Values, ", "))
	}

	t.Print()
}

func printVariants(variants []api.ProductVariant, locations map[string]string) {
	t := tabby.New()
	t.AddHeader("ID", "Title", "SKU", "Barcode", "Price", "Compare At Price", "Inventory")

	for _, v := range variants {
		t.AddLine(v.ID, v.Title(), v.SKU, v.Barcode, v.Price, v.CompareAtPrice, v.InventoryDisplay(locations))
	}

	t.Print()
}

// PrintFormatted outputs the products in the tabular format used by `products ls`.
func PrintFormatted(products []api.Product, fieldsToPrint []string, locationMaps ...map[string]string) {
	var locations map[string]string
	if len(locationMaps) > 0 {
		locations = locationMaps[0]
	}

	var normalizedFieldsToPrint []string
	for _, field := range fieldsToPrint {
		normalizedFieldsToPrint = append(normalizedFieldsToPrint, cmd.NormalizeField(field))
	}

	for _, product := range products {
		t := tabby.New()
		s := reflect.ValueOf(&product).Elem()

		for i := 0; i < s.NumField(); i++ {
			field := s.Type().Field(i).Name
			normalizedField := cmd.NormalizeField(field)
			label := displayFieldName(field)

			if normalizedField == "attributes" || normalizedField == "variants" || normalizedField == "images" || normalizedField == "options" || normalizedField == "hasonlydefaultvariant" {
				continue
			}

			if len(fieldsToPrint) > 0 {
				if isFieldToPrint(normalizedField, normalizedFieldsToPrint) {
					t.AddLine(label, s.Field(i).Interface())
				}
			} else {
				if normalizedField == "freeshipping" || normalizedField == "requiresshipping" {
					continue
				}
				t.AddLine(label, s.Field(i).Interface())
			}
		}

		t.Print()

		showAll := len(fieldsToPrint) == 0
		options := product.GetOptions()

		if (showAll && !product.HasOnlyDefaultVariant() && len(options) > 0) || (!showAll && isFieldToPrint("options", normalizedFieldsToPrint)) {
			fmt.Println("Options")
			printOptions(options)
			fmt.Print("\n")
		}

		if showAll || isFieldToPrint("variants", normalizedFieldsToPrint) {
			fmt.Println("Variants")
			printVariants(product.Variants, locations)
			fmt.Print("\n")
		}

		cmd.PrintSeparator()
	}
}

func listProducts(ctx context.Context, c *cli.Command) error {
	ids, skus, err := cmd.ParseIDArgs(c.Args().Slice(), "a product id")
	if err != nil {
		return err
	}

	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	var fields []string
	if len(c.String("fields")) > 0 {
		fields = strings.Split(c.String("fields"), ",")
	}

	var products []api.Product

	if len(ids) > 0 || len(skus) > 0 {
		var failures []string
		for _, id := range ids {
			p, err := client.GetProduct(ctx, id)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%d: %s", id, err))
				continue
			}
			products = append(products, *p)
		}
		for _, sku := range skus {
			p, err := client.GetProductBySKU(ctx, sku)
			if err != nil {
				failures = append(failures, fmt.Sprintf("sku:%s: %s", sku, err))
				continue
			}
			products = append(products, *p)
		}
		if len(failures) > 0 && len(products) == 0 {
			return fmt.Errorf("Cannot retrieve product(s): %s", strings.Join(failures, ", "))
		}
	} else {
		params := &api.ListProductsParams{
			PerPage: int(c.Int64("limit")),
		}
		if status := c.String("status"); status != "" {
			switch strings.ToLower(status) {
			case "active", "published", "true":
				t := true
				params.Published = &t
			case "draft", "archived", "unpublished", "false":
				f := false
				params.Published = &f
			}
		}
		if len(fields) > 0 {
			params.Fields = strings.Join(fields, ",")
		}

		prods, err := client.ListProducts(ctx, params)
		if err != nil {
			return err
		}
		products = prods
	}

	if c.Bool("jsonl") {
		if err := PrintJSONL(products); err != nil {
			return err
		}
	} else {
		locMap := make(map[string]string)
		if locs, err := client.ListLocations(ctx); err == nil && len(locs) > 0 {
			for _, loc := range locs {
				id := strings.TrimSpace(loc.ID.String())
				name := loc.DisplayName()
				if id != "" {
					if name == "" {
						if len(locs) == 1 || loc.IsDefault {
							name = "Default"
						} else {
							name = id
						}
					}
					locMap[id] = name
				}
			}
		}

		PrintFormatted(products, fields, locMap)
	}

	return nil
}

func deleteProducts(ctx context.Context, c *cli.Command) error {
	var ids []string

	if c.NArg() > 0 {
		for i := 0; i < c.NArg(); i++ {
			ids = append(ids, c.Args().Get(i))
		}
	} else {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("Cannot read from stdin: %s", err)
		}

		for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				ids = append(ids, line)
			}
		}
	}

	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	for _, idStr := range ids {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting product %s: invalid id: %s\n", idStr, err)
			continue
		}

		err = client.DeleteProduct(ctx, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting product %d: %s\n", id, err)
			continue
		}

		fmt.Printf("Deleted %d\n", id)
	}

	return nil
}

func init() {
	productFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "fields",
			Aliases: []string{"f"},
			Usage:   "Comma separated list of fields to output",
			Sources: cli.EnvVars("TNDT_PRODUCT_FIELDS"),
		},
		&cli.Int64Flag{
			Name:    "limit",
			Aliases: []string{"l"},
			Value:   10,
		},
		&cli.StringFlag{
			Name:    "status",
			Aliases: []string{"s"},
			Usage:   "Filter products by status (active/published or draft/archived/unpublished)",
		},
		&cli.BoolFlag{
			Name:    "jsonl",
			Aliases: []string{"j"},
			Usage:   "Output the products in JSONL format",
		},
		cmd.APIVersionFlag,
	}

	Cmd = cli.Command{
		Name:  "products",
		Usage: "Do things with products",
		Commands: []*cli.Command{
			{
				Name:      "ls",
				Usage:     "List some of a shop's products or the products matching the given IDs and/or 'sku:VALUE' arguments",
				ArgsUsage: "[ID|sku:VALUE [ID|sku:VALUE ...]]",
				Flags:     append(cmd.Flags, productFlags...),
				Action:    listProducts,
			},
			{
				Name:        "delete",
				Usage:       "Delete products by ID",
				ArgsUsage:   "[ID [ID ...]]",
				Description: "If IDs are not given they're read from stdin one per line",
				Flags:       append(cmd.Flags, cmd.APIVersionFlag),
				Action:      deleteProducts,
			},
		},
	}
}
