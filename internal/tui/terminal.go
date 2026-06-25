package tui

import "os"

func restoreTerminal() {
	_, _ = os.Stdout.WriteString("\x1b[?25h\x1b[0m")
}
