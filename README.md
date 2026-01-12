# Paku Booking Service

## 1. Propósito
El Booking Service gestiona la disponibilidad por cupos (capacity-based booking) para servicios (ej. baños de mascotas).

Características clave:
- Cupos por día y slot (AM / PM).
- Bloqueo temporal de cupos (holds) antes del pago.
- Confirmación de booking solo después del pago.
- Expiración automática de holds.
- Administración de capacidad por día o rangos.
- Soporta almacenamiento en memoria y PostgreSQL.
- Eventos de dominio mediante outbox pattern.

## 2. Modelo mental (resumen rápido)
Capacidad (DaySlot) → Hold (bloqueo temporal) → Booking (confirmado)

Reglas clave:
- Nunca se confirma un booking sin pasar por un hold.
- Nunca se consume capacidad directamente al confirmar.
- El hold es la única forma de reservar cupo.

## 3. Casos de uso (Use Cases)

### 3.1 Consultar disponibilidad
- Use case: AvailabilityUseCase
- Entrada:
    - service_id
    - location_id (opcional)
    - from, to (YYYY-MM-DD)
- Salida: lista de días con slot (AM/PM), total, reserved, available
- No modifica estado (solo lectura).

### 3.2 Crear hold (bloqueo temporal)
- Use case: CreateHoldUseCase
- Flujo:
  1. Validar entrada.
  2. Reservar capacidad (anti-overbooking).
  3. Crear hold con expires_at.
  4. Emitir evento `HOLD_CREATED`.
- Regla: `qty <= 0` → normaliza a 1. Si no hay cupo → error.

### 3.3 Cancelar hold
- Use case: CancelHoldUseCase
- Comportamiento idempotente.
- Si hold active → cancelar y liberar cupo (o marcar expirado).
- Emitir `HOLD_CANCELED` o `HOLD_EXPIRED`.

### 3.4 Confirmar booking
- Use case: ConfirmBookingUseCase
- Flujo:
  1. Validar hold y payment_id.
  2. Verificar hold existe, está ACTIVE y no vencido.
  3. Confirmar hold, crear booking.
  4. Emitir `BOOKING_CONFIRMED`.
- Idempotencia: booking_id determinístico (hold_id + payment_id).

### 3.5 Expirar holds vencidos
- Use case: ExpireHoldsUseCase (job background).
- Buscar holds ACTIVE vencidos, marcarlos EXPIRED, liberar capacidad, emitir `HOLD_EXPIRED`.

### 3.6 Administración de capacidad
- Setear capacidad puntual (AdminSetCapacityUseCase): `total >= reserved`.
- Ajustar capacidad por rango (AdminAdjustCapacityUseCase): nunca dejar `total < reserved`.
- Cerrar días (AdminCloseDaysUseCase): setear `total = 0` solo si `reserved == 0`.

## 4. Reglas de negocio
- Toda reserva de capacidad es atómica (Tx).
- Holds expiran según `expires_at`.
- Operaciones idempotentes donde corresponde.
- Eventos se guardan en la misma transacción (outbox).

## 5. Outbox / Eventos
Eventos emitidos:
- `HOLD_CREATED`
- `HOLD_CANCELED`
- `HOLD_EXPIRED`
- `BOOKING_CONFIRMED`

El outbox persiste mensajes (JSON) para publicación asíncrona.

## 6. Storage
Implementaciones:
- `internal/storage/memory` (dev)
- `internal/storage/postgres` (producción)

El dominio y los usecases no dependen de la implementación concreta.

## 7. Migraciones / Esquema (referencia)
El paquete `internal/storage/postgres` incluye una constante `Schema` con SQL mínimo para crear tablas:
- `day_slots(service_id, location_id, date, slot, total, reserved, updated_at)`
- `holds(id, service_id, location_id, date, slot, qty, status, expires_at, created_at, updated_at)`
- `bookings(id, hold_id, payment_id, service_id, location_id, date, slot, qty, status, created_at)`
- `outbox(id, type, payload jsonb, status, attempts, last_error, created_at, sent_at)`

> Nota: las migraciones no se aplican automáticamente. Ejecutar el SQL manualmente o con tu migrator preferido.

## 8. Cómo ejecutar (desarrollo)
- Go 1.20+
- Ajustar variable de entorno `DATABASE_DSN` si usas Postgres.
- Comandos típicos:
  - `go test ./...`
  - `go run ./cmd/your-app`

## 9. Qué NO hace este servicio
- No procesa pagos.
- No maneja usuarios.
- No maneja horarios por hora/minuto.
- No decide precios.
- No notifica directamente (solo emite eventos).

## 10. Estado actual
- Funcional end-to-end (MVP).
- Memory + Postgres listos.
- Tests unitarios: pendientes (siguiente hito).