package respell

import (
	"reflect"
	"testing"
)

func TestFromIPA(t *testing.T) {
	for _, c := range []struct {
		ipa, lang string
		want      []string
	}{
		{"/ˌlæmbərˈɡiːni/", "", []string{"LAM-bər-GHEE-nee"}},
		{"[ferˈruttʃo lamborˈɡiːni]", "it", []string{"feh-ROO-choh", "lahm-bor-GHEE-nee"}},
		{"[ˈɛntso]", "it", []string{"EN-tsoh"}},
		{"[ferˈraːri]", "it", []string{"feh-RAH-ree"}},
		{"[maraˈnɛllo]", "it", []string{"mah-rah-NEH-loh"}},
		{"[ˈkuntaʃ]", "it", []string{"KOON-tahsh"}},
		{"/ˈkɪləˌmiːtər/", "en", []string{"KI-lə-MEE-tər"}},
		{"/kɪˈlɒmɪtər/", "", []string{"ki-LO-mi-tər"}},
		{"[xeˈsus]", "es", []string{"kheh-SOOS"}},
		{"/ˈaɪ.dl̩/", "", []string{"EYE-dəl"}},
		{"[ʃɑ̃ˈtal]", "fr", []string{"shahn-TAHL"}},
	} {
		words, err := FromIPA(c.ipa, c.lang)
		if err != nil {
			t.Errorf("%s: %v", c.ipa, err)
			continue
		}
		var got []string
		for _, w := range words {
			got = append(got, w.Respelling())
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("FromIPA(%q, %q) = %q, want %q", c.ipa, c.lang, got, c.want)
		}
	}
	if _, err := FromIPA("/ˈ€/", ""); err == nil {
		t.Error("a symbol that isn't IPA was accepted")
	}
}

func TestFromRespelling(t *testing.T) {
	words, err := FromRespelling("LAM-bər-GHEE-nee, En-Zoh")
	if err != nil {
		t.Fatal(err)
	}
	want := []Word{
		{Source: "LAM-bər-GHEE-nee", Syllables: []Syllable{{"lam", Primary}, {"bər", Unstressed}, {"ghee", Primary}, {"nee", Unstressed}}},
		{Source: "En-Zoh", Syllables: []Syllable{{"en", Unstressed}, {"zoh", Unstressed}}},
	}
	if !reflect.DeepEqual(words, want) {
		t.Errorf("FromRespelling = %+v, want %+v", words, want)
	}
}

func TestCandidates(t *testing.T) {
	for _, c := range []struct {
		respelling string
		want       []string
	}{
		{"EN-tsoh", []string{"Entsoh", "Entso", "En tsoh", "Ent soh"}},
		{"LAM-bər-GHEE-nee", []string{"Lambergheenee", "Lambergheeny", "Lamber gheenee", "Lamber gheeny", "Lam ber ghee nee"}},
		{"feh-ROO-choh", []string{"Fehroochoh", "Fehroocho", "Feh roochoh", "Feh roocho", "Feh roo choh", "Feroochoh"}},
		{"SOH-fə", []string{"Sohfuh", "Soh fuh", "Sohfa"}},
	} {
		words, err := FromRespelling(c.respelling)
		if err != nil {
			t.Fatal(err)
		}
		if got := Candidates(words[0]); !reflect.DeepEqual(got, c.want) {
			t.Errorf("Candidates(%s) = %q, want %q", c.respelling, got, c.want)
		}
	}
}

func TestSuggest(t *testing.T) {
	got, err := Suggest("/ˌlæmbərˈɡiːni/ LAM-bər-GHEE-nee, Italian: [ferˈruttʃo lamborˈɡiːni]", "")
	if err != nil {
		t.Fatal(err)
	}
	type row struct{ kind, lang, source, respelling string }
	var rows []row
	for _, s := range got {
		rows = append(rows, row{s.Kind, s.Lang, s.Source, s.Respelling})
		if len(s.Candidates) == 0 || len(s.Candidates) > MaxCandidates {
			t.Errorf("%s has %d candidates", s.Source, len(s.Candidates))
		}
	}
	want := []row{
		{"ipa", "", "ˌlæmbərˈɡiːni", "LAM-bər-GHEE-nee"},
		{"ipa", "it", "ferˈruttʃo", "feh-ROO-choh"},
		{"ipa", "it", "lamborˈɡiːni", "lahm-bor-GHEE-nee"},
		{"respelling", "", "LAM-bər-GHEE-nee", "LAM-bər-GHEE-nee"},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Errorf("Suggest rows = %+v, want %+v", rows, want)
	}

	// Bare input: IPA when it has IPA letters, in the language given.
	for _, c := range []struct{ in, lang, kind, respelling string }{
		{"ˈɛntso", "it", "ipa", "EN-tsoh"},
		{"En-Zoh", "", "respelling", "en-zoh"},
		{"Entso", "", "respelling", "entso"},
	} {
		got, err := Suggest(c.in, c.lang)
		if err != nil || len(got) != 1 || got[0].Kind != c.kind || got[0].Respelling != c.respelling {
			t.Errorf("Suggest(%q) = %+v, %v", c.in, got, err)
		}
	}
}
