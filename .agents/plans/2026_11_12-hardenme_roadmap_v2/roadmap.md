# hardenme: roadmap v2

> Actualizado el 2026-11-12 (segunda revisión): la interfaz "mágica" y el contenido sobre robo de móvil pasan al frente.

Fecha: 2026-11-12. Sustituye a [`roadmap.md` del 2026-11-06](../2026_11_06-hardenme_security_auditor/roadmap.md). Documento interno, sin fechas ni duraciones: se ordena por prioridad y cada etapa tiene su "hecho cuando".

## 1. Finalidad

Herramientas fáciles de usar que revisan solas tu PC con Windows, tu Mac, tu Android o tu iPhone y contestan tres preguntas:

1. **¿Qué puedo mejorar?** Ajustes de seguridad flojos, ordenados por lo que más te importa a ti.
2. **¿Hay señales de que ya estoy comprometido?** Cosas instaladas o activadas que no deberían estar ahí.
3. **¿Por qué?** Cada aviso explica qué podría pasar y cómo se arregla, en lenguaje llano.

La segunda pregunta es nueva. Hasta ahora hardenme solo miraba ajustes.

## 2. Principios

- **Que parezca magia.** Ver todos los fallos y arreglarlos en un par de clics. Si hace falta leer un manual, no está terminado.
- **Fácil antes que completo.** Si una persona no técnica no puede usarlo sola, no cuenta como hecho.
- **Automático donde el sistema lo permite, y honesto donde no.** Lo que no se puede leer se pregunta, y nunca se da por bueno.
- **Siempre el porqué.** Ningún aviso sin explicación y sin pasos.
- **Señales, no veredictos.** hardenme no es un antivirus. Dice "esto es poco habitual, ¿lo reconoces?", nunca "estás infectado" ni "estás limpio".
- **Sin altas ni cuentas** para instalarlo ni para usarlo. Esta regla choca con Android e iOS (ver apartado 3) y es una decisión tuya.
- **Privado.** Nada sale del dispositivo. La revisión no cambia nada; los arreglos, solo con permiso.
- **Un sistema publicado antes de empezar el siguiente.**

## 3. Qué se puede hacer de forma automática en cada sistema

Esto condiciona todo el roadmap, así que va primero.

| Sistema | Leer ajustes solo | Buscar señales de compromiso | Cómo llega al usuario | Límite principal |
|---|---|---|---|---|
| **Mac** | Sí, casi todo. Ya funciona (21 comprobaciones) | Sí, bastante: programas que arrancan solos, perfiles, cuentas, acceso remoto | Instalador de una línea y app de doble clic. Ya funciona | Buscar mi Mac y los permisos de privacidad no se pueden leer |
| **Windows** | Sí, casi todo, sin ser administrador (PowerShell y registro) | Sí: historial de Defender, programas de inicio, tareas, cuentas, acceso remoto | Un `.exe` sin firmar lo frena SmartScreen. Hay que instalar con una línea de PowerShell, como en Mac | BitLocker completo pide administrador. No tengo un Windows donde probar |
| **Android** | Solo con una app propia, o con cable y "depuración USB" | Con app o cable: apps con permisos peligrosos, apps de fuera de Play, administradores | **App:** fácil de usar, pero publicarla en Google Play exige cuenta de desarrollador. **Cable (adb):** sin altas, pero es técnico | Elegir entre fácil (con alta) o sin altas (difícil) |
| **iPhone** | Casi nada. Ninguna app puede leer los ajustes de seguridad de iOS | Muy poco: por cable se ven los perfiles instalados y si está supervisado | Preguntas guiadas, más unas pocas lecturas por cable desde un ordenador | Apple no lo permite. "Automático" en iPhone no es posible de verdad |

Consecuencia: en Mac y Windows el objetivo se cumple entero. En Android depende de una decisión. En iPhone la revisión seguirá siendo guiada, y hay que decirlo claro.

## 4. Dónde estamos

- **Publicado:** v0.1.1, revisión del Mac con instalador.
- **Hecho y sin publicar** (10 commits locales): `hardenme fix`, revisión guiada de iPhone y Android, historias de robo, informes en JSON y Markdown.
- **Sin verificar:** el `fix` real en un Mac, los menús en móviles reales, y nadie no técnico lo ha probado.

## 5. Etapas
### Etapa 0: publicar lo que ya está hecho

