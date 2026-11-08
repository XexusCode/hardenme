# hardenme: estado actual y roadmap

Fecha: 2026-11-06, actualizado el 2026-11-07. Documento interno para decidir en qué centrarse. Revisado con el mantenedor.

## 1. Lo que tenemos

| Área | Estado | Detalle |
|---|---|---|
| Investigación | ✅ Hecha | 13 proyectos analizados, licencias revisadas, catálogo de checks por sistema (`docs/research/prior-art.md`) |
| Plan | ✅ Hecho | 3 fases con contratos (`*-plan.md`). Commit `cceee85` |
| Auditor macOS | ✅ Funciona, sin commit | 21 checks en YAML, 6 de ellos con lógica en Go. Probado en un MacBook real con macOS 26.2 |
| Nota por perfil | ✅ Funciona | Pesos 3/2/1/0. Los checks no leídos no cuentan. Probado con el caso portátil frente a sobremesa |
| Textos EN/ES | ✅ | Cada check explica por qué importa, cómo arreglarlo y enlaza a Ajustes |
| Guía antes/después del robo | ✅ Borrador | EN/ES, con apartado de España. **Los menús de móvil no están verificados** |
| Docs open source | ✅ | README, CONTRIBUTING, CODE_OF_CONDUCT, SECURITY, LICENSE, CREDITS y AGENTS |
| Tests | ✅ 34 tests | Cobertura: scan 100 %, check 89 %, render 86 %, profile 65 %, cli 22 %, runner 0 % |
| Tamaño | | ~2.600 líneas de Go y ~875 líneas de guía y README |

## 2. Lo que falta

### Bloqueantes para publicar
- Commit de la fase 1.
- Confirmar el usuario u organización de GitHub (`XexusCode`?), crear el repo público y hacer push.
- CI mínima: un GitHub Action que ejecute `gofmt`, `vet` y `test` en macOS.
- Una primera release que se pueda instalar. Hoy solo funciona `go install` o `go run`, lo que exige tener Go.

### Huecos de producto (por orden de importancia)
1. **El público y la forma de uso no encajan.** El usuario objetivo no es técnico, pero hoy la herramienta es una CLI que necesita Go y Terminal. Ese usuario no la va a usar.
2. **El móvil es lo que más se roba y aún no está cubierto.** Hoy solo hay Mac. Además, Android por adb exige activar la depuración USB, otra barrera técnica.
3. **No hay salida JSON ni Markdown**, así que no se puede compartir ni comparar el informe.
4. **No hay `--fix`.** El usuario tiene que ir a Ajustes él solo, aunque tiene el enlace directo. **Pasa a ser la v0.2** (decisión del 2026-11-07).
5. **Faltan "historias de robo"**: frases que unan varios fallos en un riesgo concreto. Es la idea más diferencial y aún no existe.
6. **Faltan Windows e iOS.**

### Deuda de verificación
- Buscar mi Mac: no hay forma fiable de leerlo, así que hoy sale siempre como "revísalo tú".
- No se ha visto en un equipo real el estado "activado" de Compartir pantalla, Compartir archivos y Gestión remota.
- No se han probado los fixtures de versiones antiguas de macOS (14 y 15) ni los enlaces a Ajustes.
- Los menús de iOS y Android de la guía, sobre todo en español, están sin revisar en un dispositivo.

### Deuda técnica pequeña
- Los tests de `cli` y `runner` son escasos (22 % y 0 %).
- Con estado "no se pudo comprobar", el informe muestra el título de fallo ("Vuelve a activar SIP"), lo que puede confundir. Debería usar un título neutro.
- Falta concordancia en los textos: "1 sugerencias".

## 3. Riesgos

