// Package respell turns IPA and dictionary respellings, such as Wikipedia's
// /ˌlæmbərˈɡiːni/ and LAM-bər-GHEE-nee, into respellings for speech engines
// that read plain text, and offers several, since which one a voice says
// best is found by listening.
//
// IPA is read syllable by syllable into a respelling in the style of
// Wikipedia's pronunciation respelling key, stressed syllables in capitals.
// Other languages' IPA is brought to the nearest English sounds, so some
// detail, such as a rolled r or a doubled consonant, is lost.
package respell

import (
	"fmt"
	"strings"
	"unicode"
)

// Stress is a syllable's stress.
type Stress int

// Stress levels.
const (
	Unstressed Stress = iota
	Secondary
	Primary
)

// Syllable is one syllable of a respelling, such as "ghee", in lower case;
// "ə" stands for the unstressed vowel in "about".
type Syllable struct {
	Text   string
	Stress Stress
}

// Word is a respelled word.
type Word struct {
	// Source is the word as given, in IPA or a respelling.
	Source    string
	Syllables []Syllable
}

// Respelling writes the word in the style of Wikipedia's respelling key:
// syllables joined by hyphens, stressed ones in capitals, as LAM-bər-GHEE-nee.
func (w Word) Respelling() string {
	parts := make([]string, len(w.Syllables))
	for i, s := range w.Syllables {
		parts[i] = s.Text
		if s.Stress != Unstressed {
			parts[i] = strings.ToUpper(s.Text)
		}
	}
	return strings.Join(parts, "-")
}

// phone is one sound of IPA, or a marker between sounds.
type phone struct {
	sym    string // IPA symbol, with affricates and diphthongs joined
	vowel  bool
	nasal  bool // nasal vowel, as French ɑ̃
	marker rune // ˈ, ˌ or '.' for a syllable break; 0 for a sound
}

const (
	markPrimary   = 'ˈ'
	markSecondary = 'ˌ'
	markBreak     = '.'
)

var vowels = "aeiouyæɑɒɔɛɪʊʌəɜɝɚøœɐɨʉɯɤʏɘɵɞɶ"

func isVowel(r rune) bool { return strings.ContainsRune(vowels, r) }

// Precomposed letters a nasal tilde or accent turns some vowels into.
var precomposed = map[rune]string{
	'ã': "ã", 'õ': "õ", 'ẽ': "ẽ", 'ĩ': "ĩ", 'ũ': "ũ",
	'á': "a", 'é': "e", 'í': "i", 'ó': "o", 'ú': "u", 'à': "a", 'è': "e", 'ì': "i", 'ò': "o", 'ù': "u",
	'g': "ɡ", 'ɹ': "r", 'ɾ': "r", 'ʁ': "r", 'ʀ': "r", 'ɽ': "r", 'ɫ': "l", 'ɚ': "ər", 'ɝ': "ɜr",
}

// pairs joins two sounds written as two letters into one.
var pairs = map[string]bool{
	"tʃ": true, "dʒ": true, "ts": true, "dz": true,
	"eɪ": true, "aɪ": true, "ɔɪ": true, "aʊ": true, "oʊ": true, "əʊ": true,
	"ɪə": true, "ɛə": true, "eə": true, "ʊə": true, "ɔʏ": true,
}

// parseIPA reads one word of IPA into sounds and markers.
func parseIPA(word string) ([]phone, error) {
	var b strings.Builder
	for _, r := range word {
		if s, ok := precomposed[r]; ok {
			b.WriteString(s)
		} else {
			b.WriteRune(r)
		}
	}
	var out []phone
	for _, r := range b.String() {
		last := len(out) - 1
		switch {
		case r == markPrimary || r == markSecondary || r == markBreak:
			out = append(out, phone{marker: r})
		case r == '̃': // nasal
			if last >= 0 && out[last].vowel {
				out[last].nasal = true
			}
		case r == '̯': // non-syllabic: a vowel used as a glide
			if last >= 0 && out[last].vowel {
				switch out[last].sym {
				case "i", "ɪ", "y":
					out[last] = phone{sym: "j"}
				case "u", "ʊ":
					out[last] = phone{sym: "w"}
				}
			}
		case r == '̩' || r == '̍': // syllabic consonant: n̩ is ən
			if last >= 0 && !out[last].vowel && out[last].marker == 0 {
				c := out[last]
				out[last] = phone{sym: "ə", vowel: true}
				out = append(out, c)
			}
		case r == 'ː' || r == 'ˑ' || r == '͡' || r == '͜' || r == '(' || r == ')' || r == '‿' || r == 'ʔ':
			// Length, tie bars, optional sounds and glottal stops: English
			// respelling doesn't mark them.
		case unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Lm, r):
			// Other diacritics and modifiers, such as aspiration.
		case unicode.IsLetter(r):
			p := phone{sym: string(r), vowel: isVowel(r)}
			if last >= 0 && out[last].marker == 0 && !out[last].nasal && pairs[out[last].sym+p.sym] {
				out[last].sym += p.sym
				continue
			}
			out = append(out, p)
		case unicode.IsSpace(r):
		default:
			return nil, fmt.Errorf("unexpected %q in IPA %q", r, word)
		}
	}
	return out, nil
}