**Qué:**
- El README instala desde el `install.sh` de la última release, no desde `main`. **Hecho.**
- Prueba real de `hardenme fix` (aplicar y deshacer) en el Mac del mantenedor.
- Una sola versión con todo lo pendiente.

**Por qué:** dos versiones de trabajo no las tiene nadie. Cuanto más se acumula, más difícil es saber qué rompe qué.

**Análisis:**
- El bloqueo es un acoplamiento: el instalador de `main` llama a `fix`, que la versión publicada no tiene. Apuntar al instalador de la release lo elimina para siempre.
- La prueba real necesita tres ventanas de contraseña del mantenedor. No hay forma de evitarlo sin una máquina virtual.
- Riesgo: publicar la revisión del móvil con menús sin comprobar. Se mitiga marcándola como "beta" en el README hasta la etapa 3.

**Hecho cuando:** quien instale hoy recibe `fix`, la revisión del móvil y las historias, y `main` se puede actualizar sin romper instalaciones.

### Etapa 1: el momento "mágico" en el Mac

**Qué:** al hacer doble clic, hardenme abre una página en el navegador, servida por el propio programa, donde todo pasa en un par de clics:
- la nota en grande y "lo que podría pasar hoy";
- todos los fallos, cada uno con su porqué;
- **un botón "Arréglalo todo"**: una sola ventana de contraseña de macOS, y la nota sube a la vista;
- **los arreglos manuales se marcan solos**: hardenme abre el ajuste exacto y vuelve a comprobar cada pocos segundos; cuando cambias el interruptor, la casilla se pone en verde sin pulsar nada;
- un botón para deshacer.

**Por qué:** es lo que hace que alguien diga "no sabía que una app podía hacer esto". Hoy hay que leer texto en una ventana negra y escribir "1,3".

**Análisis:**
- **Casi todo existe:** el motor de `fix`, el registro para deshacer, el informe en JSON y las historias. Falta la página y un servidor pequeño.
- **Alternativas descartadas:**
  - menú mejorado en Terminal: sigue siendo Terminal;
  - app nativa (Wails, Tauri): necesita firma para no asustar, y eso exige altas.
- **La misma página servirá** para Windows y para la revisión web del móvil.
- **Riesgo de seguridad:** un servidor local es superficie de ataque, y este puede cambiar ajustes.
  - Escucha solo en `127.0.0.1`, en un puerto al azar.
  - Cada petición lleva un código secreto de un solo uso, y se comprueba la cabecera `Host`.
  - Solo ejecuta arreglos del catálogo, nunca comandos que lleguen de la página.
  - Se cierra solo cuando se cierra la pestaña.
- **Límite de la magia:** hoy solo 6 ajustes se arreglan solos. FileVault y Buscar mi Mac nunca serán de un clic; para ellos la magia es la casilla que se marca sola.
- **Después, sobre esta base:**
  - historias con datos reales del equipo ("tienes 4 claves SSH y sesión abierta en...");
  - revisión semanal con aviso si la nota baja;
  - más arreglos automáticos verificados (Gatekeeper, servicios remotos, bloqueo de pantalla).

**Hecho cuando:** una persona hace doble clic, ve su nota, pulsa "Arréglalo todo", escribe su contraseña una vez y ve subir la nota, sin ver la Terminal.

**Estado (2026-11-12):** primera versión hecha (`hardenme ui`, sin publicar).
- Probado en el Mac del mantenedor: la página con datos reales, guardar el perfil, cerrar, y el doble clic sin Terminal.
- Probado solo con tests, no en real: el botón "Arréglalo todo", deshacer y la casilla que se marca sola. Necesitan el visto bueno del mantenedor porque cambian ajustes.
- Pendiente: historias con datos reales, revisión semanal y más arreglos automáticos.

### Etapa 2: "si te roban el móvil"

**Qué:**
- Cuatro páginas nuevas en la guía, en español e inglés:
  - qué pueden hacer con tu móvil bloqueado;
  - si pueden llegar a desbloquearlo;
  - huella y cara: cuánto protegen;
  - si es seguro tener la app del banco en el móvil.
- **Simulador de robo:** después de la revisión del móvil, la respuesta para tu caso. "Si te roban este móvil ahora, bloqueado pueden leer tus códigos SMS y usar tu SIM; no pueden entrar en tu banco." Cada frase sale de lo que contestaste.
- Preguntas nuevas sobre el banco: clave distinta a la del móvil, límites, avisos de movimientos.

