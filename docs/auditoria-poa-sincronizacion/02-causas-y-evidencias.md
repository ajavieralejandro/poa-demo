# Causas demostradas y evidencias

Leyenda: **demostrado** significa leído en el fuente citado o medido en el laboratorio local. **Hipótesis** es una explicación que encaja y que no se ejecutó contra el servidor público.

Laboratorio: proceso aislado `go run ./cmd/server` en `127.0.0.1:18080`, sin `.env`, con los defaults (dificultad 16, supply 1000, intervalo 2000 ms). Primera corrida 09:12:28–09:12:53. Segunda corrida, con reinicio de proceso, 09:14. Al terminar, el puerto quedó libre. No se llamó a `24.199.80.211`.

## Backend

### Arranque

`main` en `cmd/server/main.go` carga la config, construye el engine, crea el hub y lanza `runMiner` antes de bloquearse en `ListenAndServe`. No hay bandera de pausa.

`engine.New` (`internal/engine/engine.go`) crea un solo bloque con `blockchain.NewGenesisBlock`, reparte el supply entre 100 direcciones (`validatorNodes`) con `miners.NewDeterministicPool` y guarda la cadena en memoria. El test `TestMineNextAppendsVerifiedBlock` parte de un bloque y solo llega a dos después de llamar a `MineNext`. El minado automático no está en el engine: está en `runMiner`.

El génesis no pasa por `Mine` ni por las tres firmas. `consensus.Verify` retorna nil si `Height == 0`. En el laboratorio el génesis salió así:

```json
{"header":{"prev":"0000…0000","miner":"0","height":0,"nonce":0},"signatures":[],"hash":"7955cb2de90dd9efc6df9fdbf5f5d10c114f4135a9a6b52db1003be749e32f7a","verified":true}
```

`verified: true` lo pone `NewBlockView` para todo bloque que ya está en la cadena. No vuelve a verificar al serializar.

### Bucle y `BLOCK_INTERVAL_MS`

`runMiner` llama a `MineNext`, publica el JSON y después espera `cfg.BlockInterval`. Si `MineNext` falla, registra el error y espera un segundo. El intervalo es la pausa posterior a un bloque aceptado, no un reloj que incluya el minado.

En el laboratorio el log fue `dificultad 16, supply 1000, intervalo 2s` y un bloque aceptado cada unos 2 segundos (altura 1 a las 09:12:28, altura 13 a las 09:12:53). El minado de 16, implementado como dos bytes cero, fue rápido en esta máquina. Una dificultad más alta alargaría cada vuelta por encima del intervalo. No se cambió la dificultad para la prueba.

### `MineNext`, selección, firmas y verificación

`MineNext` es el único escritor. Toma el lock, avanza `cursor`, clona la cadena, suelta el lock y mina fuera. Al volver, solo inserta si `next.Header.Prev == chain.Tip`.

La ronda real está en `mineAndSeal`:

1. `node.Mine` → `consensus.Mine`. Busca un nonce cuyo hash cumpla `cumpleDificultad`. Devuelve el bloque con las tres firmas vacías. El `Select` que hace ahí usa la cabecera del padre y solo sirve para saltar un nonce si ese sorteo devuelve `ErrMultySign`.
2. `consensus.Select(nextBlock.Header, chain)` elige tres satoshis sobre la cabecera nueva.
3. Los dos primeros dueños pasan por `Node.SignBlock`. El tercero pasa por `Node.ProposeBlock`, que exige dos firmas previas, firma y llama a `consensus.Verify`.
4. `mineAndSeal` verifica otra vez.

`SignBlock` no llama a `signing.Sign`. Escribe `Signature{Signer: dirección, Value: satoshi}`. `signing.Sign` y `signing.Verify` siguen devolviendo `ErrNotImplemented`. La “firma” del JSON es la atribución del satoshi, no bytes de una firma criptográfica.