// collapse drops the first of two identical consonants, and a t or d
// before the affricate it starts: Italian [ˈmatto] and [ˈpattʃo] are said
// with one consonant in English respelling.
func collapse(ps []phone) []phone {
	var out []phone
	for i, p := range ps {
		if p.marker == 0 && !p.vowel {
			next := i + 1
			for next < len(ps) && ps[next].marker != 0 {
				next++
			}
			if next < len(ps) && !ps[next].vowel {
				n := ps[next].sym
				if n == p.sym || (p.sym == "t" || p.sym == "d") && len(n) > len(p.sym) && strings.HasPrefix(n, p.sym) {
					continue
				}
			}
		}
		out = append(out, p)
	}
	return out
}

// onsets are consonant clusters that can start an English syllable.
var onsets = map[string]bool{
	"pl": true, "pr": true, "bl": true, "br": true, "tr": true, "dr": true, "kl": true, "kr": true,
	"ɡl": true, "ɡr": true, "fl": true, "fr": true, "θr": true, "ʃr": true, "sp": true, "st": true,
	"sk": true, "sm": true, "sn": true, "sw": true, "sl": true, "tw": true, "kw": true, "dw": true,
	"spr": true, "str": true, "skr": true, "spl": true, "skw": true,
}

type syllable struct {
	onset, coda []phone
	nucleus     phone
	stress      Stress
}

// syllables splits a word's sounds into syllables. Stress marks and breaks
// start a syllable; between vowels, the longest cluster that can start an
// English syllable goes to the next one.
func syllables(ps []phone) []syllable {
	// Chunks between markers, each starting with its marker's stress.
	type chunk struct {
		stress Stress
		ps     []phone
	}
	chunks := []chunk{{}}
	for _, p := range ps {
		if p.marker != 0 {
			c := chunk{}
			switch p.marker {
			case markPrimary:
				c.stress = Primary
			case markSecondary:
				c.stress = Secondary
			}
			chunks = append(chunks, c)
			continue
		}
		chunks[len(chunks)-1].ps = append(chunks[len(chunks)-1].ps, p)
	}

	var out []syllable
	for _, c := range chunks {
		var nuclei []int
		for i, p := range c.ps {
			if p.vowel {
				nuclei = append(nuclei, i)
			}
		}
		if len(nuclei) == 0 {
			// No vowel: the sounds end the syllable before, or start the next.
			if len(out) > 0 {
				out[len(out)-1].coda = append(out[len(out)-1].coda, c.ps...)
			} else if len(c.ps) > 0 {
				out = append(out, syllable{onset: c.ps, stress: c.stress})
			}
			continue
		}
		for k, n := range nuclei {
			s := syllable{nucleus: c.ps[n]}
			if k == 0 {
				s.onset = c.ps[:n]
				s.stress = c.stress
			} else {
				prev := nuclei[k-1]
				between := c.ps[prev+1 : n]
				split := len(between) // where the next onset starts
				if split > 0 {
					split--
					for split > 0 && onsets[symbols(between[split-1:])] {
						split--
					}
				}
				out[len(out)-1].coda = between[:split]
				s.onset = between[split:]
			}
			out = append(out, s)
		}
		last := nuclei[len(nuclei)-1]
		out[len(out)-1].coda = append(out[len(out)-1].coda, c.ps[last+1:]...)
	}
	// A leading onset with no vowel joins the first syllable.
	if len(out) > 1 && out[0].nucleus.sym == "" {
		out[1].onset = append(out[0].onset, out[1].onset...)
		out = out[1:]
	}
	return out
}

func symbols(ps []phone) string {
	var b strings.Builder
	for _, p := range ps {
		b.WriteString(p.sym)
	}
	return b.String()
}
