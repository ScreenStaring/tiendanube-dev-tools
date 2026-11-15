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

	Cmd = cli.Command{
		Name:  "scripts",
		Usage: "Script utilities",
		Commands: []*cli.Command{
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
