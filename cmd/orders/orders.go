package orders

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cheynewallace/tabby"
	"github.com/urfave/cli/v3"

	"github.com/ScreenStaring/tiendanube-dev-tools/api"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd"
)

var Cmd cli.Command

// OrderFilter holds the criteria used to find orders.
type OrderFilter struct {
	IDs    []int64
	SKUs   []string
	Names  []string
	Status string
}

func parseOrderArgs(c *cli.Command) (OrderFilter, error) {
	var filter OrderFilter
	var rest []string

	for _, arg := range c.Args().Slice() {
		if strings.HasPrefix(strings.ToLower(arg), "name:") {
			name := arg[5:]
			if len(name) == 0 {
				return filter, fmt.Errorf("Name value missing after 'name:'")
			}
			filter.Names = append(filter.Names, name)
			continue
		}

		rest = append(rest, arg)
	}

	ids, skus, err := cmd.ParseIDArgs(rest, "an order id")
	if err != nil {
		return filter, err
	}

	filter.IDs = ids
	filter.SKUs = skus

	return filter, nil
}

func truncate(val string) string {
	max := 25

	if len(val) < max {
		return val
	}

	cut := val[0:max]

	if len(cut) < len(val) {
		cut += "…"
	}

	return cut
}

func printLineItems(lines []api.OrderProduct, defaultStatus string) {
	x := tabby.New()
	x.AddHeader("ID", "Product ID", "Variant ID", "SKU", "Title", "Quantity", "Status")

	for _, line := range lines {
		status := line.Status
		if status == "" {
			status = defaultStatus
		}
		x.AddLine(
			line.ID,
			line.ProductID,
			line.VariantID,
			line.SKU.String(),
			truncate(line.Name),
			line.Quantity,
			status,
		)
	}

	x.Print()
}

func printOrders(orders []api.Order) {
	for _, order := range orders {
		t := tabby.New()
		t.AddLine("Id", order.ID)
		t.AddLine("Name", order.DisplayName())
		t.AddLine("Created At", order.CreatedAt)
		t.AddLine("Updated At", order.UpdatedAt)
		t.AddLine("Canceled At", order.CancelledAt.String())
		t.AddLine("Closed At", order.ClosedAt.String())
		t.AddLine("Financial Status", order.DisplayFinancialStatus())
		t.AddLine("Fulfillment Status", order.DisplayFulfillmentStatus())

		note := order.Note
		if len(order.Note) > 0 {
			note = fmt.Sprintf("%q", order.Note)
		}

		t.AddLine("Note", note)
		t.Print()

		fmt.Println("Line Items")
		printLineItems(order.Products, order.DisplayFulfillmentStatus())
		fmt.Print("\n")

		cmd.PrintSeparator()
	}
}

func printJSONL(orders []api.Order) error {
	for _, o := range orders {
		line, err := json.Marshal(o)
		if err != nil {
			return fmt.Errorf("failed to marshal order: %w", err)
		}
		fmt.Println(string(line))
	}
	return nil
}

