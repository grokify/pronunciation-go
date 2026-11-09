package respell

import (
	"fmt"
	"strings"
	"unicode"
)

// English is the language whose IPA is respelled as written; other
// languages' sounds are brought to the nearest English ones.
const English = "en"

// consonants respells consonants and affricates.
var consonants = map[string]string{
	"b": "b", "d": "d", "f": "f", "h": "h", "k": "k", "l": "l", "m": "m", "n": "n", "p": "p",
	"s": "s", "t": "t", "v": "v", "w": "w", "z": "z", "r": "r", "ɡ": "g", "c": "k", "q": "k",
	"θ": "th", "ð": "dh", "ʃ": "sh", "ʒ": "zh", "tʃ": "ch", "dʒ": "j", "ŋ": "ng", "j": "y",
	"x": "kh", "χ": "kh", "ç": "kh", "ts": "ts", "dz": "dz", "ɲ": "ny", "ʎ": "ly", "β": "b",
	"ɣ": "g", "ɟ": "g", "ʝ": "y", "ɱ": "m", "ɰ": "w", "ʋ": "v", "ɥ": "w", "ʍ": "wh", "ɕ": "sh",
	"ʑ": "zh", "ʂ": "sh", "ʐ": "zh", "ɦ": "h", "ɬ": "l",
}

// englishVowels respells English IPA vowels; closed is the respelling in a
// syllable ending in a consonant, where it differs.
var englishVowels = map[string]struct{ open, closed string }{
	"æ": {"a", "a"}, "a": {"ah", "ah"}, "ɑ": {"ah", "ah"}, "ɒ": {"o", "o"}, "ɔ": {"aw", "aw"},
	"e": {"eh", "e"}, "ɛ": {"eh", "e"}, "ɪ": {"i", "i"}, "i": {"ee", "ee"}, "ʊ": {"uu", "uu"},
	"u": {"oo", "oo"}, "ʌ": {"uh", "u"}, "ə": {"ə", "ə"}, "ɜ": {"ur", "ur"}, "eɪ": {"ay", "ay"},
	"aɪ": {"y", "y"}, "ɔɪ": {"oy", "oy"}, "aʊ": {"ow", "ow"}, "oʊ": {"oh", "oh"}, "əʊ": {"oh", "oh"},
	"o": {"oh", "oh"}, "ɪə": {"eer", "eer"}, "ɛə": {"air", "air"}, "eə": {"air", "air"},
	"ʊə": {"oor", "oor"}, "y": {"ew", "ew"}, "ʏ": {"ew", "ew"}, "ø": {"ur", "ur"}, "œ": {"ur", "ur"},
	"ɐ": {"uh", "uh"}, "ɨ": {"i", "i"}, "ʉ": {"oo", "oo"}, "ɘ": {"ə", "ə"}, "ɵ": {"ur", "ur"},
	"ɯ": {"oo", "oo"}, "ɤ": {"uh", "uh"}, "ɔʏ": {"oy", "oy"}, "ɞ": {"ur", "ur"}, "ɶ": {"ur", "ur"},
}

// foreignVowels respells other languages' pure vowels, which English
// respelling writes long: Italian [o] is "oh", not the "o" of "lot".
var foreignVowels = map[string]struct{ open, closed string }{
	"a": {"ah", "ah"}, "ɑ": {"ah", "ah"}, "æ": {"a", "a"}, "e": {"eh", "e"}, "ɛ": {"eh", "e"},
	"i": {"ee", "ee"}, "ɪ": {"ee", "i"}, "o": {"oh", "o"}, "ɔ": {"oh", "o"}, "u": {"oo", "oo"},
	"ʊ": {"oo", "uu"},
}

// rVowels respells a vowel followed by r in the same syllable.
var rVowels = map[string]string{
	"ɑ": "ar", "a": "ar", "ɔ": "or", "o": "or", "ɜ": "ur", "ə": "ər", "ɪə": "eer", "ɛə": "air",
	"eə": "air", "ʊə": "oor", "ɛ": "err", "e": "air", "ɪ": "irr", "i": "eer", "ʌ": "urr",
	"æ": "arr", "ɒ": "orr", "u": "oor", "ʊ": "oor", "eɪ": "air", "aɪ": "yr", "aʊ": "owr",
}

