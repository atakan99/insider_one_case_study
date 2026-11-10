package content

import (
	"errors"
	"strings"
)

// ParseVars checks the variables of one field. It returns the count of
// {{first_name}}. Each "{" and "}" must be part of an exact {{first_name}}.
func ParseVars(s string) (int, error) {
	n := 0
	for i := 0; i < len(s); {
		if strings.HasPrefix(s[i:], VarFirstName) {
			n++
			i += len(VarFirstName)
			continue
		}
		if s[i] == '{' || s[i] == '}' {
			return 0, errors.New("only the variable " + VarFirstName + " is allowed; found an unknown variable or a stray brace")
		}
		i++
	}
	return n, nil
}

// Render replaces each {{first_name}} with name. It does no escaping.
func Render(s, name string) string {
	return strings.ReplaceAll(s, VarFirstName, name)
}