| Riesgo | Impacto | Cómo mitigarlo |
|---|---|---|
| Nadie no técnico lo usa porque es una CLI | Alto: invalida el objetivo | Instalador de un comando con instrucciones y capturas, lanzador de doble clic, y probarlo con personas no técnicas antes de cada versión |
| Gatekeeper bloquea el binario sin firmar | Alto: el usuario cree que es un virus | Se instala con `curl`, que no pone cuarentena, y la app se genera en local. Sin `.pkg` descargable sin firmar |
| Pegar `curl ... \| sh` da desconfianza | Medio | Script corto y legible, enlazado en el README, con verificación SHA-256 y sin sudo |
| Falsos "todo bien" | Alto: la confianza es el producto | La regla de "nunca dar por buena una lectura fallida" ya está aplicada. Hay que añadir pruebas con dispositivos reales y alguien que lo revise |
| Abarcar 4 sistemas a la vez | Medio: nada queda terminado | Un sistema terminado y publicado antes de empezar el siguiente |
| Los cambios de macOS y Android rompen probes | Medio | Fixtures por versión, CI en macOS e issues de "falso positivo" |
| Licencias | Bajo | Ya documentado: CREDITS, third_party y las reglas de CONTRIBUTING |

## 4. Roadmap

**Principio:** cada versión tiene que poder usarla alguien real. Mejor terminar un sistema y publicarlo que tener cuatro a medias.

**Métrica norte:** que una persona no técnica pase de "no sé si estoy protegido" a "he arreglado mis 3 cosas más importantes" sin ayuda de nadie.

### Decisiones tomadas (2026-11-06)
- **Público:** gente con pocos conocimientos técnicos. **La instalación fácil es un requisito, no un extra.**
- **Formato:** la CLI es el producto principal. **La web queda para el final.** Un navegador no puede leer los ajustes del dispositivo, así que la web solo sería un checklist de autoevaluación, y eso aporta poco frente a lo que ya existe.
- **Orden:** primero el checklist guiado del móvil; Android por adb queda para después.
- **Sin estimaciones de tiempo:** el roadmap se ordena por prioridad, no por fechas.
- **Aplicar mejoras antes que el móvil** (2026-11-07): el usuario elige qué mejoras quiere y hardenme las aplica. Completa la experiencia en Mac, que es lo único que existe hoy. El checklist del móvil pasa a la v0.3. Plan: [`2026_11_07-hardenme_fix-plan.md`](../2026_11_07-hardenme_fix/2026_11_07-hardenme_fix-plan.md).

