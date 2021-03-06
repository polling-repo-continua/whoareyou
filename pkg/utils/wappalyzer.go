package utils

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/ameenmaali/whoareyou/pkg/config"
	"github.com/ameenmaali/whoareyou/pkg/matcher"
)

const WAPPALYZER_SOURCE_URL = "https://raw.githubusercontent.com/AliasIO/wappalyzer/master/src/apps.json"

func FetchWappalyzerData(conf *config.Config) (map[string]matcher.AppMatch, error) {
	wappalyzerData := map[string]matcher.AppMatch{}
	resp, err := SendRequest(WAPPALYZER_SOURCE_URL, conf)
	if err != nil {
		return wappalyzerData, err
	}

	responseBody := make(map[string]map[string]map[string]interface{})

	err = json.Unmarshal(resp.Body, &responseBody)

	for _, value := range responseBody {
		for app, apps := range value {
			match := matcher.Matcher{
				Cookies:         nil,
				Icon:            "",
				Headers:         nil,
				ResponseContent: nil,
				Script:          nil,
				JavaScript:      nil,
				Meta:            nil,
			}

			wapp := matcher.AppMatch{
				Name:    app,
				Website: "",
				Matches: &match,
			}

			if apps["website"] != nil {
				wapp.Website = apps["website"].(string)
			}

			if apps["icon"] != nil {
				match.Icon = apps["icon"].(string)
			}

			if apps["html"] != nil {
				if err := stringOrSliceHandler(apps["html"], &match.ResponseContent); err != nil {
					if conf.DebugMode {
						conf.Utils.PrintRed(os.Stderr, "error parsing wappalyzer html data", err)
					}
				}
			}

			if apps["headers"] != nil {
				if err := mapHandler(apps["headers"], &match.Headers); err != nil {
					if conf.DebugMode {
						conf.Utils.PrintRed(os.Stderr, "error parsing wappalyzer header data", err)
					}
				}
			}

			if apps["cookies"] != nil {
				if err := mapHandler(apps["cookies"], &match.Cookies); err != nil {
					if conf.DebugMode {
						conf.Utils.PrintRed(os.Stderr, "error parsing wappalyzer cookie data", err)
					}
				}
			}

			if apps["script"] != nil {
				if err := stringOrSliceHandler(apps["script"], &match.Script); err != nil {
					if conf.DebugMode {
						conf.Utils.PrintRed(os.Stderr, "error parsing wappalyzer script data", err)
					}
				}
			}

			if apps["js"] != nil {
				if err := mapHandler(apps["js"], &match.JavaScript); err != nil {
					if conf.DebugMode {
						conf.Utils.PrintRed(os.Stderr, "error parsing wappalyzer js data", err)
					}
				}
			}

			if apps["meta"] != nil {
				if err := mapHandler(apps["meta"], &match.Meta); err != nil {
					if conf.DebugMode {
						conf.Utils.PrintRed(os.Stderr, "error parsing wappalyzer meta data", err)
					}
				}
			}
			wappalyzerData[strings.ToLower(wapp.Name)] = wapp
		}
	}

	return wappalyzerData, nil
}

func stringOrSliceHandler(value interface{}, matchResult *[]*regexp.Regexp) error {
	errorCount := 0
	matchError := ""

	var matches []*regexp.Regexp

	re, err := stringToRegex(value)
	if err != nil {
		errorCount += 1
		matchError += err.Error() + "\n"
	}
	matches = append(matches, re)

	matches, err = sliceToRegexSlice(value, matches)
	if err != nil {
		errorCount += 1
		matchError += err.Error() + "\n"
	}

	// If both conversions fail, mark as an error and move on
	if errorCount >= 2 {
		return errors.New(matchError)
	} else {
		*matchResult = append(matches)
	}
	return nil
}

func mapHandler(value interface{}, matchResult *map[string]*regexp.Regexp) error {
	headerMap, err := mapToRegexMap(value)
	if err != nil {
		return err
	} else {
		*matchResult = headerMap
	}
	return nil
}
