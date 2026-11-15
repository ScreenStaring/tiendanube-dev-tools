package metafields

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/ScreenStaring/tiendanube-dev-tools/api"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd"
)

var Cmd cli.Command

func parseIDArgsOnly(args []string, name string) ([]int64, error) {
	ids := make([]int64, len(args))
	for i, arg := range args {
		id, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s id '%s' invalid: must be an int", name, arg)
		}
		ids[i] = id
	}
	return ids, nil
}

func reverseMetafields(metafields []api.Metafield) {
	for i, j := 0, len(metafields)-1; i < j; i, j = i+1, j-1 {
		metafields[i], metafields[j] = metafields[j], metafields[i]
	}
}

func printJSONL(metafields []api.Metafield) error {
	for _, metafield := range metafields {
		line, err := json.Marshal(metafield)
		if err != nil {
			return fmt.Errorf("failed to marshal metafield: %w", err)
		}
		fmt.Println(string(line))
	}
	return nil
}

func printFormatted(metafields []api.Metafield) {
	items := make([]cmd.MetafieldPrintable, len(metafields))
	for i, mf := range metafields {
		items[i] = cmd.MetafieldPrintable{
			ID:          mf.ID,
			Namespace:   mf.Namespace,
			Key:         mf.Key,
			Description: mf.Description,
			Value:       mf.Value,
			CreatedAt:   mf.CreatedAt,
			UpdatedAt:   mf.UpdatedAt,
		}
	}
	cmd.PrintMetafields(items)
}

func printMetafields(metafields []api.Metafield, jsonl bool) error {
	if jsonl {
		return printJSONL(metafields)
	}
	printFormatted(metafields)
	return nil
}

func productAction(ctx context.Context, c *cli.Command) error {
	if c.NArg() == 0 {
		return errors.New("Product id required")
	}

	ids, skus, err := cmd.ParseIDArgs(c.Args().Slice(), "a product id")
	if err != nil {
		return err
	}

	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	ns := c.String("namespace")
	key := c.String("key")

	var metafields []api.Metafield
	var failures []string

	for _, id := range ids {
		mfs, err := client.ListMetafields(ctx, "products", &api.ListMetafieldsParams{
			OwnerID:   id,
			Namespace: ns,
			Key:       key,
		})
		if err != nil {
			failures = append(failures, fmt.Sprintf("%d: %s", id, err))
			continue
		}
		metafields = append(metafields, mfs...)
	}

	for _, sku := range skus {
		prod, err := client.GetProductBySKU(ctx, sku)
		if err != nil {
			failures = append(failures, fmt.Sprintf("sku:%s: not found", sku))
			continue
		}
		mfs, err := client.ListMetafields(ctx, "products", &api.ListMetafieldsParams{
			OwnerID:   prod.ID,
			Namespace: ns,
			Key:       key,
		})
		if err != nil {
			failures = append(failures, fmt.Sprintf("sku:%s: %s", sku, err))
			continue
		}
		metafields = append(metafields, mfs...)
	}

	if c.Bool("reverse") {
		reverseMetafields(metafields)
	}

	if err := printMetafields(metafields, c.Bool("jsonl")); err != nil {
		return err
	}

	if len(failures) > 0 && len(metafields) == 0 {
		return fmt.Errorf("Cannot retrieve metafield(s): %s", strings.Join(failures, ", "))
	}

	return nil
}

func variantAction(ctx context.Context, c *cli.Command) error {
	if c.NArg() == 0 {
		return errors.New("Variant id required")
	}

	ids, skus, err := cmd.ParseIDArgs(c.Args().Slice(), "a variant id")
	if err != nil {
		return err
	}

	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	ns := c.String("namespace")
	key := c.String("key")

	var metafields []api.Metafield
	var failures []string

	for _, id := range ids {
		mfs, err := client.ListMetafields(ctx, "product_variants", &api.ListMetafieldsParams{
			OwnerID:   id,
			Namespace: ns,
			Key:       key,
		})
		if err != nil {
			failures = append(failures, fmt.Sprintf("%d: %s", id, err))
			continue
		}
		metafields = append(metafields, mfs...)
	}

	for _, sku := range skus {
		prod, err := client.GetProductBySKU(ctx, sku)
		if err != nil {
			failures = append(failures, fmt.Sprintf("sku:%s: not found", sku))
			continue
		}

		var matchedVariant *api.ProductVariant
		for i := range prod.Variants {
			if string(prod.Variants[i].SKU) == sku {
				matchedVariant = &prod.Variants[i]
				break
			}
		}

		if matchedVariant == nil {
			failures = append(failures, fmt.Sprintf("sku:%s: variant not found", sku))
			continue
		}

		mfs, err := client.ListMetafields(ctx, "product_variants", &api.ListMetafieldsParams{
			OwnerID:   matchedVariant.ID,
			Namespace: ns,
			Key:       key,
		})
		if err != nil {
			failures = append(failures, fmt.Sprintf("sku:%s: %s", sku, err))
			continue
		}
		metafields = append(metafields, mfs...)
	}

	if c.Bool("reverse") {
		reverseMetafields(metafields)
	}

	if err := printMetafields(metafields, c.Bool("jsonl")); err != nil {
		return err
	}

	if len(failures) > 0 && len(metafields) == 0 {
		return fmt.Errorf("Cannot retrieve metafield(s): %s", strings.Join(failures, ", "))
	}

	return nil
}

func customerAction(ctx context.Context, c *cli.Command) error {
	if c.NArg() == 0 {
		return errors.New("Customer id required")
	}

	ids, err := parseIDArgsOnly(c.Args().Slice(), "Customer")
	if err != nil {
		return err
	}

	return listResourceMetafields(ctx, c, "customers", ids)
}

