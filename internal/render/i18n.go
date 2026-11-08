package render

import "github.com/XexusCode/hardenme/internal/check"

// messages holds the report's own strings (check texts live in the YAML).
var messages = map[check.Lang]map[string]string{
	check.EN: {
		"profile":            "Profile",
		"laptop":             "laptop",
		"desktop":            "desktop",
		"phone":              "phone",
		"leaves_home":        "goes outside",
		"stays_home":         "stays at home",
		"shared":             "shared",
		"sensitive":          "work/banking data",
		"risk_normal":        "normal risk",
		"risk_elevated":      "elevated risk",
		"score":              "Security score",
		"no_score":           "Security score: not enough data to score",
		"summary":            "%d to fix, %d %s, %d to check yourself",
		"suggestion_one":     "suggestion",
		"suggestion":         "suggestions",
		"sec_essential":      "FIX FIRST (essential for you)",
		"sec_recommended":    "RECOMMENDED",
		"sec_optional":       "SUGGESTIONS (optional for you)",
		"sec_manual":         "CHECK IT YOURSELF",
		"sec_unknown":        "COULD NOT CHECK",
		"sec_pass":           "ALREADY GOOD",
		"sec_na":             "NOT APPLICABLE TO YOUR PROFILE",
		"now":                "Now",
		"why":                "Why it matters",
		"how":                "How to fix it",
		"open":               "Open",
		"command":            "Or in Terminal",
		"edit_profile":       "Not right? Change your profile with: hardenme profile",
		"edit_profile_phone": "Not right? Answer the questions about your phone again with: hardenme checklist %s --profile",
		"next_phone":         "When you change something, run hardenme checklist %s again: I will only ask what was not right.",
		"label_A":            "Excellent",
		"label_B":            "Good",
		"label_C":            "Fair",
		"label_D":            "Needs work",
		"label_F":            "At risk",
		"scoring_note":       "Checks are weighted by your profile: essential x3, recommended x2, optional x1. Items you have to check yourself do not count.",
	},
	check.ES: {
		"profile":            "Perfil",
		"laptop":             "portátil",
		"desktop":            "sobremesa",
		"phone":              "móvil",
		"leaves_home":        "sale de casa",
		"stays_home":         "no sale de casa",
		"shared":             "compartido",
		"sensitive":          "datos de trabajo/banco",
		"risk_normal":        "riesgo normal",
		"risk_elevated":      "riesgo elevado",
		"score":              "Nota de seguridad",
		"no_score":           "Nota de seguridad: no hay datos suficientes para puntuar",
		"summary":            "%d por arreglar, %d %s, %d para revisar tú",
		"suggestion_one":     "sugerencia",
		"suggestion":         "sugerencias",
		"sec_essential":      "ARREGLA PRIMERO (esencial para ti)",
		"sec_recommended":    "RECOMENDADO",
		"sec_optional":       "SUGERENCIAS (opcional para ti)",
		"sec_manual":         "REVÍSALO TÚ",
		"sec_unknown":        "NO SE HA PODIDO COMPROBAR",
		"sec_pass":           "YA ESTÁ BIEN",
		"sec_na":             "NO APLICA A TU PERFIL",
		"now":                "Ahora",
		"why":                "Por qué importa",
		"how":                "Cómo arreglarlo",
		"open":               "Abrir",
		"command":            "O en Terminal",
		"edit_profile":       "¿No es correcto? Cambia tu perfil con: hardenme profile",
		"edit_profile_phone": "¿No es correcto? Vuelve a contestar las preguntas sobre tu móvil con: hardenme checklist %s --profile",
		"next_phone":         "Cuando cambies algo, vuelve a ejecutar hardenme checklist %s: solo te preguntaré lo que no estaba bien.",
		"label_A":            "Excelente",
		"label_B":            "Bien",
		"label_C":            "Aceptable",
		"label_D":            "Mejorable",
		"label_F":            "En riesgo",
		"scoring_note":       "Cada punto pesa según tu perfil: esencial x3, recomendado x2, opcional x1. Lo que tienes que revisar tú no cuenta.",
	},
}

func msg(lang check.Lang, key string) string {
	if s, ok := messages[lang][key]; ok {
		return s
	}
	return messages[check.EN][key]
}

// plural returns the singular form (key + "_one") when n is 1.
func plural(lang check.Lang, key string, n int) string {
	if n == 1 {
		return msg(lang, key+"_one")
	}
	return msg(lang, key)
}
