# Estado actual y versiones

Fecha de la lectura local: 19 de octubre de 2026.

No existe `AGENTS.md` en `poa-demo` ni en `poa-front`. Las instrucciones de build salen de `README.md`, `package.json`, `.env.example` y del código de `cmd/server`.

No se accedió a `24.199.80.211`, a `/var/www/poa-demo` ni a `/var/www/poa-front`. Todo lo que sigue sobre producción es una lista de lectura, no una comprobación.

## Backend local auditado

| Dato | Valor |
| --- | --- |
| Ruta | `C:\Users\Dell\Desktop\Javi\Ciberseguridad\poa-demo` |
| Remoto | `https://github.com/ajavieralejandro/poa-demo.git` |
| Rama | `main`, tracking `origin/main` |
| HEAD | `6ec36945468f9e387de1633659fe3126758a7da5` |
| Fecha | 2026-10-19 08:55:01 +0200 |
| Mensaje | `api changes to deploy` |
| Autor | Javier Alejandro Amorosi Gómez |
| Árbol de trabajo | Limpio. `git status -sb` → `## main...origin/main` |
| Fetch | `git fetch --no-prune origin`. `origin/main` coincide con HEAD. No se usó prune, merge ni push |
| Go | `go 1.27` en `go.mod`. El laboratorio compiló con `go version go1.27.0 windows/amd64` |
| Otras ramas locales | Ninguna. Solo `main` y `remotes/origin/main` |

Ese commit agrega `cmd/server`, `internal/api`, `internal/config` e `internal/engine`. Es el código que esta auditoría ejecutó en local.

`README.md` del backend quedó atrasado: todavía dice que minería, selección y firmas devuelven error y que el comando de la demo es `go run ./cmd/demo`. El servidor real de la API es `go run ./cmd/server`. `.env.example` sí lo documenta.

No hay `.env` en el árbol local. El proceso de prueba usó los defaults de `config.Load`: puerto 8080 (el laboratorio lo pisó con `PORT=18080`), `DIFFICULTY=16`, `INITIAL_SUPPLY=1000`, `BLOCK_INTERVAL_MS=2000`.

## Frontend local auditado

| Dato | Valor |
| --- | --- |
| Ruta | `C:\Users\Dell\Desktop\Javi\Ciberseguridad\poa-front` |
| Remoto | `https://github.com/ajavieralejandro/poa-front.git` |
| Rama | `main`, tracking `origin/main` |
| HEAD | `39a12ffe211d05c05e8b4b0da0dc087ba0f2da6c` |
| Fecha | 2026-10-18 12:08:03 +0200 |
| Mensaje | `skeleton made` |
| Árbol de trabajo | Sucio. No se modificó en esta auditoría |
| Fetch | `git fetch --no-prune origin`. `origin/main` coincide con HEAD |

Cambios locales preservados, sin stage:

- Modificados: `.env.example`, `.gitignore`, `src/env.d.ts`, `src/hooks/useBlockchainSimulation.ts`, `src/services/blockchainApi.ts`, `src/types/blockchain.ts`
- Sin seguimiento: `src/config/`, `src/hooks/useBlockStream.ts`, `src/services/blockStream.ts`

No hay `.env`. `VITE_API_BASE_URL` vacío significa mock.

Esos cambios locales ya describen el stream y leen `supply` y `difficulty` del snapshot. El commit `39a12ff` no lo hace: fija `POA_NETWORK` y, al pulsar Iniciar, llama a `POST /api/v1/blocks`. El síntoma observado (dificultad 4 y POST 404) coincide con ese commit. El árbol sucio, si se construyera con URL de API, mostraría 16 y no haría ese POST. No está commiteado y no controla al minero.

## Bundle local `dist/`

`poa-front/dist` está en `.gitignore`. Sus archivos tienen fecha 2026-10-18 09:41, anterior al commit `skeleton made` de las 12:08.

Ese bundle es otra interfaz:

