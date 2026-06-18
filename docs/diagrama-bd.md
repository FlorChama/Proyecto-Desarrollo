# Diagrama de la Base de Datos

Modelo relacional del sistema (MySQL, mapeado con GORM). Cuatro entidades:
**User**, **Event**, **Ticket** y **Payment**, con sus claves foráneas.

```mermaid
erDiagram
    USER ||--o{ TICKET : "compra"
    EVENT ||--o{ TICKET : "tiene"
    TICKET ||--|| PAYMENT : "genera"
    USER ||--o{ PAYMENT : "realiza"

    USER {
        uint id PK
        string name
        string email "único"
        string password "hash SHA-256"
        string role "client | admin"
        datetime created_at
        datetime updated_at
        datetime deleted_at "soft delete"
    }

    EVENT {
        uint id PK
        string title
        string description
        datetime date
        string extra_dates "fechas adicionales"
        int duration "minutos"
        string venue
        int capacity
        int available
        string category
        string image_url
        float price
        float vip_price
        string status "active | cancelled"
        datetime created_at
        datetime updated_at
        datetime deleted_at "soft delete"
    }

    TICKET {
        uint id PK
        uint user_id FK
        uint event_id FK
        string status "active | cancelled | transferred"
        string qr_code "QR en base64"
        string ticket_type "general | vip"
        float price
        datetime created_at
        datetime updated_at
        datetime deleted_at "soft delete"
    }

    PAYMENT {
        uint id PK
        uint ticket_id FK
        uint user_id FK
        float amount
        string method "credit_card | debit_card | modo | mercadopago"
        string status "approved | pending | failed"
        datetime created_at
        datetime updated_at
        datetime deleted_at "soft delete"
    }
```

## Relaciones
- **User → Ticket** (1:N): un usuario puede tener muchos tickets. `ticket.user_id` → `user.id`
- **Event → Ticket** (1:N): un evento puede tener muchos tickets. `ticket.event_id` → `event.id`
- **Ticket → Payment** (1:1): cada ticket genera un pago. `payment.ticket_id` → `ticket.id`
- **User → Payment** (1:N): un usuario puede tener muchos pagos. `payment.user_id` → `user.id`

## Notas de diseño
- **Soft delete:** todas las entidades usan `deleted_at` (gorm.Model), las bajas no borran físicamente.
- **Cancelación de eventos:** se hace por campo `status = cancelled`, preservando los tickets existentes.
- **Traspaso:** al traspasar, el ticket original pasa a `transferred` y se crea un ticket nuevo para el destinatario con un QR nuevo.