func listOrdersAction(ctx context.Context, c *cli.Command) error {
	filter, err := parseOrderArgs(c)
	if err != nil {
		return err
	}

	status := "open"
	if c.IsSet("status") && c.String("status") != "" {
		status = c.String("status")
	}
	filter.Status = status

	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	var orders []api.Order
	limit := int(c.Int64("limit"))
	if limit <= 0 {
		limit = 10
	}

	if len(filter.IDs) > 0 || len(filter.SKUs) > 0 || len(filter.Names) > 0 {
		var failures []string
		seenIDs := make(map[int64]bool)

		// 1. Fetch by ID
		for _, id := range filter.IDs {
			order, err := client.GetOrder(ctx, id)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%d: %s", id, err))
				continue
			}
			if !seenIDs[order.ID] {
				seenIDs[order.ID] = true
				orders = append(orders, *order)
			}
		}

		// 2. Fetch by Name
		for _, name := range filter.Names {
			trimmedName := strings.TrimPrefix(name, "#")
			params := &api.ListOrdersParams{
				Q:      trimmedName,
				Status: "any",
			}
			list, err := client.ListOrders(ctx, params)
			if err != nil {
				failures = append(failures, fmt.Sprintf("name:%s: %s", name, err))
				continue
			}

			var matched []api.Order
			for _, o := range list {
				if strings.EqualFold(o.Name, name) ||
					strings.EqualFold(o.DisplayName(), name) ||
					strings.EqualFold(o.DisplayName(), "#"+trimmedName) ||
					fmt.Sprintf("%d", o.Number) == trimmedName ||
					strings.EqualFold(o.ContactName, name) {
					matched = append(matched, o)
				}
			}

			if len(matched) == 0 {
				failures = append(failures, fmt.Sprintf("name:%s: no matching order found", name))
				continue
			}

			for _, o := range matched {
				if !seenIDs[o.ID] {
					seenIDs[o.ID] = true
					orders = append(orders, o)
				}
			}
		}

		// 3. Fetch by SKU
		if len(filter.SKUs) > 0 {
			fetchLimit := 50
			if limit > fetchLimit {
				fetchLimit = limit
			}
			params := &api.ListOrdersParams{
				PerPage: fetchLimit,
			}
			if filter.Status != "any" {
				params.Status = filter.Status
			}
			list, err := client.ListOrders(ctx, params)
			if err != nil {
				for _, sku := range filter.SKUs {
					failures = append(failures, fmt.Sprintf("sku:%s: %s", sku, err))
				}
			} else {
				for _, sku := range filter.SKUs {
					var foundForSKU []api.Order
					for _, o := range list {
						for _, p := range o.Products {
							if strings.EqualFold(p.SKU.String(), sku) {
								foundForSKU = append(foundForSKU, o)
								break
							}
						}
					}
					if len(foundForSKU) == 0 {
						failures = append(failures, fmt.Sprintf("sku:%s: no matching order found", sku))
						continue
					}
					for _, o := range foundForSKU {
						if !seenIDs[o.ID] {
							seenIDs[o.ID] = true
							orders = append(orders, o)
						}
					}
				}
			}
		}

		if len(failures) > 0 && len(orders) == 0 {
			return fmt.Errorf("Cannot retrieve order(s): %s", strings.Join(failures, ", "))
		}
	} else {
		params := &api.ListOrdersParams{
			PerPage: limit,
		}
		if filter.Status != "any" {
			params.Status = filter.Status
		}
		ords, err := client.ListOrders(ctx, params)
		if err != nil {
			return fmt.Errorf("Cannot list orders: %s", err)
		}
		orders = ords
	}

	if len(orders) == 0 {
		if !c.Bool("jsonl") {
			fmt.Println("No orders")
		}
		return nil
	}

	if c.Bool("jsonl") {
		return printJSONL(orders)
	}

	printOrders(orders)
	return nil
}

func init() {
	apiVersionFlag := cmd.APIVersionFlag

	ordersFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "status",
			Aliases: []string{"s"},
			Usage:   "Order status to filter (open, closed, cancelled, any), defaults to 'open'",
		},
		&cli.Int64Flag{
			Name:    "limit",
			Aliases: []string{"l"},
			Usage:   "Maximum number of orders to return, must be <= 250",
			Value:   10,
		},
		&cli.BoolFlag{
			Name:    "jsonl",
			Aliases: []string{"j"},
			Usage:   "Output the orders in JSONL format",
		},
		apiVersionFlag,
	}

	Cmd = cli.Command{
		Name:    "orders",
		Aliases: []string{"order"},
		Usage:   "Information about orders",
		Commands: []*cli.Command{
			{
				Name:      "ls",
				Aliases:   []string{"list"},
				Usage:     "List the shop's orders or the orders matching the given IDs, 'sku:VALUE' and/or 'name:VALUE' arguments",
				ArgsUsage: "[ID|sku:VALUE|name:VALUE [ID|sku:VALUE|name:VALUE ...]]",
				Flags:     append(cmd.Flags, ordersFlags...),
				Action:    listOrdersAction,
			},
		},
	}
}
