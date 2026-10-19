# Comportamiento objetivo, carreras y decisiones

## Objetivo

El proceso Go es la única autoridad de la cadena y del estado `paused` / `running`. React muestra esa foto y los eventos posteriores. No mina cuando la API falla.

Una instancia nueva, al empezar a escuchar, está en pausa y ya tiene tres bloques: el génesis y dos bloques que produjo `MineNext` y que pasaron por `consensus.Verify`. Abrir otra pestaña o recargar hace `GET /api/v1/chain` y pinta esa misma memoria. Esas lecturas no reinician la demo ni piden un bloque nuevo.

Iniciar y reanudar ponen `running` en el servidor. El bucle de `runMiner` solo llama a `MineNext` en ese estado. Pausar pasa a `paused` y no programa otra ronda. Reiniciar es un POST con `confirm: true`, compartido por todos los navegadores: nueva generación, otra vez tres bloques, otra vez pausa.

Una ronda cuya generación, capturada al empezar, no coincide con la generación bajo el lock al confirmar, no entra en `chain.Blocks`, no mueve `Tip` y no llega a `Hub.Publish`.

Cada generación admite como máximo `MAX_BLOCKS` bloques, génesis incluido. El default es 1000, así que las alturas válidas son 0..999. `INITIAL_SUPPLY` sigue siendo la cantidad de satoshis y no se reduce al sortearlos. Al confirmar el bloque que llena el cupo, la cadena queda entera, el estado pasa a `paused` con `stopReason=block_limit` y `limitReached=true`, y no se programa otra ronda. Start en ese estado responde 409 `block_limit_reached`. Un reinicio confirmado abre otra generación de tres bloques, con el tope otra vez disponible.

Los bloques nuevos llegan por el evento `block` sin recargar la página. El modo mock (URL vacía) puede seguir existiendo para trabajar la interfaz sin Go. Un error contra una URL configurada no cae en `mineBlock`.

## Decisiones

### 1. Ronda ya empezada al pausar

Recomendación: descartarla antes de insertar.

`MineNext` suelta el mutex durante el PoW. La pausa toma el mismo mutex, pone `paused` y responde. Cuando la ronda vuelve a tomar el lock, ve `paused` y tira el candidato. El tip que vio el cliente en la respuesta de pausa no se mueve un segundo después.

La alternativa, dejar terminar la ronda, publica un bloque después de que ambas pestañas ya muestran “En pausa”. En una demo de cátedra esa carrera se lee como un fallo del botón.

El coste es tirar un nonce ya encontrado. Es aceptable: la dificultad 16 del laboratorio encontró bloque en menos de un segundo.

### 2. Reinicio y ronda vieja

Recomendación: generación monótona, chequeada bajo el lock junto con el tip.

El chequeo actual `prev == tip` no alcanza. El génesis no tiene entropía y el reparto de satoshis es determinista. Dos procesos locales produjeron el mismo prefijo de hash en la altura 1 (`0000c59bd0a7`). Una ronda que extendiera una cadena reconstruida con los mismos dos bloques podría pasar el chequeo de tip. La generación corta ese caso aunque los hashes coincidan.

El reinicio siembra los dos bloques dentro de la operación de reset, con la generación nueva ya instalada, y solo entonces responde y emite `status`. Quien estuviera minando la generación anterior no publica en el hub de la nueva.

### 3. Intervalo del slider

Recomendación: el slider deja de presentarse como control del minero. El panel muestra `intervalMs` del snapshot.

Cambiar `BLOCK_INTERVAL_MS` en caliente no hace falta para el incidente. Si más adelante se quiere, sería otro campo del POST de start, no un estado local de React.

### 4. Nombres de validadores

Recomendación: en modo API, mostrar `signer` y `miner` tal como llegan (`"1"`, `"89"`). No superponer Aurora, Borealis y el resto.

Esos seis nombres y sus rangos viven en `POA_NETWORK` y no salen de Go. Go reparte 1000 satoshis entre las direcciones 1..100. Mezclarlos hace parecer que el comité del bloque es el del mock. Un mapa de nombres en el servidor puede esperar. No bloquea la sincronización.

### 5. Dos reinicios a la vez

Recomendación: los dos son válidos. Cada uno sube la generación y deja tres bloques en pausa. El último que toma el lock gana. Los clientes que vean un `status` con generación mayor tiran su foto y cargan esa. Ese caso no usa 409. El 409 queda reservado a `block_limit_reached`.

### 6. Persistencia

Recomendación: no agregar archivo ni base de datos.

Recargar el navegador solo necesita la RAM del proceso, y eso ya es compartido entre pestañas. Reiniciar `poa-backend.service`, o reemplazar el binario, borra la cadena. Con el cambio, ese reinicio de proceso es una demo nueva: pausada, generación 1, tres bloques. Hay que decirlo en el deploy. Sobrevivir a un restart de systemd sería otro proyecto.

### 7. CORS y Apache