**Por qué:** la gente no busca "ajustes de seguridad"; busca "qué pasa si me roban el móvil". Es la puerta de entrada al producto.

**Análisis:**
- **Viabilidad:** alta. El simulador es una ampliación de las historias de robo: añade historias "tranquilizadoras" (lo que no pueden hacer) junto a las de riesgo.
- **Riesgo: dar datos falsos o que caducan.** Cada afirmación se comprueba con fuentes oficiales (Apple, Google, INCIBE, Banco de España) y la página las enlaza.
- **Riesgo: asustar o tranquilizar de más.** Se dice lo que hace un ladrón de calle y, aparte, lo que solo hace alguien con medios forenses.
- **No es asesoramiento bancario:** se explican riesgos y se remite a cada banco para sus límites y bloqueos.

**Estado (2026-11-12):** las cuatro páginas están escritas en español e inglés, con cada dato contrastado y sus fuentes enlazadas. Lo que no se pudo confirmar en una fuente oficial se dejó fuera o va marcado como consejo propio. 
- **Simulador de robo, hecho:** el informe separa "si te roban este móvil hoy" (riesgos) de "lo que un ladrón no podría hacer" (lo que ya te protege). Cada frase sale de las respuestas. También en el Mac y en la página del navegador.
- **Preguntas sobre el banco, hechas:** bloqueo propio de la app y clave distinta del código del móvil, en iPhone y Android. Sus consejos no tienen fuente oficial y van marcados así en la guía.
- Sin verificar en un móvil real, como el resto de la revisión del móvil.

**Hecho cuando:** alguien que no sabe de seguridad lee la respuesta a "¿qué pasa si me roban el móvil ahora?" con su propia configuración y sabe qué tres cosas cambiar.

### Etapa 3: probarlo con personas reales

**Qué:**
- 2 o 3 personas no técnicas instalan y usan hardenme en su Mac sin ayuda.
- Recorrer la revisión de iPhone y la de Android en móviles reales, en español.
- Apuntar dónde se atascan y qué no entienden.

**Por qué:** es la única forma de saber si es "fácil de usar". Todo lo demás del roadmap se apoya en esa suposición.

**Análisis:**
- Es el punto pendiente más antiguo y no cuesta código.
- Lo que salga aquí puede cambiar la interfaz y los textos. Va después de la etapa 1 para probar ya con la interfaz nueva.
- Riesgo: que se quede sin hacer otra vez porque depende de terceros. Basta con una sola persona para empezar.

**Hecho cuando:** hay una lista escrita de atascos reales, y los menús del móvil están corregidos o confirmados.

### Etapa 4: señales de compromiso en el Mac

**Qué:** una sección nueva del informe, "Cosas que conviene que reconozcas", con lo que hay en el Mac y no suele estar:
- programas que arrancan solos (LaunchAgents, LaunchDaemons, ítems de inicio) que no son de Apple ni de una app conocida y firmada;
- perfiles de configuración instalados;
- cuentas de usuario y administradores que no esperas;
- herramientas de acceso remoto instaladas o activas;
- claves SSH autorizadas, y cambios en el archivo `hosts`, el proxy o el DNS;
- extensiones del sistema;
- protección de macOS (XProtect) desactualizada.

**Por qué:** es la segunda pregunta de la finalidad, y el Mac es donde ya hay motor, instalador y usuarios.

**Análisis:**
- **Viabilidad:** alta. Casi todo se lee sin contraseña, con comandos del sistema.
- **El riesgo es el falso positivo.** Un Mac normal tiene decenas de programas de arranque legítimos. Un aviso por cada uno haría la herramienta inútil.
  - Mitigación: listar solo lo que no está firmado por un desarrollador identificado, o lo que coincide con una lista corta de herramientas de control remoto y vigilancia.
  - Y decirlo como pregunta: "¿Instalaste tú esto?", con el porqué y cómo quitarlo.
- **El riesgo contrario es la falsa tranquilidad.** No encontrar nada no significa estar limpio. El informe lo tiene que decir.
- **No cuenta para la nota.** Son hallazgos para revisar, no ajustes. Mezclarlos falsearía la nota.
- **Fuera:** leer los permisos de privacidad (macOS no deja sin acceso total al disco) y analizar programas en ejecución como un antivirus.
- **Modelo de datos:** hace falta un tipo nuevo, "hallazgo", distinto de "comprobación": puede devolver varios elementos, cada uno con nombre, ruta y motivo.

