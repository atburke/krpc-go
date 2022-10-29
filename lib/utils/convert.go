package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/ztrue/tracerr"
)

// EncodeBool encodes a bool in the Protobuf format.
func EncodeBool(b bool) []byte {
	if b {
		return proto.EncodeVarint(1)
	}
	return proto.EncodeVarint(0)
}

// DecodeBool decodes a bool from the Protobuf format.
func DecodeBool(d []byte) (bool, error) {
	b, size := proto.DecodeVarint(d)
	if size == 0 {
		return false, tracerr.Errorf("Failed to decode bool: %v", d)
	}
	return b != 0, nil
}

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
	matches := summaryRE.FindStringSubmatch(docData)
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
	summary = strings.ToLower(summary[:1]) + summary[1:]
	return prefix + summary, nil
}
