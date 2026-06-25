# Guía de Defensa Oral — TicketHub (Go + React + Docker)

> Guía pensada para defender el proyecto **sin saber Go de antes**. Está ordenada por los temas que te avisaron: la funcionalidad nueva (QR + traspaso), el panel de admin, Docker, y el arreglo de los tests. Al final de cada tema hay un bloque **"Si te preguntan…"** con respuestas listas para decir.

---

## 0. El proyecto en una frase

Es un sistema tipo **Ticketek**: vende entradas para eventos.
- **Backend** en **Go** (lenguaje) con **Gin** (framework web) y **GORM** (para hablar con la base de datos).
- **Base de datos**: **MySQL**.
- **Frontend** en **React**.
- Todo **dockerizado** (se levanta con un comando).
- Hay **dos roles**: **Cliente** (compra, cancela, traspasa entradas) y **Administrador** (crea/edita/cancela eventos y ve reportes).

---

## 1. Go en 10 minutos (lo mínimo para defender)

Go es un lenguaje compilado. Estas son las cosas que vas a ver en TU código y cómo explicarlas:

### 1.1 Paquetes (`package`)
Cada carpeta es un **paquete**. Arriba de cada archivo dice a qué paquete pertenece, por ejemplo `package services`. Para usar código de otro paquete se hace `import`.

```go
package services            // este archivo es del paquete "services"

import "ticketek-backend/dao"   // traigo el paquete dao para usarlo
```

### 1.2 Structs (son como "clases" / fichas de datos)
Un `struct` agrupa campos. Por ejemplo, una entrada (Ticket):

```go
type Ticket struct {
    UserID  uint    // de quién es la entrada
    EventID uint    // a qué evento
    Status  string  // "active", "cancelled" o "transferred"
    QRCode  string  // el QR en texto
    Price   float64 // precio
}
```

### 1.3 Funciones y métodos
- **Función**: `func sumar(a int, b int) int { return a + b }`
- **Método**: una función "pegada" a un struct. Se reconoce por el paréntesis **antes** del nombre (el *receiver*):

```go
func (s *TicketService) Cancel(...) { ... }
//      ^^^^^^^^^^^^^^^^ esto significa: "Cancel es un método del TicketService"
```

### 1.4 Punteros: `*` y `&`
- `*Ticket` = "un puntero a Ticket" = la **dirección** donde vive el dato, no una copia.
- `&ticket` = "dame la dirección de ticket".
- **Por qué se usan**: para **modificar el original** (no una copia) y para no copiar datos grandes. Si ves `*` y `&`, decí: *"trabajo con la referencia al dato real para poder modificarlo"*.

### 1.5 Manejo de errores (¡clave en Go!)
Go **no usa try/catch**. Las funciones devuelven **dos cosas**: el resultado y un error. Se chequea con `if err != nil`:

```go
ticket, err := s.ticketDAO.FindByID(ticketID)
if err != nil {
    return nil, domain.ErrEntradaNoEncontrada  // si hubo error, corto acá
}
// si llegué hasta acá, todo salió bien y uso "ticket"
```

> `:=` crea una variable nueva. `nil` es "nada / vacío".

### 1.6 Las dos librerías estrella
- **Gin** = el framework web. Maneja las rutas (URLs) y el objeto `c *gin.Context`, que tiene la petición que entra y la respuesta que sale.
- **GORM** = el ORM. Traduce tus structs de Go a **tablas de MySQL**. Así no escribís SQL a mano:
  - `db.Create(&x)` → INSERT
  - `db.First(&x, id)` → SELECT por id
  - `db.Where("...").Find(&lista)` → SELECT con condición
  - `db.Save(&x)` → UPDATE
  - `db.Delete(&x)` → DELETE

---

## 2. Arquitectura: el patrón MVC en capas

El backend está separado en capas. Cada **petición HTTP** atraviesa las capas de arriba hacia abajo. **Esto es lo más importante para empezar la defensa**, porque ordena todo lo demás.

