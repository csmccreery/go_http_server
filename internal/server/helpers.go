package server

import (
	"strings"
)

func CleanProfaneWords(chirp string, profaneMap map[string]bool) string {
	fields := strings.Fields(chirp)
	for i := 0; i < len(fields); i++ {
		word := strings.ToLower(fields[i])
		_, ok := profaneMap[word]
		if ok {
			fields[i] = "****"
		}
	}

	return strings.Join(fields, " ")
}
