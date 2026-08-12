# Decisiones — TP1

## 1. Por qué Git no pudo resolver el conflicto solo

Git no puede resolver automáticamente un conflicto cuando **dos ramas modifican exactamente la misma línea** de un archivo. En este caso, tanto `feature/titulo-a` como `feature/titulo-b` cambiaron la primera línea del `README.md`:

- `feature/titulo-a` → `# Proyecto IngSoft3 - versión A`
- `feature/titulo-b` → `# Proyecto IngSoft3 - versión B`

Git puede fusionar cambios automáticamente cuando tocan líneas o archivos **distintos** (hace un merge de 3 vías comparando con el ancestro común). Pero cuando dos ramas modifican **la misma línea**, Git no tiene forma de saber cuál versión es la "correcta": esa es una decisión de **contenido**, no de algoritmo.

Para que nunca hubiera aparecido este conflicto, las dos ramas deberían haber modificado líneas distintas del mismo archivo, o directamente archivos distintos. Con trabajo en paralelo sobre las mismas líneas, el conflicto es inevitable y su resolución siempre requiere una decisión humana.

---

## 2. Problemas encontrados y cómo se solucionaron

### `gh pr merge` no acepta `--yes`
Al ejecutar `gh pr merge 1 --squash --delete-branch --yes` se obtuvo un error: `unknown flag: --yes`.
**Solución**: La CLI de `gh` en versiones más recientes no requiere ese flag; simplemente se omitió y el merge funcionó sin confirmación interactiva gracias a que el PR cumplía todas las condiciones.

### Advertencia de `zsh` en el comando de la release
Al crear la release v1.0.0 con `gh release create`, apareció `zsh: command not found: .gitignore` como advertencia. Esto se debió a que el shell intentó interpretar parte del texto de las notas. **No afectó la release**: la URL de la release se creó correctamente (`https://github.com/mateopappa/ingsoft3-tp01/releases/tag/v1.0.0`).

### Estado inicial `UNKNOWN` del PR en conflicto
Inmediatamente después de mergear el PR A, la API de GitHub devolvió `"mergeable": "UNKNOWN"` para el PR B. Esto es normal: GitHub necesita unos segundos para recalcular el estado de mergeo. Al esperar y volver a consultar, el estado cambió a `"mergeable": "CONFLICTING"`.

---

## 3. Declaración de uso de IA

Este TP fue completado con asistencia del agente de IA **Antigravity** (basado en Google DeepMind), que ejecutó los comandos de Git y la CLI de GitHub en nombre del usuario.

Los resultados se verificaron en cada paso de la siguiente manera:

- **Autenticación**: se verificó con `gh auth status` que el usuario `mateopappa` estaba autenticado.
- **Branch protection**: se verificó que el push directo a `main` fue rechazado con el mensaje `protected branch hook declined` antes de continuar.
- **PRs y conflicto**: se verificó el estado `CONFLICTING` vía `gh pr view 3 --json mergeable,mergeStateStatus` y se observaron los marcadores de conflicto en el contenido del archivo antes de resolverlos.
- **Release**: se verificó la URL de la release publicada (`https://github.com/mateopappa/ingsoft3-tp01/releases/tag/v1.0.0`).

Todos los conceptos de la guía (qué es una rama, cómo funciona un PR, por qué ocurre un conflicto, qué es SemVer) fueron comprendidos y pueden ser defendidos oralmente.
