package shell

import (
	"errors"
	"strings"

	"github.com/elastic/beats/libbeat/common/match"
)

var (
	errDoesntmatch     = errors.New("Check failed")
	errNoneisMatched   = errors.New("None is matched")
	errCriticalMatched = errors.New("The critical is matched")

	Ok       = "ok"
	Critical = "critical"
)

func makeValidator(config *Config) OutputCheck {
	checks := make(map[string]OutputCheck)
	for _, ok := range config.Check.Response.Ok {
		checks[Ok+ok.String()] = checkOutput(ok)
	}

	for _, critical := range config.Check.Response.Critical {
		checks[Critical+critical.String()] = checkOutput(critical)
	}

	return checkAll(checks)
}

func checkAll(checks map[string]OutputCheck) OutputCheck {

	return func(output string) error {
		for key, check := range checks {
			if err := check(output); err == nil {
				if strings.Index(key, Ok) == 0 {
					return nil
				}
				if strings.Index(key, Critical) == 0 {
					return errCriticalMatched
				}
			}
		}
		return errNoneisMatched
	}
}

type OutputCheck func(string) error

func checkOutput(c match.Matcher) OutputCheck {
	return func(output string) error {
		if c.MatchString(output) {
			return nil
		}
		return errDoesntmatch
	}
}
