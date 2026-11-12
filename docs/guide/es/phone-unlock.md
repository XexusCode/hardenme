# ¿Pueden desbloquear mi móvil?

[English](../en/phone-unlock.md)

Respuesta corta: **un ladrón normal no puede adivinar el código de un móvil actualizado**. Lo que pasa de verdad es otra cosa: que te vea teclearlo, o que te quite el móvil ya desbloqueado.

## Adivinando el código: casi imposible

El propio chip de seguridad del móvil limita los intentos, y no se puede saltar desde fuera.

- **iPhone:** los tres primeros fallos no tienen espera. Luego hay que esperar 1 minuto, 5 minutos, 15 minutos, 1 hora, 3 horas y 8 horas. Al décimo fallo el iPhone se queda bloqueado (y se borra, si activaste "Borrar datos").
- **Android:** las esperas también crecen con cada fallo. En las versiones más recientes llegan a horas y días, y a los 20 fallos ya no se admite ninguno más.

Con un código de 6 cifras hay un millón de combinaciones. Con esas esperas, probarlas no es viable. Con 4 cifras hay diez mil, y además es mucho más fácil de ver de reojo.

## Con herramientas forenses: a veces, y no un ladrón de calle

Existen empresas que venden herramientas para sacar datos de móviles. Lo que se sabe por la prensa especializada:

- Funcionan peor cuanto más nuevo y actualizado está el móvil. Documentos filtrados en 2024 mostraban que una de ellas no podía con buena parte de los iPhone modernos.
- Un móvil recién encendido, que aún no se ha desbloqueado ni una vez, tiene todos los datos cifrados y es casi imposible de leer sin el código.
- Por eso un iPhone que lleva 72 horas sin desbloquearse se reinicia solo, según han comprobado investigadores.

Conclusión práctica: **mantén el móvil actualizado**. Es tu mejor defensa contra esto.

## Lo que pasa de verdad

### Te ven teclear el código y luego te roban

En 2023 se documentaron en Estados Unidos robos así en bares: alguien mira cómo tecleas el código y después te quita el móvil. Con el código, el ladrón cambiaba la contraseña de la cuenta de Apple, desactivaba Buscar, pagaba con Apple Pay y entraba en apps del banco con las contraseñas guardadas.

**Qué hacer:**

- **iPhone: activa la Protección en caso de robo** (Ajustes > Face ID y código). Lejos de tus sitios habituales:
  - ver contraseñas guardadas, borrar el iPhone o desactivar el modo Perdido pide tu cara o tu huella, y el código ya no vale;
  - cambiar la contraseña de la cuenta de Apple o el código del iPhone, además, espera una hora.
- **Android: activa la Verificación de identidad** si tu móvil la tiene (Ajustes > Google > Todos los servicios > Protección antirrobo). Fuera de tus sitios de confianza, cambiar el bloqueo, la cuenta de Google o ver contraseñas guardadas pide tu huella o tu cara.
- **Teclea el código lo menos posible en público.** Usa la huella o la cara, y si tienes que teclear, tápalo como el PIN del cajero.
- **Usa 6 cifras o más.**

### Te lo arrancan de la mano desbloqueado

El ladrón tiene el móvil abierto y solo necesita que no se bloquee.

- **Android: activa el Bloqueo antirrobo.** Detecta el tirón y la huida y bloquea la pantalla.
- **Bloqueo automático corto** (30 segundos o 1 minuto) en cualquier móvil.
- Si pasa, bloquéalo a distancia enseguida: icloud.com/find, android.com/find o android.com/lock.

## Resumen

| Situación | ¿Pueden entrar? |
|---|---|
| Bloqueado, código de 6 cifras, actualizado | Un ladrón normal, no |
| Bloqueado, móvil viejo sin actualizar | Más riesgo frente a herramientas especializadas |
| Te vieron el código | Sí, salvo lo que protegen la Protección en caso de robo o la Verificación de identidad |
| Te lo quitaron desbloqueado | Sí, hasta que se bloquee |

Sigue en [Huella y cara: cuánto protegen](biometrics.md).

## Fuentes

- Apple: [seguridad de los códigos](https://support.apple.com/guide/security/passcodes-sec20230a10d/web), [Protección en caso de robo](https://support.apple.com/es-es/120340).
- Android: [límite de intentos](https://source.android.com/docs/security/features/authentication/rate-limiting), [protección antirrobo](https://support.google.com/android/answer/15146908?hl=es).
- Prensa: [robos tras ver el código (resumen de MacRumors del reportaje del Wall Street Journal)](https://www.macrumors.com/2023/02/24/iphone-stolen-passcodes-report/), [reinicio del iPhone a los 3 días (TechCrunch)](https://techcrunch.com/2024/11/14/new-apple-security-feature-reboots-iphones-after-3-days-researchers-confirm), [qué móviles resisten (404 Media)](https://www.404media.co/leaked-docs-show-what-phones-cellebrite-can-and-cant-unlock/).