`consensus.Verify` comprueba el padre contra el tip, la dificultad y que cada firma válida esté en el sorteo de `Select(block.Header, chain)`. El comentario del código dice que el sorteo usa al padre. La llamada usa la cabecera del bloque nuevo, la misma que usó `mineAndSeal`. En el laboratorio los bloques se verificaron y se insertaron. El comentario está desfasado. Eso no detuvo la cadena.

`Node.ProposeBlock` recibe `txs == nil`. No hay transferencia de satoshis en la demo HTTP.

### Concurrencia

`Snapshot` toma `RLock` y puede correr mientras `MineNext` mina fuera del lock. Ve la cadena ya confirmada, sin el bloque a medio buscar.

Un segundo `MineNext` concurrente fallaría el chequeo de tip. Hoy solo hay un `runMiner`, así que esa carrera no ocurre. El hueco aparece si un reinicio reemplaza la cadena mientras una ronda sigue viva: el chequeo de tip no basta cuando la cadena nueva es determinista y el padre coincide. Más abajo está la evidencia de que el bloque 1 se repite entre procesos.

`cursor` avanza antes de saber si la ronda tendrá éxito. Una ronda fallida cambia el minero de la siguiente. No explica la altura congelada.

### Persistencia

La cadena vive en `Engine.chain`, un mapa en RAM. No hay escritura de bloques a disco. El único `os.WriteFile` del módulo está en el test de config.

Al matar el proceso y arrancar otro:

- Primera instancia, a los 3 s: 3 bloques, altura 2, hash `0000eefd71d8…`.
- Segunda instancia, primera lectura: 2 bloques, altura 1, hash `0000c59bd0a7…`, el mismo prefijo del bloque 1 de la primera instancia.

Reiniciar Go pierde la cadena y vuelve a minar desde el génesis. Recargar el navegador no hace eso: el proceso sigue y la cadena crece.

### Hub y SSE

`Server.Handler` registra solo:

- `GET /api/v1/chain`
- `GET /api/v1/stream`

`handleStream` exige `http.Flusher`, manda `: connected`, y por cada payload escribe:

```text
event: block
data: {JSON}\n\n
```

Cada 15 s manda `: ping`. No envía `id:`. No retransmite historia. No hay evento de estado.

`Hub.Publish` copia el payload a cada cliente. Si el canal del cliente (buffer 8) no acepta, el `select` cae en `default` y ese cliente pierde el evento. El minero no espera. Esta omisión está en el código. El laboratorio no la provocó: el cliente leyó a tiempo y el ritmo era de un bloque cada ~2 s.

`runMiner` publica después de que `MineNext` retorna, es decir, después de insertar. Un bloque rechazado por el tip no se anuncia.

### Qué pasa al reiniciar el proceso

Se destruye la memoria. El proceso nuevo ejecuta `engine.New` y `runMiner` otra vez. No queda pausado. La primera lectura del laboratorio, en las dos instancias, ya traía el génesis y el bloque 1, porque el bloque 1 se emitió en el mismo segundo en que el log anunció la API.

## Frontend

Hay dos clientes en el disco. El que coincide con el síntoma es el commit. El otro es el árbol local.

### Commit `39a12ff` (origen remoto)

`fetchChainSnapshot` (`src/services/blockchainApi.ts`):

- Sin `VITE_API_BASE_URL`, espera 160 ms y clona `src/data/mockChain.json` (6 bloques, alturas 0 a 5, minados por `scripts/generate-mock.ts` con dificultad local 4).
- Con URL, hace `GET {base}/api/v1/chain` y castea el JSON a `ChainSnapshot` sin validar la forma.

`useBlockchainSimulation.publish` guarda los bloques, pero asigna `configRef` y `setConfig` a `POA_NETWORK` (`difficulty: 4`, `totalSupply: 1000`, seis validadores con nombre). El estado pasa a `paused`. No lee `snapshot.difficulty` ni `snapshot.supply`: en ese commit `ChainSnapshot` solo declara `blocks`.

