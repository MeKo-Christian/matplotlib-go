package pyplot

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/style"
)

// RCParams returns the active rcParam-style defaults.
func RCParams() style.Params {
	return style.CurrentParams()
}

// RC applies rcParam-style overrides to the active defaults. When group is non-empty,
// keys in params are prefixed with "group." unless already fully qualified.
func RC(group string, params style.Params) error {
	_, err := style.UpdateParams(qualifyRCParams(group, params))
	return err
}

// RCContext applies temporary rcParam overrides and returns a restore function.
func RCContext(params style.Params) (func(), error) {
	restore, _, err := style.PushContext(params)
	if err != nil {
		return nil, err
	}
	return restore, nil
}

// RCDefaults restores the active defaults to the library baseline.
func RCDefaults() {
	style.ResetDefaults()
}

// LoadRCFile loads a Matplotlib-style rc file into the active defaults.
func LoadRCFile(path string) error {
	_, err := style.LoadRCFile(path)
	return err
}

// LoadDefaultRCFile loads the first rc file found in the default search path.
func LoadDefaultRCFile() (string, error) {
	path, _, err := style.LoadDefaultRCFile()
	return path, err
}

func qualifyRCParams(group string, params style.Params) style.Params {
	if len(params) == 0 {
		return nil
	}

	qualified := make(style.Params, len(params))
	group = strings.ToLower(strings.TrimSpace(group))
	for key, value := range params {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if group != "" && !strings.Contains(normalizedKey, ".") {
			normalizedKey = group + "." + normalizedKey
		}
		qualified[normalizedKey] = value
	}
	return qualified
}