**Estado (2026-11-13):** primera versión hecha, con la lista corta.
- Busca: apps de acceso remoto y de vigilancia, programas de arranque sin firma o que lanzan un script, perfiles, gestión remota, administradores de más, claves SSH, proxy y redirecciones en `hosts`.
- En el Mac del mantenedor (8 programas de arranque, todos firmados) la sección sale vacía, que era el criterio.
- El caso contrario solo está probado con tests. El formato de `profiles list` con un perfil instalado no se ha capturado en un Mac real.
- Fuera por ahora: extensiones del sistema y antigüedad de XProtect.

**Hecho cuando:** en un Mac limpio la sección sale vacía o con muy pocos elementos, y en uno con una herramienta de acceso remoto y un perfil instalados, los dos aparecen con su explicación.

### Etapa 5: revisión del móvil en una página web

**Qué:** la revisión guiada de iPhone y Android como página estática, generada desde los mismos YAML y alojada en GitHub Pages. Se abre en el propio móvil.

**Por qué:** hoy revisar el móvil exige tener un Mac y abrir Terminal. Quien solo tiene móvil, que es lo más común y lo más robado, se queda fuera.

**Análisis:**
- **Viabilidad:** alta. Son preguntas y respuestas; un navegador lo hace igual de bien que la CLI.
- **Sin altas:** GitHub Pages va con la cuenta que ya existe.
- **Privacidad:** todo en el navegador, sin servidor ni analítica. Las respuestas se guardan solo en el móvil.
- **Límite honesto:** no es automática. Es la mejor opción posible en iPhone, y la opción sin altas en Android.
- Comparte interfaz con la etapa 1.

**Estado (2026-11-13):** página hecha, sin publicar.
- `hardenme site` genera un solo `index.html` (117 KB) con iPhone y Android, en español e inglés, sin cargar nada de fuera.
- La nota y las historias las calcula el navegador. La página lleva 48 casos calculados por el motor de Go y los comprueba con `#selftest`: pasan 48 de 48 en Chrome.
- Revisadas en capturas a tamaño de móvil: inicio, pregunta y resultado. Sin probar en un móvil real ni en Safari.
- Para publicarla: subir a `main` y activar una vez GitHub Pages (Settings > Pages > Source: GitHub Actions).

**Hecho cuando:** alguien abre un enlace en su móvil, contesta, y ve su nota, sus historias de robo y qué cambiar, sin instalar nada.

### Etapa 6: Windows

**Qué:** la misma revisión que en Mac: ajustes, señales de compromiso y porqués.
- Ajustes: cifrado (BitLocker), Defender, cortafuegos, actualizaciones, bloqueo de pantalla, control de cuentas, escritorio remoto, arranque seguro, cuenta administradora de diario.
- Señales: amenazas detectadas por Defender, programas de inicio y tareas programadas, administradores locales, herramientas de acceso remoto, exclusiones de Defender.
- Instalación con una línea de PowerShell y acceso directo en el escritorio.

**Por qué:** "PC" es el primer sistema de la finalidad y el más usado por gente no técnica.

**Análisis:**
- **Viabilidad técnica:** alta. Go compila para Windows, y el catálogo de la investigación ya lista las comprobaciones.
- **Problema principal: no hay un Windows donde verificar.** La regla del proyecto es no inventar salidas.
  - Los ejecutores de Windows de GitHub Actions dan salidas reales, gratis.
  - Para probar la experiencia completa hace falta una máquina virtual o un PC prestado.
- **Instalación:** un `.exe` descargado con el navegador lo frena SmartScreen ("Windows protegió su PC"). Lo descargado con PowerShell no suele llevar esa marca, igual que `curl` en Mac. **Hay que verificarlo.**
- **Arreglos automáticos:** después, y solo los verificados, como en Mac.
- Reutiliza la interfaz de la etapa 1, así que nace sin Terminal.

**Hecho cuando:** una persona no técnica instala hardenme en su PC siguiendo solo las instrucciones y entiende su informe.

### Etapa 7: Android automático

**Qué:** leer el móvil de verdad en vez de preguntar: bloqueo, cifrado, fecha del parche, Play Protect, apps de fuera de Play, y señales de vigilancia (apps con accesibilidad, administradores de dispositivo, lectores de notificaciones).

**Por qué:** en Android sí es posible, y las señales de programas espía son justo lo que una persona no puede ver sola.

