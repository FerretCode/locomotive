package reconstruct_discord

import "strings"

// ref: https://gist.github.com/thomasbnt/b6f455e2c7d743b796917fa3c205f812
func getColor(severity string) int {
	severity = strings.ToLower(severity)

	color := 2303786 // black

	switch severity {
	case "info":
		color = 16777215 // white
	case "err", "error":
		color = 15548997 // red
	case "warn":
		color = 16776960 // yellow
	case "debug":
		color = 9807270 // grey
	}

	return color
}
