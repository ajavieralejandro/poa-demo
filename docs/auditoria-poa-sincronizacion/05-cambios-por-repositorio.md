# Cambios por repositorio, en orden

Ninguno de estos cambios está hecho. El orden importa: el frontend no puede pausar un minero que no tiene estado.

Preservar el árbol sucio de `poa-front` al implementar. Esas ediciones ya traen el `EventSource` y la lectura de `difficulty`; hay que corregirlas, no revertirlas a ciegas ni desplegarlas como están. Hoy ponen la UI en “en curso” al cargar y no hablan con ningún control del servidor.

## 1. Backend `poa-demo`

Partir de `6ec3694`.

1. Estado en `Engine`: `status` (`paused` / `running`), `generation` (`uint64`, arranca en 1), `interval`, `maxBlocks`, `limitReached` y `stopReason`. `Snapshot` los devuelve junto con los bloques, el supply, la dificultad, `blockCount` y `height`. `config.Load` lee `MAX_BLOCKS` (default 1000, mínimo 3), aparte de `INITIAL_SUPPLY`.
2. Semilla de arranque, antes de `ListenAndServe`: `engine.New` y dos `MineNext`. Si uno falla, el proceso termina y no escucha. El estado queda `paused`. `runMiner` no llama a `MineNext` mientras esté en pausa.
3. Separar la confirmación del bucle. Un método interno inserta solo si `prev == tip` y `blockCount < maxBlocks`. Si el insertado completa `maxBlocks`, deja `paused`, `limitReached=true` y `stopReason=block_limit`. `MineNext`, usado por `runMiner`, además exige que la generación capturada al empezar la ronda siga vigente y que `status == running`. Si falla cualquiera, no inserta y no publica. La semilla de arranque y la del reinicio no pasan por ese bucle: llaman a la confirmación interna mientras arman la generación, antes de responder y antes de que el bucle vea esa generación. Con `MAX_BLOCKS=3` la semilla ya deja el tope marcado.
4. `runMiner` solo llama a `MineNext` cuando el estado es `running` y `blockCount < maxBlocks`. Después de un bloque aceptado que no llena el cupo, espera `BLOCK_INTERVAL_MS`. Después de una ronda descartada, no publica. Al llenar el cupo publica el bloque y luego el estado, y la goroutine queda bloqueada en una señal hasta un start válido o el apagado. No hace poll. Start con `limitReached` responde 409 `{"error":"block_limit_reached","maxBlocks":N,"blockCount":N}` y no despierta al minero.
5. `NewChainView` agrega `intervalMs`, `status`, `generation`, `height`, `maxBlocks`, `blockCount`, `limitReached` y `stopReason`. `NewBlockView` agrega `generation` al JSON del evento. El bloque guardado en la cadena no necesita ese campo: se rellena al serializar. El evento `status` lleva los mismos campos de tope.
6. Rutas nuevas en `Server.Handler`, con la interfaz de control al lado de `ChainSource`:
   - `POST /api/v1/simulation/start`
   - `POST /api/v1/simulation/pause`
   - `POST /api/v1/simulation/reset` exige `{"confirm": true}`
7. `handleStream` lee `generation`, `after` y `Last-Event-ID`. Retransmite desde `orderedBlocks` y luego sigue el hub. Si la generación no coincide, emite `status` y no retransmite bloques viejos.
8. Cada cambio de estado hace `Publish` de un evento `status`. El evento `block` solo sale de una inserción confirmada de la generación actual.
9. CORS: métodos `GET, POST, OPTIONS`.
10. Tests en `internal/config`, `internal/engine` y `internal/api`. El tope se prueba con un `MAX_BLOCKS` chico (4 o 5 alcanza) y puede usar dificultad 0: exactamente N bloques contando el génesis, el último en altura N-1, ausencia del bloque siguiente, cadena estable después del tope, start repetido y dos confirmaciones concurrentes sin pasar de N, `supply` igual al inicial, reset a tres bloques en una generación nueva con `limitReached=false`. Las pruebas de integración del stream siguen en dificultad 16, con un máximo por encima de los pocos bloques que esperan; no minan 1000 bloques. También siguen la semilla en pausa, el descarte al pausar, el rechazo de una ronda de la generación anterior, el replay `after` y el 404 de `POST /api/v1/blocks`. `MAX_BLOCKS` menor que 3 falla en `config.Load`.

`cmd/demo` puede seguir llamando a `MineNext` directo para el stress test. No es el servidor de la cátedra. No hace falta que respete la pausa de la UI.

## 2. Frontend `poa-front`

Sobre el árbol de trabajo, no sobre un revert a `39a12ff`.

1. Borrar el uso de `requestNextBlock` en el efecto de simulación. Borrar el POST de `blockchainApi.ts` o dejar la función sin llamadores y sin fallback a `mineBlock`.
2. Tipos: `ChainSnapshot` con `intervalMs`, `status`, `generation`, `height`, `maxBlocks`, `blockCount`, `limitReached` y `stopReason` obligatorios cuando `dataSource === "api"`. Validar en runtime antes de `setBlocks`. Si falta un campo, estado `error` y no pintar la respuesta como cadena buena.
3. El estado de los botones sale de `snapshot.status`, no de un `useState` paralelo. Carga y recarga: `GET /chain`, suscripción con `after`. No llaman a reset.
4. Iniciar y pausar: los POST correspondientes. Iniciar queda deshabilitado si `limitReached`. Un 409 `block_limit_reached` no altera la foto. Reiniciar: confirmación en el navegador y luego `{"confirm": true}`. Sigue disponible con el tope alcanzado.
5. `useBlockStream`: query `generation` y `after`, evento `status`, regla de generación del documento 04, cierre al desmontar. Al abrir (`onopen`), repetir el GET para cerrar la carrera que el `after` no hubiera cubierto si el query se armó con una altura vieja.
6. Métricas: `Supply: {supply} satoshis`, `Máximo: {maxBlocks} bloques`, `Bloques: {blockCount} / {maxBlocks}`, dificultad e intervalo salen del snapshot. La dificultad se muestra como el entero de Go. Con `limitReached`, el texto fijo es `Límite de la demo alcanzado. Reiniciá para comenzar otra ejecución.` La grilla no oculta bloques. Los nombres de `POA_NETWORK.validators` no se usan en modo API.
7. El slider deja de controlar el minado. Se puede dejar fijo, mostrando `intervalMs`.
8. Un fallo de red conserva los bloques solo con un aviso de que son la última foto conocida. No llama a `scripts` ni a `consensus.mineBlock`.
9. El modo mock (URL vacía) sigue pausado al cargar, con `mockChain.json`, y su minado local no se mezcla con una respuesta de Go.

## 3. Lo que no entra en este cambio

- Apache y el `ProxyPass`.
- Persistencia en disco.
- Firmas criptográficas de `internal/signing` (siguen sin implementarse; el JSON sigue siendo signer + satoshiId).
- Cambiar el algoritmo de `cumpleDificultad`.
- Mapa de nombres de los 100 nodos.
- Alterar `DIFFICULTY` para que la interfaz “se vea en movimiento”. El fallo era de sincronización, con bloques cada ~2 s en el laboratorio.
- Recortar una cadena ya minada, en memoria o en el servidor público, para que parezca que el tope ya estaba. El límite evita el bloque siguiente. No borra los anteriores. Un reinicio de proceso sigue perdiendo la RAM; no es una herramienta para podar la demo en curso.
