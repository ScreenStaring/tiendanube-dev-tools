package store

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/cheynewallace/tabby"
	"github.com/urfave/cli/v3"

	"github.com/ScreenStaring/tiendanube-dev-tools/api"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd"
)

var Cmd cli.Command

// StoreInfo holds formatted information about the store matching shopify-dev-tools's shop info.
type StoreInfo struct {
	ID                  string
	Name                string
	Description         string
	Type                string
	Email               string
	CustomerEmail       string
	CreatedAt           string
	UpdatedAt           string
	Address             string
	Country             string
	CountryCode         string
	Currency            string
	Domain              string
	OriginalDomain      string
	Domains             string
	Phone               string
	WhatsappPhoneNumber string
	BusinessID          string
	BusinessName        string
	BusinessAddress     string
	CustomerAccounts    string
	PlanName            string
	CurrentTheme        string
	MainLanguage        string
	AdminLanguage       string
	Timezone            string
	IanaTimezone        string
	Logo                string
	Facebook            string
	Instagram           string
	Twitter             string
	Youtube             string
	Tiktok              string
	Pinterest           string
	Blog                string
}

// Some low-budget formatting
func formatField(field string) string {
	re := regexp.MustCompile("([a-z])([A-Z])")
	name := strings.Replace(field, "API", "Api", 1)
	name = re.ReplaceAllString(name, "$1 $2")
	name = strings.Replace(name, " At", "", 1)

	return name
}

func findStore(ctx context.Context, client *api.Client) (*StoreInfo, error) {
	s, err := client.GetStore(ctx)
	if err != nil {
		return nil, err
	}

	domain := s.URLWithProtocol
	if domain == "" && len(s.Domains) > 0 {
		domain = s.Domains[0]
	}
	if domain == "" {
		domain = s.OriginalDomain
	}

	return &StoreInfo{
		ID:                  s.ID.String(),
		Name:                s.Name.String(),
		Description:         s.Description.String(),
		Type:                s.Type.String(),
		Email:               s.Email,
		CustomerEmail:       s.ContactEmail,
		CreatedAt:           s.CreatedAt,
		UpdatedAt:           s.UpdatedAt,
		Address:             s.Address.String(),
		Country:             s.Country,
		CountryCode:         s.Country,
		Currency:            s.MainCurrency,
		Domain:              domain,
		OriginalDomain:      s.OriginalDomain,
		Domains:             strings.Join(s.Domains, ", "),
		Phone:               s.Phone.String(),
		WhatsappPhoneNumber: s.WhatsappPhoneNumber.String(),
		BusinessID:          s.BusinessID.String(),
		BusinessName:        s.BusinessName.String(),
		BusinessAddress:     s.BusinessAddress.String(),
		CustomerAccounts:    s.CustomerAccounts.String(),
		PlanName:            s.PlanName,
		CurrentTheme:        s.CurrentTheme.String(),
		MainLanguage:        s.MainLanguage,
		AdminLanguage:       s.AdminLanguage,
		Timezone:            s.TimezoneID,
		IanaTimezone:        s.TimezoneID,
		Logo:                s.Logo.String(),
		Facebook:            s.Facebook.String(),
		Instagram:           s.Instagram.String(),
		Twitter:             s.Twitter.String(),
		Youtube:             s.Youtube.String(),
		Tiktok:              s.Tiktok.String(),
		Pinterest:           s.Pinterest.String(),
		Blog:                s.Blog.String(),
	}, nil
}

func infoAction(ctx context.Context, c *cli.Command) error {
	client, err := cmd.NewClient(c)
	if err != nil {
		return err
	}

	info, err := findStore(ctx, client)
	if err != nil {
		return fmt.Errorf("Cannot get info for store: %s", err)
	}

	t := tabby.New()
	s := reflect.ValueOf(info).Elem()

	for i := 0; i < s.NumField(); i++ {
		t.AddLine(formatField(s.Type().Field(i).Name), s.Field(i).Interface())
	}

	t.Print()

	return nil
}

func init() {
	apiVersionFlag := cmd.APIVersionFlag

	Cmd = cli.Command{
		Name:  "store",
		Usage: "Information about the given store",
		Commands: []*cli.Command{
			{
				Name:   "info",
				Usage:  "Information about the store",
				Flags:  append(cmd.Flags, apiVersionFlag),
				Action: infoAction,
			},
		},
	}
}
