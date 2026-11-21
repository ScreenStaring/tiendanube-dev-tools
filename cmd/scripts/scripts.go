package scripts

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cheynewallace/tabby"
	"github.com/urfave/cli/v3"

	"github.com/ScreenStaring/tiendanube-dev-tools/api"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd"
)

var Cmd cli.Command

// Match https://foo.com or //foo.com
var scriptURL = regexp.MustCompile(`(?i)\A(?:https:)?//[\da-z]`)

func printFormatted(scripts []api.Script) {
	for _, s := range scripts {
		t := tabby.New()
		t.AddLine("Id", s.ID)
		if s.Name != "" {
			t.AddLine("Name", s.Name)
		}
		t.AddLine("Src", s.GetSrc())
		t.AddLine("Event", s.Event)
		if s.Location != "" {
			t.AddLine("Location", s.Location)
		}
		if s.Status != "" {
			t.AddLine("Status", s.Status)
		}
		t.AddLine("Created", s.CreatedAt)
		t.AddLine("Updated", s.UpdatedAt)
		t.Print()

		cmd.PrintSeparator()
	}
}

func printJSONL(scripts []api.Script) error {
	for _, s := range scripts {
		line, err := json.Marshal(s)
		if err != nil {
			return fmt.Errorf("failed to marshal script: %w", err)
		}
		fmt.Println(string(line))
	}
	return nil
}

// parseQueryParams converts a list of "key=value" arguments into a JSON object string
// (e.g. {"key1":"val1","key2":"val2"}).
// Returns an empty string if no arguments or only empty entries are provided.
func parseQueryParams(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}

	params := make(map[string]string)
	for _, raw := range args {
		part := strings.TrimSpace(raw)
		if part == "" {
			continue
		}
		idx := strings.Index(part, "=")
		if idx == -1 {
			return "", fmt.Errorf("Query parameter '%s' is invalid: must be in key=value format", part)
		}
		key := strings.TrimSpace(part[:idx])
		if key == "" {
			return "", fmt.Errorf("Query parameter '%s' is invalid: key cannot be empty", part)
		}
		val := strings.TrimSpace(part[idx+1:])
		params[key] = val
	}

	if len(params) == 0 {
		return "", nil
	}

	b, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("failed to encode query parameters to JSON: %w", err)
	}
	return string(b), nil
}

