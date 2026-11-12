# hardenme: roadmap v2

Fecha: 2026-11-12. Sustituye a [`roadmap.md` del 2026-11-06](../2026_11_06-hardenme_security_auditor/roadmap.md). Documento interno, sin fechas ni duraciones: se ordena por prioridad y cada etapa tiene su "hecho cuando".

## 1. Finalidad

Herramientas fáciles de usar que revisan solas tu PC con Windows, tu Mac, tu Android o tu iPhone y contestan tres preguntas:

1. **¿Qué puedo mejorar?** Ajustes de seguridad flojos, ordenados por lo que más te importa a ti.
2. **¿Hay señales de que ya estoy comprometido?** Cosas instaladas o activadas que no deberían estar ahí.
3. **¿Por qué?** Cada aviso explica qué podría pasar y cómo se arregla, en lenguaje llano.

La segunda pregunta es nueva. Hasta ahora hardenme solo miraba ajustes.

## 2. Principios

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
- El README instala desde el `install.sh` de la última release, no desde `main`.
- Prueba real de `hardenme fix` (aplicar y deshacer) en el Mac del mantenedor.
- Una sola versión con todo lo pendiente.

**Por qué:** dos versiones de trabajo no las tiene nadie. Cuanto más se acumula, más difícil es saber qué rompe qué.

**Análisis:**
- El bloqueo es un acoplamiento: el instalador de `main` llama a `fix`, que la versión publicada no tiene. Apuntar al instalador de la release lo elimina para siempre.
- La prueba real necesita tres ventanas de contraseña del mantenedor. No hay forma de evitarlo sin una máquina virtual.
- Riesgo: publicar la revisión del móvil con menús sin comprobar. Se mitiga marcándola como "beta" en el README hasta la etapa 1.

**Hecho cuando:** quien instale hoy recibe `fix`, la revisión del móvil y las historias, y `main` se puede actualizar sin romper instalaciones.

### Etapa 1: probarlo con personas reales

**Qué:**
- 2 o 3 personas no técnicas instalan y usan hardenme en su Mac sin ayuda.
- Recorrer la revisión de iPhone y la de Android en móviles reales, en español.
- Apuntar dónde se atascan y qué no entienden.

**Por qué:** es la única forma de saber si es "fácil de usar". Todo lo demás del roadmap se apoya en esa suposición.

**Análisis:**
- Es el punto pendiente más antiguo y no cuesta código.
- Lo que salga aquí puede reordenar las etapas 3 y 4. Por eso va antes que ellas.
- Riesgo: que se quede sin hacer otra vez porque depende de terceros. Basta con una sola persona para empezar.

**Hecho cuando:** hay una lista escrita de atascos reales, y los menús del móvil están corregidos o confirmados.

### Etapa 2: señales de compromiso en el Mac

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

**Hecho cuando:** en un Mac limpio la sección sale vacía o con muy pocos elementos, y en uno con una herramienta de acceso remoto y un perfil instalados, los dos aparecen con su explicación.

### Etapa 3: una interfaz que no sea la Terminal

**Qué:** al hacer doble clic, hardenme abre una página en el navegador, servida por el propio programa en el ordenador, con la nota, las historias, los avisos y botones para arreglar.

**Por qué:** hoy hay que leer texto en una ventana negra y escribir "1,3". Para el público objetivo eso ya es una barrera, y en Windows lo será más.

**Análisis:**
- **Alternativas:**
  - menú mejorado en Terminal (librería `huh`): barato, pero sigue siendo Terminal;
  - app nativa (Wails, Tauri): necesita firma para no asustar, y eso exige altas;
  - **página local en el navegador: la misma interfaz vale para Mac y Windows, sin firmar nada.** Es la recomendada.
- **Reutiliza** el informe en JSON que ya existe, y la misma página sirve para la etapa 4.
- **Riesgo de seguridad:** un servidor local es superficie de ataque. Tiene que escuchar solo en `127.0.0.1`, con un código de un solo uso en la dirección, y cerrarse al terminar.
- La Terminal se queda para quien la prefiera.
- Depende de la etapa 1: si las personas reales se manejan bien con la Terminal, esta etapa baja de prioridad.

**Hecho cuando:** una persona hace doble clic, ve su nota y sus avisos en el navegador y arregla uno con un botón, sin ver la Terminal.

### Etapa 4: revisión del móvil en una página web

**Qué:** la revisión guiada de iPhone y Android como página estática, generada desde los mismos YAML y alojada en GitHub Pages. Se abre en el propio móvil.

**Por qué:** hoy revisar el móvil exige tener un Mac y abrir Terminal. Quien solo tiene móvil, que es lo más común y lo más robado, se queda fuera.

**Análisis:**
- **Viabilidad:** alta. Son preguntas y respuestas; un navegador lo hace igual de bien que la CLI.
- **Sin altas:** GitHub Pages va con la cuenta que ya existe.
- **Privacidad:** todo en el navegador, sin servidor ni analítica. Las respuestas se guardan solo en el móvil.
- **Límite honesto:** no es automática. Es la mejor opción posible en iPhone, y la opción sin altas en Android.
- Comparte interfaz con la etapa 3, así que el orden entre las dos es flexible.

**Hecho cuando:** alguien abre un enlace en su móvil, contesta, y ve su nota, sus historias de robo y qué cambiar, sin instalar nada.

### Etapa 5: Windows

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
- Va después de la etapa 3 porque la Terminal de Windows es todavía menos amigable que la del Mac.

**Hecho cuando:** una persona no técnica instala hardenme en su PC siguiendo solo las instrucciones y entiende su informe.

### Etapa 6: Android automático

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

### Etapa 7: iPhone por cable

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
2. **Interfaz:** ¿te parece bien la página local en el navegador como interfaz principal en Mac y Windows?
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
- **Las señales de compromiso son la parte más arriesgada.** Un falso positivo asusta y una falsa tranquilidad es peor. Si la etapa 2 no consigue una sección casi vacía en un Mac normal, hay que recortarla en vez de publicarla.
- **Son cinco superficies** (Mac, Windows, web, adb, cable de iPhone) para una sola persona. El principio de "uno publicado antes del siguiente" es lo único que lo hace viable; si se rompe, nada queda terminado.
- **La etapa 3 es una apuesta.** Si la etapa 1 muestra que el problema no es la Terminal sino los textos o la instalación, hay que cambiarla por eso.
- **Windows sin máquina propia** es el mayor riesgo de calidad. Sin una forma de probar la experiencia completa, no debería publicarse como estable.
- **Orden discutible:** Windows llega tarde para ser el primer sistema de la finalidad. Va detrás porque sin interfaz fácil y sin máquina de pruebas saldría peor que el Mac. Si consigues un PC pronto, puede adelantarse a la etapa 3.
