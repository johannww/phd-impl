package setup

import "strings"

func isAlreadySetupError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already exists")
}
