# hardenme

[English](README.md) | **Español**

> Comprueba lo seguro que está tu dispositivo, en un lenguaje que se entiende, y prepárate para el día en que te lo roben.

`hardenme` es una herramienta gratuita y de código abierto para la terminal pensada para **gente sin conocimientos técnicos**. Revisa los ajustes de seguridad de tu dispositivo, te da una **nota del 0 al 100** (de la A a la F) y, para cada problema, te explica **por qué importa y cómo arreglarlo**, en español o en inglés.

La nota depende de **tu situación**. Cifrar el disco es imprescindible en un portátil que llevas a todas partes, pero opcional en un ordenador de sobremesa que nunca sale de casa. `hardenme` te hace unas pocas preguntas sencillas y le da a cada comprobación el peso que le toca.

> **Estado del proyecto: en desarrollo inicial, solo macOS.** Android, Windows e iOS están en la [hoja de ruta](#hoja-de-ruta).

## Por qué

Hay herramientas de seguridad muy buenas, pero ninguna junta todo esto:

- **Multiplataforma**: una sola herramienta para tu Mac, tu portátil con Windows y tu móvil.
- **Pensada para gente no técnica**: sin jerga, con pasos claros y enlaces directos al ajuste que hay que tocar.
- **Una nota que tiene en cuenta tu perfil**: lo que importa en el portátil de alguien que va en metro cada día no es lo mismo que en el ordenador familiar del salón.
- **Centrada en el robo**: qué preparar antes de que te roben el dispositivo, y qué hacer en los primeros minutos, horas y días después.
- **En español y en inglés.**

## Lo que prometemos sobre tu privacidad

- **Solo lectura.** Únicamente lee ajustes. Nunca cambia nada por su cuenta.
- **Sin sudo ni contraseña de administrador.** Nunca.
- **Sin red.** No envía nada a ningún sitio. Ni telemetría, ni cuentas, ni nube.
- **Código abierto** (MIT), para que cualquiera pueda comprobar estas promesas.

## Instalación

Los binarios listos para descargar llegarán en la fase 3. De momento necesitas tener [Go](https://go.dev/dl/) instalado:

```sh
go install github.com/XexusCode/hardenme@latest
```

O clona el repositorio y ejecútalo directamente:

```sh
git clone https://github.com/XexusCode/hardenme.git
cd hardenme
go run . scan
```

## Uso

```sh
# Cuéntale a hardenme tu situación (sin opciones, te pregunta paso a paso)
hardenme profile
hardenme profile --portable --leaves-home --sensitive-data --risk normal

# Revisa este dispositivo
hardenme scan
hardenme scan --lang es
hardenme scan --only filevault,firewall

# Lee las guías sobre robos en la terminal
hardenme guide theft-before --lang es
hardenme guide theft-after --lang es
```

Tu perfil se guarda en `~/.config/hardenme/profile.json`. Las preguntas son:

| Opción | Pregunta |
|---|---|
| `--portable` | ¿Es un portátil? (se detecta solo cuando se puede) |
| `--leaves-home` | ¿Lo sacas de casa? |
| `--shared` | ¿Lo usa más gente? |
| `--sensitive-data` | ¿Tiene datos del trabajo o del banco? |
| `--risk normal\|elevated` | ¿Es probable que alguien vaya a por ti en concreto? |

### Ejemplo de resultado

Un fragmento de una ejecución real en un MacBook (macOS 26.2):

```text
hardenme · macOS 26.2
Perfil: portátil, sale de casa, riesgo normal

Nota de seguridad: 57/100 (F, En riesgo)
7 por arreglar, 1 sugerencias, 2 para revisar tú

ARREGLA PRIMERO (esencial para ti) (5)

✗ Activa el cifrado del disco (FileVault)
  Ahora: FileVault is Off.
  Por qué importa:
    Sin FileVault, quien robe tu Mac puede sacar el disco o arrancarlo desde
    otra unidad y leer todos tus archivos, contraseñas y fotos sin saber tu
    contraseña. [...]
  Cómo arreglarlo:
    1. Abre Ajustes del Sistema > Privacidad y seguridad.
    2. Baja hasta FileVault y pulsa "Activar". [...]
  Abrir: open "x-apple.systempreferences:com.apple.settings.PrivacySecurity.extension"

✗ Instala la actualización de macOS pendiente
  Ahora: 26.2, pending: macOS Tahoe 26.7.2
  [...]

REVÍSALO TÚ (2)
? Comprueba que Buscar mi Mac está activado
  [...]

YA ESTÁ BIEN (10)
✓ Gatekeeper revisa las apps que abres
✓ macOS se actualiza solo
[...]
```

### Códigos de salida

| Código | Significado |
|---|---|
| `0` | Todo lo que te aplica está bien |
| `1` | Ha fallado al menos una comprobación imprescindible o recomendada |
| `2` | Solo hay avisos o cosas opcionales |

## Cómo se calcula la nota

Cada comprobación tiene un **peso** según tu perfil:

| Importancia para ti | Peso |
|---|---|
| Imprescindible | 3 |
| Recomendada | 2 |
| Opcional | 1 |
| No aplica | 0 |

La nota es lo que has conseguido entre lo que podías conseguir, por 100:

```
nota = Σ(conseguido × peso) / Σ(máximo × peso) × 100
```

Letras: **A** 90 o más, **B** de 80 a 89, **C** de 70 a 79, **D** de 60 a 69, **F** menos de 60.

Algunas comprobaciones dan puntos a medias. Las que no se han podido leer (**desconocido**) o que tienes que comprobar tú (**manual**) **no cuentan**, ni a favor ni en contra. Una comprobación que no se puede ejecutar nunca se da por buena.

## Guías

- [Antes de que te roben el dispositivo](docs/guide/es/theft-before.md)
- [Me han robado el dispositivo: qué hago ahora](docs/guide/es/theft-after.md)

## Hoja de ruta

- **Fase 1 (ahora):** revisión de macOS, nota según tu perfil y guías sobre robos en español e inglés.
- **Fase 2:** revisión de Android por USB (`adb`, sin root), resultados en JSON y Markdown, e "historias de robo" que explican qué podría hacer un ladrón con tu configuración actual.
- **Fase 3:** revisión de Windows, lista guiada para iOS, binarios para descargar y fórmula de Homebrew.

## Cómo colaborar

Toda ayuda es bienvenida, sobre todo comprobaciones nuevas, correcciones y traducciones. Y puedes escribirnos en español sin problema. Lee [CONTRIBUTING.md](CONTRIBUTING.md) y nuestro [Código de conducta](CODE_OF_CONDUCT.md). Para avisar de un fallo de seguridad, mira [SECURITY.md](SECURITY.md).

## Créditos

`hardenme` se apoya en el trabajo de muchos proyectos de código abierto. Los tienes en [CREDITS.md](CREDITS.md).

## Licencia

[MIT](LICENSE)
