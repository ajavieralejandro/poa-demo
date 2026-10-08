# Blockchain académica — Proof of Activity

Primera versión local de una demo de cátedra. El modelo ya está definido. Minería, selección, firmas y alta de bloques siguen pendientes: esas funciones devuelven error y no confirman un bloque.

## Alcance

- Un solo proceso.
- Tres posiciones de validación.
- Satoshis individuales, cada uno con un propietario.
- Cadena lineal, con un único tip.
- Sin emisión nueva ni comisiones.

Proof of Activity elige satoshis y después consulta a sus dueños. No usa una lista previa de candidatos. Un mismo nodo puede minar, firmar un cupo y transaccionar.

## Límites

Estas tres extensiones no están diseñadas:

- Bifurcaciones: hay un solo mapa de dueños. Otro hijo del mismo padre no tiene un estado propio.
- Reorganización: `Append` solo acepta un bloque cuyo padre es el tip.
- Red: «desconectado» significa que este proceso no llama a firmar. No hay mensajes ni pares.

## Estructura

| Ruta | Rol |
| --- | --- |
| `cmd/demo/main.go` | Anuncia que la ronda todavía no se ejecuta. |
| `internal/blockchain/types.go` | `SatoshiID`, `Address`, `Header`, `Signature`, `Tx`, `Block` y `Chain`. |
| `internal/blockchain/chain.go` | Bytes canónicos, génesis, hash de cabecera y alta lineal. |
| `internal/consensus/poa.go` | `Mine` y `Select`. |
| `internal/signing/signer.go` | Clave local, `Sign`, `Verify` y `CreateTx`. |

La clave privada queda en `signing.Local`. No entra en bloques, transacciones ni en la génesis. Una dirección entera no reemplaza a la clave pública. El hash de la cabecera no reemplaza la verificación de cada firma.

`NewLinear` asigna el satoshi `i` a `owners[i]` y deja la génesis como tip. Ese bloque inicial no pasa por las tres firmas.

## Reparto

1. Tipos, `NewLinear` y los mensajes canónicos: ya están en este commit.
2. Integrante A: algoritmo de `Sign` y `Verify` en `internal/signing`.
3. Integrante B: `Mine`, `Select` y completar `Append` en `internal/consensus` e `internal/blockchain`.
4. Los dos: la ronda en `cmd/demo`, cuando esas funciones ya rechacen de verdad y no antes con un bloque simulado.

## Comandos

```powershell
go run ./cmd/demo
go test ./...
```