func createAction(ctx context.Context, c *cli.Command) error {
	if c.NArg() == 0 {
		return fmt.Errorf("Script id is required")
	}
	if c.NArg() > 1 {
		return fmt.Errorf("Too many arguments: expected a single script id")
	}

	scriptID, err := strconv.ParseInt(c.Args().First(), 10, 64)
	if err != nil {
		return fmt.Errorf("Script id '%s' is invalid: must be an int", c.Args().First())
	}

	queryParams, err := parseQueryParams(c.StringSlice("query"))
	if err != nil {
		return err
	}

	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	created, err := client.CreateScript(ctx, &api.CreateScriptRequest{
		ScriptID:    scriptID,
		QueryParams: queryParams,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Script created: %d\n", created.ID)
	return nil
}

func updateAction(ctx context.Context, c *cli.Command) error {
	if c.NArg() == 0 {
		return fmt.Errorf("Script id is required")
	}
	if c.NArg() > 1 {
		return fmt.Errorf("Too many arguments: expected a single script id")
	}

	scriptID, err := strconv.ParseInt(c.Args().First(), 10, 64)
	if err != nil {
		return fmt.Errorf("Script id '%s' is invalid: must be an int", c.Args().First())
	}

	queryParams, err := parseQueryParams(c.StringSlice("query"))
	if err != nil {
		return err
	}

	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	updated, err := client.UpdateScript(ctx, scriptID, &api.UpdateScriptRequest{
		ScriptID:    scriptID,
		QueryParams: queryParams,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Script updated: %d\n", updated.ID)
	return nil
}

func deleteAction(ctx context.Context, c *cli.Command) error {
	if c.NArg() == 0 {
		return fmt.Errorf("Script id is required")
	}
	if c.NArg() > 1 {
		return fmt.Errorf("Too many arguments: expected a single script id")
	}

	scriptID, err := strconv.ParseInt(c.Args().First(), 10, 64)
	if err != nil {
		return fmt.Errorf("Script id '%s' is invalid: must be an int", c.Args().First())
	}

	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	err = client.DeleteScript(ctx, scriptID)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %d\n", scriptID)
	return nil
}

func listScripts(ctx context.Context, c *cli.Command) error {
	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	var scripts []api.Script

	if c.NArg() > 0 {
		var failures []string
		// We may need all scripts to find by URL
		var allScripts []api.Script
		var allFetched bool

		for _, arg := range c.Args().Slice() {
			arg = strings.TrimSpace(arg)
			if scriptURL.MatchString(arg) {
				if !allFetched {
					all, err := client.ListScripts(ctx, nil)
					if err != nil {
						failures = append(failures, fmt.Sprintf("url:%s: %s", arg, err))
						continue
					}
					allScripts = all
					allFetched = true
				}

				var found []api.Script
				for _, s := range allScripts {
					if s.GetSrc() == arg || strings.Contains(s.GetSrc(), arg) {
						found = append(found, s)
					}
				}

				if len(found) == 0 {
					failures = append(failures, fmt.Sprintf("Cannot find script with URL %s", arg))
					continue
				}
				scripts = append(scripts, found...)
			} else {
				id, err := strconv.ParseInt(arg, 10, 64)
				if err != nil {
					return fmt.Errorf("Script id '%s' is invalid: must be an int or URL", arg)
				}

				s, err := client.GetScript(ctx, id)
				if err != nil {
					failures = append(failures, fmt.Sprintf("%d: %s", id, err))
					continue
				}
				scripts = append(scripts, *s)
			}
		}

		if len(failures) > 0 && len(scripts) == 0 {
			return fmt.Errorf("Cannot retrieve script(s): %s", strings.Join(failures, ", "))
		}
	} else {
		scrs, err := client.ListScripts(ctx, nil)
		if err != nil {
			return fmt.Errorf("Cannot list scripts: %s", err)
		}
		scripts = scrs
	}

	if len(scripts) == 0 {
		if !c.Bool("jsonl") {
			fmt.Println("No scripts")
		}
		return nil
	}

	if c.Bool("jsonl") {
		return printJSONL(scripts)
	}

	printFormatted(scripts)
	return nil
}

func init() {
	apiVersionFlag := cmd.APIVersionFlag

	listFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "jsonl",
			Aliases: []string{"j"},
			Usage:   "Output the scripts in JSONL format",
		},
		apiVersionFlag,
	}

	queryFlags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "query",
			Aliases: []string{"q"},
			Usage:   "Script query parameters in key=value format (can be specified multiple times)",
		},
		apiVersionFlag,
	}

	deleteFlags := []cli.Flag{
		apiVersionFlag,
	}

	Cmd = cli.Command{
		Name:  "scripts",
		Usage: "Script utilities",
		Commands: []*cli.Command{
			{
				Name:                      "create",
				Usage:                     "Associate a script with the store",
				ArgsUsage:                 "ID",
				DisableSliceFlagSeparator: true,
				Flags:                     append(cmd.Flags, queryFlags...),
				Action:                    createAction,
			},
			{
				Name:                      "update",
				Usage:                     "Update an existing script association",
				ArgsUsage:                 "ID",
				DisableSliceFlagSeparator: true,
				Flags:                     append(cmd.Flags, queryFlags...),
				Action:                    updateAction,
			},
			{
				Name:      "delete",
				Usage:     "Delete a script by ID",
				ArgsUsage: "ID",
				Flags:     append(cmd.Flags, deleteFlags...),
				Action:    deleteAction,
			},
			{
				Name:      "ls",
				Aliases:   []string{"list"},
				Usage:     "List scripts for the given store or scripts matching the given IDs or URLs",
				ArgsUsage: "[ID|URL [ID|URL ...]]",
				Flags:     append(cmd.Flags, listFlags...),
				Action:    listScripts,
			},
		},
	}
}