- El texto visible habla de `/api/v1/chain` y `/api/v1/blocks`. No contiene `EventSource` ni `/api/v1/stream`.
- El cuerpo del POST usa `prevHash`, no `prev`.
- Los bloques embebidos son planos (`height`, `prevHash`, `winners`, `difficulty: 4`), no el `header` actual.
- Vite dejó la base de API compilada como cadena vacía (`function hd(){return\`\`.replace...}`), así que ese artefacto ni siquiera llama a Go.

No es el código de `39a12ff` ni el árbol de trabajo. No se puede tratar como lo que hay en `/var/www/poa-front`. El botón “Iniciar simulación” está en el fuente actual (`SimulationControls`); ese `dist` no contiene la palabra “simulación”.

## Código desplegado

Referencias que dio el pedido, no verificadas aquí:

| Pieza | Ruta o nombre |
| --- | --- |
| Código backend | `/var/www/poa-demo` |
| Binario | `/var/www/poa-demo/poa-backend` |
| Servicio | `poa-backend.service` |
| Frontend | `/var/www/poa-front` |
| URL | `http://24.199.80.211` |
| Proxy | Apache reenvía `/api/` a `http://127.0.0.1:8080/api/` |

El 404 de `/api/v1/blocks` se reprodujo en un proceso Go local. No hace falta cambiar Apache para inventar esa ruta.

## Comandos de lectura en el servidor

Ejecutarlos en la máquina, sin reiniciar el servicio y sin `systemctl restart`. Son de inspección.

```bash
systemctl status poa-backend.service --no-pager
systemctl show poa-backend.service \
  -p FragmentPath -p ExecStart -p Environment -p EnvironmentFiles \
  -p WorkingDirectory -p User
systemctl cat poa-backend.service
sha256sum /var/www/poa-demo/poa-backend
cd /var/www/poa-demo && git rev-parse HEAD && git status -sb && git log -1 --format='%H %ci %s'
cd /var/www/poa-front && git rev-parse HEAD && git status -sb && git log -1 --format='%H %ci %s'
grep -R -l -a "api/v1" /var/www/poa-front --include='*.js'
grep -a -o -E "api/v1/[A-Za-z]+|Iniciar simulación|EventSource|prevHash|difficulty" \
  /var/www/poa-front/assets/*.js /var/www/poa-front/dist/assets/*.js
curl -sS http://127.0.0.1:8080/api/v1/chain
curl -sS -D - -o /tmp/poa-post.txt -X POST http://127.0.0.1:8080/api/v1/blocks \
  -H 'Content-Type: application/json' \
  --data '{"prev":"aa","height":0}'
curl -sS -N --max-time 3 http://127.0.0.1:8080/api/v1/stream
```

Comparar el `HEAD` de `/var/www/poa-demo` con `6ec36945468f9e387de1633659fe3126758a7da5`. Si el bundle servido contiene `EventSource`, el frontend desplegado incluye el árbol local o un build posterior. Si contiene `POST` a `/api/v1/blocks` y no contiene `EventSource`, coincide con `39a12ff`.

Para saber si Apache bufferiza el stream, leer el virtual host. No cambiarlo en esta fase:

```bash
apachectl -S
grep -R -n "ProxyPass /api" /etc/apache2 /etc/httpd
```

`X-Accel-Buffering: no`, que el handler Go ya envía, es una cabecera de nginx. Apache no la interpreta.

## Instrucciones de build reales

Backend, desde la raíz del módulo:

```powershell
go test ./...
go build -o poa-backend ./cmd/server
```

El binario escucha `:$PORT` y mina en el mismo proceso. `go run ./cmd/demo` es el stress test de 100 bloques: no es el servidor de la interfaz.

Frontend:

```powershell
npm ci
npm run build
```

`vite build` incrusta `VITE_API_BASE_URL` en el bundle. Cadena vacía: mock, sin red. Con URL: el commit `39a12ff` usa `GET /api/v1/chain` y `POST /api/v1/blocks`. No hay modo “ruta relativa = mismo origen”. Un build de producción tiene que llevar el origen público, por ejemplo `http://24.199.80.211`, si Apache publica API y estáticos en ese host.
