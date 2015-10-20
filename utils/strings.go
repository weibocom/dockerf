package utils

import "strings"

// delete all blank or empty element
func TrimSplit(str, split string) []string {
	splits := strings.Split(str, split)
	if len(splits) == 0 {
		return splits
	}
	slice := make([]string, len(splits))
	idx := 0
	for _, s := range splits {
		ts := strings.TrimSpace(s)
		if ts == "" {
			continue
		}
		slice[idx] = ts
		idx++
	}
	return slice[0:idx]
}
