package locations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/cheynewallace/tabby"
	"github.com/urfave/cli/v3"

	"github.com/ScreenStaring/tiendanube-dev-tools/api"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd"
)

var Cmd cli.Command

// formatAddress returns a human-readable one-line string for a location address.
func formatAddress(addr *api.LocationAddress) string {
	if addr == nil {
		return ""
	}

	var streetLine []string
	if addr.Street != "" {
		streetLine = append(streetLine, addr.Street)
	}
	if addr.Number != "" {
		streetLine = append(streetLine, addr.Number)
	}
	if addr.Floor != "" {
		streetLine = append(streetLine, addr.Floor)
	}

	var parts []string
	if len(streetLine) > 0 {
		parts = append(parts, strings.Join(streetLine, " "))
	} else if addr.Reference != "" {
		parts = append(parts, addr.Reference)
	}
	if addr.Locality != "" {
		parts = append(parts, addr.Locality)
	}
	if addr.City != "" {
		parts = append(parts, addr.City)
	}
	if addr.Zipcode != "" {
		parts = append(parts, addr.Zipcode)
	}

	return strings.Join(parts, ", ")
}

func listLocations(ctx context.Context, c *cli.Command) error {
	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	var locations []api.Location

	if c.NArg() > 0 {
		for _, rawID := range c.Args().Slice() {
			id := strings.TrimSpace(rawID)
			loc, err := client.GetLocation(ctx, id)
			if err != nil {
				var apiErr *api.APIError
				if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
					return fmt.Errorf("Cannot get location %s: no location found with that id", rawID)
				}
				return fmt.Errorf("Cannot get location %s: %s", rawID, err)
			}
			if loc == nil {
				return fmt.Errorf("Cannot get location %s: no location found with that id", rawID)
			}
			locations = append(locations, *loc)
		}
	} else {
		locations, err = client.ListLocations(ctx)
		if err != nil {
			return fmt.Errorf("Cannot list locations: %s", err)
		}
	}

	if len(locations) == 0 {
		if !c.Bool("jsonl") {
			fmt.Println("No locations")
		}
		return nil
	}

	if c.Bool("jsonl") {
		if err := printLocationsJSONL(locations); err != nil {
			return err
		}
	} else {
		printLocations(locations)
	}

	return nil
}

func printLocationsJSONL(locations []api.Location) error {
	for _, loc := range locations {
		line, err := json.Marshal(loc)
		if err != nil {
			return fmt.Errorf("failed to marshal location: %w", err)
		}

		fmt.Println(string(line))
	}
	return nil
}

func printLocations(locations []api.Location) {
	t := tabby.New()
	for _, loc := range locations {
		t.AddLine("ID", loc.ID.String())
		t.AddLine("Name", loc.DisplayName())
		t.AddLine("Default", loc.IsDefault)
		if loc.Priority != 0 {
			t.AddLine("Priority", loc.Priority)
		}
		t.AddLine("Address", formatAddress(loc.Address))
		t.AddLine("Created At", loc.CreatedAt)
		t.AddLine("Updated At", loc.UpdatedAt)
		t.Print()

		cmd.PrintSeparator()
	}
}

func init() {
	apiVersionFlag := cmd.APIVersionFlag

	listFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "jsonl",
			Aliases: []string{"j"},
			Usage:   "Output the locations in JSONL format",
		},
	}

	Cmd = cli.Command{
		Name:  "locations",
		Usage: "Do things with locations",
		Commands: []*cli.Command{
			{
				Name:      "ls",
				Usage:     "List the shop's locations or the locations given by the specified IDs",
				ArgsUsage: "[ID [ID ...]]",
				Flags:     append(cmd.Flags, append(listFlags, apiVersionFlag)...),
				Action:    listLocations,
			},
		},
	}
}
