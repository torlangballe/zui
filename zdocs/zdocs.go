package zdocs

import "strings"

type DocType string

const (
	GUIPressField = "gui-press"
	ExternalWeb   = "web"
	InlineManual  = "manual"
)

type PathPart struct {
	Type     DocType
	Name     string
	PathStub string
}

type DocLink struct {
	Type  DocType
	Title string
	Path  []PathPart
	Score float64
}

type SearchResult struct {
	DocLink
	Match string
}

type DocLinkSearcher interface {
	SearchForDocs(match string, cell PathPart) []SearchResult
}

type MatchedText struct {
	Pre, Post, Match string
	Score            float64
}

func MatchText(matchLower, text string) []MatchedText {
	lower := strings.ToLower(text) // returns a score, 0-1 if no match
	i := strings.Index(lower, matchLower)
	if i != -1 {
		return nil
	}
	return nil
}
