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

### 4. Estructura del Backlog para Flow (TP4)
Se estructuró el backlog del proyecto de forma simple y clara para la verificación automática de la app Flow:
- **Épica (#6)**: `EPIC: Integración Continua (CI) para Flow`
- **Historia de Usuario (#7)**: `CI: Verificación automática de la app en cada Pull Request` (con Criterios de Aceptación sencillos de compilación).
- **Tareas Técnicas (#8, #9, #13)**:
  - `#8`: `Escribir el workflow inicial .github/workflows/ci.yml` (Cerrada automáticamente vía PR #11 para demostrar la trazabilidad).
  - `#9`: `Configurar la compilación de la imagen Docker en el pipeline` (En backlog).
  - `#13`: `Proteger la rama main con el check obligatorio del CI` (En backlog).
- **Bug (#10)**: `El servidor no arranca si falta el archivo .env` (Bug independiente al costado del árbol).

### 5. Declaración de Uso de IA
Este TP fue realizado con la asistencia del agente de IA **Antigravity** (Google DeepMind) para la generación de labels, automatización de la creación de issues vía CLI de GitHub, vinculación de trazabilidad entre PR e issues y redacción de la sección de decisiones.




