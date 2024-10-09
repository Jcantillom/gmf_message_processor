package utils

import (
	"strings"
)

// ReplacePlaceholders reemplaza los placeholders en el texto con los valores proporcionados en params.
func ReplacePlaceholders(text string, params map[string]string) string {
	// Iterar sobre todos los parámetros y reemplazar los placeholders
	for key, value := range params {
		// Reemplazar el placeholder con el valor real
		text = strings.ReplaceAll(text, key, value)
	}
	return text
}
