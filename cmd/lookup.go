package cmd

import (
	"errors"
	"time"

	"github.com/majd/ipatool/v2/pkg/appstore"
	"github.com/spf13/cobra"
)

// nolint:wrapcheck
func lookupCmd() *cobra.Command {
	var (
		appID            int64
		bundleID         string
		country          string
		platformValue    string
		includeVersionID bool
		bulk             bool
	)

	cmd := &cobra.Command{
		Use:   "lookup",
		Short: "Look up an App Store app without signing in",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if appID <= 0 && bundleID == "" {
				return errors.New("either a positive app ID or a bundle identifier must be specified")
			}

			if country == "" {
				return errors.New("country code must not be empty")
			}

			platform, err := appstore.ParsePlatform(platformValue)
			if err != nil {
				return err
			}

			if bulk {
				return lookupBulk(cmd, appstore.LookupInput{
					AppID: appID, BundleID: bundleID, Platform: platform,
					IncludeVersionID: includeVersionID,
				}, appstore.Countries(), 250*time.Millisecond)
			}

			output, err := dependencies.AppStore.Lookup(appstore.LookupInput{
				AppID: appID, BundleID: bundleID, Country: country,
				Platform: platform, IncludeVersionID: includeVersionID,
			})
			if err != nil {
				return err
			}

			event := dependencies.Logger.Log().Object("app", output.App)

			if includeVersionID {
				event.Str("externalVersionID", output.ExternalVersionID)
			}

			event.Bool("success", true).Send()

			return nil
		},
	}

	cmd.Flags().BoolVar(&bulk, "bulk", false, "Check availability in every supported country")
	cmd.Flags().Int64VarP(&appID, "app-id", "i", 0, "ID of the target app")
	cmd.Flags().StringVarP(&bundleID, "bundle-identifier", "b", "", "Bundle identifier of the target app (overrides the app ID)")
	cmd.Flags().StringVarP(&country, "country", "c", "US", "App Store country code")
	cmd.Flags().BoolVar(&includeVersionID, "get-version-id", false, "Include the current external version ID")
	cmd.Flags().StringVar(&platformValue, "platform", "", "Platform to look up: iphone (iOS), ipad (iPadOS), appletv (tvOS), watchos, visionos, or macos")

	cmd.MarkFlagsMutuallyExclusive("bulk", "country")

	return cmd
}

// lookupBulk spaces requests to avoid sending a burst to the public catalog.
// nolint:wrapcheck
func lookupBulk(cmd *cobra.Command, input appstore.LookupInput, countries []string, interval time.Duration) error {
	for index, country := range countries {
		if index > 0 {
			timer := time.NewTimer(interval)
			select {
			case <-cmd.Context().Done():
				timer.Stop()

				return cmd.Context().Err()
			case <-timer.C:
			}
		}

		if err := cmd.Context().Err(); err != nil {
			return err
		}

		input.Country = country
		output, err := dependencies.AppStore.Lookup(input)
		event := dependencies.Logger.Log().Str("country", country)

		switch {
		case errors.Is(err, appstore.ErrAppNotFound):
			event.Str("availability", "unavailable")
		case err != nil:
			event.Str("availability", "unknown").Err(err)
		default:
			event.Str("availability", "available").Object("app", output.App)
			if input.IncludeVersionID {
				event.Str("externalVersionID", output.ExternalVersionID)
			}
		}

		event.Send()
	}

	return nil
}
