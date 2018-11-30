package util

import (
	"fmt"
	"strings"
)

func BuildCmd(dir, command string, args ...string) string {
	if dir != "" {
		return fmt.Sprintf("cd %v && %v %v", dir, command, strings.Join(args, " "))
	}

	return fmt.Sprintf("%v %v", command, strings.Join(args, " "))
}
