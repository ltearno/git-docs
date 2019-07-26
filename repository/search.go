package repository

import (
	"strings"
	"unicode"
)

func tagsContainText(tags []string, q string) bool {
	if q == "" {
		return true
	}

	exactMatch := unicode.IsUpper([]rune(q)[0])
	q = strings.ToLower(q)

	if exactMatch {
		for _, tag := range tags {
			if strings.ToLower(tag) == q {
				return true
			}
		}
	} else {
		for _, tag := range tags {
			if strings.Contains(strings.ToLower(tag), q) {
				return true
			}
		}
	}

	return false
}

func tagsMatchSearch(tags []string, q string) bool {
	q = strings.TrimSpace(q)
	if q == "" {
		return true
	} else if strings.HasPrefix(q, "!") {
		q = strings.TrimSpace(q[1:])
		return !tagsMatchSearch(tags, q)
	} else if strings.HasPrefix(q, "&") {
		q = strings.TrimSpace(q[1:])
		separatorPos := strings.Index(q, " ")
		if separatorPos == 0 {
			return false
		}
		return tagsMatchSearch(tags, q[:separatorPos]) && tagsMatchSearch(tags, q[separatorPos+1:])
	} else if strings.HasPrefix(q, "|") {
		q = strings.TrimSpace(q[1:])
		separatorPos := strings.Index(q, " ")
		if separatorPos == 0 {
			return false
		}
		return tagsMatchSearch(tags, q[:separatorPos]) || tagsMatchSearch(tags, q[separatorPos+1:])
	} else {
		return tagsContainText(tags, q)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
