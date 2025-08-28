package config

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

func PrintStartupBanner() {
	fmt.Println(getBanner() + "\n")
	fmt.Printf("Basic Auth mode: %v\n", GetBasicAuthMode())
	fmt.Printf("HTTP mode      : %v\n", GetHttpMode())
	fmt.Printf("Readonly mode  : %v\n", GetReadonlyMode())
	fmt.Printf("Sinkhole mode  : %v\n", GetSinkholeMode())
	fmt.Printf("Path           : %s\n", GetPathDataFolder())

	if GetBasicAuthMode() {
		fmt.Printf("Username       : %s\n", GetBasicAuthUsername())
		fmt.Printf("Password       : %s\n", GetBasicAuthPassword())
	}

	if GetHttpMode() {
		fmt.Printf("Listening on   : http://0.0.0.0:%d\n", GetPortToListenOn())
	} else {
		if pathTLSCertFile == "" || pathTLSKeyFile == "" {
			fmt.Printf("TLS cert       : self-signed\n")
			fmt.Printf("TLS key        : self-signed\n")
		} else {
			fmt.Printf("TLS cert       : %s\n", pathTLSCertFile)
			fmt.Printf("TLS key        : %s\n", pathTLSKeyFile)
		}

		fmt.Printf("Listening on   : https://0.0.0.0:%d\n", GetPortToListenOn())
	}

	fmt.Println("")
}

func centerTextWithWhitespaces(text string, maxWidth int) string {
	textLength := utf8.RuneCountInString(text)
	if textLength >= maxWidth {
		return text
	}

	totalPadding := maxWidth - textLength
	leftPadding := int(math.Ceil((float64(totalPadding) / 2.0)))
	rightPadding := totalPadding - leftPadding

	return strings.Repeat(" ", leftPadding) + text + strings.Repeat(" ", rightPadding)
}

func getBanner() string {
	return strings.Join(
		[]string{
			"┌──────────────────────────────────────┐",
			"│                Ablage                │",
			fmt.Sprintf(
				"│%s│",
				centerTextWithWhitespaces("v"+VersionString, 38),
			),
			"└──────────────────────────────────────┘",
		},
		"\n",
	)
}