// nasalVowels respells nasal vowels, as French [ɑ̃].
var nasalVowels = map[string]string{
	"a": "ahn", "ɑ": "ahn", "ɔ": "ohn", "o": "ohn", "ɛ": "an", "e": "an", "œ": "un", "ə": "un",
	"i": "een", "u": "oon",
}

// FromIPA respells IPA in a language (a BCP 47 tag such as en or it-IT;
// empty is English), word by word. Slashes and brackets around it are
// left out.
func FromIPA(ipa, lang string) ([]Word, error) {
	english := lang == "" || strings.EqualFold(strings.SplitN(lang, "-", 2)[0], English)
	ipa = strings.Trim(strings.TrimSpace(ipa), "/[]")
	var words []Word
	for _, w := range strings.FieldsFunc(ipa, func(r rune) bool { return unicode.IsSpace(r) || r == '|' || r == '‖' || r == ',' }) {
		ps, err := parseIPA(w)
		if err != nil {
			return nil, err
		}
		word := Word{Source: w}
		for _, s := range syllables(collapse(ps)) {
			word.Syllables = append(word.Syllables, Syllable{Text: respellSyllable(s, english), Stress: s.stress})
		}
		if len(word.Syllables) > 0 {
			words = append(words, word)
		}
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("no sounds in IPA %q", ipa)
	}
	return words, nil
}

func respellSyllable(s syllable, english bool) string {
	var b strings.Builder
	for _, p := range s.onset {
		b.WriteString(consonant(p))
	}
	coda := s.coda
	if s.nucleus.sym != "" {
		var v string
		switch {
		case s.nucleus.nasal && nasalVowels[s.nucleus.sym] != "":
			v = nasalVowels[s.nucleus.sym]
		case len(coda) > 0 && coda[0].sym == "r" && rVowels[s.nucleus.sym] != "":
			v = rVowels[s.nucleus.sym]
			coda = coda[1:]
		default:
			v = vowel(s.nucleus.sym, len(coda) > 0, english)
			if s.nucleus.sym == "aɪ" && len(s.onset) == 0 {
				v = "eye"
			}
		}
		// A g before e, i or y is hard: "ghee", not "gee".
		if strings.HasSuffix(b.String(), "g") && v != "" && strings.ContainsAny(v[:1], "eiy") {
			b.WriteString("h")
		}
		b.WriteString(v)
	}
	for _, p := range coda {
		b.WriteString(consonant(p))
	}
	return b.String()
}

func consonant(p phone) string {
	if c, ok := consonants[p.sym]; ok {
		return c
	}
	return p.sym
}

func vowel(sym string, closed, english bool) string {
	if !english {
		if v, ok := foreignVowels[sym]; ok {
			if closed {
				return v.closed
			}
			return v.open
		}
	}
	if v, ok := englishVowels[sym]; ok {
		if closed {
			return v.closed
		}
		return v.open
	}
	return sym
}

// FromRespelling reads a respelling such as LAM-bər-GHEE-nee or En-Zoh:
// hyphens split syllables, spaces split words, and a syllable in capitals
// is stressed.
func FromRespelling(s string) ([]Word, error) {
	var words []Word
	for _, w := range strings.Fields(s) {
		w = strings.TrimFunc(w, func(r rune) bool { return !unicode.IsLetter(r) })
		if w == "" {
			continue
		}
		word := Word{Source: w}
		for _, part := range strings.Split(w, "-") {
			if part == "" {
				continue
			}
			stress := Unstressed
			if len([]rune(part)) > 1 && part == strings.ToUpper(part) && part != strings.ToLower(part) {
				stress = Primary
			}
			word.Syllables = append(word.Syllables, Syllable{Text: strings.ToLower(part), Stress: stress})
		}
		words = append(words, word)
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("no respelling in %q", s)
	}
	return words, nil
}