func collectionAction(ctx context.Context, c *cli.Command) error {
	if c.NArg() == 0 {
		return errors.New("Collection id required")
	}

	ids, err := parseIDArgsOnly(c.Args().Slice(), "Collection")
	if err != nil {
		return err
	}

	return listResourceMetafields(ctx, c, "categories", ids)
}

func orderAction(ctx context.Context, c *cli.Command) error {
	if c.NArg() == 0 {
		return errors.New("Order id required")
	}

	ids, err := parseIDArgsOnly(c.Args().Slice(), "Order")
	if err != nil {
		return err
	}

	return listResourceMetafields(ctx, c, "orders", ids)
}

func pageAction(ctx context.Context, c *cli.Command) error {
	if c.NArg() == 0 {
		return errors.New("Page id required")
	}

	ids, err := parseIDArgsOnly(c.Args().Slice(), "Page")
	if err != nil {
		return err
	}

	return listResourceMetafields(ctx, c, "pages", ids)
}

func listResourceMetafields(ctx context.Context, c *cli.Command, ownerResource string, ids []int64) error {
	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	ns := c.String("namespace")
	key := c.String("key")

	var metafields []api.Metafield
	var failures []string

	for _, id := range ids {
		mfs, err := client.ListMetafields(ctx, ownerResource, &api.ListMetafieldsParams{
			OwnerID:   id,
			Namespace: ns,
			Key:       key,
		})
		if err != nil {
			failures = append(failures, fmt.Sprintf("%d: %s", id, err))
			continue
		}
		metafields = append(metafields, mfs...)
	}

	if c.Bool("reverse") {
		reverseMetafields(metafields)
	}

	if err := printMetafields(metafields, c.Bool("jsonl")); err != nil {
		return err
	}

	if len(failures) > 0 && len(metafields) == 0 {
		return fmt.Errorf("Cannot retrieve metafield(s): %s", strings.Join(failures, ", "))
	}

	return nil
}

func parseMetafieldID(arg string) (int64, error) {
	arg = strings.TrimSpace(arg)
	if idx := strings.Index(arg, "@"); idx != -1 {
		arg = strings.TrimSpace(arg[:idx])
	}
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid metafield ID %q: must be numeric ID", arg)
	}
	return id, nil
}

func deleteAction(ctx context.Context, c *cli.Command) error {
	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	var inputs []string
	if c.NArg() > 0 {
		inputs = c.Args().Slice()
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				inputs = append(inputs, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading from stdin: %w", err)
		}
	}

	for _, input := range inputs {
		id, err := parseMetafieldID(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting metafield %s: %s\n", input, err)
			continue
		}

		err = client.DeleteMetafield(ctx, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting metafield %d: %s\n", id, err)
			continue
		}

		fmt.Printf("Deleted %d\n", id)
	}

	return nil
}

func init() {
	metafieldFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "key",
			Aliases: []string{"k"},
			Usage:   "Find metafields with the given key",
		},
		&cli.StringFlag{
			Name:    "namespace",
			Aliases: []string{"n"},
			Usage:   "Find metafields with the given namespace",
		},
		&cli.BoolFlag{
			Name:    "jsonl",
			Aliases: []string{"j"},
			Usage:   "Output the metafields in JSONL format",
		},
		&cli.BoolFlag{
			Name:    "reverse",
			Aliases: []string{"r"},
			Usage:   "Reverse the order of the results",
		},
	}

	Cmd = cli.Command{
		Name:  "metafields",
		Usage: "Metafield utilities",
		Commands: []*cli.Command{
			{
				Name:        "delete",
				Usage:       "Delete one or more metafields",
				ArgsUsage:   "[ID [ID ...]]",
				Description: "If IDs are not given they're read from stdin one per line",
				Flags:       append(cmd.Flags, cmd.APIVersionFlag),
				Action:      deleteAction,
			},
			{
				Name:      "product",
				Usage:     "List metafields for the given product(s)",
				ArgsUsage: "[ID|sku:VALUE [ID|sku:VALUE ...]]",
				Flags:     append(append(cmd.Flags, metafieldFlags...), cmd.APIVersionFlag),
				Action:    productAction,
			},
			{
				Name:      "variant",
				Usage:     "List metafields for the given variant(s)",
				ArgsUsage: "[ID|sku:VALUE [ID|sku:VALUE ...]]",
				Flags:     append(append(cmd.Flags, metafieldFlags...), cmd.APIVersionFlag),
				Action:    variantAction,
			},
			{
				Name:      "customer",
				Usage:     "List metafields for the given customer",
				ArgsUsage: "[ID [ID ...]]",
				Flags:     append(append(cmd.Flags, metafieldFlags...), cmd.APIVersionFlag),
				Action:    customerAction,
			},
			{
				Name:      "collection",
				Usage:     "List metafields for the given collection(s)",
				ArgsUsage: "[ID [ID ...]]",
				Flags:     append(append(cmd.Flags, metafieldFlags...), cmd.APIVersionFlag),
				Action:    collectionAction,
			},
			{
				Name:      "orders",
				Usage:     "List metafields for the given order(s)",
				ArgsUsage: "[ID [ID ...]]",
				Flags:     append(append(cmd.Flags, metafieldFlags...), cmd.APIVersionFlag),
				Action:    orderAction,
			},
			{
				Name:      "page",
				Usage:     "List metafields for the given page(s)",
				ArgsUsage: "[ID [ID ...]]",
				Flags:     append(append(cmd.Flags, metafieldFlags...), cmd.APIVersionFlag),
				Action:    pageAction,
			},
		},
	}
}
