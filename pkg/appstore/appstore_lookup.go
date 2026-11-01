package appstore

import (
	"errors"
	"fmt"
	gohttp "net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/majd/ipatool/v2/pkg/http"
)

var ErrAppNotFound = errors.New("app not found")

type LookupInput struct {
	Account  Account
	AppID    int64
	BundleID string
	Platform Platform
	// Country overrides the account storefront and allows lookup without an account.
	Country          string
	IncludeVersionID bool
}

type LookupOutput struct {
	App               App
	ExternalVersionID string
}

func (t *appstore) Lookup(input LookupInput) (LookupOutput, error) {
	countryCode := strings.ToUpper(strings.TrimSpace(input.Country))

	var err error

	if countryCode == "" {
		countryCode, err = countryCodeFromStoreFront(input.Account.StoreFront)
		if err != nil {
			return LookupOutput{}, fmt.Errorf("failed to resolve the country code: %w", err)
		}
	} else {
		storefront, ok := storeFronts[countryCode]
		if !ok {
			return LookupOutput{}, fmt.Errorf("unsupported country code %q", input.Country)
		}

		input.Account.StoreFront = storefront
	}

	var request http.Request
	if input.BundleID == "" && input.AppID > 0 {
		request, err = t.lookupIDsRequest([]string{strconv.FormatInt(input.AppID, 10)}, countryCode, input.Platform)
	} else {
		request, err = t.lookupRequest(input.BundleID, countryCode, input.Platform)
	}

	if err != nil {
		return LookupOutput{}, fmt.Errorf("failed to create lookup request: %w", err)
	}

	res, err := t.searchClient.Send(request)
	if err != nil {
		return LookupOutput{}, fmt.Errorf("request failed: %w", err)
	}

	if res.StatusCode != gohttp.StatusOK {
		return LookupOutput{}, NewErrorWithMetadata(errors.New("invalid response"), res)
	}

	if len(res.Data.Results) == 0 {
		return LookupOutput{}, ErrAppNotFound
	}

	if input.Platform == PlatformWatchOS && !slices.Contains(res.Data.Results[0].Platforms, PlatformWatchOS) {
		return LookupOutput{}, ErrAppNotFound
	}

	output := LookupOutput{App: res.Data.Results[0]}

	if input.IncludeVersionID {
		platform := input.Platform

		if platform == "" {
			platform = PlatformIPhone
		}

		if platform == PlatformMacOS {
			output.ExternalVersionID, err = t.lookupLatestMacOSExternalVersionID(input.Account, output.App)
		} else {
			output.ExternalVersionID, err = t.lookupLatestExternalVersionID(input.Account, output.App, platform)
		}

		if err != nil {
			return LookupOutput{}, fmt.Errorf("failed to look up current external version ID: %w", err)
		}
	}

	return output, nil
}

func (t *appstore) lookupRequest(bundleID, countryCode string, platform Platform) (http.Request, error) {
	url, err := t.lookupURL(bundleID, countryCode, platform)
	if err != nil {
		return http.Request{}, err
	}

	return http.Request{
		URL:            url,
		Method:         http.MethodGET,
		ResponseFormat: http.ResponseFormatJSON,
	}, nil
}

func (t *appstore) lookupURL(bundleID, countryCode string, platform Platform) (string, error) {
	entity, err := platform.lookupEntity()
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("entity", entity)
	params.Add("limit", "1")
	params.Add("media", "software")
	params.Add("bundleId", bundleID)
	params.Add("country", countryCode)

	return fmt.Sprintf("https://%s%s?%s", iTunesAPIDomain, iTunesAPIPathLookup, params.Encode()), nil
}