```
Navegador (React)
      │  pide algo por HTTP (ej: POST /api/tickets)
      ▼
┌─────────────────────────────────────────────────────────┐
│ middleware/   → revisa el token JWT (¿quién sos? ¿podés?) │
├─────────────────────────────────────────────────────────┤
│ controllers/  → recibe la petición HTTP, lee el body,     │
│                 llama al service y arma la respuesta       │
├─────────────────────────────────────────────────────────┤
│ services/     → la LÓGICA DE NEGOCIO (reglas, validaciones)│
├─────────────────────────────────────────────────────────┤
│ dao/          → habla con la base de datos (queries GORM)  │
├─────────────────────────────────────────────────────────┤
│ domain/       → las entidades (User, Event, Ticket) y      │
│                 los errores del negocio                    │
└─────────────────────────────────────────────────────────┘
      │
      ▼
   MySQL
```

- **`utils/`**: ayudas reutilizables (hashing de contraseñas, generar JWT, generar QR, armar respuestas JSON).
- **`main.go`**: el punto de arranque. Conecta a la base, crea el admin inicial y arma el servidor con todas las rutas.

**Frase para arrancar el oral:** *"El backend está hecho en MVC por capas: el controller recibe la petición HTTP, delega la lógica al service, el service usa el dao para tocar la base, y el domain define las entidades. El middleware valida el token JWT antes de llegar al controller."*

### Ejemplo concreto: comprar una entrada (`POST /api/tickets`)
1. **middleware** valida el token → saca tu `user_id`.
2. **controller** (`Buy`) lee del body el `event_id`, tipo de entrada y método de pago.
3. **service** (`Buy`) aplica las reglas: ¿existe el evento?, ¿está cancelado?, ¿hay stock?, calcula el precio **en el servidor** (no confía en el cliente), descuenta una entrada, genera el QR y registra el pago.
4. **dao** hace los INSERT/UPDATE en MySQL.
5. **controller** devuelve la entrada creada con código `201 Created`.

---

## 3. ⭐ Funcionalidad nueva (Bonus): Entradas con QR + Traspaso

Esta es la estrella del oral. Tu bonus tiene **dos partes**: el **código QR** que se genera al comprar, y el **traspaso** de la entrada a otra persona (que regenera el QR y cambia de titular).

> Nota: el bonus originalmente incluía email, pero **se sacó el email** (último commit) y quedó **solo el QR + traspaso**. Si preguntan: *"simplificamos el bonus a QR y traspaso para enfocarnos en la lógica del dominio"*.

### 3.1 Cómo se genera el QR — `utils/qr.go`

```go
// Genera el código QR como una imagen en base64 a partir de un texto único.
func GenerateQRCode(data string) (string, error) {
    png, err := qrcode.Encode(data, qrcode.Medium, 256) // crea el PNG del QR (256px)
    if err != nil {
        return "", fmt.Errorf("error generando QR: %w", err)
    }
    encoded := base64.StdEncoding.EncodeToString(png)   // paso la imagen a texto base64
    return "data:image/png;base64," + encoded, nil      // formato que el navegador muestra directo
}

// Arma el texto único que va DENTRO del QR.
func GenerateTicketQRData(ticketID, userID, eventID uint) string {
    return fmt.Sprintf("TICKET-%d-USER-%d-EVENT-%d", ticketID, userID, eventID)
}
```

**Explicación para decir en voz alta:**
- Usamos la librería **`go-qrcode`** (`github.com/skip2/go-qrcode`).
- El QR **no guarda una imagen en disco**: se genera un PNG, se convierte a **base64** y se devuelve como un *data URL* (`data:image/png;base64,...`). Eso permite que el `<img>` del frontend lo muestre directo, sin servir un archivo.
- El **contenido** del QR es un texto único por entrada: `TICKET-{id}-USER-{titular}-EVENT-{evento}`. Identifica de quién es la entrada y para qué evento.
- **Por qué importa en el traspaso:** cuando una entrada cambia de dueño, el `USER` que va dentro del QR cambia, entonces **se regenera el QR** → el QR viejo deja de ser válido para el nuevo titular. Eso garantiza la integridad del traspaso.

