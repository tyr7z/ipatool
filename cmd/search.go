package cmd

import (
	"errors"
	"strings"

	"github.com/majd/ipatool/v2/pkg/appstore"
	"github.com/spf13/cobra"
)

// nolint:wrapcheck
func searchCmd() *cobra.Command {
	var (
		limit         int64
		platformValue string
		country       string
	)

	cmd := &cobra.Command{
		Use:   "search <term>",
		Short: "Search for iOS, iPadOS, tvOS, watchOS, visionOS, and macOS apps available on the App Store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var account appstore.Account
			if cmd.Flags().Changed("country") {
				if strings.TrimSpace(country) == "" {
					return errors.New("country code must not be empty")
				}
			} else {
				infoResult, err := dependencies.AppStore.AccountInfo()
				if err != nil {
					country = "US"
					dependencies.Logger.Verbose().Err(err).Msg("Using US storefront because no account is available")
				} else {
					account = infoResult.Account
				}
			}

			platform, err := appstore.ParsePlatform(platformValue)
			if err != nil {
				return err
			}

			output, err := dependencies.AppStore.Search(appstore.SearchInput{
				Account:  account,
				Country:  country,
				Term:     args[0],
				Limit:    limit,
				Platform: platform,
			})
			if err != nil {
				return err
			}

			dependencies.Logger.Log().
				Int("count", output.Count).
				Array("apps", appstore.Apps(output.Results)).
				Send()

			return nil
		},
	}

	cmd.Flags().StringVarP(&country, "country", "c", "", "App Store country code (defaults to the account storefront, or US without an account)")
	cmd.Flags().Int64VarP(&limit, "limit", "l", 5, "maximum amount of search results to retrieve; visionOS supports up to 12")
	cmd.Flags().StringVar(&platformValue, "platform", "", "Platform to search: iphone (iOS), ipad (iPadOS), appletv (tvOS), watchos, visionos, or macos")

	return cmd
}
