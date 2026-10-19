# Auditoría de sincronización y control — demo PoA

Diagnóstico de solo lectura, hecho el 19 de octubre de 2026. No se modificó la lógica de la demo, no se añadió ningún endpoint y no se tocó el servidor público `http://24.199.80.211`.

Los cambios locales del frontend quedaron como estaban. El backend local sigue limpio en `main`.

| Documento | Contenido |
| --- | --- |
| [01-estado-y-versiones.md](01-estado-y-versiones.md) | Repositorios, commits, árbol local y comandos para leer producción |
| [02-causas-y-evidencias.md](02-causas-y-evidencias.md) | Flujo actual, causas demostradas y laboratorio |
| [03-contrato-actual-y-propuesto.md](03-contrato-actual-y-propuesto.md) | JSON real, unidades de dificultad y contrato propuesto |
| [04-decisiones-y-sincronizacion.md](04-decisiones-y-sincronizacion.md) | Comportamiento objetivo, carreras y decisión recomendada |
| [05-cambios-por-repositorio.md](05-cambios-por-repositorio.md) | Orden de implementación, sin aplicarla |
| [06-matriz-de-aceptacion.md](06-matriz-de-aceptacion.md) | Comprobaciones de la implementación futura |
| [07-build-deploy-rollback.md](07-build-deploy-rollback.md) | Build, deploy y rollback posteriores |

## 1. Qué funciona

- `GET /api/v1/chain` devuelve la cadena en memoria, con `supply` y `difficulty`. En el laboratorio local, con la configuración por defecto, respondió `supply: 1000` y `difficulty: 16`.
- `GET /api/v1/stream` abre un `text/event-stream`, escribe el comentario `: connected` y emite el evento nombrado `block` con el JSON de un bloque ya verificado.
- `Engine.MineNext` mina, selecciona tres satoshis, anota tres firmantes y solo entonces avanza el tip. Dos lecturas paralelas del mismo proceso vieron el mismo tip.
- El frontend commiteado (`39a12ff`) carga ese JSON y dibuja `header.height`. Por eso la cadena aparece. El mock local, cuando no hay `VITE_API_BASE_URL`, también arranca en pausa y mina solo en el navegador.
- Recargar el navegador no reinicia Go. Vuelve a leer la cadena que el proceso ya acumuló.

## 2. Qué está roto y por qué

La demo no tiene un estado compartido controlable. El proceso Go mina solo desde que arranca, en memoria, y el navegador no puede pararlo ni reiniciarlo.

- La altura se queda quieta porque el frontend publicado en git, en `useBlockchainSimulation`, solo pide la cadena al cargar. No abre el stream. Los bloques siguientes nacen en `runMiner` (`cmd/server/main.go`) y no llegan a React hasta la próxima carga.
- “Iniciar simulación” hace `POST /api/v1/blocks`. Esa ruta no existe. Go responde `404 page not found`. El hook guarda el error, vuelve a `paused` y deja en pantalla la cadena anterior.
- “Pausar” y “Reiniciar” cambian estado de React. No hay endpoint que pause el minero ni que reemplace la cadena.
- Una instancia nueva no nace pausada con génesis y dos bloques. `engine.New` deja solo el génesis y `runMiner` empieza a minar antes de que alguien pulse Iniciar. En el laboratorio, la primera lectura ya traía altura 1.
- La dificultad 4 del panel es el constante `POA_NETWORK.difficulty` del cliente commiteado. El backend usa otro algoritmo y otro número: 16 significa dos bytes cero de SHA-256. No son la misma unidad. El detalle está en el contrato.
- Reiniciar el proceso Go borra la cadena. No hay archivo ni base de datos. El laboratorio lo mostró: al matar el proceso y volver a levantarlo, la altura regresó a 1 y el primer bloque repetió el mismo prefijo de hash.

El árbol de trabajo del frontend, sin commitear, ya escucha el evento `block` y dejaría de llamar a `POST /blocks` cuando hay URL de API. Ese cambio no está en `origin/main`, no controla al minero y no debe desplegarse como solución.

## 3. Qué hay que implementar primero

Primero el backend, porque el navegador no puede gobernar un minero que no tiene estado.

1. Dejar el proceso pausado al escuchar, con exactamente tres bloques producidos por `MineNext`: génesis y dos bloques verificados.
2. Añadir generación de cadena y comprobarla, bajo el mismo mutex, antes de insertar y antes de publicar.
3. Exponer inicio, pausa y reinicio confirmado, con `MAX_BLOCKS` (default 1000, mínimo 3) aplicado en ese coordinador. El tope es la longitud de la generación, no el supply. El contrato está en el documento 03.
4. Después, cambiar React para obedecer ese estado, abrir el stream, mostrar el cupo de bloques y eliminar el minado mock como respuesta a un error de API.

## 4. Cómo comprobaremos que quedó resuelto

La matriz del documento 06 es la lista ejecutable. El mínimo que cierra el incidente:

- Un proceso recién arrancado responde `status: "paused"`, `height: 2` y tres bloques, y sigue igual después de varios segundos sin pulsar Iniciar.
- Dos pestañas leen esa misma cadena. Recargar no añade bloques ni cambia la generación.
- Iniciar hace crecer la cadena en ambas pestañas por el evento `block`, sin recargar.
- Pausar impide el bloque siguiente. Una ronda que ya estaba en curso no se inserta después de la pausa.
- Reiniciar, con confirmación, sube la generación, vuelve a tres bloques y pausa. Una ronda de la generación anterior no aparece en la cadena nueva ni en el stream.
- `POST /api/v1/blocks` deja de existir en el cliente. Un 404 u otro fallo de API muestra error y no mina un bloque local.
- El panel muestra la dificultad 16 del servidor, con el significado de bytes cero de SHA-256, y el intervalo que realmente usa Go.
- Con `MAX_BLOCKS=1000`, la generación se detiene en 1000 bloques (altura 999), conserva la cadena, publica el bloque y después el estado `block_limit`, y un start posterior responde 409. Un reinicio confirmado vuelve a tres bloques. Las pruebas automáticas del tope usan un N chico; las de sincronización siguen en dificultad 16.
