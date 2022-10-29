package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceXMLLink(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "link with T",
			input:    `junk text <see cref=\"T:Some.Referenced.Thing\" /> junk text`,
			expected: "junk text [Some.Referenced.Thing] junk text",
		},
		{
			name:     "link with M",
			input:    `junk text <see cref=\"M:Some.Referenced.Thing\" /> junk text`,
			expected: "junk text [Some.Referenced.Thing] junk text",
		},
		{
			name:     "no link",
			input:    "junk text",
			expected: "junk text",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, ReplaceXMLLink(tc.input))
		})
	}
}

func TestStripTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		input    string
		expected string
	}{
		{
			name:     "short tag",
			tag:      "c",
			input:    "this is <c>some text</c> right here",
			expected: "this is some text right here",
		},
		{
			name:     "longer tag",
			tag:      "longertag",
			input:    "this is <longertag>some text</longertag> right here",
			expected: "this is some text right here",
		},
		{
			name:     "replace multiple",
			tag:      "c",
			input:    "this <c>is</c> some <c>text</c> right here",
			expected: "this is some text right here",
		},
		{
			name:     "ignore non-matching tags",
			tag:      "c",
			input:    "this <c>is</c> some <d>text</d> right here",
			expected: "this is some <d>text</d> right here",
		},
		{
			name:     "no tags",
			tag:      "c",
			input:    "this is some text right here",
			expected: "this is some text right here",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, StripTag(tc.input, tc.tag))
		})
	}
}

func TestStripParamRef(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "param ref",
			input:    `junk text <paramref name=\"thing\" /> junk text`,
			expected: "junk text thing junk text",
		},
		{
			name:     "no param ref",
			input:    "junk text",
			expected: "junk text",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, StripParamRef(tc.input))
		})
	}
}

func TestParseXMLDocumentation(t *testing.T) {
	tests := []struct {
		input    string
		prefix   string
		expected string
	}{
		{
			input:    `<doc>\n<summary>\nGet the alarm with the given <paramref name=\"name\" />, or <c>null</c>\nif no alarms have that name. If more than one alarm has the name,\nonly returns one of them.\n</summary>\n<param name=\"name\">Name of the alarm to search for.</param>\n</doc>`,
			prefix:   "GetAlarmWithName will ",
			expected: "GetAlarmWithName will get the alarm with the given name, or nil if no alarms have that name. If more than one alarm has the name, only returns one of them.",
		},
		{
			input:    `<doc>\n<summary>\nRepresents an alarm. Obtained by calling\n<see cref=\"M:KerbalAlarmClock.Alarms\" />,\n<see cref=\"M:KerbalAlarmClock.AlarmWithName\" /> or\n<see cref=\"M:KerbalAlarmClock.AlarmsWithType\" />.\n</summary>\n</doc>`,
			prefix:   "Alarm ",
			expected: "Alarm represents an alarm. Obtained by calling [KerbalAlarmClock.Alarms], [KerbalAlarmClock.AlarmWithName] or [KerbalAlarmClock.AlarmsWithType].",
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("test %v", i), func(t *testing.T) {
			out, err := ParseXMLDocumentation(tc.input, tc.prefix)
			require.NoError(t, err)
			require.Equal(t, tc.expected, out)
		})
	}
}
