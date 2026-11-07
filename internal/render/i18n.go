package render

import "github.com/XexusCode/hardenme/internal/check"

// messages holds the report's own strings (check texts live in the YAML).
var messages = map[check.Lang]map[string]string{
	check.EN: {
		"profile":         "Profile",
		"laptop":          "laptop",
		"desktop":         "desktop",
		"leaves_home":     "goes outside",
		"stays_home":      "stays at home",
		"shared":          "shared",
		"sensitive":       "work/banking data",
		"risk_normal":     "normal risk",
		"risk_elevated":   "elevated risk",
		"score":           "Security score",
		"no_score":        "Security score: not enough data to score",
		"summary":         "%d to fix, %d suggestions, %d to check yourself",
		"sec_essential":   "FIX FIRST (essential for you)",
		"sec_recommended": "RECOMMENDED",
		"sec_optional":    "SUGGESTIONS (optional for you)",
		"sec_manual":      "CHECK IT YOURSELF",
		"sec_unknown":     "COULD NOT CHECK",
		"sec_pass":        "ALREADY GOOD",
		"sec_na":          "NOT APPLICABLE TO YOUR PROFILE",
		"now":             "Now",
		"why":             "Why it matters",
		"how":             "How to fix it",
		"open":            "Open",
		"command":         "Or in Terminal",
		"edit_profile":    "Not right? Change your profile with: hardenme profile",
		"label_A":         "Excellent",
		"label_B":         "Good",
		"label_C":         "Fair",
		"label_D":         "Needs work",
		"label_F":         "At risk",
		"scoring_note":    "Checks are weighted by your profile: essential x3, recommended x2, optional x1. Items you have to check yourself do not count.",
	},
	check.ES: {
		"profile":         "Perfil",
		"laptop":          "portátil",
		"desktop":         "sobremesa",
		"leaves_home":     "sale de casa",
		"stays_home":      "no sale de casa",
		"shared":          "compartido",
		"sensitive":       "datos de trabajo/banco",
		"risk_normal":     "riesgo normal",
		"risk_elevated":   "riesgo elevado",
		"score":           "Nota de seguridad",
		"no_score":        "Nota de seguridad: no hay datos suficientes para puntuar",
		"summary":         "%d por arreglar, %d sugerencias, %d para revisar tú",
		"sec_essential":   "ARREGLA PRIMERO (esencial para ti)",
		"sec_recommended": "RECOMENDADO",
		"sec_optional":    "SUGERENCIAS (opcional para ti)",
		"sec_manual":      "REVÍSALO TÚ",
		"sec_unknown":     "NO SE HA PODIDO COMPROBAR",
		"sec_pass":        "YA ESTÁ BIEN",
		"sec_na":          "NO APLICA A TU PERFIL",
		"now":             "Ahora",
		"why":             "Por qué importa",
		"how":             "Cómo arreglarlo",
		"open":            "Abrir",
		"command":         "O en Terminal",
		"edit_profile":    "¿No es correcto? Cambia tu perfil con: hardenme profile",
		"label_A":         "Excelente",
		"label_B":         "Bien",
		"label_C":         "Aceptable",
		"label_D":         "Mejorable",
		"label_F":         "En riesgo",
		"scoring_note":    "Cada punto pesa según tu perfil: esencial x3, recomendado x2, opcional x1. Lo que tienes que revisar tú no cuenta.",
	},
}

func msg(lang check.Lang, key string) string {
	if s, ok := messages[lang][key]; ok {
		return s
	}
	return messages[check.EN][key]
}