### Ahora (v0.1): publicar el Mac con instalación fácil
- [x] Commit de la fase 1 (`8c46765`).
- [x] Crear el repo público en GitHub y hacer push ([XexusCode/hardenme](https://github.com/XexusCode/hardenme)). Releases v0.1.0 y v0.1.1 (arreglo del idioma cuando falta `LANG`).
- [x] CI con GitHub Actions: `gofmt`, `vet` y `test` en `macos-latest`.
- [x] Release con GoReleaser: binario universal de macOS, `checksums.txt` e `install.sh` (Homebrew queda para más adelante). Se adelanta de la fase 3 porque sin esto nadie puede usarlo.
- [ ] **Instalación sin darse de alta en ningún sitio** (decidido el 2026-11-07: nada de Apple Developer Program). Por eso **no hay `.pkg` notarizado**: un `.pkg` sin firmar descargado desde el navegador lo bloquea Gatekeeper, y para abrirlo hay que ir a Ajustes > Privacidad y seguridad > "Abrir igualmente", que parece malware. En su lugar:
  - [x] **Instalador de un solo comando** (`install.sh`, publicado en el repo y en cada release):
    - Lo que se descarga con `curl` no lleva la marca de cuarentena, así que Gatekeeper no lo bloquea.
    - El script descarga el binario universal de la release, comprueba su checksum SHA-256 y lo instala en `~/.local/bin` (sin sudo).
    - Crea la app lanzadora `~/Applications/hardenme.app`. Al generarse en local, tampoco lleva cuarentena.
    - Explica en EN/ES qué hace en cada paso.
  - [x] **Lanzador de doble clic** `hardenme.app`: abre Terminal y ejecuta `hardenme scan`. Es lo que usará la persona no técnica después de instalar: Terminal solo se usa una vez, para pegar el comando de instalación.
  - [x] **README con la instalación paso a paso** (las capturas, cuando exista la primera release): cómo abrir Terminal, qué pegar y qué vas a ver. Con un enlace al script, para quien quiera leerlo antes.
  - [ ] **Homebrew** (`brew install XexusCode/tap/hardenme`) para usuarios técnicos. Solo necesita GitHub, que ya existe.
  - [ ] Probar la instalación en un Mac limpio, sin Go ni Homebrew, con una persona no técnica.
- [x] Arreglos pequeños: título neutro para "no se pudo comprobar" y "no aplica", y concordancia de plurales ("1 sugerencia").
- [ ] Ejecutarlo en 2 o 3 Macs de personas no técnicas (amigos o familia, con macOS 14 y 15), sin ayuda, y anotar dónde se atascan.

**Hecho cuando:** una persona no técnica instala hardenme siguiendo solo las instrucciones y entiende el informe sin ayuda.

### Siguiente (v0.2): elegir mejoras y aplicarlas
Plan detallado: [`2026_11_07-hardenme_fix-plan.md`](../2026_11_07-hardenme_fix/2026_11_07-hardenme_fix-plan.md).
- [ ] Verificar en un Mac real cada arreglo automático: aplicarlo, comprobarlo y deshacerlo. Lo que no se pueda verificar queda como "guiado".
- [ ] `hardenme fix`: lista numerada de mejoras; la persona elige cuáles, ve qué va a cambiar y confirma.
- [ ] Una sola ventana de contraseña de macOS por tanda. hardenme nunca ve la contraseña y solo ejecuta comandos fijos que van dentro del programa.
- [ ] Los arreglos guiados (FileVault, Buscar mi Mac, actualizaciones...) abren el ajuste exacto, esperan y lo vuelven a comprobar.
- [ ] Al terminar, la nota de antes y la de ahora. Deshacer con `hardenme fix --undo`. Opción `--dry-run`.
- [ ] El doble clic en la app hace el escaneo y luego ofrece los arreglos.

**Hecho cuando:** una persona no técnica hace doble clic, elige qué arreglar, escribe su contraseña una vez y ve la nota nueva. Y puede deshacerlo.

### Después (v0.3): el móvil, sin barreras técnicas
- [ ] **Checklist de móvil guiado, iOS y Android, en la CLI** (`hardenme checklist ios|android`). Preguntas sí/no con nota por perfil, usando los textos ya escritos en la investigación. Con esto se cubre lo más robado sin adb ni cables. Antes estaba en la fase 3; se adelanta.
- [ ] **Historias de robo**: unir varios fallos en un riesgo concreto, por ejemplo "si te roban el Mac ahora mismo, podrían leer tus archivos y no podrías borrarlo".
- [ ] Salidas `--format json|markdown` para compartir el informe.
- [ ] Revisar los menús de la guía en un iPhone y un Android reales.

**Hecho cuando:** alguien revisa la seguridad de su móvil con hardenme y sabe qué hacer si se lo roban.

### Después (v0.4): Android automático por adb
Es la fase 2 del plan, ahora para usuarios avanzados:
- Checks adaptados de HARDAX.
- Probar las claves sin verificar en un Android real.
- Los checks automáticos sustituyen a las preguntas del checklist cuando hay un móvil conectado.

### Más adelante (v0.5+)
- **Windows** (checks adaptados de HardeningKitty, leyendo el estado real).
- **Lecturas de iOS por USB** (libimobiledevice o go-ios) para completar el checklist.
- **App de escritorio**: una app de menú en macOS, si el lanzador no basta.
- **Web (al final)**: un checklist estático generado desde los YAML, para usarlo sin instalar nada. Solo es autoevaluación, porque el navegador no puede leer los ajustes.
- **Comunidad**: plantillas de issues ("check nuevo", "falso positivo"), etiquetas `good first issue` y traducciones.

## 5. Cambios frente al plan original

| Plan original | Propuesta | Por qué |
|---|---|---|
| Fase 2: Android por adb | Pasa a v0.4 | Exige depuración USB, una barrera para el público objetivo |
| Fase 3: checklist iOS | Se adelanta a v0.3 y se amplía a Android | Cubre el dispositivo más robado sin barreras técnicas |
| Fase 3: GoReleaser y CI | Se adelanta a v0.1 | Sin binario no hay usuarios |
| `--fix`, en "más adelante" | Pasa a v0.2 | Elegir y aplicar mejoras completa la experiencia en Mac |
| Historias de robo, en la fase 2 | Pasa a v0.3 | Es lo más diferencial del proyecto |

Si se aceptan estos cambios, hay que actualizar las fases del plan con `/codely-plan-create` o editando el `*-plan.md`.
