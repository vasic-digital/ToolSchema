package tools

import (
	"path/filepath"
	"regexp"
	"strings"
)

// ValidatePath checks if a path is safe for use in commands.
// It prevents path traversal and shell injection attacks.
func ValidatePath(path string) bool {
	if path == "" {
		return false
	}

	// Prevent path traversal
	clean := filepath.Clean(path)
	if strings.Contains(clean, "..") {
		return false
	}

	// Prevent shell metacharacters
	dangerous := []string{";", "&", "|", "$", "`", "(", ")", "{", "}", "<", ">", "\n", "\r"}
	for _, char := range dangerous {
		if strings.Contains(path, char) {
			return false
		}
	}

	return true
}

// ValidateSymbol checks if a symbol name is safe for use in grep patterns.
func ValidateSymbol(symbol string) bool {
	if symbol == "" {
		return false
	}

	// Only allow alphanumeric, underscore, and common symbol characters
	validSymbol := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return validSymbol.MatchString(symbol)
}

// SanitizePath cleans a path and returns an error if it's unsafe.
func SanitizePath(path string) (string, bool) {
	if !ValidatePath(path) {
		return "", false
	}
	return filepath.Clean(path), true
}

// ValidateGitRef checks if a git reference is safe.
func ValidateGitRef(ref string) bool {
	if ref == "" {
		return false
	}

	// Only allow safe characters in git refs
	validRef := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
	return validRef.MatchString(ref)
}

// ValidateCommandArg checks if an argument is safe for shell commands.
func ValidateCommandArg(arg string) bool {
	if arg == "" {
		return true // Empty is safe
	}

	// Prevent shell metacharacters
	dangerous := []string{";", "&", "|", "$", "`", "(", ")", "{", "}", "<", ">", "\n", "\r", "\\"}
	for _, char := range dangerous {
		if strings.Contains(arg, char) {
			return false
		}
	}

	return true
}
