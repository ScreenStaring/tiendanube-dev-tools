package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/cheynewallace/tabby"
	"github.com/urfave/cli/v3"

	"github.com/ScreenStaring/tiendanube-dev-tools/api"
)

var Flags []cli.Flag
var accessTokenCommand = regexp.MustCompile(`\A\s*<\s*(.+)\z`)

// APIVersionFlag is the shared --api-version flag. Its Destination writes the
// process-wide api.ActiveAPIVersion.
var APIVersionFlag = &cli.StringFlag{
	Name:        "api-version",
	Usage:       "API version to use; default is 2025-03",
	Value:       api.DefaultAPIVersion,
	Destination: &api.ActiveAPIVersion,
}

func NewClient(c *cli.Command) (*api.Client, error) {
	store := c.String("store")
	if store == "" {
		return nil, fmt.Errorf("Store ID is required: specify with --store, TNDT_STORE_ID, or TIENDANUBE_STORE_ID")
	}

	rawToken := c.String("access-token")
	token := LookupAccessToken(store, rawToken)
	if token == "" {
		return nil, fmt.Errorf("Access token is required: specify with --access-token, TNDT_ACCESS_TOKEN, or TIENDANUBE_ACCESS_TOKEN")
	}

	opts := []api.ClientOption{
		api.WithVerbose(c.Bool("verbose")),
	}

	if c.IsSet("api-version") {
		opts = append(opts, api.WithAPIVersion(c.String("api-version")))
	}

	if ua := c.String("user-agent"); ua != "" {
		opts = append(opts, api.WithUserAgent(ua))
	}

	return api.NewClient(store, token, opts...), nil
}

func ParseIntAt(c *cli.Command, pos int) (int64, error) {
	return strconv.ParseInt(c.Args().Get(pos), 10, 64)
}

func PrintSeparator() {
	fmt.Printf("%s\n", strings.Repeat("-", 20))
}

// ParseIDArgs splits args into numeric IDs and 'sku:VALUE' arguments. idDesc
// names the resource for the error message, e.g. "a product id".
func ParseIDArgs(args []string, idDesc string) ([]int64, []string, error) {
	var ids []int64
	var skus []string

	for _, arg := range args {
		if strings.HasPrefix(strings.ToLower(arg), "sku:") {
			sku := arg[4:]
			if len(sku) == 0 {
				return nil, nil, fmt.Errorf("SKU value missing after 'sku:'")
			}
			skus = append(skus, sku)
			continue
		}

		id, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("Argument '%s' invalid: must be %s or 'sku:VALUE'", arg, idDesc)
		}
		ids = append(ids, id)
	}

	return ids, skus, nil
}

// NormalizeField returns a --fields name in the form used to match it against
// a field or column name: lowercase with spaces and underscores removed.
func NormalizeField(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "")
	return strings.ReplaceAll(s, "_", "")
}

// MetafieldPrintable is the shape metafields convert into so print logic lives once.
type MetafieldPrintable struct {
	ID          interface{}
	Namespace   interface{}
	Key         interface{}
	Description interface{}
	Value       interface{}
	Type        interface{}
	CreatedAt   interface{}
	UpdatedAt   interface{}
}

// PrintMetafields outputs formatted metafields separated by dashed lines.
func PrintMetafields(items []MetafieldPrintable) {
	t := tabby.New()
	for _, mf := range items {
		if mf.ID != nil {
			t.AddLine("Id", mf.ID)
		}
		t.AddLine("Namespace", mf.Namespace)
		t.AddLine("Key", mf.Key)
		t.AddLine("Description", mf.Description)
		t.AddLine("Value", mf.Value)
		if mf.Type != nil && mf.Type != "" {
			t.AddLine("Type", mf.Type)
		}
		t.AddLine("Created", mf.CreatedAt)
		t.AddLine("Updated", mf.UpdatedAt)
		t.Print()
		PrintSeparator()
	}
}

func LookupAccessToken(store, token string) string {
	match := accessTokenCommand.FindStringSubmatch(token)
	if len(match) == 0 {
		return token
	}

	fields := strings.Fields(match[1])
	if len(fields) == 0 {
		return token
	}

	cmdArgs := append(fields[1:], store)
	out, err := exec.Command(fields[0], cmdArgs...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "access token command failed: %s\n", err)
		os.Exit(2)
	}

	return strings.TrimSuffix(string(out), "\n")
}

// BaseFlags returns a fresh slice of global flags for the CLI.
func BaseFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "verbose",
			Usage: "Output Tiendanube API request/response",
		},
		&cli.StringFlag{
			Name:     "store",
			Usage:    "Tiendanube store ID to perform command against",
			Required: true,
			Sources: cli.EnvVars(
				"TNDT_STORE_ID",
				"TIENDANUBE_STORE_ID",
			),
		},
		&cli.StringFlag{
			Name:  "access-token",
			Usage: "Tiendanube access token for store",
			Sources: cli.EnvVars(
				"TNDT_ACCESS_TOKEN",
				"TIENDANUBE_ACCESS_TOKEN",
			),
		},
		&cli.StringFlag{
			Name:    "user-agent",
			Aliases: []string{"ua"},
			Usage:   "Tiendanube User-Agent header for API requests",
			Sources: cli.EnvVars(
				"TNDT_USER_AGENT",
			),
		},
	}
}

func init() {
	Flags = BaseFlags()
}
