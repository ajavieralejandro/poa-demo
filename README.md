# Blockchain académica — Proof of Activity

Esqueleto en Go para una demo de cátedra. Define los modelos, las interfaces y el orden de las fases de Proof of Activity (prueba de actividad). No implementa el consenso, la criptografía ni la validación.

Proof of Activity no es Proof of Authority. Tampoco consiste en esperar la firma de todos los usuarios: quiénes participan y qué umbral alcanza todavía no está definido.

## Objetivo

Dejar una base mínima, comprensible y compilable para repartir el trabajo. Los identificadores públicos serán UUID. Un UUID no es una clave privada, y `hash(UUID || contenido)` no es una firma digital.

## Estructura

| Ruta | Rol |
| --- | --- |
| `cmd/demo/main.go` | Ejecutable. Solo anuncia que la implementación está pendiente. |
| `internal/blockchain/types.go` | `Transaction`, `Signature`, `Validator`, `Block` y `Chain`. |
| `internal/blockchain/chain.go` | Hash de bloque, validación e incorporación a la cadena. |
| `internal/consensus/poa.go` | Selección de validadores y las cuatro fases. |
| `internal/signing/signer.go` | Firmar y verificar mensajes, sin algoritmo fijo. |

Solo se usa la biblioteca estándar.

## Modelos

- `Transaction`: identificador, emisor, destinatario, monto y una firma.
- `Signature`: identificador público del firmante y los bytes de la firma.
- `Validator`: identificador público de un validador.
- `Block`: transacciones, hash anterior, hash propio, firmas del bloque y `ChildBlock`.
- `Chain`: bloques indexados por hash y hash de la cabeza.

`ChildBlock` es una simplificación para una cadena lineal: guarda el sucesor único. No modela bifurcaciones ni ramas competidoras.

## Fases previstas

1. Minar un encabezado candidato con prueba de trabajo.
2. Elegir validadores según su participación.
3. Recoger las firmas de esos validadores.
4. Finalizar el bloque e incorporarlo a la cadena.

Ninguna fase está implementada. Los métodos concretos devuelven `ErrNotImplemented`. No producen un hash, una firma ni un bloque válido.

## Firmas y hash

La interfaz de firma no fija el algoritmo: más adelante se puede elegir Ed25519, ECDSA u otro. No queda impuesto RSA.

Un hash final puede comprometer el contenido y el conjunto de firmas. Ese hash no sustituye la verificación de cada firma. Si en una prueba posterior se usa `hash(UUID || contenido)`, será solo una simulación, sin autenticidad criptográfica.

Esta etapa no genera claves, UUID, hashes ni firmas.

## Decisiones pendientes

- Algoritmo de firma.
- Cómo se seleccionan los validadores y cuál es el umbral.
- Serialización canónica.
- Qué bytes exactos se firman y qué entra en el hash del bloque.

## Comandos

Desde este directorio:

```powershell
go run ./cmd/demo
go test ./...
```

## Reparto sugerido

1. Firmas: algoritmo, serialización canónica y bytes que se firman (`internal/signing`).
2. Cadena: hash del bloque, validación de transacciones y bloques, e incorporación (`internal/blockchain`).
3. Consenso: las cuatro fases, el criterio de participación y el umbral (`internal/consensus`). El umbral no es «firman todos».
4. Integración: enlazar las fases en `cmd/demo` cuando esas piezas existan, sin abrir la red.

Entre dos integrantes, una persona puede tomar la firma y la serialización, y la otra el consenso y la cadena.
