package rotor

import "strings"

type Subject struct {
	Parts      []string
	GroupDepth int
}

func ParseSubject(subject string) Subject {
	parts := strings.Split(subject, ".")
	groupDepth := 0

	for i, part := range parts {
		if part == "*" {
			groupDepth = i
			break
		}
	}

	return Subject{
		Parts:      parts,
		GroupDepth: groupDepth,
	}
}

func (s Subject) Group() string {
	if s.GroupDepth == 0 {
		return ""
	}

	return strings.Join(s.Parts[:s.GroupDepth], ".")
}

func (s Subject) String() string {
	return strings.Join(s.Parts, ".")
}
