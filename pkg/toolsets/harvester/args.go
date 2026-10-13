package harvester

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

var k8sResourceName = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// explicitStringArg reports whether key was provided, including an empty string.
func explicitStringArg(req mcp.CallToolRequest, key string) (string, bool) {
	args := req.GetArguments()
	val, ok := args[key]
	if !ok || val == nil {
		return "", false
	}
	s, ok := val.(string)
	if !ok {
		return "", false
	}
	return s, true
}

func argPresent(req mcp.CallToolRequest, key string) bool {
	args := req.GetArguments()
	val, ok := args[key]
	if !ok || val == nil {
		return false
	}
	if s, ok := val.(string); ok && strings.TrimSpace(s) == "" {
		return false
	}
	return true
}

// optionalBoolArg reads a bool argument. Strings "true"/"false" and JSON bools are accepted.
// set is false when the argument is absent.
func optionalBoolArg(req mcp.CallToolRequest, key string) (value bool, set bool, err error) {
	if !argPresent(req, key) {
		return false, false, nil
	}
	switch v := req.GetArguments()[key].(type) {
	case bool:
		return v, true, nil
	case string:
		b, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return false, true, fmt.Errorf("%s must be true or false", key)
		}
		return b, true, nil
	default:
		return false, true, fmt.Errorf("%s must be true or false", key)
	}
}

func applyHaltAfterRestore(req mcp.CallToolRequest, spec map[string]interface{}) error {
	halt, set, err := optionalBoolArg(req, "halt_after_restore")
	if err != nil {
		return err
	}
	if set && halt {
		spec["haltAfterRestore"] = true
	}
	return nil
}

func requireResourceName(kind, name string) error {
	if !k8sResourceName.MatchString(name) || len(name) > 63 {
		return fmt.Errorf("%s %q is not a valid Kubernetes name", kind, name)
	}
	return nil
}
