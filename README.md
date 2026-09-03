# Flow — Personal Activity Tracker

[![CI](https://github.com/mateopappa/ingsoft3-tp01/actions/workflows/ci.yml/badge.svg)](https://github.com/mateopappa/ingsoft3-tp01/actions/workflows/ci.yml)

Repositorio oficial del proyecto práctico de **Ingeniería de Software III (DevOps) — 2026**.

Aplicación web contenerizada de 3 capas para seguimiento de tiempo y productividad personal, con pipelines de integración continua declarativos y branch protection activa.

---

## 🚀 Arquitectura del Sistema

- **Frontend**: SPA Vanilla con CSS moderno servida por servidor web **Nginx:alpine** actuando además como reverse proxy (`/api/` hacia el backend).
- **Backend**: API REST de alto rendimiento en **Go 1.24** con compilación estática multi-stage (`CGO_ENABLED=0`, binario de 23.5 MB sobre Alpine).
- **Base de Datos**: **PostgreSQL 16** con volumen Docker administrado para persistencia de datos.
- **CI / DevOps**: Pipeline as Code con **GitHub Actions** (`.github/workflows/ci.yml`), Docker Buildx, caché GHA y Required Status Checks estrictos.

---

## 🛠️ Ejecución Rápida

Para clonar y levantar el entorno completo localmente en 2 comandos:

```bash
# 1. Copiar plantilla de variables de entorno
cp .env.example .env

# 2. Construir e iniciar contenedores
docker compose up -d --build
```

### URLs de Acceso:
- **Frontend SPA**: [http://localhost:3000](http://localhost:3000)
- **Healthcheck Directo**: [http://localhost:8080/api/health](http://localhost:8080/api/health)
- **API a través de Nginx**: [http://localhost:3000/api/health](http://localhost:3000/api/health)

---

## 📚 Documentación de Trabajos Prácticos

- [decisiones.md](decisiones.md): Registro acumulativo de decisiones técnicas y justificaciones de arquitectura (TP1 a TP4).
- [defensa-oral.md](defensa-oral.md): Ayudamemoria y preguntas clave para la defensa oral del coloquio.
- [evidencias.md](evidencias.md): Registro fotográfico y comandos de verificación de TPs previos.