No importa `useBlockStream`. No hay `EventSource` en ese commit.

El efecto de minado corre cuando `status === "running"`. `play` pone `running` si había bloques y config. Cada vuelta llama a `requestNextBlock`. Con URL de API, esa función hace:

```json
POST {base}/api/v1/blocks
{"prev":"<hash del tip>","height":<altura del tip>}
```

Si la respuesta no es `ok`, lanza `El backend no pudo minar el bloque`. El efecto pone el error y vuelve a `paused`. No agrega un bloque. No llama a `mineBlock` cuando la URL existe: el mock solo se usa si la base está vacía. Un 404 deja la cadena ya cargada en pantalla y muestra el error. Eso es el fallo de “datos anteriores visibles como actuales”.

`pause` solo baja el estado local a `paused`, lo que corta el temporizador del efecto. `reset` vuelve a `fetchChainSnapshot` y publica lo que el servidor tenga en ese momento. No pide a Go que vacíe la cadena.

El slider `intervalMs` nace en 2000 y solo cambia `intervalRef` del navegador. Go no lo recibe.

`App` muestra como altura `blocks.at(-1).header.height`. `ChainView` muestra aparte `blocks.length`. Con una cadena lineal desde el génesis, la cantidad de bloques es la altura más uno. No hay altura máxima en ninguno de los dos lados.

### Árbol de trabajo, sin commitear

`readAppConfig` define `dataSource: "api"` cuando la URL no está vacía.

`publish` arma la config con `snapshot.supply` y `snapshot.difficulty`, y si la fuente es API pone el estado en `running` de inmediato. El botón Iniciar queda deshabilitado (`disabled={status !== "paused"}`) y el encabezado dice “En curso” sin que nadie haya pulsado.

`useBlockStream` se suscribe solo si `dataSource === "api"` y `status === "running"`. `openBlockStream` escucha el evento `block`, filtra con `isBlock` (exige `hash` string, `header.height` number, `header.prev` string y `signatures` array) y cierra el `EventSource` al desmontar o cuando `enabled` pasa a falso.

Al recibir un bloque, si el hash ya está, lo ignora. Si `prev` no es el hash del tip local o la altura no crece, llama a `resync` y no lo agrega. `resync` vuelve a pedir `GET /chain` y reemplaza el estado. No incrementa `epoch`. Si el refetch falla, escribe el error y conserva los bloques viejos.

El efecto que llama a `requestNextBlock` retorna enseguida cuando `dataSource === "api"`. Por eso este árbol, con URL, no hace el POST. La función `requestNextBlock` sigue existiendo.

`pause` cierra el stream local. El backend sigue en `runMiner`. Al volver a `running`, un efecto pide `resync` y la interfaz salta a la altura que Go acumuló durante la pausa.

Esto arreglaría la altura congelada si el stream llega al navegador, y dejaría intacto el minado autónomo. No cumple el inicio pausado ni el reinicio compartido.

### Controles, en una frase

| Control | Commit `39a12ff` con API | Árbol local con API | Backend actual |
| --- | --- | --- | --- |
| Carga | `GET /chain`, queda en pausa, dificultad 4 | `GET /chain`, pasa a “en curso”, dificultad del JSON | Sigue minando |
| Iniciar | `POST /blocks` | Botón deshabilitado | No existe la ruta |
| Pausar | Corta el temporizador de React | Cierra el `EventSource` | El bucle sigue |
| Reiniciar | Otro `GET /chain` | Otro `GET /chain` | La cadena no se reemplaza |
| Slider | Ritmo del mock o de la espera tras el POST | No afecta al bucle de API | `BLOCK_INTERVAL_MS` fijo al arrancar |

Origen de las métricas en el commit, que es el que muestra 4:

