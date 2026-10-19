# Contrato actual y contrato propuesto

No se agregaron rutas. Esta sección fija el contrato que la implementación siguiente debe cumplir.

## Lo que Go emite hoy

`GET /api/v1/chain` → 200, `application/json`, un objeto `ChainView`:

| Campo | Go | JSON medido |
| --- | --- | --- |
| `blocks` | `[]BlockView` | Array desde el génesis hasta el tip |
| `supply` | `uint32` | `1000` con el default |
| `difficulty` | `int` | `16` con el default |

No viajan `status`, `intervalMs`, `generation` ni `height` de la cadena. La altura hay que leerla del último bloque.

`BlockView` (`internal/api/dto.go`, `NewBlockView`):

| Campo | Tipo JSON | Origen |
| --- | --- | --- |
| `header.prev` | string hex de 64 caracteres, minúsculas | `Header.Prev` |
| `header.miner` | string decimal | `Address` (`uint32`) con `strconv.FormatUint` |
| `header.height` | número | `uint64` |
| `header.nonce` | número | `uint64` |
| `signatures[].signer` | string decimal | `Signature.Signer`. Se omiten las firmas con `Signer == 0` |
| `signatures[].satoshiId` | número | `Signature.Value` |
| `hash` | string hex de 64 | `Header.Hash()`, SHA-256 de prev, miner, height y nonce |
| `verified` | `true` | Constante al proyectar un bloque ya insertado |

El supply de la cabecera no entra en el hash ni en el JSON del bloque. El cliente lo ve una sola vez, en la raíz del snapshot.

Ejemplo real de altura 1, laboratorio `18080`:

```json
{
  "header": {
    "prev": "7955cb2de90dd9efc6df9fdbf5f5d10c114f4135a9a6b52db1003be749e32f7a",
    "miner": "1",
    "height": 1,
    "nonce": 61440
  },
  "signatures": [
    {"signer": "89", "satoshiId": 889},
    {"signer": "24", "satoshiId": 231},
    {"signer": "85", "satoshiId": 848}
  ],
  "hash": "0000c59bd0a7342bfb4ed873c4584e10e0988fb923b9d614b1641401bdcee21d",
  "verified": true
}
```

Tres firmas, firmantes numéricos, hash con cuatro ceros hex. El génesis tiene `signatures: []`, `miner: "0"` y un hash sin prefijo de ceros.

`GET /api/v1/stream`:

```text
: connected

event: block
data: {un BlockView}

: ping
```

El evento se llama `block`. El data es el mismo objeto que un elemento de `blocks`, sin generación. No hay `id:`.

Cualquier otro método o ruta, incluido `POST /api/v1/blocks` y `GET /api/v1/blocks`, cae en el 404 de `net/http`: cuerpo `404 page not found\n`, `Content-Type: text/plain; charset=utf-8`. CORS anuncia `Access-Control-Allow-Methods: GET, OPTIONS`.

## Lo que el TypeScript cree

Fuente actual del árbol de trabajo, `src/types/blockchain.ts`. El commit `39a12ff` es más estrecho: `ChainSnapshot` solo tiene `blocks`, y no existe `StreamedBlock`.

| Tipo | Campos | Encaje con Go |
| --- | --- | --- |
| `Header` | `prev`, `miner`, `height`, `nonce` | Encaja con `HeaderView` |
| `Signature` | `signer`, `satoshiId` | Encaja. Go manda ids decimales; el mock manda nombres (`"Aurora"`) |
| `Block` | `header`, `signatures`, `hash`, `verified` | Encaja con `BlockView` |
| `ChainSnapshot` | `blocks`, `supply?`, `difficulty?` | Encaja con `ChainView`. En el commit, supply y difficulty se ignoran |
| `MineRequest` | `prev`, `height` | No tiene receptor. `height` es la altura del tip anterior, no la del bloque nuevo |
| `ChainConfig` | `totalSupply`, `difficulty`, `validators[]` | Los validadores (Aurora…Flux y rangos de satoshis) solo existen en el cliente |

`fetchChainSnapshot` no valida en runtime. Un JSON inesperado se guarda igual. `isBlock`, solo en el stream del árbol local, rechaza un data ilegible y no toca la cadena; un snapshot malo sí puede reemplazarla.

No hay campo de generación. Dos cadenas distintas con la misma forma son indistinguibles para React.

