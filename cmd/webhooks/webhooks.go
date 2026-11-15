package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cheynewallace/tabby"
	"github.com/urfave/cli/v3"

	"github.com/ScreenStaring/tiendanube-dev-tools/api"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd"
)

var Cmd cli.Command

func printFormatted(webhooks []api.Webhook) {
	t := tabby.New()
	for _, webhook := range webhooks {
		t.AddLine("Id", webhook.ID)
		t.AddLine("Address", webhook.URL)
		t.AddLine("Event", webhook.Event)
		t.AddLine("Created", webhook.CreatedAt)
		t.AddLine("Updated", webhook.UpdatedAt)
		t.Print()

		cmd.PrintSeparator()
	}
}

func printJSONL(webhooks []api.Webhook) error {
	for _, webhook := range webhooks {
		line, err := json.Marshal(webhook)
		if err != nil {
			return fmt.Errorf("failed to marshal webhook: %w", err)
		}

		fmt.Println(string(line))
	}
	return nil
}

func getEventArg(c *cli.Command) string {
	return strings.TrimSpace(c.String("event"))
}

func getAddressArg(c *cli.Command) string {
	return strings.TrimSpace(c.String("address"))
}

func createAction(ctx context.Context, c *cli.Command) error {
	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	event := getEventArg(c)
	address := getAddressArg(c)

	if event == "" {
		return fmt.Errorf("event is required")
	}
	if address == "" {
		return fmt.Errorf("address is required")
	}

	if c.Bool("one") {
		existing, err := client.ListWebhooks(ctx, nil)
		if err != nil {
			return err
		}
		for _, w := range existing {
			if w.Event == event && w.URL == address {
				return fmt.Errorf("Webhook already exists for event %s at %s", event, address)
			}
		}
	}

	created, err := client.CreateWebhook(ctx, &api.CreateWebhookRequest{
		Event: event,
		URL:   address,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Webhook created: %d\n", created.ID)
	return nil
}

func deleteAction(ctx context.Context, c *cli.Command) error {
	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	var toDelete []api.Webhook

	if c.Bool("all") {
		all, err := client.ListWebhooks(ctx, nil)
		if err != nil {
			return err
		}
		toDelete = all
	} else {
		if c.Args().Len() == 0 {
			return fmt.Errorf("You must supply a webhook id or event")
		}

		var allWebhooks []api.Webhook
		var fetchedAll bool

		for _, arg := range c.Args().Slice() {
			if id, err := strconv.ParseInt(arg, 10, 64); err == nil {
				toDelete = append(toDelete, api.Webhook{ID: id})
			} else {
				if !fetchedAll {
					var listErr error
					allWebhooks, listErr = client.ListWebhooks(ctx, nil)
					if listErr != nil {
						return fmt.Errorf("Cannot list webhooks for event %s: %s", arg, listErr)
					}
					fetchedAll = true
				}
				for _, w := range allWebhooks {
					if w.Event == arg {
						toDelete = append(toDelete, w)
					}
				}
			}
		}
	}

	if len(toDelete) == 0 {
		return fmt.Errorf("No webhooks found")
	}

	seen := make(map[int64]bool)
	var unique []api.Webhook
	for _, w := range toDelete {
		if !seen[w.ID] {
			seen[w.ID] = true
			unique = append(unique, w)
		}
	}

	for _, w := range unique {
		err := client.DeleteWebhook(ctx, w.ID)
		if err != nil {
			return err
		}
	}

	fmt.Printf("%d webhook(s) deleted\n", len(unique))
	return nil
}

func updateAction(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("You must supply a webhook id to update")
	}

	id, err := strconv.ParseInt(c.Args().Get(0), 10, 64)
	if err != nil {
		return fmt.Errorf("Argument '%s' invalid: must be a webhook id", c.Args().Get(0))
	}

	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	var req api.UpdateWebhookRequest
	if c.IsSet("address") {
		req.URL = getAddressArg(c)
	}
	if c.IsSet("event") {
		req.Event = getEventArg(c)
	}

	if req.URL == "" && req.Event == "" {
		return fmt.Errorf("You must supply at least one option to update")
	}

	_, err = client.UpdateWebhook(ctx, id, &req)
	if err != nil {
		return err
	}

	fmt.Println("Webhook updated")
	return nil
}

func parseWebhookIDArgs(args []string) ([]int64, error) {
	seen := make(map[int64]bool)
	var ids []int64
	for _, arg := range args {
		id, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Argument '%s' invalid: must be a webhook id", arg)
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func listAction(ctx context.Context, c *cli.Command) error {
	ids, err := parseWebhookIDArgs(c.Args().Slice())
	if err != nil {
		return err
	}

	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	address := getAddressArg(c)
	event := getEventArg(c)

	var hooks []api.Webhook
	if len(ids) > 0 {
		var failures []string
		for _, id := range ids {
			hook, err := client.GetWebhook(ctx, id)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%d: %s", id, err))
				continue
			}
			hooks = append(hooks, *hook)
		}
		if len(failures) > 0 && len(hooks) == 0 {
			return fmt.Errorf("Cannot retrieve webhook(s): %s", strings.Join(failures, ", "))
		}
	} else {
		params := &api.ListWebhooksParams{}
		if address != "" {
			params.URL = address
		}
		if event != "" {
			params.Event = event
		}

		all, err := client.ListWebhooks(ctx, params)
		if err != nil {
			return err
		}
		hooks = all
	}

	if address != "" || event != "" {
		var filtered []api.Webhook
		for _, h := range hooks {
			if address != "" && h.URL != address {
				continue
			}
			if event != "" && h.Event != event {
				continue
			}
			filtered = append(filtered, h)
		}
		hooks = filtered
	}

	if c.Bool("jsonl") {
		if err := printJSONL(hooks); err != nil {
			return err
		}
	} else {
		printFormatted(hooks)
	}

	return nil
}

func init() {
	apiVersionFlag := cmd.APIVersionFlag

	createFlags := []cli.Flag{
		&cli.StringFlag{
			Name:     "address",
			Required: true,
			Aliases:  []string{"a", "url"},
			Usage:    "URL endpoint where webhook POST notifications will be delivered",
		},
		&cli.BoolFlag{
			Name:    "one",
			Aliases: []string{"1"},
			Usage:   "Error if a webhook already exists for this event and address",
		},
		&cli.StringFlag{
			Name:     "event",
			Required: true,
			Aliases:  []string{"e", "topic", "t"},
			Usage:    "Event name that will trigger the webhook (e.g. order/created, product/updated)",
		},
		apiVersionFlag,
	}

	updateFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"a", "url"},
			Usage:   "Updated URL endpoint",
		},
		&cli.StringFlag{
			Name:    "event",
			Aliases: []string{"e", "topic", "t"},
			Usage:   "Updated event name",
		},
		apiVersionFlag,
	}

	deleteFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "Delete all webhooks for the store",
		},
		apiVersionFlag,
	}

	listFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"a", "url"},
			Usage:   "Only show webhooks with the given address",
		},
		&cli.BoolFlag{
			Name:    "jsonl",
			Aliases: []string{"j"},
			Usage:   "Output the webhooks in JSONL format",
		},
		&cli.StringFlag{
			Name:    "event",
			Aliases: []string{"e", "topic", "t"},
			Usage:   "Only show webhooks with the given event",
		},
		apiVersionFlag,
	}

	Cmd = cli.Command{
		Name:  "webhooks",
		Usage: "Webhook utilities",
		Commands: []*cli.Command{
			{
				Name:   "create",
				Flags:  append(cmd.Flags, createFlags...),
				Action: createAction,
				Usage:  "Create a webhook for the given shop",
			},
			{
				Name:      "delete",
				ArgsUsage: "[event or webhook ID ...]",
				Flags:     append(cmd.Flags, deleteFlags...),
				Action:    deleteAction,
				Usage:     "Delete the given webhook",
			},
			{
				Name:      "update",
				ArgsUsage: "[webhook ID]",
				Flags:     append(cmd.Flags, updateFlags...),
				Action:    updateAction,
				Usage:     "Update the given webhook",
			},
			{
				Name:      "ls",
				ArgsUsage: "[webhook ID ...]",
				Flags:     append(cmd.Flags, listFlags...),
				Action:    listAction,
				Usage:     "List the shop's webhooks",
			},
		},
	}
}