- Supply: `POA_NETWORK.totalSupply` (1000). Coincide con el default de Go por azar. Si el servidor usara otro `INITIAL_SUPPLY`, el panel seguiría en 1000.
- Dificultad: `POA_NETWORK.difficulty` (4). El JSON trae 16 y el commit lo descarta.
- Intervalo: `useState(2000)`. No viene del servidor.
- Altura: último `header.height` recibido en el snapshot. No se actualiza hasta otra carga o un POST exitoso.

## Causas del incidente

### Demostradas

1. **El minero no depende del navegador.** `runMiner` arranca con el proceso. En el laboratorio, sin ningún stream abierto, la cadena pasó de 2 bloques (altura 1) a 5 bloques (altura 4) en 6 segundos. `POST /api/v1/blocks` respondió `404 page not found` (19 bytes, `Content-Type: text/plain`) y las dos lecturas paralelas siguieron en la misma altura 2, el mismo hash. El POST no crea bloques.
2. **El cliente commiteado no escucha `block`.** Por eso una altura ya pintada no se mueve, mientras Go sigue. Al recargar, `GET /chain` devuelve lo acumulado. Eso reproduce “al recargar aparecen bloques de más” sin reiniciar el proceso.
3. **Iniciar llama a una ruta que Go no registra.** El 404 está en `Server.Handler` y se midió con un cuerpo JSON válido (`{"prev":"aa","height":0}`) contra `127.0.0.1:18080`. `GET /api/v1/blocks` también responde 404. Apache no interviene en ese resultado.
4. **La dificultad 4 y la 16 no son el mismo parámetro.** El panel del commit hardcodea 4. El algoritmo de cada lado está comparado en el documento de contrato. El laboratorio mostró hashes de bloque minado con al menos cuatro ceros hex (`0000c59bd0a7342b…`) y uno con cinco (`00000b1e7f1a666c…`). El génesis (`7955cb2d…`) no cumple ese prefijo, porque la dificultad no se le aplica.
5. **No hay persistencia.** Reiniciar el proceso local devolvió la cadena a altura 1. Dos clientes simultáneos del mismo proceso vieron el mismo tip (`0000eefd71d8…`, altura 2).
6. **El stream no recupera huecos.** La primera suscripción recibió alturas 5, 6, 7 y 8. Tras cerrarla, el snapshot llegó a altura 10. La reconexión recibió 11, 12 y 13. Las alturas 9 y 10 no se retransmitieron. Tampoco se retransmitieron los bloques anteriores al `: connected`.
7. **Un error de API deja la cadena vieja como si siguiera siendo la actual.** El `catch` del efecto de minado no vacía `blocks`. `fail` y el `catch` de `resync` tampoco.

### Hipótesis, no medidas en producción

- El bundle de `/var/www/poa-front` es el de `39a12ff` construido con `VITE_API_BASE_URL` apuntando al origen público. Encaja con el POST y con la dificultad 4. Hay que confirmarlo con el `grep` del documento de versiones. El `dist/` local de las 09:41 del 18 de octubre no es esa interfaz.
- Si el sitio desplegado ya llevara el árbol sucio, la altura no debería congelarse mientras el stream llegue, y no habría POST. El síntoma descrito apunta al commit, no a ese árbol.
- Apache puede acumular la respuesta SSE porque el handler solo marca `X-Accel-Buffering`. No se observó en local: `curl` recibió `: connected` y los eventos `block` al momento. Si en producción un cliente con `EventSource` se conecta y no ve eventos, el virtual host es el siguiente lugar a mirar. No es la causa del POST 404.

## Carrera ya visible en local

Entre el `GET /chain` y el `Subscribe` del stream, `runMiner` puede confirmar bloques. Esos bloques están en el snapshot siguiente y no salen en el stream de quien se suscribió tarde. El árbol local, al ver un `prev` que no encaja, pide otro snapshot. Si ese refetch pierde contra el minero, puede guardar una foto ya superada y esperar al próximo evento para notar el hueco. Con el servidor en pausa, ese “próximo evento” no llega y la pestaña se queda corta. El commit ni siquiera tiene ese refetch.