Los números JSON de `height` y `nonce` son `float64` en el decoder de Go y `number` en JS. Con la dificultad 16 los nonces medidos cupieron en enteros exactos (el mayor visto en la corrida corta fue 129664). Un nonce por encima de 2^53 dejaría de ser exacto en el navegador. No se vio en el laboratorio. Conviene tratar altura y nonce como enteros decimales en string si más adelante la dificultad sube mucho. No es parte del arreglo mínimo.

## Dificultad: no son el mismo número

### Backend, `consensus.cumpleDificultad`

```go
bytesRequeridos := bits / 8
for i := 0; i < bytesRequeridos; i++ {
    if hash[i] != 0 { return false }
}
```

`DIFFICULTY` es un entero de 0 a 256. Solo cuenta el cociente entero de dividir por 8. El resto se tira. 16, 17 y 23 exigen lo mismo: dos bytes iniciales en cero sobre el SHA-256 de la cabecera. Dos bytes cero son cuatro caracteres hex `0`. Un quinto cero puede aparecer por azar, como el hash `00000b1e…` del laboratorio, y sigue siendo válido. 8 y 15 exigen un solo byte. 24 exige tres.

`.env.example` ya lo dice: “Solo cuentan los múltiplos de 8 (cada 8 bits = un byte cero)”. El hash incluye prev, miner, height y nonce. No incluye supply ni firmas.

El génesis no se mina. Su hash no tiene por qué empezar con ceros. En el laboratorio no lo hace.

### Frontend, `findNonce` en `src/services/consensus.ts`

`POA_NETWORK.difficulty` vale 4. `findNonce` exige que `hashHex(material)` empiece por `"0".repeat(4)`. `hashHex` es un mezclador propio (cuatro estados de 32 bits y ocho rondas), no SHA-256. El material es `altura|prev|miner|satoshis`. El nonce es el contador de ese bucle, no el nonce de Go.

El 4 cuenta nibbles hexadecimales del hash de juguete. No cuenta bits de SHA-256. Por eso 16 en Go y 4 en React no se equivalen, aunque un bloque de Go con dificultad 16 se vea con cuatro ceros hex. Esos cuatro ceros son la representación de dos bytes cero. El 4 del panel es el constante que `publish` del commit `39a12ff` pisa encima del JSON.

Si el árbol local llegara a usar `snapshot.difficulty` (16) dentro de `findNonce`, pediría 16 ceros hex del hash de juguete y el mock no encontraría nonce. Otra razón para no reutilizar ese número dentro del minero del navegador.

### Qué debe mostrar la interfaz

Cuando la fuente es API, el número visible es el entero de Go, hoy 16, y el texto tiene que decir que son bits agrupados de a 8: hacen falta `difficulty/8` bytes cero de SHA-256. No hace falta traducir 16 a 4 para “unificar” las métricas. El algoritmo del mock queda solo en el modo sin URL.

## Altura y cantidad de bloques

No hay tope de altura. `App` muestra la altura del tip. `ChainView` muestra `blocks.length`. En una cadena contigua desde el génesis:

```text
cantidad = altura + 1
```

El laboratorio lo cumplió (2 bloques y altura 1; 5 bloques y altura 4; 14 bloques y altura 13). El objetivo de “3 bloques totales” es altura 2, no altura 3.

Si el cliente pierde eventos, su altura es la del último bloque que sí aplicó, y `blocks.length` es la de su array local. Las dos cifras se congelan juntas. El servidor sigue en otra altura. Hoy no hay forma de ver esa diferencia porque el panel no recibe la altura del servidor por separado.

## Errores que conservan datos viejos

- `POST` fallido en el commit: `setError` + `setStatus("paused")`. Los bloques siguen.
- `resync` fallido en el árbol local: `setError`. Los bloques siguen y el estado sigue en `running`.
- `onError` del stream: escribe “Se interrumpió el stream…” si no había otro error. No marca la foto como vieja.
- `ChainView` solo sustituye la cadena por el mensaje de error cuando `blocks.length === 0`. Con bloques previos, el error queda en el lateral y la grilla sigue pareciendo la cadena viva.

## Contrato propuesto

Misma familia de rutas `/api/v1`, mismo `BlockView`, mismo evento `block`. Se agregan campos y tres POST de control. `POST /api/v1/blocks` no se implementa: el cliente deja de llamarlo.

### Foto de la cadena

`GET /api/v1/chain` responde 200:

```json
{
  "blocks": [ "BlockView, igual que hoy" ],
  "supply": 1000,
  "difficulty": 16,
  "intervalMs": 2000,
  "status": "paused",
  "generation": 1,
  "height": 2,
  "maxBlocks": 1000,
  "blockCount": 3,
  "limitReached": false,
  "stopReason": ""
}
```