### 3.2 El traspaso — `services/ticket_service.go` (función `Transfer`)

Esta es la parte que **ajustamos en esta entrega**. La idea del traspaso:
- El dueño **A** le pasa la entrada a **B** (por email).
- La entrada **deja de aparecer** en "Mis Entradas" de A.
- Se le crea una entrada **nueva y activa a B**, con **QR nuevo**.
- Si la entrada después **vuelve** a A, se **limpia** el registro viejo de traspaso de A.

```go
func (s *TicketService) Transfer(ticketID, ownerID uint, req domain.TransferRequest) (*domain.Ticket, error) {
    ticket, err := s.ticketDAO.FindByID(ticketID)
    if err != nil {
        return nil, domain.ErrEntradaNoEncontrada       // la entrada no existe → 404
    }
    if ticket.UserID != ownerID {
        return nil, domain.ErrSinPermiso                 // no es tuya → 403
    }
    if ticket.Status != domain.TicketStatusActive {
        return nil, errors.New("la entrada no está activa")  // ya cancelada/traspasada → 400
    }

    targetUser, err := s.userDAO.FindByEmail(req.TargetEmail)
    if err != nil {
        return nil, domain.ErrUsuarioNoEncontrado        // el destinatario no existe → 404
    }
    if targetUser.ID == ownerID {
        return nil, errors.New("no podés traspasar la entrada a vos mismo")
    }

    // Si la entrada "vuelve" al destinatario (antes él la había traspasado),
    // borramos su registro viejo de traspaso para ese evento, así no le queda duplicado.
    if err := s.ticketDAO.DeleteTransferredByUserAndEvent(targetUser.ID, ticket.EventID); err != nil {
        return nil, errors.New("error al limpiar entradas traspasadas")
    }

    // Marcamos la entrada original como "traspasada": deja de ser de A
    // (no le aparece más en "Mis Entradas") pero queda el registro del traspaso.
    ticket.Status = domain.TicketStatusTransferred
    if err := s.ticketDAO.Update(ticket); err != nil {
        return nil, errors.New("error al traspasar la entrada")
    }

    // Generamos un QR NUEVO para el nuevo titular.
    qrData := utils.GenerateTicketQRData(ticket.ID, targetUser.ID, ticket.EventID)
    newQR, err := utils.GenerateQRCode(qrData)
    if err != nil {
        return nil, fmt.Errorf("error generando nuevo QR: %w", err)
    }

    // Creamos la entrada nueva para B, conservando tipo y precio.
    newTicket := &domain.Ticket{
        UserID:     targetUser.ID,
        EventID:    ticket.EventID,
        Status:     domain.TicketStatusActive,
        QRCode:     newQR,
        TicketType: ticket.TicketType,
        Price:      ticket.Price,
    }
    if err := s.ticketDAO.Create(newTicket); err != nil {
        return nil, errors.New("error al crear entrada para el nuevo titular")
    }
    return newTicket, nil
}
```

**El "estado" de una entrada** (`Status`) puede ser:
- `active` → válida y usable.
- `cancelled` → anulada por el dueño.
- `transferred` → fue traspasada (registro histórico, ya no es del dueño).

### 3.3 Lo que aparece (y lo que NO) en "Mis Entradas" — `dao/ticket_dao.go`

El arreglo clave para que la entrada **no aparezca más** en la cuenta de quien la traspasó:

```go
// "Mis Entradas": trae las entradas del usuario, PERO excluye las traspasadas.
func (d *TicketDAO) FindByUserID(userID uint) ([]domain.Ticket, error) {
    var tickets []domain.Ticket
    err := d.db.Preload("Event").
        Where("user_id = ? AND status <> ?", userID, domain.TicketStatusTransferred).
        Find(&tickets).Error
    return tickets, err
}

// Borra el registro de traspaso de un usuario para un evento (cuando la entrada "vuelve").
func (d *TicketDAO) DeleteTransferredByUserAndEvent(userID, eventID uint) error {
    return d.db.
        Where("user_id = ? AND event_id = ? AND status = ?", userID, eventID, domain.TicketStatusTransferred).
        Delete(&domain.Ticket{}).Error
}
```

