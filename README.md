# 🎫 TicketHub — Sistema de Gestión de Eventos y Entradas

Sistema tipo Ticketek desarrollado como Práctico Integrador 2026 — Desarrollo de Software, Facultad de Ingeniería UCC.

## Descripción

TicketHub permite explorar eventos, comprar entradas, cancelarlas y traspasar su titularidad a otros usuarios. Los administradores pueden crear y gestionar eventos y ver métricas de ocupación.

## Tecnologías Utilizadas

| Capa | Tecnología |
|------|-----------|
| Backend | Go 1.22, Gin, GORM |
| Base de datos | MySQL 8.0 |
| ORM | GORM (con driver MySQL) |
| Autenticación | JWT (golang-jwt/jwt v5) |
| QR Code | skip2/go-qrcode |
| Email | gopkg.in/gomail.v2 |
| Frontend | React 18, Vite, React Router v6 |
| HTTP Client | Axios |
| DevOps | Docker, Docker Compose, Nginx |

## Requisitos Previos

- Go >= 1.22
- Node.js >= 20
- MySQL 8.0
- Docker y Docker Compose (opcional)

## Instalación y Uso

### Opción 1 — Docker Compose (recomendado)

```bash
git clone <url-del-repo>
cd PROYECTO\ DESARROLLO
cp backend/.env.example backend/.env   # configurá las variables
docker-compose up --build
```

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

### Opción 2 — Local

#### Base de datos
```sql
CREATE DATABASE ticketek_db;
```

#### Backend
```bash
cd backend
cp .env.example .env       # editá con tus datos
go mod tidy
go run main.go
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

Frontend disponible en http://localhost:5173

### Ejecutar tests
```bash
cd backend
go test ./tests/... -v -cover
```

## Diagrama de Base de Datos

![DB Diagram](docs/db_diagram.png)

## Endpoints de la API

| Método | Ruta | Auth | Descripción |
|--------|------|------|-------------|
| POST | /api/auth/register | No | Registro de usuario |
| POST | /api/auth/login | No | Login, retorna JWT |
| GET | /api/events | No | Listar eventos (filtros: search, category, available) |
| GET | /api/events/:id | No | Detalle de evento |
| POST | /api/tickets | Cliente | Comprar entrada |
| GET | /api/tickets/my | Cliente | Mis entradas |
| DELETE | /api/tickets/:id | Cliente | Cancelar entrada |
| POST | /api/tickets/:id/transfer | Cliente | Traspasar entrada |
| POST | /api/admin/events | Admin | Crear evento |
| PUT | /api/admin/events/:id | Admin | Editar evento |
| DELETE | /api/admin/events/:id | Admin | Cancelar evento |
| GET | /api/admin/events/:id/report | Admin | Reporte de ocupación |

## Decisiones de Diseño

**1. Soft delete para eventos:** En lugar de eliminar físicamente los registros de la tabla `events`, se actualiza el campo `status` a `cancelled`. Esto preserva la integridad referencial de los tickets existentes y permite auditar el historial de eventos cancelados sin perder datos.

**2. Generación del QR post-insert:** El código QR se genera con el ID del ticket recién creado como parte del string único (`TICKET-{id}-USER-{id}-EVENT-{id}`). Esto requiere un segundo `UPDATE` tras el `INSERT`, pero garantiza que el QR esté ligado al ID real de la base de datos, evitando colisiones y haciendo el código escaneable trazable al registro exacto.

**3. Nuevo QR en traspaso:** Al transferir un ticket, se genera un nuevo código QR para el nuevo titular, invalidando el anterior. Esto impide que el titular original use una copia del QR anterior tras el traspaso.

## Bonus Track — Notificaciones por Email con QR

Al comprar una entrada, el sistema envía un email al comprador con el QR embebido. Al traspasar un ticket, se genera un nuevo QR y se notifica al nuevo titular por email (el QR viejo queda inválido). El QR se almacena como campo `qr_code` en la tabla `tickets`.