Si el bundle y la API salen del mismo origen (`http://24.199.80.211` y `/api/` por el proxy), el navegador no usa CORS. Los defaults actuales solo listan `localhost:5173` y `127.0.0.1:5173`. Un front abierto desde otro origen, o un `VITE_API_BASE_URL` con el puerto 8080, fallaría el preflight. Hay que leer `ALLOWED_ORIGINS` de la unit antes de desplegar. No se cambia Apache para crear `/api/v1/blocks`.

### 8. Tope de bloques

El tope vive en el coordinador de la demo, en los dos puntos que ya serializan la cadena: antes de lanzar la ronda y bajo el lock antes de confirmar. No se implementa achicando `INITIAL_SUPPLY`, marcando satoshis como usados ni cambiando `Select`.

Al insertar el bloque número `MAX_BLOCKS`:

- Se conserva el mapa completo. No se recortan los bloques antiguos para simular el límite.
- `status=paused`, `limitReached=true`, `stopReason=block_limit`.
- Se publica primero el evento `block` y después el `status`.
- La goroutine del minero queda bloqueada en una señal, no en un bucle que despierte para mirar el reloj. Start con el tope alcanzado responde 409 y no dispara esa señal.

`MAX_BLOCKS=3` es válido: el arranque ya nace en el tope, porque la semilla son tres bloques. Start responde 409 hasta un reinicio confirmado. Un valor menor que 3 no arranca el proceso: la semilla no cabría.

Las pruebas del tope usan un `MAX_BLOCKS` chico y pueden usar dificultad 0. Las de sincronización del stream siguen en dificultad 16 y no tienen que producir 1000 bloques: les basta un máximo mayor que los pocos bloques que esperan.

## Sincronización mínima

El servidor retransmite desde memoria. El cliente obedece la generación. No hay cola externa ni base.

### Carrera entre el snapshot y el stream

Orden del cliente al montar y al reconectar:

1. `GET /api/v1/chain`. Guardar `generation`, `height`, `status`, `blocks`.
2. Abrir `EventSource` en `/api/v1/stream?generation={generation}&after={height}`.
3. Aplicar los `block` retransmitidos y los vivos solo si `generation` coincide y `header.height == alturaLocal + 1` y `header.prev` es el hash local.
4. Ante cualquier otro bloque, o ante un `status` de otra generación, repetir el paso 1 y reabrir el stream con el `after` nuevo.

El hueco que hoy existe entre el GET y el `Subscribe` lo cierra `after`: el servidor manda los bloques confirmados con altura mayor antes de pasar al vivo. Si uno se cuela entre la retransmisión y el vivo, el cliente lo ve fuera de orden y repite el GET. Con la cadena en pausa, el GET y el stream devuelven la misma altura y no nace un bloque en el medio.

### Eventos perdidos al desconectarse

Medido en local: alturas 9 y 10 ocurrieron con el socket cerrado y no volvieron en la reconexión. Con `after` igual a la última altura aplicada, el servidor las reescribe desde `orderedBlocks`. El buffer de 8 y el `default` de `Publish` dejan de ser la única recuperación: un cliente lento pierde el vivo y, al ver un hueco o al reconectar, pide `after`.

Conviene subir el buffer solo si el replay fallara. No hace falta para el arreglo mínimo. El replay cubre la pérdida.

### Duplicados y desorden

- Mismo hash: se ignora.
- Misma generación y altura `tip+1` con `prev` correcto: se agrega.
- Altura menor o igual, o `prev` distinto: no se agrega; se pide la foto.
- Generación menor que la local: se ignora. Es un evento atrasado de un reinicio ya visto.
- Generación mayor: se reemplaza la foto entera. La altura nueva puede ser 2 aunque la local fuera 40. Eso es un reinicio, no un desorden.

### Reconexión después de reiniciar

El `status` del reinicio llega a quien siga conectado. Quien esté caído se entera en el próximo GET o porque el stream, al abrirse con la generación vieja, recibe `status` de la nueva. En los dos casos reemplaza bloques y vuelve a suscribirse con `after=2` (o la altura que traiga la foto).

El tope viaja en ese mismo snapshot y en el `status`. Una pestaña que se conecta después de los 1000 bloques recibe `limitReached=true`, `stopReason=block_limit`, `blockCount=1000` y los 1000 bloques. No depende de haber visto el evento en vivo.

`Last-Event-ID` con forma `generación:altura` permite que el reintento automático de `EventSource` pida el mismo replay sin esperar a un hueco. Hay que leerlo en `handleStream`. Si el header viene mal formado, se ignora y se trata como suscripción en vivo; el cliente igual hará el GET inicial.

### Dos navegadores

Los dos leen y se suscriben al mismo `Engine` y al mismo `Hub`. Iniciar en uno pone `running` para los dos: el otro recibe `status` y habilita la espera de bloques. Pausar en uno detiene el bucle para los dos. Reiniciar en uno, con el diálogo de confirmación de esa pestaña, cambia la cadena de todos. No hay sesión por navegador. Es la demo compartida que se pidió.

El botón Reiniciar de React pide confirmación en el navegador y solo entonces manda `{ "confirm": true }`. Cerrar el diálogo no llama al servidor.