> `status <> ?` significa "estado distinto de". `Preload("Event")` trae también los datos del evento de cada entrada (es la relación de GORM).

### 3.4 ✅ Si te preguntan sobre la funcionalidad

- **"¿Cómo se genera el QR?"** → *"Con la librería go-qrcode genero un PNG, lo paso a base64 y lo devuelvo como data URL para que el navegador lo muestre directo. El contenido es un texto único TICKET-USER-EVENT."*
- **"¿Por qué base64 y no un archivo?"** → *"Para no tener que guardar ni servir imágenes; va embebido en el JSON de la respuesta."*
- **"¿Qué pasa con el QR al traspasar?"** → *"Se regenera, porque el USER que va adentro cambia; así el QR viejo deja de servir y se garantiza la integridad (la consigna pide que cambie de titular de forma íntegra)."*
- **"¿Cómo garantizás que la entrada cambie de titular?"** → *"La original pasa a estado 'transferred' (y deja de verse en Mis Entradas del dueño) y se crea una entrada nueva 'active' para el destinatario."*
- **"¿Qué pasa si la entrada vuelve a la persona original?"** → *"Antes de crearle la entrada nueva, borro su registro viejo de traspaso para ese evento, así no le queda duplicado."*
- **"¿Dónde está la lógica?"** → *"En services/ticket_service.go, función Transfer. Las queries están en dao/ticket_dao.go y el QR en utils/qr.go."*

---

## 4. Panel de Administrador

El admin tiene una sección protegida. **Mostralo logueándote como admin** y recorré estas acciones (todas pegan a endpoints `/api/admin/...` que exigen token de admin):

| Acción en la pantalla | Endpoint backend | Qué hace |
|---|---|---|
| Listar eventos | `GET /api/admin/events` | Trae **todos** los eventos (incluso cancelados) |
| Crear evento | `POST /api/admin/events` | Alta de un evento nuevo |
| Editar evento | `PUT /api/admin/events/:id` | Actualiza atributos |
| Cancelar evento | `DELETE /api/admin/events/:id` | Lo marca como cancelado (**soft delete**) |
| Ver reporte | `GET /api/admin/events/:id/report` | Métricas: capacidad, vendidas, canceladas, disponibles, compradores |

### 4.1 Cómo se protege el admin (seguridad por roles)
Dos middlewares en cadena (`middleware/auth_middleware.go`):
1. **`AuthRequired`**: exige un token JWT válido. Si no hay o es inválido → **401**.
2. **`AdminRequired`**: además, el rol dentro del token tiene que ser `admin`. Si es cliente → **403**.

```go
admin := api.Group("/admin")
admin.Use(middleware.AuthRequired(), middleware.AdminRequired()) // primero loguea, después chequea rol
```

> **Dato importante de seguridad:** un usuario que se registra **siempre** es Cliente. El rol Admin **no** se puede elegir en el registro: se crea con un *seed* inicial en `main.go` (función `seedAdmin`). Así nadie se auto-asciende a admin.

### 4.2 El reporte (soft delete + recuento) — `services/ticket_service.go`
El reporte recorre **todas** las entradas del evento y cuenta:
- `active`/otras → **vendidas**
- `cancelled` → **canceladas**
- arma la lista de **compradores** únicos.

> Un bug que se arregló: antes el contador de canceladas siempre daba 0, porque la query solo traía las activas. Se cambió por `FindAllByEventID` (trae todos los estados) y se recalcula. **Esto es bueno mencionarlo como "decisión de diseño / hallazgo".**

