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






