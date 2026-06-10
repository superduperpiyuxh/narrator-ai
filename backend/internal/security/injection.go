package security

import (
	"regexp"
	"strings"
)

var injectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ignore\s+(previous|all|above)\s+instructions`),
	regexp.MustCompile(`(?i)disregard\s+(your|all)\s+(rules|instructions)`),
	regexp.MustCompile(`(?i)you\s+are\s+now\s+in\s+(admin|developer|dan)\s+mode`),
	regexp.MustCompile("(?i)```system"),
	regexp.MustCompile(`(?i)\[INST\]`),
	regexp.MustCompile(`(?i)###\s*Instruction:`),
	regexp.MustCompile(`(?i)base64:[A-Za-z0-9+/=]{20,}`),
	regexp.MustCompile(`(?i)pretend\s+(to\s+be|you\s+are)`),
	regexp.MustCompile(`(?i)act\s+as\s+(if|a)`),
	regexp.MustCompile(`(?i)system\s*:\s*`),
}

var zeroWidthChars = regexp.MustCompile(`[\x{200b}-\x{200f}\x{2028}-\x{202f}\x{2060}-\x{206f}]`)
var htmlComments = regexp.MustCompile(`<!--.*?-->`)

func DetectInjection(input string) bool {
	for _, pattern := range injectionPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

func SanitizeInput(input string, maxLength int) string {
	if len(input) > maxLength {
		input = input[:maxLength]
	}

	input = strings.TrimSpace(input)

	input = zeroWidthChars.ReplaceAllString(input, "")

	input = htmlComments.ReplaceAllString(input, "")

	return input
}

func WrapInXML(content, label string) string {
	return "<" + label + ">\n" + content + "\n</" + label + ">"
}
