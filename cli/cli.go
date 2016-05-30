package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func ShowUsage() {
	fmt.Println("\nUSAGE:\n")
	flag.PrintDefaults()
	fmt.Println()
}

func FormatCommandPrefix(command string) string {
	return fmt.Sprintf("%s command", command)
}

func FormatError(prefix string, err error) string {
	if prefix == "" {
		prefix = "Unknown"
	}

	return fmt.Sprintf("%s error: %v\n", prefix, err)
}

func LogError(prefix string, err error) {
	log.Println(FormatError(prefix, err))
}

func FatalError(prefix string, err error) {
	Exit(FormatError(prefix, err))
}

func Exit(reason string) {
	if reason != "" {
		fmt.Fprintln(os.Stderr, reason)
	}

	os.Exit(1)
}
