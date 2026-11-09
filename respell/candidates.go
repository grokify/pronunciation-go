package respell

import (
	"regexp"
	"strings"
	"unicode"
)

// MaxCandidates is the most respellings Candidates offers.
const MaxCandidates = 8

// Candidates offers respellings of a word for a speech engine, most likely
// first. Voices read unfamiliar words by their own rules, so the same
// sounds are written several ways: as one word, split where the stress
// falls (voices tend to stress a word's first syllable), syllable by
// syllable, with simpler endings, and with a consonant cluster split.
func Candidates(w Word) []string {
	if len(w.Syllables) == 0 {
		return nil
	}
	texts := make([]string, len(w.Syllables))
	for i, s := range w.Syllables {
		texts[i] = plain(s.Text)
	}
	var out []string
	seen := map[string]bool{}
	add := func(s string) {
		s = strings.TrimSpace(s)
		k := strings.ToLower(s)
		if s == "" || seen[k] || len(out) >= MaxCandidates {
			return
		}
		seen[k] = true
		out = append(out, capitalize(s))
	}

	joined := strings.Join(texts, "")
	add(joined)
	add(simpleEnding(joined))
	// A new word at each stressed syllable after the first.
	var groups []string
	for i, t := range texts {
		if i > 0 && w.Syllables[i].Stress != Unstressed {
			groups = append(groups, "")
		}
		if len(groups) == 0 {
			groups = append(groups, "")
		}
		groups[len(groups)-1] += t
	}
	add(strings.Join(groups, " "))
	add(simpleEnding(strings.Join(groups, " ")))
	add(strings.Join(texts, " "))
	// Split a cluster that starts a syllable: "En tsoh" as "Ent soh".
	for i := 1; i < len(texts); i++ {
		if c := leadingConsonants(texts[i]); len(c) >= 2 && !digraphs[c[:2]] {
			split := append([]string{}, texts...)
			split[i-1] += c[:1]
			split[i] = texts[i][1:]
			add(strings.Join(split, " "))
			break
		}
	}
	// An "eh" before a consonant reads as "e": Fehroocho as Feroocho.
	add(ehBeforeConsonant.ReplaceAllString(joined, "e$1"))
	// The unstressed vowel as "a", as in "sofa".
	add(strings.Join(schwaAsA(w), ""))
	return out
}

// plain writes a syllable without "ə", which voices can't read.
func plain(s string) string {
	s = strings.ReplaceAll(s, "ər", "er")
	return strings.ReplaceAll(s, "ə", "uh")
}

func schwaAsA(w Word) []string {
	out := make([]string, len(w.Syllables))
	for i, s := range w.Syllables {
		t := strings.ReplaceAll(s.Text, "ər", "er")
		out[i] = strings.ReplaceAll(t, "ə", "a")
	}
	return out
}

var ehBeforeConsonant = regexp.MustCompile(`eh([bcdfgjklmnpqrstvwxz])`)

// simpleEnding drops a final "h" or doubled vowel: "Entsoh" as "Entso",
// "Lambergheenee" as "Lambergheeny".
func simpleEnding(s string) string {
	switch {
	case strings.HasSuffix(s, "ee") && len(s) > 3:
		return s[:len(s)-2] + "y"
	case strings.HasSuffix(s, "oh") || strings.HasSuffix(s, "ah"):
		return s[:len(s)-1]
	}
	return s
}

// digraphs are letter pairs that spell one sound, which a split would break.
var digraphs = map[string]bool{
	"ch": true, "sh": true, "th": true, "dh": true, "zh": true, "kh": true, "gh": true,
	"ng": true, "ny": true, "ly": true, "wh": true,
}

func leadingConsonants(s string) string {
	i := strings.IndexFunc(s, func(r rune) bool { return strings.ContainsRune("aeiouy", r) })
	if i < 0 {
		return s
	}
	return s[:i]
}

func capitalize(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// Suggestion is the respellings suggested for one word of a source.
type Suggestion struct {
	// Kind is "ipa" or "respelling".
	Kind string `json:"kind"`
	// Lang is the IPA's language, such as it; empty for English.
	Lang string `json:"lang,omitempty"`
	// Source is the word as given.
	Source string `json:"source"`
	// Respelling is the word in the style of Wikipedia's respelling key.
	Respelling string `json:"respelling"`
	// Candidates are respellings for a speech engine, most likely first.
	Candidates []string `json:"candidates"`
}

// Languages are the names Wikipedia writes before IPA, as in
// "Italian: [ferˈruttʃo]", by language tag.
var Languages = map[string]string{
	"english": "en", "italian": "it", "spanish": "es", "french": "fr", "german": "de",
	"portuguese": "pt", "dutch": "nl", "swedish": "sv", "norwegian": "no", "danish": "da",
	"polish": "pl", "czech": "cs", "russian": "ru", "japanese": "ja", "catalan": "ca",
}

var (
	bracketed  = regexp.MustCompile(`(?:([A-Za-z]+)\s*:\s*)?(/[^/]+/|\[[^\]]+\])`)
	respelled  = regexp.MustCompile(`^[\p{L}ə]+(?:-[\p{L}ə]+)+$`)
	ipaLetters = "æɑɒɔɛɪʊʌəɜɝɚøœɐɨʉɯɤʏθðʃʒŋɡɲʎɾɹʁχçβɣˈˌː"
)

// Suggest reads text such as Wikipedia's "/ˌlæmbərˈɡiːni/ LAM-bər-GHEE-nee,
// Italian: [ferˈruttʃo lamborˈɡiːni]", or a respelling someone tried, and
// suggests respellings for each word. IPA between slashes is English; IPA
// in brackets is in the language named before it, or lang. Hyphenated words
// outside them are respellings. Text with neither is read as IPA when it
// has IPA letters, and as a respelling otherwise.
func Suggest(text, lang string) ([]Suggestion, error) {
	var out []Suggestion
	addWords := func(kind, l string, words []Word) {
		for _, w := range words {
			out = append(out, Suggestion{Kind: kind, Lang: l, Source: w.Source, Respelling: w.Respelling(), Candidates: Candidates(w)})
		}
	}
	rest := text
	for _, m := range bracketed.FindAllStringSubmatch(text, -1) {
		rest = strings.Replace(rest, m[0], " ", 1)
		l := lang
		if strings.HasPrefix(m[2], "/") {
			l = ""
		}
		if name := strings.ToLower(m[1]); Languages[name] != "" {
			l = Languages[name]
			if l == English {
				l = ""
			}
		}
		words, err := FromIPA(m[2], l)
		if err != nil {
			return nil, err
		}
		addWords("ipa", l, words)
	}
	var tries []string
	for _, f := range strings.Fields(rest) {
		f = strings.TrimFunc(f, func(r rune) bool { return !unicode.IsLetter(r) && r != 'ə' })
		if respelled.MatchString(f) {
			tries = append(tries, f)
		}
	}
	if len(tries) > 0 {
		words, err := FromRespelling(strings.Join(tries, " "))
		if err != nil {
			return nil, err
		}
		addWords("respelling", "", words)
	}
	if len(out) == 0 && strings.TrimSpace(text) != "" {
		var words []Word
		var err error
		kind := "respelling"
		if strings.ContainsAny(text, ipaLetters) {
			kind = "ipa"
			words, err = FromIPA(text, lang)
		} else {
			words, err = FromRespelling(text)
		}
		if err != nil {
			return nil, err
		}
		l := ""
		if kind == "ipa" {
			l = lang
		}
		addWords(kind, l, words)
	}
	return out, nil
}
