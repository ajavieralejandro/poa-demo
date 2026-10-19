# Matriz de aceptación

Las filas 1 a 15 se prueban contra un proceso local recién arrancado, con los defaults (dificultad 16, supply 1000, `MAX_BLOCKS=1000`, intervalo 2000 ms), y después contra el despliegue. No bajar la dificultad para dar por buena la sincronización, ni hace falta minar los 1000 bloques en esa pasada. Las filas 16 a 20 cubren el tope con un N chico. No apuntar las pruebas de reinicio al proceso público hasta tener backup del binario y una ventana acordada: reiniciar Go borra la cadena. No se recorta la cadena que ya esté en memoria.

| # | Situación | Resultado que cierra el incidente |
| --- | --- | --- |
| 1 | Proceso recién escuchando, sin pulsar nada, durante al menos 10 s | `GET /chain` estable: `status=paused`, `generation=1`, `height=2`, `blockCount=3`, `maxBlocks=1000`, `limitReached=false`, `stopReason=""`. El tercero enlaza `prev` con el segundo. Los dos no génesis traen tres firmas y pasan la regla de bytes cero |
| 2 | Recargar el navegador y abrir una segunda pestaña | Las dos muestran los mismos tres hashes y la misma generación. La cantidad de bloques sigue en 3 |
| 3 | `POST /api/v1/blocks` | 404. La cadena no gana un bloque |
| 4 | Iniciar en la pestaña A | Las dos pestañas pasan a `running` sin recargar. Llegan eventos `block` con `generation=1` y alturas 3, 4, 5… El `prev` de cada uno es el hash anterior. La grilla crece |
| 5 | Pausar mientras una ronda ya está buscando nonce | La respuesta queda `paused`. Durante otros 10 s no aparece un bloque nuevo ni en `GET /chain` ni en el stream. La altura de la respuesta de pausa es la definitiva |
| 6 | Reanudar | Vuelve a `running` y la altura siguiente es la anterior más uno, misma generación |
| 7 | Cerrar el stream 5 s y reconectar con `after` en la última altura pintada | El stream entrega los bloques nacidos durante el corte, en orden, y después los nuevos. No hace falta recargar para alcanzar el tip |
| 8 | Reiniciar con el diálogo cancelado | No hay POST. Generación y bloques iguales |
| 9 | Reiniciar confirmado en una pestaña, con la otra abierta y con una ronda en curso | `generation=2`, `status=paused`, tres bloques nuevos. La otra pestaña reemplaza la cadena al recibir `status`. Ningún bloque de la generación 1 aparece después del `status` de la 2. Un bloque que se estaba minando no aumenta la altura |
| 10 | Segundo reinicio inmediato | `generation=3`, otra vez tres bloques en pausa. La pestaña que todavía procesaba la generación 2 la descarta |
| 11 | Apagar la red del navegador o parar Go a mitad de la demo | Error visible. La última foto queda marcada como no actual. No se agregan bloques con hashes del mock (`mineBlock` / `hashHex`) |
| 12 | URL de API vacía, `npm run dev` | Sigue el mock, pausado, sin llamar a `/api` |
| 13 | Panel al arrancar | `Supply: 1000 satoshis`, `Máximo: 1000 bloques`, `Bloques: 3 / 1000`, dificultad 16, intervalo 2000, altura igual a `header.height` del tip. La grilla muestra los 3 bloques. Los firmantes se ven como los decimales del JSON |
| 14 | `go test ./...` del backend | Cubre semilla, pausa, descarte por generación, replay `after`, el 404 de `POST /api/v1/blocks` y el tope con N chico. La integración del stream permanece en dificultad 16 y no mina 1000 bloques |
| 15 | Dos `GET /chain` paralelos durante `running` | Misma generación. Las alturas difieren como máximo en el bloque que acaba de confirmarse. Ninguna de las dos vuelve a una generación anterior |
| 16 | Test con `MAX_BLOCKS=N` pequeño, N>=4, dificultad 0 | Tras start hay exactamente N bloques, génesis incluido. El último tiene altura N-1. No existe altura N. `supply` sigue en el valor configurado. `limitReached=true`, `stopReason=block_limit`, `status=paused` |
| 17 | Esperar después del tope, repetir start y lanzar dos confirmaciones con la cadena en N-1 | La cadena no pasa de N. El segundo start responde 409 `block_limit_reached` y el cuerpo de `GET /chain` no cambia. Las dos confirmaciones concurrentes dejan un solo bloque nuevo |
| 18 | Dos pestañas abiertas cuando se confirma el bloque N | Las dos reciben primero el `block` de altura N-1 y después el `status` con `limitReached=true`. Iniciar queda deshabilitado. Se lee `Límite de la demo alcanzado. Reiniciá para comenzar otra ejecución.` La grilla sigue mostrando los N bloques |
| 19 | Cerrar el stream en el tope y volver a entrar, o recargar | `GET /chain` y el `status` inicial traen `limitReached=true`, `blockCount=N` y los N bloques. No hace falta haber visto el evento en vivo |
| 20 | Reinicio confirmado con el tope ya alcanzado | Generación nueva, `blockCount=3`, `height=2`, `limitReached=false`, `stopReason=""`, `supply` igual. Start vuelve a minar. La generación anterior no recibe ese bloque |

Comprobación manual corta del enlace, además de los tests:

```text
bloque[0].header.height == 0
bloque[i].header.prev == bloque[i-1].hash
bloque[i].header.height == i
último.height == body.height
len(bloque) == body.height + 1
len(bloque) == body.blockCount
len(bloque) <= body.maxBlocks
```

El génesis puede no empezar con ceros. Los bloques de altura mayor que 0, con dificultad 16, empiezan con `0000`.