### 4.3 ✅ Si te preguntan sobre el admin
- **"¿Cómo cancelás un evento?"** → *"Con soft delete: no lo borro físicamente, le cambio el estado a 'cancelled'. Así no pierdo el histórico ni las entradas asociadas, y deja de aparecer en el catálogo público."*
- **"¿Cómo evitás que un cliente entre al admin?"** → *"Con dos middlewares: AuthRequired valida el token (401 si falta) y AdminRequired valida el rol (403 si es cliente)."*
- **"¿Cómo se vuelve admin alguien?"** → *"No desde el registro (siempre crea Cliente). El admin se provisiona con un seed inicial en main.go."*
- **"¿Qué muestra el reporte?"** → *"Capacidad total, vendidas, canceladas, disponibles y la lista de compradores."*

---

## 5. Docker

Docker empaqueta cada parte en un **contenedor** para que el proyecto **corra igual en cualquier máquina**, sin instalar Go, Node ni MySQL a mano. Hay **3 contenedores**: base de datos, backend y frontend, orquestados por **docker-compose**.

### 5.1 `docker-compose.yml` (el director de orquesta)

```yaml
services:
  db:                      # 1) MySQL
    image: mysql:8.0
    ports: ["3310:3306"]   # afuera 3310 → adentro 3306
    healthcheck: ...       # avisa cuando MySQL está listo

  backend:                 # 2) la API en Go
    build: ./backend       # se construye con backend/Dockerfile
    ports: ["8080:8080"]
    environment:           # le paso config por variables de entorno
      DB_HOST: db          # se conecta al contenedor "db" por su nombre
      JWT_SECRET: ...
    depends_on:
      db: { condition: service_healthy }   # espera a que MySQL esté sano

  frontend:                # 3) React servido por nginx
    build: ./frontend
    ports: ["3000:80"]     # entrás por localhost:3000
    depends_on: [backend]
```

**Puntos para explicar:**
- **`depends_on` + `healthcheck`**: el backend **espera** a que MySQL esté listo antes de arrancar (si no, fallaría la conexión).
- **Variables de entorno**: la config (host de la base, secreto del JWT, etc.) se inyecta desde afuera; el código las lee con `os.Getenv(...)`. Así no hay credenciales "hardcodeadas".
- **Red interna**: los contenedores se ven entre sí **por nombre** (`DB_HOST: db`), no por IP.
- **Volumen `db_data`**: persiste los datos de MySQL aunque apagues los contenedores.

### 5.2 Los `Dockerfile` (multi-stage build)

**Backend** (`backend/Dockerfile`):
```dockerfile
FROM golang:1.22-alpine AS builder   # etapa 1: imagen con Go para COMPILAR
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download                  # baja las dependencias
COPY . .
RUN go build -o ticketek-backend .   # compila a un único ejecutable

FROM alpine:latest                   # etapa 2: imagen mínima, SIN Go
COPY --from=builder /app/ticketek-backend .   # copio SOLO el ejecutable
EXPOSE 8080
CMD ["./ticketek-backend"]           # lo ejecuto
```

**Por qué "multi-stage" (multietapa):** compilo en una imagen grande (con Go) pero la imagen final solo lleva el **ejecutable** sobre Alpine (Linux mínimo). Resultado: imagen **chica y segura** (no viaja el código fuente ni el compilador).

**Frontend** (`frontend/Dockerfile`): misma idea: etapa 1 con Node hace `npm run build` (genera los archivos estáticos), etapa 2 los sirve con **nginx**.

### 5.3 ✅ Si te preguntan sobre Docker
- **"¿Para qué Docker?"** → *"Para que el proyecto corra igual en cualquier máquina sin instalar nada a mano; con un comando levanto base, backend y frontend."*
- **"¿Qué es multi-stage?"** → *"Compilo en una imagen con las herramientas y la imagen final solo lleva el ejecutable, así queda chica."*
- **"¿Cómo se conecta el backend a la base?"** → *"Por el nombre del servicio 'db' en la red interna de compose, usando variables de entorno; y espera con depends_on + healthcheck a que MySQL esté listo."*
- **"¿Y si se apaga? ¿se pierden los datos?"** → *"No, MySQL usa un volumen (db_data) que los persiste."*

