package subject

import (
	"errors"
	"strings"
)

type Subject struct {
	Parts []string
}

var (
	ErrMultipleWildcards = errors.New("multiple wildcards in subject")
	ErrWildcardNotLast   = errors.New("wildcard not at the end of subject")
)

func Parse(subject string) (Subject, error) {
	parts := strings.Split(subject, ".")

	for i, part := range parts {
		if part == "*" {
			if i != len(parts)-1 {
				return Subject{}, ErrWildcardNotLast
			}
		}
	}

	return Subject{
		Parts: parts,
	}, nil
}

func (s Subject) HasWildcard() bool {
	return s.Parts[len(s.Parts)-1] == "*"
}

func (s Subject) String() string {
	return strings.Join(s.Parts, ".")
}
