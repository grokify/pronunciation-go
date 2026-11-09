# Pronunciation Go

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go Reference][reference-svg]][reference-url]

Go packages for teaching speech engines how to say words.

Browser voices and many text-to-speech engines take plain text only, with no
phonetic markup, so the way to fix a name they say wrong is to respell it:
"Lamborgheeny" for "Lamborghini". Finding a respelling a voice reads as
intended takes listening. `respell` turns what dictionaries publish, IPA and
respellings, into several candidates to play and choose from.

## Packages

| Package | Purpose |
|---------|---------|
| [`respell`](respell/) | IPA and dictionary respellings to respellings for speech engines |

## respell

```go
import "github.com/grokify/pronunciation-go/respell"

// Text as Wikipedia writes it: English IPA, its respelling, and IPA in
// another language.
got, err := respell.Suggest("/ˌlæmbərˈɡiːni/ LAM-bər-GHEE-nee, Italian: [ferˈruttʃo lamborˈɡiːni]", "")
if err != nil {
	return err
}
for _, s := range got {
	fmt.Println(s.Source, s.Respelling, s.Candidates)
}
```

| Input | Respelling | Candidates |
|-------|------------|------------|
| `/ˌlæmbərˈɡiːni/` | LAM-bər-GHEE-nee | Lambergheenee, Lambergheeny, Lamber gheenee, … |
| Italian `[ˈɛntso]` | EN-tsoh | Entsoh, Entso, En tsoh, Ent soh |
| Italian `[ferˈruttʃo]` | feh-ROO-choh | Fehroochoh, Fehroocho, Feh roochoh, … |
| Spanish `[xeˈsus]` | kheh-SOOS | Khehsoos, Kheh soos, Khesoos |

- **IPA** is read syllable by syllable into a respelling in the style of
  [Wikipedia's pronunciation respelling key][respelling-key], stressed
  syllables in capitals. IPA between slashes is English; IPA in brackets is
  in the language named before it, as Wikipedia writes `Italian: [...]`.
- **Other languages** are brought to the nearest English sounds: pure
  vowels written long ("oh", not the "o" of "lot"), doubled consonants
  single, nasal vowels with an n. A rolled r or a doubled consonant is lost.
- **Candidates** write the same sounds several ways, since voices read
  unfamiliar words by their own rules: as one word, split where the stress
  falls, syllable by syllable, with simpler endings, and with a consonant
  cluster split. Which one a voice says best is found by listening.

`FromIPA` and `FromRespelling` return the syllables, and `Candidates` the
respellings for one word, for callers that want them separately.

## Development

```sh
go test ./...
golangci-lint run
```

## License

MIT

[go-ci-svg]: https://github.com/grokify/pronunciation-go/actions/workflows/go-ci.yaml/badge.svg?branch=main
[go-ci-url]: https://github.com/grokify/pronunciation-go/actions/workflows/go-ci.yaml
[go-lint-svg]: https://github.com/grokify/pronunciation-go/actions/workflows/go-lint.yaml/badge.svg?branch=main
[go-lint-url]: https://github.com/grokify/pronunciation-go/actions/workflows/go-lint.yaml
[reference-svg]: https://pkg.go.dev/badge/github.com/grokify/pronunciation-go
[reference-url]: https://pkg.go.dev/github.com/grokify/pronunciation-go
[respelling-key]: https://en.wikipedia.org/wiki/Help:Pronunciation_respelling_key