| Campo nuevo | Significado |
| --- | --- |
| `intervalMs` | `BLOCK_INTERVAL_MS` efectivo del proceso. Solo lectura |
| `status` | `"paused"` o `"running"` |
| `generation` | Entero que empieza en 1 y sube solo en un reinicio exitoso. Identifica la cadena, no el bloque |
| `height` | `header.height` del tip. Con tres bloques es 2. En el tope es `maxBlocks - 1` |
| `maxBlocks` | `MAX_BLOCKS`. Longitud máxima de esta generación, génesis incluido. No es el supply |
| `blockCount` | `len(blocks)`. En una cadena contigua es `height + 1` |
| `limitReached` | `true` solo cuando `blockCount == maxBlocks` después de confirmar ese último bloque |
| `stopReason` | `""` si la pausa es de arranque o del botón. `"block_limit"` cuando el tope detuvo la demo |

`blocks` sigue ordenado de génesis a tip y trae la cadena completa. Llegar al tope no esconde ni borra bloques. `height` tiene que coincidir con el último elemento. Una respuesta de arranque en pausa, con `MAX_BLOCKS` mayor que 3, tiene `blockCount == 3`, `limitReached == false` y `stopReason == ""`.

`INITIAL_SUPPLY` y `MAX_BLOCKS` viajan en campos distintos (`supply` y `maxBlocks`). El default de ambos es 1000 y no se interpretan como la misma magnitud.

### Stream

`GET /api/v1/stream?generation={n}&after={altura}`

- Sin query, equivale a suscribirse en vivo a la generación actual, sin historia.
- Con `generation` distinta de la actual, el primer evento es `status` con la generación vigente. El cliente reemplaza su estado con `GET /chain` y vuelve a abrir el stream.
- Con la generación actual y `after` menor que el tip, el servidor escribe, en orden, los `BlockView` ya confirmados con `header.height > after`, y después sigue en vivo. La cadena está en memoria: la retransmisión no necesita otra infraestructura.
- Cada evento lleva `id: {generation}:{height}` para que un `EventSource` mande `Last-Event-ID` al reconectar. El servidor interpreta ese id con la misma regla que la query.

Evento de bloque, igual de compatible con el `data` actual, más la generación:

```text
id: 1:3
event: block
data: {"header":{},"signatures":[],"hash":"","verified":true,"generation":1}
```

`generation` extra no rompe a `isBlock`. Un cliente viejo que ignore el campo sigue leyendo el bloque. El cliente nuevo descarta el evento si la generación no es la suya.

Evento de estado, al iniciar, pausar, reiniciar y al abrir un stream con generación distinta:

```text
id: 2:2
event: status
data: {"status":"paused","generation":2,"height":2,"supply":1000,"difficulty":16,"intervalMs":2000,"maxBlocks":1000,"blockCount":3,"limitReached":false,"stopReason":""}
```

El mismo objeto de estado, con `maxBlocks`, `blockCount`, `limitReached` y `stopReason`, sale al pausar, al iniciar, al reiniciar, al alcanzar el tope y al abrir un stream. El `id` usa la altura del tip de esa generación. No anuncia un bloque.

Al confirmar el bloque que completa `maxBlocks`, el orden es fijo: primero `event: block` de ese bloque, después `event: status` con `status=paused`, `blockCount=maxBlocks`, `height=maxBlocks-1`, `limitReached=true` y `stopReason=block_limit`. La cadena publicada es la completa.

El comentario `: connected` y el `: ping` cada 15 s se mantienen. `Cache-Control: no-cache` y `X-Accel-Buffering: no` se mantienen. CORS pasa a anunciar `GET, POST, OPTIONS`.

### Control

Los tres POST devuelven el mismo JSON que `GET /api/v1/chain` y publican un evento `status`. Son idempotentes, salvo el reinicio, que siempre crea otra generación.

`POST /api/v1/simulation/start`

- Cuerpo vacío.
- Si ya estaba `running` y `blockCount < maxBlocks`, responde 200 con la foto actual y no mina un bloque extra por el hecho de repetir la llamada.
- Si estaba `paused` y `blockCount < maxBlocks`, pasa a `running` y el bucle puede empezar la ronda siguiente.
- Si `limitReached` es verdadero, responde 409 con JSON y no modifica la cadena, la generación ni el estado:

```json
{"error":"block_limit_reached","maxBlocks":1000,"blockCount":1000}
```