---

## 6. ⭐ El arreglo de los tests (por qué daba 0% y ahora da 86.7%)

Este tema te lo van a preguntar sí o sí. Entendelo bien porque es **conceptual de cómo mide cobertura Go**.

### 6.1 Qué es la cobertura
**Cobertura (coverage)** = qué porcentaje de las líneas de tu código fueron **ejecutadas** por los tests. Más cobertura = más código probado.

### 6.2 Por qué ANTES daba 0%
Antes **todos los tests estaban juntos en una carpeta `tests/`** y probaban la app **por HTTP de punta a punta** (eso técnicamente son tests de **integración**).

El problema es **cómo mide Go la cobertura con este comando**:

```
go test ./... -coverprofile=coverage.out
```

> Go mide la cobertura **paquete por paquete**: a cada paquete lo evalúa contra **sus propios archivos de test** (`_test.go` que estén en esa misma carpeta).

Como **`services/`, `controllers/`, `dao/` no tenían ningún `_test.go` adentro**, Go los reportaba en **0%** — aunque la carpeta `tests/` los ejecutara por dentro, esa ejecución **no se le atribuye** a esos paquetes con este comando. Resultado: **0.0%**.

(El plan viejo lo "esquivaba" usando `-coverpkg=./services/...,./controllers/...`, que **fuerza** a medir esos paquetes desde afuera. Pero tu comando nuevo **no** usa `-coverpkg`, así que daba 0.)

### 6.3 Qué hice para que dé 86.7%
**Moví/creé los tests DENTRO de cada paquete** (tests unitarios de caja blanca), que es la forma idiomática en Go y la única que hace funcionar ese comando sin `-coverpkg`:

| Paquete | Archivo de test creado | Qué prueba | Cobertura |
|---|---|---|---|
| `utils/` | `utils_test.go` | hashing, JWT, QR, respuestas | 91.7% |
| `dao/` | `dao_test.go` | queries contra SQLite | 100% |
| `services/` | `services_test.go` | toda la lógica de negocio | 89.8% |
| `controllers/` | `controllers_test.go` | endpoints HTTP con `httptest` (status codes éxito/error) | 89.4% |
| `middleware/` | `auth_middleware_test.go` | validación de token y rol | 100% |
| `main` | `main_test.go` | seed del admin y armado del router | — |

Además hice **un refactor chico**: saqué el armado del router de `main()` a una función `SetupRouter(db)` para poder **testear las rutas** sin levantar el servidor real ni MySQL.

**Total: 86.7%** (supera el 80% que pide la consigna).

### 6.4 ¿Esto cumple la consigna?
Sí. La consigna (punto 5, Testing) pide **tres cosas** y están las tres:
1. **Pruebas unitarias** (`testing`) → en services, dao, utils, middleware. ✓
2. **Pruebas de integración de controllers con `httptest`** (status codes de éxito y error) → en `controllers_test.go`. ✓
3. **80% de cobertura** en services y controllers → 86.7%. ✓

> La consigna **no exige** una carpeta `tests/`; eso era una decisión de organización. De hecho, separar unitarias (en cada paquete) de integración (httptest) está **más alineado** con cómo la consigna distingue ambos tipos.

### 6.5 Por qué los tests usan SQLite y no MySQL
Para que los tests sean **rápidos y aislados**, cada test abre una base **SQLite** temporal (archivo propio) y la migra con GORM. Como usamos GORM, el mismo código anda con MySQL (producción) y SQLite (tests). En `main.go`, `initDB` y `main` quedan en 0% porque **necesitan MySQL real y levantar el servidor**, que no se testea como unidad.

