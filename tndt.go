package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ScreenStaring/tiendanube-dev-tools/api"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd/locations"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd/metafields"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd/orders"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd/products"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd/scripts"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd/store"
	"github.com/ScreenStaring/tiendanube-dev-tools/cmd/webhooks"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:                   "tndt",
		Usage:                  "Tiendanube / Nuvemshop Development Tools",
		Version:                api.Version,
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			&products.Cmd,
			&locations.Cmd,
			&metafields.Cmd,
			&store.Cmd,
			&webhooks.Cmd,
			&scripts.Cmd,
			&orders.Cmd,
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