- Respuesta 200 con la foto cuando el start sí cambia o confirma el estado `running`. Error inesperado al persistir el estado en memoria: 500, cuerpo `text/plain` corto, sin media cadena.

`POST /api/v1/simulation/pause`

- Cuerpo vacío.
- Pasa a `paused`. No programa otra ronda.
- Una ronda que ya está fuera del lock se descarta al volver a tomarlo si el estado ya no es `running` o si la generación cambió. No se inserta y no se publica. La respuesta de la pausa ya muestra el tip quieto.
- Repetir la pausa responde 200 con la misma foto.

`POST /api/v1/simulation/reset`

- Cuerpo obligatorio:

```json
{ "confirm": true }
```

- Sin ese campo, o con otro valor: 400 y cuerpo `text/plain` `confirmacion requerida`. La cadena no cambia.
- Con `confirm: true`: bajo el mismo mutex, sube `generation`, sustituye la cadena por un engine nuevo y ejecuta dos `MineNext` de esa generación antes de responder. `status` queda `paused`. La respuesta trae tres bloques, `blockCount: 3`, `limitReached: false`, `stopReason: ""` y la generación nueva. `supply` sale igual que antes del reinicio. El tope no se salta borrando bloques de la generación anterior: esa cadena se reemplaza entera porque el reinicio fue confirmado, y la nueva vuelve a poder iniciarse.
- Si la semilla de dos bloques falla, el proceso no publica la generación a medias: responde 500 y conserva la cadena anterior. Esta es la única falla en la que el reinicio no ocurre.
- La ronda que hubiera empezado con la generación vieja ve el número nuevo bajo el lock y se tira, aunque su `prev` coincida con el tip nuevo. El laboratorio mostró que el bloque 1 es determinista entre procesos (`0000c59bd0a7…` en las dos instancias). Un chequeo de solo `prev == tip` aceptaría esa ronda dentro de la cadena reconstruida.

No hay `POST /api/v1/blocks`. Quien lo llame sigue recibiendo 404. El cliente nuevo no lo llama.

### Errores que el cliente debe tratar

| Respuesta | Qué hace React |
| --- | --- |
| Red caída o estado distinto de 2xx en `GET /chain` | Estado `error`. Conserva la última foto solo si la marca como no actualizada (el texto de error visible y la altura etiquetada como última conocida). No llama a `mineBlock` |
| `event: status` con otra generación | Reemplaza bloques, altura, estado e intervalo con un `GET /chain` de esa generación |
| `event: block` con otra generación, hash repetido, o altura que no es `tip+1` | No lo agrega. Pide la foto. Si la foto trae la misma generación y una altura mayor, reemplaza |
| 400 en el reinicio | No toca la cadena. Muestra que faltó la confirmación |
| 409 `block_limit_reached` | No toca la cadena. El botón Iniciar sigue deshabilitado. El texto de límite permanece |
| 404 de cualquier ruta | Error visible. Sin minado local |

### Tope de la demo

`MAX_BLOCKS` se carga en la config, con default 1000, y se rechaza si es menor que 3. No interviene en `consensus.Select` ni en `Owners`. Los satoshis se pueden volver a sortear. El número de bloques de la generación y el número de satoshis son contadores distintos.

El coordinador lo aplica dos veces:

- Antes de programar una ronda, si `blockCount >= maxBlocks`, no llama a `MineNext`.
- Bajo el lock, antes de confirmar el candidato. Si la cadena ya tiene `maxBlocks`, el candidato se descarta. Si este candidato es el que completa el cupo, se inserta, se pasa a `paused`, se marcan `limitReached` y `stopReason`, y se publican el bloque y el estado en ese orden.

Ese segundo chequeo es el que cierra start repetido y dos rondas que hayan pasado el primer chequeo con la cadena en `maxBlocks - 1`. Después del bloque que llena el cupo, el bucle no queda en un poll: espera una señal de arranque o el apagado del proceso. Start con el tope alcanzado no emite esa señal.

La interfaz, en modo API, muestra `Bloques: {blockCount} / {maxBlocks}`, `Supply: {supply} satoshis` y `Máximo: {maxBlocks} bloques`. Con `limitReached`, muestra `Límite de la demo alcanzado. Reiniciá para comenzar otra ejecución.` y deshabilita Iniciar. La grilla sigue listando todos los bloques. Reiniciar confirmado sigue disponible. Recargar u otra pestaña leen el mismo `limitReached` en el snapshot.
