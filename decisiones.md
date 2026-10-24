# Enlaces de este TP (TP6 — CD)

> Completar / actualizar tras el primer deploy real a Render. Quien corrige navega desde acá.

| Qué | URL |
|---|---|
| Paquete público backend (GHCR) | `ghcr.io/mateopappa/ingsoft3-tp01-backend` — etiqueta `sha-<commit-del-merge>` |
| Paquete público frontend (GHCR) | `ghcr.io/mateopappa/ingsoft3-tp01-frontend` — etiqueta `sha-<commit-del-merge>` |
| Cadena 1 — PR con tests verdes y «Entrar al registry» **salteado** | _(pegar corrida del PR de implementación TP6)_ |
| Cadena 2 — corrida de `main` con «Construir y publicar» **último** paso propio | _(pegar corrida de push a main)_ |
| URL QA (front) | _(https://…-front-qa.onrender.com)_ |
| URL QA (api) | _(https://…-api-qa.onrender.com)_ |
| URL PROD (front) | _(https://…-front-prod.onrender.com)_ |
| URL PROD (api) | _(https://…-api-prod.onrender.com)_ |

**Checklist de cableado (una vez, en tu cuenta):** ver § TP6 → «Setup manual pendiente».

---

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

---

## TP2 — Contenedores: La App del Semestre

### 1. Elección de la Aplicación
Se seleccionó la aplicación **Flow — Personal Activity Tracker** para el desarrollo de la materia.
- **Justificación según criterios**:
  - **Arquitectura de 3 capas**: Frontend estático en Nginx, Backend REST API en Go 1.24 y Base de datos relacional PostgreSQL 16.
  - **Rapidez y Cero Fricción**: Se compila en milisegundos, inicia en menos de un segundo y requiere un consumo mínimo de RAM/CPU.
  - **Cero dependencias pagas**: Funciona 100% de manera nativa y containerizada sin requerir tarjetas de crédito ni servicios externos costosos.
  - **Idoneidad para CI/CD**: Es ideal para implementar pipelines de CI (tests unitarios + linter), escaneo de vulnerabilidades, empaquetado de imágenes multi-stage y despliegues continuos (CD) en los TPs 4 a 9.

### 2. Decisiones de Contenerización y Dockerfiles
- **Backend (`backend/Dockerfile`)**:
  - **Multi-stage build**: Se utilizó una etapa de compilación (`golang:1.24-alpine`) donde se descargan dependencias y se compila el ejecutable estático con `CGO_ENABLED=0` y `-ldflags="-s -w"`.
  - **Etapa de ejecución mínima**: La imagen final parte de `alpine:3.20` conteniendo únicamente el binario compilado y certificados CA.
  - **Seguridad**: Se creó un usuario sin privilegios (`appuser:appgroup`) para ejecutar el proceso en lugar de `root`.
  - **Reducción de tamaño**: La imagen final pesa solo **23.5 MB**, comparada contra los >800 MB que pesaría una imagen con el SDK completo de Go.
- **Frontend (`frontend/Dockerfile` & `nginx.conf`)**:
  - Se utiliza `nginx:alpine` para servir la SPA estática.
  - Se configuró `nginx.conf` como un proxy inverso de la API REST (`location /api/ { proxy_pass http://backend:8080; }`), resolviendo peticiones cross-origin en el cliente sin necesidad de habilitar CORS ni hardcodear URLs absolutas.

### 3. Estrategia de Persistencia y Redes
- **Persistencia**: La base de datos PostgreSQL utiliza un volumen administrado por Docker (`db_data:/var/lib/postgresql/data`). Esto garantiza que los datos de las actividades persistan entre reinicios (`docker compose down` seguido de `docker compose up`), manteniéndose aislados de la capa de escritura efímera del contenedor.
- **Red interna y DNS**: Los contenedores se comunican dentro de la red privada de Compose mediante resolución DNS por nombre de servicio (el backend se conecta a `db:5432`).
- **Healthcheck**: El servicio `backend` depende de la condición `service_healthy` del contenedor `db`, ejecutando `pg_isready` para evitar fallos de conexión al arrancar el servidor antes de que PostgreSQL acepte conexiones.

### 4. Declaración de Uso de IA
Este TP contó con el soporte del agente de IA **Antigravity** (Google DeepMind) para auditar la estructura de archivos, verificar comandos de Compose y redactar la documentación. Todos los manifiestos (`Dockerfile`, `docker-compose.yml`, `nginx.conf`) fueron verificados ejecutando compilaciones, pruebas de persistencia y suites unitarias (`make test`) localmente.

---

## TP3 — Planificación DevOps

### 1. Duración del Sprint e Iteraciones
- **Elección**: Se fijó una duración de **2 semanas** para la iteración (`Sprint 1`).
- **Justificación**: Se alineó con el calendario de entregas de la materia y el ritmo de desarrollo de incrementos de CI/CD. Permite tener metas alcanzables y acotadas sin acumular trabajo en progreso innecesario.

### 2. Límite de Trabajo en Progreso (WIP Limit)
- **Elección**: Se configuró un límite de **2 ítems simultáneos** en la columna *In Progress*.
- **Justificación**: El objetivo central de la filosofía Kanban/DevOps es "empezar menos, terminar más". Establecer un WIP limit de 2 evita la multitarea excesiva, reduce el cambio de contexto y visibiliza cuellos de botella (poniendo el contador en rojo si se sobrepasa) antes de continuar incorporando nuevas tareas.

### 3. Diagnóstico de la Historia Mal Escrita
- **Historia analizada**: *"Como desarrollador quiero crear la tabla usuarios para guardar los datos."*
- **Diagnóstico (Por qué está mal escrita)**:
  1. **Rol incorrecto**: El desarrollador es quien implementa la solución, no el cliente o usuario final que percibe el valor de negocio.
  2. **Es una Tarea Técnica disfrazada**: Crear una tabla de base de datos no es una capacidad funcional de usuario observable; es un paso de implementación técnica (debería ser un Task bajo una Story).
  3. **Beneficio trivial**: "Para guardar los datos" no explica el valor de negocio real ni ayuda al Product Owner a priorizar el backlog.
- **Reescritura correcta**: *"Como usuario registrado quiero poder guardar mis datos personales en mi perfil para que queden almacenados de forma segura entre sesiones."*

### 4. Estructura del Backlog y Jerarquía
Se estructuró el backlog del proyecto siguiendo la jerarquía canónica de la materia:
- **Épica (#6)**: `EPIC: Pipeline DevOps completo para mi app`
- **Historia de Usuario (#7)**: `CI: build y tests automáticos en cada PR` (con 4 Criterios de Aceptación verificables).
- **Tareas Técnicas (#8, #9)**:
  - `#8`: `Escribir el workflow de build y tests` (Cerrada automáticamente mediante el PR #11 que incorporó `.github/workflows/ci.yml` demostrando trazabilidad).
  - `#9`: `Publicar el reporte de tests como artefacto` (Queda abierta en To Do para continuar en los TP4/TP5).
- **Bug (#10)**: `El front carga sin la lista cuando el back todavía no responde` (Bug independiente al costado del árbol).

### 5. Declaración de Uso de IA
Este TP fue realizado con la asistencia del agente de IA **Antigravity** (Google DeepMind) para la generación de labels, automatización de la creación de issues vía CLI de GitHub, vinculación de trazabilidad entre PR e issues y redacción de la sección de decisiones.

---

## TP4 — CI: Pipelines as Code

### 1. Estructura Elegida del Pipeline
- **Jobs definidos**: `build-backend` y `build-frontend`.
- **Justificación de los jobs**: La aplicación Flow está compuesta por dos componentes contenerizados independientes: el backend REST en Go y el frontend SPA servido por Nginx. No existe un contenedor monolítico único; por lo tanto, el pipeline modela la realidad de la arquitectura del software.
- **Por qué corren en paralelo**:
  - En GitHub Actions, los jobs se ejecutan en máquinas virtuales separadas y limpias (`ubuntu-latest`).
  - Al no existir una dependencia causal entre la construcción de la imagen de Nginx y la compilación del binario en Go, no hay motivo para serializarlos (`needs`).
  - Ejecutar en paralelo minimiza el tiempo de ciclo (Lead Time) y otorga feedback rápido al desarrollador: la duración total de la corrida es determinada por el job más lento ($\approx 19$s con caché) y no por la suma acumulativa de ambos.
  - Los jobs no comparten sistema de archivos ni memoria; cada uno baja su copia limpia del código con `actions/checkout@v6`.

### 2. Estrategia de Caché de Capas (GHA)
- **Mecanismo adoptado**: Se utilizó el backend de caché nativo de GitHub Actions provisto por Docker Buildx (`cache-from: type=gha` y `cache-to: type=gha,mode=max`).
- **Aislamiento por Scope**: Se configuró `scope=backend` para el backend y `scope=frontend` para el frontend. Esto evita condiciones de carrera donde un job sobreescriba los metadatos de caché del otro en el almacenamiento del repositorio.
- **Capas reutilizadas**:
  - En `build-backend`: Gracias a la estructura multi-stage del Dockerfile del TP2, se copian primero `go.mod` y `go.sum` antes de hacer `RUN go mod download`. Cuando un commit altera código en `cmd/` o `internal/`, las capas base del SDK y la descarga de dependencias externas se recuperan en estado `CACHED`. Esto redujo el tiempo del job de 54 segundos en la corrida inicial a 19 segundos en la segunda corrida (reducción del 65%).
  - En `build-frontend`: La capa base de `nginx:alpine` y la configuración estática se reutilizan inmediatamente.
- **Propiedad fundamental del caché**: El caché es una optimización efímera y no transaccional. GitHub puede desalojarlo en cualquier momento por políticas de cuota (límite de 10 GB por repo) o tiempo de inactividad (7 días). Si el caché desaparece, el pipeline **no falla**: se degrada graciosamente reconstruyendo las capas desde cero, tardando únicamente unos segundos más. No existen dependencias ocultas atadas al estado del runner.

### 3. Construcción vía Dockerfile vs Compilación Ad-Hoc
- **Decisión**: El pipeline delega la construcción estrictamente en `docker/build-push-action@v7` apuntando al `Dockerfile` de cada servicio, sin ejecutar comandos sueltos de compilación (`go build` o `npm`) en el runner de GitHub.
- **Justificación**:
  1. **Principio de Fuente Única de Verdad**: Si el pipeline compilara con herramientas del runner y luego el despliegue usara un Dockerfile, existirían dos definiciones divergentes del build que inevitablemente generarían inconsistencias ("compila en el CI pero falla en el contenedor").
  2. **Paridad de Entornos**: El Dockerfile asegura que la compilación se realice exactamente bajo el SDK definido (`golang:1.24-alpine`), con las banderas de compilación estática (`CGO_ENABLED=0`) y empaquetado en Alpine minimal con el usuario `appuser`. Probar el build en CI es probar exactamente el mismo artefacto inmutable que se distribuye a producción.

### 4. Required Status Checks y Demostración del Gate
- **Configuración de protección**: Se aplicó sobre `main` la regla:
  - `required_status_checks.contexts`: `["build-backend", "build-frontend"]`.
  - `strict: true` (Require branches to be up to date before merging).
  - `enforce_admins: true`.
  - `required_approving_review_count: 0` (adecuado a desarrollo individual).
- **Demostración práctica documentada**:
  1. **Rojo**: En la rama `feature/demo-gate` (PR #18), se introdujo deliberadamente un import no existente en `backend/cmd/server/main.go`. El job `build-backend` falló (`conclusion: FAILURE`) y la API de GitHub reportó `mergeStateStatus: BLOCKED`, impidiendo físicamente el merge.
  2. **Efecto de `strict: true`**: Mientras PR #18 estaba bloqueado, se abrió el PR #19 (`docs/muestra-del-freno`). Tras solucionar y mergear el PR #18, el PR #19 pasó a estado `mergeStateStatus: BEHIND`, obligando a presionar "Update branch" para re-validar los checks contra el nuevo `main` antes de habilitar el merge.
  3. **Verde y Merge**: Se corrigió el código (`fix: saca el import que no existe`), el workflow se disparó automáticamente, ambos jobs concluyeron en `SUCCESS` (`mergeStateStatus: CLEAN`) y el PR fue mergeado exitosamente.

### 5. Problemas Encontrados y Soluciones
- **Ausencia de `CACHED` en ejecuciones solapadas**: Al disparar commits casi consecutivos, la segunda corrida iniciaba antes de que la primera finalizara la exportación de su caché (`cache-to: mode=max`). Se resolvió esperando la finalización de la primera corrida antes de enviar el commit `--allow-empty`.
- **Visibilidad de Checks en la API de GitHub**: Para configurar los Required Status Checks, los nombres de contexto (`build-backend`, `build-frontend`) deben haber corrido previamente al menos una vez en el repositorio. Se completó primero la corrida del PR #17 y luego se aplicó la política vía API sin inconvenientes.

### 6. Declaración de Uso de IA
Este trabajo práctico fue completado con la asistencia del agente de IA **Antigravity** (Google DeepMind), que colaboró en la redacción del workflow de GitHub Actions, la automatización de la API de protección de ramas, la ejecución de la prueba de rotura/fix y la elaboración de la documentación técnica. Todos los pasos fueron verificados mediante la CLI `gh`, los logs de GitHub Actions y pruebas locales de Docker.

---

## TP5 — Calidad automatizada: tests, coverage y el umbral que frena un merge

### 1. Qué lógica elegí testear y por qué

Elegí las **reglas de negocio de los services** (no handlers ni SQL): son el lugar donde un bug duele de verdad en Flow.

| Regla | Por qué duele si se rompe |
|---|---|
| Duración de actividad `> 0` | Deja entrar sesiones basura y corrompe totales del dashboard |
| Categoría ajena / inexistente | Filtra datos entre usuarios (aislamiento) |
| No borrar categoría en uso | Evita romper FK / historial |
| Un solo timer activo por usuario | Invariante central del producto (HTTP 409) |
| Password ≥ 8 / email válido | Puerta de auth |
| `FormatDuration` y % por categoría | Lo que el usuario ve en el dashboard |

Los tests viven al lado del código (`internal/*/service_test.go`), con estructura AAA y nombres que describen el comportamiento.

**Frontend** (Dockerfile separado → los 4 tests + 3 técnicas sí aplican): extraje lógica pura a `frontend/src/lib/` (`format`, `validate`, `api`, `filters`) y la reutiliza `static/js/app.js` vía ES modules. Así se testea sin DOM.

### 2. Stack Go + vitest (tabla «Tu stack, de un vistazo»)

| Logro | Herramienta elegida |
|---|---|
| Dónde viven los tests | Backend: `*_test.go` en el mismo paquete lógico (`activities_test`, etc.). Frontend: `src/lib/*.test.js` |
| Test parametrizado | Go: table-driven `t.Run` · Front: `it.each` |
| Doble (mock) | Go: structs manuales que implementan interfaces inyectadas · Front: `vi.fn()` sobre `fetch` |
| Medir cobertura | Go: `go test -coverprofile` · Front: `vitest run --coverage` (v8) |
| Umbral que rompe el build | Go: script `backend/scripts/check-coverage.sh` compara el total y hace `exit 1` · Front: `coverage.thresholds` en `vitest.config.js` |
| Qué entra en la cuenta | Go: filtro del profile a `*service.go` · Front: `coverage.include: ['src/lib/**/*.js']` |

**Branch coverage en Go:** no existe en `go test -cover` (sólo sentencias). Queda declarado acá y en el Summary del job (`Branch coverage: N/A`). En frontend sí se reporta rama (vitest/v8).

### 3. Refactor para poder mockear

Antes, `Service` guardaba `*Repository` y `*categories.Service` concretos: no se podía reemplazar la BD en un unit test sin Postgres.

**Qué cambié:** introduje puertos (interfaces) inyectados por constructor:

- `activities.ActivityStore` + `CategoryFinder`
- `categories.CategoryStore`
- `timer.TimerStore` + `CategoryFinder`
- `dashboard.MetricsStore`
- `auth.UserStore` (además del `CategorySeeder` que ya existía)

Los `*Repository` reales siguen cumpliendo las interfaces; `main.go` no cambia su cableado. Los tests fabrican mocks/stubs y **verifican la interacción** (p. ej. que `CreateActivity` del store **no** se llama si la categoría es inválida).

### 4. Umbral de coverage: número, métrica y porqué

**Medición real (suite completa, local):**

| Lado | Métrica | Valor medido | Umbral |
|---|---|---|---|
| Backend | Statement coverage sobre `*service.go` | **~78%** | **70%** (statements) |
| Backend | Branch | N/A (Go) | — |
| Frontend | Lines | **~98%** | **80%** |
| Frontend | Branches | **~89%** | **75%** |

**Por qué esos números:** el umbral del backend está anclado a la medición real (~78) y frena si baja, sin ser inalcanzable. Para subirlo a 85 habría que cubrir CRUD triviales del service (`DeleteActivity`, `ListActivities`, `SeedDefaultCategories`) que hoy sólo delegan al repo. En frontend el umbral es más holgado respecto al 98% actual porque el gate tiene que fallar ante lógica nueva sin tests, no ante ruido de ramas menores.

Prueba (corrida con summary + artefacto): ver links de Actions pegados abajo tras el merge del PR de implementación.

### 5. Qué dejé afuera de la cuenta (y por qué)

**Backend — fuera del umbral:**

| Qué | Por qué |
|---|---|
| `cmd/server/main.go` (arranque) | Cableado; si falla, la app no levanta |
| `internal/*/handler.go` | Adaptadores HTTP; se prueban en e2e (TP7) |
| `internal/*/repository.go` | SQL/infra; unit tests no deben abrir Postgres |
| `internal/models/*` | Structs de datos sin reglas |
| `migrations/*`, `database/*`, `config/*`, middleware | Arranque / generado / infra |

**Frontend — fuera del umbral:**

| Qué | Por qué |
|---|---|
| `static/js/app.js` | Controlador DOM (listeners, render); sin jsdom a propósito |
| `index.html`, CSS, nginx | No es lógica de negocio |

Excluir eso **no** es esconder reglas: las reglas viven en `service.go` / `src/lib`.

### 6. Por qué coverage alto no garantiza calidad (ejemplo propio)

Un test que llame `FormatDuration(60)` y **no haga assert** sumaría línea al 100% de esa función y no detectaría si mañana devolviera `"1h"`. Coverage mide ejecución, no verificación. Mis tests de timer con mock fallan si alguien quita el check de «ya hay timer activo» aunque el coverage del archivo siga alto: el assert sobre `createCalls == 0` es la verificación que la métrica no ve.

### 7. Ejercicio del camino sin cubrir (rama de código)

- **Línea:** `frontend/src/lib/validate.js` — `if (!activityDate) { return { ok: false, error: 'Activity date is required' }; }` (aprox. líneas 33–35).
- **Entrada que la recorrería:** `validateActivityForm({ description: 'Coding', categoryId: 'cat-1', activityDate: '', hours: 1, minutes: 0 })`.
- **Decisión:** **no lo agregué** en esta entrega. El reporte ya marca esa rama en rojo; dejarla visible documenta el límite de la métrica. Si mañana alguien borra el check de fecha, el umbral de branches aún puede sostenerse por otras ramas — por eso el test de borde importa más que el porcentaje.

(Backend: el branch `ActivityDate == ""` → default `time.Now()` en `CreateActivity` tampoco tiene test dedicado; misma decisión.)

### 8. Quality gate y required checks

Los jobs siguen llamándose `build-backend` y `build-frontend` (required desde TP4). Ahora, adentro de cada uno:

1. Construyen la imagen final (como TP4).
2. Construyen `target: test` con `load: true`.
3. `docker run` ejecuta la suite + umbral; si el número no alcanza, el job falla → merge bloqueado.

No hay jobs nuevos: el freno de cobertura reutiliza los mismos required checks.

### 9. Demostración del freno (dos PRs)

| Evidencia | URL |
|---|---|
| Corrida verde con coverage en Summary + artefactos (`coverage-backend`, `coverage-frontend`) | https://github.com/mateopappa/ingsoft3-tp01/actions/runs/40000635837 |
| Corrida **roja por umbral** (compila; suite vieja verde; coverage corta) | https://github.com/mateopappa/ingsoft3-tp01/actions/runs/40001080013 |
| Secuencia rojo → tests → verde → merge | https://github.com/mateopappa/ingsoft3-tp01/pull/22 |
| Freno vigente (PR **abierto** en rojo hasta la defensa) | https://github.com/mateopappa/ingsoft3-tp01/pull/23 |

**Detalle de la corrida roja** (`40001080013`):
- `build-backend`: `FAIL: coverage 69.0% is below threshold 70%` (métrica: **statements** sobre `*service.go`)
- `build-frontend`: `Coverage for lines (69.48%) does not meet global threshold (80%)` (métrica: **lines**)
- Qué faltaba: tests de `GoalProgress` / `goals.js` → se agregaron en el mismo PR #22 y la corrida verde fue https://github.com/mateopappa/ingsoft3-tp01/actions/runs/40001455038

**PR de implementación** (suite + pipeline): https://github.com/mateopappa/ingsoft3-tp01/pull/21

### 10. Problemas encontrados y soluciones

| Problema | Solución |
|---|---|
| Tests viejos en `tests/unit` reimplementaban reglas sin llamar al service (coverage 0 útil) | Se reescribieron tests reales con mocks sobre interfaces |
| `go test -coverpkg` sobre el paquete entero daba ~16% (handlers/repos en 0%) | Script filtra el profile a `*service.go` antes de evaluar el umbral |
| Go 1.24 en `go.mod` vs toolchain local 1.22 | `GOTOOLCHAIN=auto` descarga 1.24; CI usa `golang:1.24-alpine` |
| Frontend sin bundler + necesidad de testear lógica | ES modules en `src/lib` + `type="module"` en `index.html`; vitest cubre sólo `src/lib` |
| Etapa de test del front necesita vitest | `npm ci` **sin** `--omit=dev` en la etapa `build`; la imagen `final` (nginx) no lleva Node |

### 11. Declaración de uso de IA

Este TP se completó con asistencia del agente de IA **Cursor (Auto/Composer)**. La IA ayudó a refactorizar interfaces, escribir la suite, el script de umbral Go, vitest y el workflow. **Verificación humana / automatizada:**

- `go test ./internal/...` y `./scripts/check-coverage.sh` (umbral 70%) en verde localmente.
- `npm run test:coverage` en verde con thresholds 80/75.
- Cada assert de la suite se puede explicar (qué regla rompe si se invierte).
- Los links de Actions/PR de la §9 se contrastan contra corridas reales, no capturas.

Si un test fue sugerido por IA y no puedo decir qué caso **no** cubre, no cuenta como entregado: por eso el camino sin cubrir de `activityDate` quedó explícito arriba.

---

## TP6 — CD: environments, aprobaciones y deployment patterns

### 0. Setup manual pendiente (bloquea URLs vivas / approve / rechazo)

Este agente **no puede** crear cuentas en Render/Neon ni environments/secrets en GitHub (403). El código del pipeline ya está; falta cablear en tu usuario `mateopappa`:

1. **Neon** — proyecto con databases `app_qa` y `app_prod`. Copiá `DATABASE_URL` de cada una (URI `postgres://…`). Flow usa `DATABASE_URL` + `AUTO_MIGRATE=true` (las tablas se crean al primer arranque).
2. **Render** — 4 Web Services Docker, Auto-Deploy **OFF**:
   - `flow-api-qa` / `flow-api-prod` — Root Directory `backend`; env: `DATABASE_URL`, `SESSION_SECRET`, `AUTO_MIGRATE=true`, `COOKIE_SECURE=true`, `APP_ENV=qa|production`
   - `flow-front-qa` / `flow-front-prod` — Root Directory `frontend`; env: `BACKEND_URL=https://<api-de-ESE-entorno>` (sin `/` final), `DNS_RESOLVER=8.8.8.8`
3. **GitHub Environments**
   - `qa` (sin reviewers) + secrets `RENDER_HOOK_API_QA`, `RENDER_HOOK_FRONT_QA`
   - `production` + **Required reviewers** = vos; *Prevent self-review* **OFF** + secrets `RENDER_HOOK_API_PROD`, `RENDER_HOOK_FRONT_PROD`
4. **Variables de repo** (Settings → Actions → Variables): `QA_API_URL`, `QA_FRONT_URL`, `PROD_API_URL`, `PROD_FRONT_URL`
5. Tras el primer push a `main` que publique: hacer **públicos** los dos packages en GHCR; `docker logout ghcr.io` y `docker pull …:sha-<commit>`
6. Flujo: merge → deploy-qa automático → Waiting for review en prod → **rechazar una vez** con motivo propio → aprobar otra → tag `v6.0.0` en el SHA de PROD + release
7. Rollback cronometrado: redeploy del SHA de `v6.0.0` vía hook; anotar minutos desde *Deploys → live*

### 1. Artefacto: por qué sólo se publica con verificación verde

La imagen sale del **mismo job** que corrió los tests, y el paso «Construir y publicar» es el **último**. Si el `docker run` de tests falla, el job muere y nunca se llega a `docker/login` ni al push. Además: sólo `push` a `main` publica (`push: ${{ github.event_name == 'push' && github.ref == 'refs/heads/main' }}`). Cadena de tres eslabones: (1) nada entra a main sin CI+coverage, (2) sólo main publica, (3) publicar es lo último del job que testeó.

Si se publicara con tests en rojo, el registry dejaría de significar «esto pasó la verificación» y pasaría a ser un depósito de builds cualquiera.

### 2. Continuous Delivery (no Deployment)

Implementamos **Continuous Delivery**: QA automático tras merge; PROD con aprobación humana (`environment: production` + required reviewer). Continuous Deployment sacaría al humano: hoy no tenemos monitoreo/alertas suficientes para eso. El contexto de la materia (un desarrollador, free tier, demos) pide el gate de negocio.

### 3. Diseño de la cadena y alcance de secrets

```
build-backend ∥ build-frontend  →  deploy-qa (environment: qa)  →  deploy-prod (environment: production)
```

- `needs` + `if: github.ref == 'refs/heads/main'` en `deploy-qa`: los PRs verifican y no despliegan.
- `deploy-prod` no repite el `if`: hereda la compuerta porque `needs: deploy-qa`.
- Secrets de hooks viven **en el environment** (`--env qa` / `--env production`): un job de QA no puede leer los hooks de PROD.
- URLs públicas van en **variables** de repo (`vars.*`), no en secrets: no son secretas y el smoke las necesita.

### 4. Qué mira el aprobador antes de aprobar

1. ¿`deploy-qa` verde y smoke OK en esa corrida?
2. ¿El commit del run es el que quiero (no otro merge que se coló)?
3. ¿El cambio es seguro para usuarios reales ahora (timing de demo/defensa)?
4. ¿Hay migraciones destructivas? (rollback de código no deshace schema)

Motivo de rechazo (ejemplo a personalizar en la corrida real): *«Smoke de QA pasó contra cold-start de la versión anterior; espero a que Render marque live el SHA de esta corrida antes de promover.»*

### 5. Free tier (letra chica)

- **Render free**: sleep tras idle → cold start; builds lentos. El smoke reintenta 30×20s.
- **Neon free**: permanente (a diferencia del Postgres de Render, que expira ~30 días) — por eso Neon.
- **GHCR**: paquetes públicos sin credenciales para quien corrige.

### 6. Garantía que se pierde porque Render reconstruye

En TP6 el hook pide a Render **rebuild desde el repo** del `&ref=$GITHUB_SHA`. Promovemos el **mismo commit**, no la **misma imagen** que el job publicó en GHCR. Puede diferir (base image, dependencia flotante). El TP7 cierra eso desplegando la imagen del registry.

### 7. Qué prueba el smoke (y qué no)

Prueba: proceso vivo (`/api/health`), BD alcanzable (`database: connected`), front sirve HTML y proxy `/health`. **No** prueba: que el commit sea el esperado (mitigación opcional: comparar `commit` del JSON con `github.sha` — el campo ya existe en health), ni flujos autenticados, ni que no quede sirviendo la versión anterior mientras Render buildea.

### 8. Config fuera de la imagen (frontend)

| En la imagen | Por variable de entorno |
|---|---|
| HTML/JS/CSS estáticos | `BACKEND_URL`, `DNS_RESOLVER` (plantilla nginx) |
| Binario Go + migraciones embebidas | `DATABASE_URL`, `SESSION_SECRET`, `APP_ENV`, `COOKIE_SECURE`, `AUTO_MIGRATE` |

Misma imagen de front sirve QA y PROD; sólo cambian las env del servicio Render.

### 9. Deployment pattern elegido + rollback

Para Flow en producción real con usuarios: **rolling** (o recrear instancia única en free tier) + **feature flags** para cambios de UX. Blue-green pide dos slots; canary pide métricas/tráfico partido que hoy no tenemos.

**Plan de rollback actual (medido tras el primer PROD):**
1. Leer SHA bueno: `gh api 'repos/mateopappa/ingsoft3-tp01/deployments?environment=production' --jq '.[0].sha'` o el de `v6.0.0`.
2. `curl` a los Deploy Hooks de PROD con `&ref=$SHA_ANTERIOR`.
3. Esperar *Deploys → live* en api y front; anotar minutos.
4. Verificar `/api/health` y el front.
5. Limitación: migraciones destructivas **no** se revierten con esto.

### 10. Fallback local

Quedan `compose.qa.yml` / `compose.prod.yml` (`name:` distintos, puertos 8080/3000 vs 8081/3001, volúmenes `db_data_qa`/`db_data_prod`) por si el free tier de Render falla. El plan C exige self-hosted runner y que Actions dispare el `compose up` — no levantarlo a mano.

### 11. Problemas / IA

- Token del agente sin permiso para crear environments/secrets → setup documentado para el dueño del repo.
- Health antes mentía `database: connected` sin pingear → ahora hace `PingContext` real.
- IA (Cursor): workflow, plantilla nginx, compose de fallback y esta sección. Verificación: orden de pasos en `ci.yml`, tests Go, y tras el cableado manual las corridas de Actions/Deployments.



