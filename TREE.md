paku-booking/
├── cmd/
│   └── api/
│       └── main.go                    # Entrypoint HTTP
├── internal/
│   ├── booking/                       # DOMINIO
│   │   ├── model.go                   # Entidades (DaySlot, Hold, Booking, OutboxMessage)
│   │   ├── errors.go                  # Errores de negocio
│   │   ├── events.go                  # Eventos de dominio (HOLD_CREATED, etc)
│   │   ├── repo.go                    # Interfaces Repository/TxRepo
│   │   ├── types.go                   # Tipos de dominio (Slot, Status)
│   │   ├── http/                      # ADAPTER HTTP (entrada)
│   │   │   ├── router.go
│   │   │   ├── handlers_common.go
│   │   │   ├── availability_handler.go
│   │   │   ├── holds_handler.go
│   │   │   ├── bookings_handler.go
│   │   │   ├── admin_capacity_handler.go
│   │   │   ├── internal_handler.go
│   │   │   ├── requests.go            # DTOs HTTP
│   │   │   └── errors_mapper.go
│   │   └── usecases/                  # CASOS DE USO
│   │       ├── availability.go
│   │       ├── create_hold.go
│   │       ├── cancel_hold.go
│   │       ├── confirm_booking.go
│   │       ├── expire_holds.go
│   │       ├── admin_set_capacity.go
│   │       ├── admin_adjust_capacity.go
│   │       └── admin_close_days.go
│   ├── storage/                       # ADAPTERS PERSISTENCIA (salida)
│   │   ├── memory/
│   │   │   ├── tx.go
│   │   │   ├── booking_repo.go
│   │   │   └── outbox_repo.go
│   │   └── postgres/
│   │       ├── db.go
│   │       ├── tx.go
│   │       ├── booking_repo.go
│   │       └── outbox_repo.go
│   ├── jobs/                          # JOBS/WORKERS
│   │   └── hold_expirer.go
│   ├── outbox/                        # INFRAESTRUCTURA OUTBOX
│   │   ├── dispatcher.go
│   │   ├── publisher.go
│   │   └── nop_publisher.go
│   ├── shared/                        # UTILIDADES
│   │   ├── clock.go
│   │   └── logger.go
│   └── config/
│       └── config.go
└── README.md