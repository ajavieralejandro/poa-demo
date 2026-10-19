# Build, deploy y rollback posteriores

Esta auditoría no desplegó. Los pasos quedan para cuando el cambio de los documentos 03 y 05 esté implementado y la matriz 06 pase en local.

Reiniciar el proceso Go **siempre** pierde la cadena en memoria. Hoy eso vuelve a un génesis que enseguida sigue minando. Después del cambio, el proceso nuevo escucha en pausa, con generación 1 y tres bloques. No recupera la demo que había en RAM. No hay archivo que salvar.

## Build local antes de copiar nada

Backend, en `poa-demo`, con el mismo Go 1.27 del módulo:

```powershell
go test ./...
go build -o poa-backend.exe ./cmd/server
```

En el servidor el artefacto se llama `poa-backend`, sin `.exe`. Cruzar la compilación desde Windows solo si `GOOS=linux` y `GOARCH=amd64` coinciden con la máquina. Si se compila allí, mejor: el binario y el `HEAD` quedan en el mismo sitio.

Frontend, en `poa-front`, con la URL del sitio, no vacía y no `127.0.0.1` (una base vacía publica el mock; `127.0.0.1` haría que el navegador del público llame a su propia máquina):

```powershell
$env:VITE_API_BASE_URL = "http://24.199.80.211"
npm ci
npm run build
```

El valor queda incrustado. Cambiarlo exige otro build.

Probar el binario nuevo en un puerto distinto de 8080, con la matriz 06, antes de sustituir `/var/www/poa-demo/poa-backend`.

## Qué leer en el servidor antes de reemplazar

Sin reiniciar:

```bash
systemctl cat poa-backend.service
sha256sum /var/www/poa-demo/poa-backend
cp -a /var/www/poa-demo/poa-backend /var/www/poa-demo/poa-backend.bak
cp -a /var/www/poa-front /var/www/poa-front.bak
```

Anotar `ExecStart`, `WorkingDirectory`, `Environment` y `EnvironmentFile`. El binario nuevo tiene que quedar en la ruta que la unit ya ejecuta. `ALLOWED_ORIGINS` debe incluir el origen del navegador si el front no es same-origin con la API. `DIFFICULTY`, `INITIAL_SUPPLY`, `MAX_BLOCKS` y `BLOCK_INTERVAL_MS` se dejan como estén, salvo que la unit no los defina: los defaults del código son 16, 1000 satoshis, 1000 bloques y 2000 ms. No bajar `MAX_BLOCKS` en el servidor para acortar la demo, y no borrar bloques de la cadena en memoria para imponer el tope: el proceso nuevo, al reiniciarse, nace pausado en tres bloques.

No editar el `ProxyPass` de Apache.

## Deploy

1. Copiar el binario nuevo sobre la ruta de `ExecStart`, o a esa ruta con otro nombre y ajustar la unit solo si el equipo ya lo hace así.
2. Publicar el contenido de `dist/` en `/var/www/poa-front`, reemplazando el `index.html` y los assets que el HTML referencia.
3. Reiniciar el servicio una vez:

```bash
sudo systemctl restart poa-backend.service
sudo systemctl status poa-backend.service --no-pager
```

Ese restart tira la cadena anterior. La demo que vea la cátedra a continuación es la nueva: pausada, tres bloques.

4. Comprobar en la propia máquina, sin pasar todavía por el navegador público:

```bash
curl -sS http://127.0.0.1:8080/api/v1/chain
sleep 5
curl -sS http://127.0.0.1:8080/api/v1/chain
```

Las dos respuestas tienen tres bloques, `height` 2, `status` `paused` y la misma `generation`. Si la segunda creció, el bucle sigue autónomo y el deploy no está aceptado.

5. Recién ahí abrir `http://24.199.80.211` en dos pestañas y recorrer las filas 2 a 11 de la matriz. El reinicio de la fila 9 sí afecta la demo compartida: hacerlo sabiendo que las dos pestañas van a volver a tres bloques.

## Rollback

1. Restaurar el binario:

```bash
cp -a /var/www/poa-demo/poa-backend.bak /var/www/poa-demo/poa-backend
sudo systemctl restart poa-backend.service
```

Ese segundo restart también pierde la cadena. El binario viejo, el de esta auditoría, al volver a escuchar mina solo desde el génesis. No reaparece la demo pausada de tres bloques ni la cadena que hubiera en el momento del rollback.

2. Restaurar el estático:

```bash
rm -rf /var/www/poa-front
cp -a /var/www/poa-front.bak /var/www/poa-front
```

El frontend no tiene cadena. Revertir `dist` no borra bloques. Sí vuelve el botón que hace `POST /api/v1/blocks` si el backup era el build de `39a12ff`.

3. Dejar Apache como estaba.

No hay migraciones que deshacer.
