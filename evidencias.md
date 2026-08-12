# Evidencias — TP1

## 1. Push directo a `main` rechazado

```
remote: error: GH006: Protected branch update failed for refs/heads/main.
remote:
remote: - Changes must be made through a pull request.
To https://github.com/mateopappa/ingsoft3-tp01.git
 ! [remote rejected] main -> main (protected branch hook declined)
error: failed to push some refs to 'https://github.com/mateopappa/ingsoft3-tp01.git'
```

GitHub rechaza el push porque `main` está protegida con `enforce_admins: true`,
lo que significa que la regla alcanza también al dueño del repositorio.
Todo cambio debe entrar a través de un Pull Request.

---

## 2. PR de la rama B no se puede mergear: conflicto

El PR #3 (`feature/titulo-b`) quedó con estado `CONFLICTING` / `DIRTY` en la API de GitHub
luego de que el PR #2 (`feature/titulo-a`) fue mergeado a `main`.

```json
{ "mergeable": "CONFLICTING", "mergeStateStatus": "DIRTY" }
```

GitHub mostraba el aviso: **"This branch has conflicts that must be resolved"**
porque ambas ramas modificaron la misma primera línea del `README.md`.

---

## 3. Marcadores del conflicto en `README.md`

Al hacer `git merge origin/main` desde la rama `feature/titulo-b`,
Git no pudo resolver automáticamente y dejó los marcadores:

```
<<<<<<< HEAD
# Proyecto IngSoft3 - versión B
=======
# Proyecto IngSoft3 - versión A
>>>>>>> origin/main
```

- `HEAD` = la versión de la rama actual (`feature/titulo-b`)
- La parte debajo de `=======` = lo que ya estaba en `main` (versión A)

Se resolvió **manualmente**: se eligió la versión B, se borraron los tres marcadores,
y se commiteó el resultado con `fix: resuelve conflicto de título tomando la versión B`.

---

## 4. Release `v1.0.0` publicada

La release `v1.0.0` fue publicada exitosamente en GitHub con el tag anotado:

- **URL**: https://github.com/mateopappa/ingsoft3-tp01/releases/tag/v1.0.0
- **Tag**: `v1.0.0` (anotado, `git tag -a`)
- **Notas**: Incluyen el resumen de todo lo que abarca esta primera versión estable.
