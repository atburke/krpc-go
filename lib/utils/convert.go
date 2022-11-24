package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ztrue/tracerr"
)

var xmlLink = regexp.MustCompile(`<see cref=\\"(T|M):([a-zA-z.]+)\\" ?/>`)

// ReplaceXMLLink attempts to convert an XML doc link to a Godoc one.
func ReplaceXMLLink(link string) string {
	return xmlLink.ReplaceAllString(link, "[$2]")
}

// StripTag removes a specified tag from some text.
func StripTag(text, tag string) string {
	re, err := regexp.Compile(fmt.Sprintf("<%v>([^<]+)</%v>", tag, tag))
	if err != nil {
		return text
	}
	return re.ReplaceAllString(text, "$1")
}

var paramRef = regexp.MustCompile(`<paramref name=\\"([a-zA-Z]+)\\" ?/>`)

// StripParamRef removes the paramref tag from some text.
func StripParamRef(text string) string {
	return paramRef.ReplaceAllString(text, "$1")
}

var summaryRE = regexp.MustCompile(`<summary>(.+)</summary>`)

// ParseXMLDocumentation parses a Go doc comment's content from
// C# XML docs.
func ParseXMLDocumentation(docData, prefix string) (string, error) {
	matches := summaryRE.FindStringSubmatch(strings.ReplaceAll(docData, "\n", " "))
	if matches == nil {
		return "", tracerr.Errorf("No summary in doc string: %v", docData)
	}
	summary := matches[1]

	summary = strings.ReplaceAll(summary, "<c>null</c>", "nil")
	summary = StripTag(summary, "c")
	summary = StripParamRef(summary)
	summary = ReplaceXMLLink(summary)
	summary = strings.ReplaceAll(summary, "\\n", " ")
	summary = strings.TrimSpace(summary)
	if prefix != "" {
		summary = strings.ToLower(summary[:1]) + summary[1:]
	}
	return prefix + summary, nil
}

// SanitizeIdentifier ensures that an identifier is valid for Go.
func SanitizeIdentifier(s string) string {
	switch s {
	case "type":
		return "t"
	case "func":
		return "f"
	default:
		return s
	}
}
