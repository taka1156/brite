package help

import (
	"strings"
	"testing"
)

func TestHelpBrite_HELP_MSG(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{name: "contains init command", command: "init"},
		{name: "contains setup command", command: "setup"},
		{name: "contains new command", command: "new"},
		{name: "contains convert command", command: "convert"},
		{name: "contains publish command", command: "publish"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(HELP_MSG, tt.command) {
				t.Errorf("HELP_MSG does not contain %q", tt.command)
			}
		})
	}
}
