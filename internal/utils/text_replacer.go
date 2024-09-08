package utils

import (
	"strings"
)

// ReplacePlaceholders reemplaza los placeholders en el texto con los valores proporcionados en params.
func ReplacePlaceholders(text string, params map[string]string) string {
	// Iterar sobre todos los parámetros y reemplazar los placeholders
	for key, value := range params {
		// Reemplazar el placeholder con el valor real
		placeholder := "&" + key
		text = strings.ReplaceAll(text, placeholder, value)
	}
	return text
}
