# hardenme

[English](README.md) | **Español**

> Comprueba lo seguro que está tu dispositivo, en un lenguaje que se entiende, y prepárate para el día en que te lo roben.

`hardenme` es una herramienta gratuita y de código abierto para la terminal pensada para **gente sin conocimientos técnicos**. Revisa los ajustes de seguridad de tu dispositivo, te da una **nota del 0 al 100** (de la A a la F) y, para cada problema, te explica **por qué importa y cómo arreglarlo**, en español o en inglés.

La nota depende de **tu situación**. Cifrar el disco es imprescindible en un portátil que llevas a todas partes, pero opcional en un ordenador de sobremesa que nunca sale de casa. `hardenme` te hace unas pocas preguntas sencillas y le da a cada comprobación el peso que le toca.

> **Estado del proyecto: en desarrollo inicial, solo macOS.** Android, Windows e iOS están en la [hoja de ruta](#hoja-de-ruta).
>
> **Esta página describe el código de `main`, que va por delante de la última versión publicada (v0.1.1).** El instalador te da la v0.1.1: la revisión en Terminal, sin la página del navegador, sin `fix` y sin la revisión del móvil. Eso llega con la próxima versión.
>
> **Beta: los arreglos automáticos.** `hardenme fix` y el botón "Arréglalo todo" están cubiertos por tests automáticos, pero aplicarlos y deshacerlos todavía no se ha probado en un Mac real. `hardenme fix --dry-run` enseña los comandos exactos sin cambiar nada. La revisión del móvil también es beta: las rutas de los menús salen de las páginas de ayuda de Apple y de Google y no se han recorrido en móviles reales.

## Por qué

Hay herramientas de seguridad muy buenas, pero ninguna junta todo esto:

- **Multiplataforma**: una sola herramienta para tu Mac, tu portátil con Windows y tu móvil.
- **Pensada para gente no técnica**: sin jerga, con pasos claros y enlaces directos al ajuste que hay que tocar.
- **Una nota que tiene en cuenta tu perfil**: lo que importa en el portátil de alguien que va en metro cada día no es lo mismo que en el ordenador familiar del salón.
- **Señales de compromiso**: en el Mac también te enseña cosas que conviene que reconozcas, como apps de acceso remoto, programas sin firma que arrancan solos, perfiles de configuración o administradores de más. Son preguntas con su porqué, no veredictos, y hardenme no es un antivirus.
- **Centrada en el robo**: qué preparar antes de que te roben el dispositivo, y qué hacer en los primeros minutos, horas y días después.
- **En español y en inglés.**

## Lo que prometemos sobre tu privacidad

- **La revisión solo lee.** `hardenme scan` nunca cambia nada ni pide contraseña.
- **Arreglos solo con tu permiso** (desde la v0.2). `hardenme fix` cambia solo lo que eliges, después de enseñarte qué va a cambiar. Cuando un cambio necesita tu contraseña, te la pide macOS en su propia ventana: hardenme no la ve. Todo cambio automático se puede deshacer con `hardenme fix --undo`.
- **Sin red.** No envía nada a ningún sitio. Ni telemetría, ni cuentas, ni nube.
- **Código abierto** (MIT), para que cualquiera pueda comprobar estas promesas.

## Instalación

No hace falta saber programar. Solo usarás Terminal **una vez**, para pegar una línea. Después, hardenme es una app que abres con doble clic.

1. **Abre Terminal.** Pulsa `Comando (⌘) + Espacio`, escribe `Terminal` y pulsa `Intro`. Aparece una ventana con texto.
2. **Copia esta línea, pégala en esa ventana y pulsa `Intro`:**

   ```sh
   curl -fsSL https://github.com/XexusCode/hardenme/releases/latest/download/install.sh | sh
   ```

3. **Espera unos segundos.** Verás cuatro pasos (descargar, comprobar, instalar y crear la app) y después "¡Listo!". No pide tu contraseña y no instala nada fuera de tu carpeta de usuario o de Aplicaciones.
4. **Se abre una ventana del Finder con hardenme seleccionado. Haz doble clic.** Se abre una página en tu navegador con tu nota, lo que podría pasar hoy y todo lo que se puede mejorar, cada cosa con su porqué. Un botón arregla lo que hardenme puede cambiar por ti (macOS te pide la contraseña una vez). Para lo demás, te abre el ajuste exacto y lo marca solo cuando lo cambias.

A partir de ahí, abre **hardenme** desde Aplicaciones o Spotlight cuando quieras revisar otra vez.

> **¿Es seguro pegar un comando?** Solo si confías en su origen. Puedes [leer el instalador](install.sh) antes: es corto y está comentado. Descarga la versión publicada en GitHub, comprueba su checksum SHA-256 y se para si algo no coincide.

**Desinstalar:**

```sh
curl -fsSL https://github.com/XexusCode/hardenme/releases/latest/download/install.sh | sh -s -- --uninstall
```

**Para desarrolladores:** `go install github.com/XexusCode/hardenme@latest`, o clona el repositorio y ejecuta `go run . scan`.

## Uso

```sh
# Cuéntale a hardenme tu situación (sin opciones, te pregunta paso a paso)
hardenme profile
hardenme profile --portable --leaves-home --sensitive-data --risk normal

# Revisa este dispositivo
hardenme scan
hardenme scan --lang es
hardenme scan --only filevault,firewall

# Mira tu nota en el navegador y arregla las cosas en un par de clics (desde la v0.2)
hardenme ui

# O hazlo en Terminal: elige las mejoras y deja que hardenme las aplique
hardenme fix
hardenme fix --dry-run      # enseña los comandos exactos, no cambia nada
hardenme fix --undo         # deshace la última tanda de cambios

# Revisa tu móvil contestando unas preguntas (desde la v0.3)
hardenme checklist ios
hardenme checklist android
hardenme checklist ios --again    # pregunta todo otra vez, no solo lo que estaba mal

# Guarda o comparte el informe (desde la v0.3)
hardenme scan --format markdown > mi-mac.md
hardenme scan --format json

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
- [Qué pueden hacer con tu móvil bloqueado](docs/guide/es/phone-locked.md)
- [¿Pueden desbloquear mi móvil?](docs/guide/es/phone-unlock.md)
- [Huella y cara: cuánto protegen](docs/guide/es/biometrics.md)
- [¿Es seguro tener la app del banco en el móvil?](docs/guide/es/bank-app.md)

## Hoja de ruta

- **v0.1 (ahora):** revisión de macOS, nota según tu perfil, guías sobre robos en español e inglés, e instalador de una línea con app de doble clic.
- **v0.2:** eliges qué mejoras quieres y hardenme las aplica. Escribes tu contraseña una vez en la ventana de macOS, ves tu nota nueva y puedes deshacerlo todo. Los ajustes que te toca cambiar a ti (como FileVault) te los abre en la página exacta y los comprueba cuando terminas.
- **v0.3:** lista guiada para iPhone y Android, "historias de robo" que explican qué podría hacer un ladrón con tu configuración actual, y resultados en JSON y Markdown.
- **v0.4:** revisión automática de Android por USB (`adb`, sin root) para usuarios avanzados.
- **Más adelante:** revisión de Windows, lectura de ajustes del iPhone por USB y fórmula de Homebrew.

## Cómo colaborar

Toda ayuda es bienvenida, sobre todo comprobaciones nuevas, correcciones y traducciones. Y puedes escribirnos en español sin problema. Lee [CONTRIBUTING.md](CONTRIBUTING.md) y nuestro [Código de conducta](CODE_OF_CONDUCT.md). Para avisar de un fallo de seguridad, mira [SECURITY.md](SECURITY.md).

## Créditos

`hardenme` se apoya en el trabajo de muchos proyectos de código abierto. Los tienes en [CREDITS.md](CREDITS.md).

## Licencia

[MIT](LICENSE)
