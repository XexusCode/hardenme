package cli

import "github.com/XexusCode/hardenme/internal/check"

var cliMessages = map[check.Lang]map[string]string{
	check.EN: {
		"first_run":          "First run: a few questions so the score fits how you use this device.",
		"no_profile_default": "No saved profile: using defaults based on this device. Run `hardenme profile` to answer the questions.",
		"saved":              "Profile saved in %s",
		"q_portable":         "Is this device a laptop?",
		"q_leaves_home":      "Do you take it outside your home (work, travel, cafés)?",
		"q_shared":           "Do other people use it (family, friends, coworkers)?",
		"q_sensitive":        "Does it hold work data or access your bank?",
		"q_risk":             "Could someone target you specifically (journalist, activist, harassment, public figure)?",
		"yes_no_default_yes": "[Y/n]",
		"yes_no_default_no":  "[y/N]",
		"guide_pages":        "Available guide pages:",
	},
	check.ES: {
		"first_run":          "Primera vez: unas preguntas para que la nota se ajuste a cómo usas este dispositivo.",
		"no_profile_default": "No hay perfil guardado: uso valores por defecto según este dispositivo. Ejecuta `hardenme profile` para responder las preguntas.",
		"saved":              "Perfil guardado en %s",
		"q_portable":         "¿Este dispositivo es un portátil?",
		"q_leaves_home":      "¿Lo sacas de casa (trabajo, viajes, cafeterías)?",
		"q_shared":           "¿Lo usan otras personas (familia, amigos, compañeros)?",
		"q_sensitive":        "¿Tiene datos del trabajo o acceso a tu banco?",
		"q_risk":             "¿Alguien podría ir a por ti en concreto (periodista, activista, acoso, persona pública)?",
		"yes_no_default_yes": "[S/n]",
		"yes_no_default_no":  "[s/N]",
		"guide_pages":        "Páginas de la guía disponibles:",
	},
}

func t(lang check.Lang, key string) string {
	if s, ok := cliMessages[lang][key]; ok {
		return s
	}
	return cliMessages[check.EN][key]
}