**Análisis: hay dos caminos y son incompatibles con alguna regla.**
- **App de Android:**
  - Es lo único realmente fácil: instalar y pulsar un botón.
  - Publicarla en Google Play exige cuenta de desarrollador (un alta, con pago único).
  - Sin Play, habría que instalar un APK activando "orígenes desconocidos", que es justo lo que hardenme desaconseja.
  - Es otro lenguaje y otro proyecto (Kotlin), aunque puede reutilizar los textos de los YAML.
- **Cable y adb desde el ordenador:**
  - Sin altas y reutiliza el motor actual.
  - Exige activar opciones de desarrollador y depuración USB: no es para el público objetivo.
  - Sirve como modo avanzado y para verificar las lecturas antes de hacer una app.
- **Recomendación:** adb primero, como modo avanzado, porque valida qué se puede leer sin comprometerse a nada. La app, solo si aceptas el alta en Google Play.

**Hecho cuando:** con el móvil conectado, las preguntas que hardenme puede contestar solo desaparecen de la revisión, y las apps con permisos de vigilancia salen listadas con su porqué.

### Etapa 8: iPhone por cable

**Qué:** con el iPhone conectado al ordenador, leer lo poco que Apple deja: versión de iOS, si tiene código, si la copia está cifrada, perfiles instalados y si está supervisado.

**Por qué:** los perfiles y la supervisión son la principal señal de compromiso visible en un iPhone.

**Análisis:**
- Cubre unas 5 de las 15 comprobaciones. El resto seguirá siendo guiado.
- Depende de una herramienta externa (libimobiledevice o go-ios) y de emparejar el iPhone.
- Detectar programas espía avanzados queda fuera: es análisis forense, y el proyecto de referencia (MVT) no se puede reutilizar por su licencia.
- Es la etapa con peor relación entre esfuerzo y resultado. Por eso va la última.

**Hecho cuando:** con el iPhone conectado, hardenme avisa de un perfil de configuración instalado y explica qué puede hacer ese perfil.

### Continuo: que no se pudra

- Cada comprobación guarda en qué versión del sistema se verificó y cuándo.
- Revisión semanal automática de los enlaces a Apple, Google y Microsoft.
- CI: análisis estático, vulnerabilidades de dependencias, revisión del instalador, y `scan` en ejecutores con macOS 14, 15 y actual, y Windows.
- Atestación de las descargas, para que el binario se pueda verificar.
- Cada sistema nuevo (iOS, macOS, Android) abre una tarea de revisión de menús.

## 6. Decisiones que son tuyas

1. **Android:** ¿aceptas un alta en Google Play para tener una app fácil, o se queda en página web y cable?
3. **Señales de compromiso:** ¿lista corta y conservadora (pocos avisos, alguno se escapa) o lista amplia (más avisos, más ruido)? Recomiendo la corta.
4. **Windows:** ¿tienes acceso a un PC o prefieres una máquina virtual?

## 7. Lo que queda fuera

- Ser un antivirus o eliminar programas maliciosos.
- Análisis forense de programas espía avanzados.
- Comprobar si tus contraseñas o tu correo están en filtraciones: exige enviar datos por internet. Se enlaza desde la guía.
- Escanear la red o el router.
- Cualquier cosa que necesite una cuenta en un servicio de hardenme.

## 8. Revisión crítica de este roadmap

- **Promete más de lo que la finalidad permite en iPhone.** "Automático" ahí no existe. Está dicho en el apartado 3, pero conviene decirlo también en el README para no decepcionar.
- **Las señales de compromiso son la parte más arriesgada.** Un falso positivo asusta y una falsa tranquilidad es peor. Si la etapa 4 no consigue una sección casi vacía en un Mac normal, hay que recortarla en vez de publicarla.
- **Son cinco superficies** (Mac, Windows, web, adb, cable de iPhone) para una sola persona. El principio de "uno publicado antes del siguiente" es lo único que lo hace viable; si se rompe, nada queda terminado.
- **La etapa 1 es una apuesta.** Si las pruebas con personas (etapa 3) muestran que el problema son los textos o la instalación, hay que corregir ahí antes de seguir.
- **Windows sin máquina propia** es el mayor riesgo de calidad. Sin una forma de probar la experiencia completa, no debería publicarse como estable.
- **Orden discutible:** Windows llega tarde para ser el primer sistema de la finalidad. Va detrás porque sin interfaz fácil y sin máquina de pruebas saldría peor que el Mac. Si consigues un PC pronto, puede adelantarse a la etapa 4.