### 6.6 ✅ Si te preguntan sobre los tests
- **"¿Por qué antes daba 0%?"** → *"Porque los tests vivían en una carpeta tests/ aparte y Go mide cobertura por paquete contra sus propios _test.go. Como services y controllers no tenían tests adentro, daban 0 con `go test ./...` sin `-coverpkg`."*
- **"¿Qué hiciste?"** → *"Escribí los tests dentro de cada paquete (unitarios), más los de integración de controllers con httptest. Y saqué el router a SetupRouter para poder testearlo. Quedó en 86.7%."*
- **"¿Diferencia unitario vs integración?"** → *"El unitario prueba una pieza aislada (un service con una base SQLite). El de integración prueba el endpoint HTTP completo con httptest y verifica los status codes."*
- **"¿Por qué SQLite en los tests?"** → *"Para que sean rápidos y aislados; GORM nos deja usar SQLite en test y MySQL en producción con el mismo código."*

---

## 7. Comandos (chuleta)

### 7.1 Levantar TODO con Docker (lo más fácil para la demo)
Desde la carpeta del proyecto (donde está `docker-compose.yml`):
```bash
docker-compose up --build      # construye y levanta base + backend + frontend
```
Luego abrís en el navegador:
- **Frontend:** http://localhost:3000
- **Backend (API):** http://localhost:8080

Para apagar todo:
```bash
docker-compose down            # apaga los contenedores (los datos quedan en el volumen)
```

### 7.2 Levantar SIN Docker (modo desarrollo)

**Backend** (desde `backend/`):
```bash
go run .            # compila y ejecuta el servidor (necesita MySQL corriendo)
# o
go build ./...      # solo compila, para verificar que todo está bien
```

**Frontend** (desde `frontend/`):
```bash
npm install         # instala dependencias (solo la primera vez)
npm run dev         # levanta en modo desarrollo → http://localhost:5173
```

### 7.3 Tests y cobertura (los comandos que te van a pedir)
Desde `backend/`:
```bash
go test ./...                                  # corre todos los tests
go test ./... -cover                           # con el % por paquete
go test ./... -coverprofile=coverage.out       # guarda el detalle de cobertura
go tool cover -func=coverage.out               # muestra cobertura por función + TOTAL
go tool cover -html=coverage.out               # abre un HTML con las líneas cubiertas
```
> El último (`-html`) es ideal para **mostrar en pantalla** qué líneas están cubiertas (verde) y cuáles no (rojo).

---

## 8. Glosario rápido (por si tiran una palabra)

- **API REST**: la forma en que el frontend le pide cosas al backend por HTTP (GET, POST, PUT, DELETE).
- **JWT**: un "carnet" digital firmado que se da al loguearte; lleva tu id y rol y vence a las 24 h. El backend lo valida sin guardar sesión en la base.
- **Hashing (SHA-256)**: la contraseña **nunca** se guarda en texto plano; se guarda su "huella" irreversible. Al loguearte, se compara la huella.
- **ORM (GORM)**: traduce structs de Go ↔ tablas de MySQL; evita escribir SQL a mano.
- **Middleware**: código que se ejecuta **antes** del controller (acá, valida token y rol).
- **Soft delete**: "borrar" cambiando un estado (a `cancelled`) en vez de eliminar físicamente.
- **base64 / data URL**: forma de meter una imagen (el QR) dentro de texto para mostrarla sin guardar archivos.
- **Contenedor / imagen (Docker)**: "caja" que empaqueta una app con todo lo que necesita para correr igual en cualquier lado.

---

## 9. Cierre: el guion de 60 segundos

> *"TicketHub es un sistema de venta de entradas tipo Ticketek. El backend está en Go con Gin y GORM, organizado en MVC por capas: controllers reciben la petición HTTP, services tienen la lógica, dao habla con MySQL y domain define las entidades; un middleware valida el JWT. Hay roles Cliente y Administrador. Mi funcionalidad extra son las entradas con QR que se pueden traspasar: al comprar se genera un QR en base64, y al traspasar la entrada cambia de titular, se regenera el QR y deja de aparecer en 'Mis Entradas' del dueño anterior. Todo está dockerizado con docker-compose (base, backend y frontend) y los tests llegan al 86.7% de cobertura: los reorganicé dentro de cada paquete para medir bien la cobertura con `go test ./...`."*

¡Mucha suerte! 🎫
