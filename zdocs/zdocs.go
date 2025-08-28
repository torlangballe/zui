package zdocs

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
)

type DocType string

// https://github.com/blevesearch/bleve
// https://kevincoder.co.za/bleve-how-to-build-a-rocket-fast-search-engine
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
	matchWords := strings.Fields(matchLower)
	zlog.Info("MatchText:", len(matchWords), matchWords)

	matchWordsLens := make([]int, len(matchWords))
	for i, w := range matchWords {
		matchWordsLens[i] = len(w)
	}
	buf := bytes.NewBuffer([]byte(lower))
	s := bufio.NewScanner(buf)
	var lines int
	for s.Scan() {
		line := s.Text()
		if s.Err() != nil {
			break
		}
		lineWords := strings.Fields(line)
		var mi, matches int
		var scoreSum float64
		for li, lw := range lineWords {
			lw = zstr.AlphaNumericASCIIOnly(lw)
			for {
				var score float64
				if strings.HasPrefix(lw, matchWords[mi]) {
					score = float64(matchWordsLens[mi]) / float64(len(lw))
					if score < 0.7 && matchWordsLens[mi] < 4 {
						score = 0 // if small stub on big word, don't use
					}
				} else {
					score = -zstr.GetLevenshteinRatio(lw, matchWords[mi])
				}
				mi++
				if score == 1 {
					scoreSum += score
					matches++
					if mi >= len(matchWords) {
						scoreSum /= float64(len(matchWords))
						zlog.Info("Score:", matchWords[mi-1], score, lw, li, matches)
						addMatch(scoreSum, lineWords, li, matches, lines)
						scoreSum = 0
						matches = 0
						mi = 0
					}
					break
				} else {
					if scoreSum != 0 {
						scoreSum /= float64(len(matchWords))
						addMatch(scoreSum, lineWords, li, matches, lines)
					}
					scoreSum = 0
					matches = 0
					if score > -0.3 && score != 0 {
						addMatch(score/float64(len(matchWords)), lineWords, li, 1, lines)
					}
					if mi >= len(matchWords) {
						mi = 0
						break
					}
				}
			}
		}
		lines++
	}
	return nil
}

func addMatch(score float64, lineWords []string, lwi, matches, line int) {
	fmt.Print(zstr.EscCyan, line, score, ": ")
	s := max(0, lwi-3-matches)
	e := min(len(lineWords), lwi+4)
	for i := s; i < e; i++ {
		if i >= lwi-matches && i < lwi {
			fmt.Print(zstr.EscYellow)
		} else {
			fmt.Print(zstr.EscGreen)
		}
		fmt.Print(lineWords[i], " ")
	}
	fmt.Println(zstr.EscNoColor)
}
