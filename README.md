# BloodLink Platform — Red Nacional Interconectada de Bancos de Sangre del Perú

Backend **Modular Monolith** construido con **Arquitectura Hexagonal (Puertos y Adaptadores)** en **Go (Gin + GORM + PostgreSQL 15)**, diseñado bajo un enfoque **Contract-First** con interoperabilidad híbrida **REST/JSON (`/api/v1`)** y **SOAP 1.1 / WSDL / UDDI (`/soap/v1`)**.

---

## 1. Arquitectura Hexagonal por Bounded Context (`backend/internal/modules/`)

Cada uno de los módulos de negocio respeta las 4 capas hexagonales (`domain/`, `ports/`, `application/`, `adapters/`):

1. **`auth`**: IAM, autenticación JWT (`HS256`), bloqueo automático de cuenta por 15 minutos tras 3 intentos fallidos consecutivos (`HTTP 423 Locked`) y control de acceso basado en roles (`RBAC`: `MEDICO`, `PERSONAL_BANCO`, `COORDINADOR`).
2. **`inventario` (`MS-Inventario`)**: Gestión de stock regional e inter-hospitalario, máquina de estados de `UnidadSangre` (`Disponible -> Reservada | Caducada`, `Reservada -> Transfundida | Disponible`), validación de vigencia de 42 días (`422 Unprocessable Entity` si ingresa vencida — `CUN-003 Flujo 2A`) y semáforo de alertas (`VERDE` > 7d, `AMARILLO` ≤ 7d, `ROJO` ≤ 3d).
3. **`emergencias` (`MS-Emergencias` + Motor de Matching)**: Orquestación transaccional (`SrvAtenderEmergencia`) con regla de prioridad local, búsqueda en red externa, cálculo geoespacial Haversine con caché en memoria thread-safe (`IGeoCache`), reserva atómica con bloqueo pesimista PostgreSQL (`SELECT ... FOR UPDATE SKIP LOCKED`), fallback por jurisdicción ante caída de mapas (`Contingencia E1`) y activación automática de convocatoria de donantes (`409 Conflict` + `StockCriticoDetectado`).
4. **`donantes` (`MS-Donantes`)**: Validación de identidad ciudadana mediante cliente SOAP 1.1 contra RENIEC (`IReniec`), cifrado simétrico **AES-256-GCM** de datos sensibles (`DNI` y `celular`) antes de persistir (`RNF-04`), hash `SHA-256` para unicidad y enmascaramiento en consultas (`****3210`, `+519****4321`).
5. **`notificaciones` (`MS-Notificaciones`)**: Suscriptor asíncrono conectado al Bus de Eventos (`IEventBus`) para `UnidadRegistrada`, `EstadoUnidadActualizado`, `UnidadProximaACaducar`, `EmergenciaReservada`, `SolicitudCancelada` y `StockCriticoDetectado` (envío de alertas SMS/Email a donantes compatibles).
6. **`reportes` (`MS-Reportes`)**: Dashboard de KPIs en tiempo real y exportación del certificado de trazabilidad clínica de bolsas de sangre (`AuditLog`) en formatos **JSON**, **CSV (Excel)** y **PDF** (`go-pdf/fpdf`).
7. **`integracion` (`MS-Integracion`)**:
   - Adaptador geoespacial Haversine + `IGeoCache`.
   - Servidor y cliente **SOAP 1.1 / WSDL** (`ServicioBloodLinkSOAP`) en `/soap/v1/interoperabilidad?wsdl` con las operaciones `ValidarIdentidadReniec` y `ConsultarYReservarStockLegacy`.
   - Directorio **UDDI** (Páginas Blancas, Amarillas y Verdes) en `/api/v1/uddi/servicios`.

---

## 2. Datos Semilla (Seed Automático Idempotente)

Al iniciar el backend se cargan automáticamente:

### Bancos de Sangre en Piura
| ID | Nombre | Institución | Coordenadas | Jurisdicción | SOAP Legado | Stock Inicial Semilla |
|---|---|---|---|---|---|---|
| `HOSP-MINSA-01` | Hospital Nacional Santa Rosa - Piura | `MINSA` | `-5.1945,-80.6328` | `Piura Oeste` | `false` | **0 unidades `O+`** (para disparar `CP-01`/`CP-02`), 3 `A-`, 2 `B+`, 2 `O-` (vencen en 4 días) |
| `HOSP-ESSALUD-04` | Hospital III José Cayetano Heredia - EsSalud | `EsSalud` | `-5.1821,-80.6190` | `Castilla` | **`true`** | **10 unidades `O+`** disponibles (a ~2.4 km de `HOSP-MINSA-01`) |
| `HOSP-MINSA-02` | Hospital de Apoyo II-2 Sullana | `MINSA` | `-4.9039,-80.6853` | `Sullana` | `false` | **6 unidades `O+`** disponibles |
| `CLIN-PRIV-01` | Clínica Miraflores - Piura | `Privada` | `-5.1880,-80.6250` | `Castilla` | `false` | **4 unidades `A+`**, **3 unidades `O-`** |

### Usuarios de Prueba (Credenciales RBAC)
| Rol | Correo | Contraseña | Usuario ID | Hospital Asignado |
|---|---|---|---|---|
| `MEDICO` | `medico@bloodlink.gob.pe` | `Medico123!` | `12` (`Dr. Carlos Mendoza`) | `HOSP-MINSA-01` |
| `PERSONAL_BANCO` | `banco@bloodlink.gob.pe` | `Banco123!` | `13` (`Lic. María Torres`) | `HOSP-MINSA-01` |
| `COORDINADOR` | `coordinador@bloodlink.gob.pe` | `Coord123!` | `14` (`Mg. Roberto Paredes`) | `HOSP-MINSA-01` |

### Padrón Simulado SOAP RENIEC (`ValidarIdentidadReniec`)
- **`76543210`** -> `JUAN CARLOS PEREZ GOMEZ` (`mayorEdad: true`) — *Listo para registrar en `CP-03`*.
- Otros DNIs válidos mayores de edad: `71234567`, `72345678`, `45678901`, `48901234`, `70112233`.
- DNI menor de edad (retorna `422 Unprocessable Entity`): `61112222` (`mayorEdad: false`).

---

## 3. Comandos en PowerShell para Levantar y Ejecutar Pruebas

### 3.1. Ejecutar Suite de Pruebas Unitarias y Cobertura (`≥ 80%`)
```powershell
cd backend
go test -v -cover ./...
```

### 3.2. Levantar Plataforma con Docker Compose (PostgreSQL 15 + Backend Go)
```powershell
cd ..
docker compose up -d --build
docker compose ps
```

Verificar estado de salud (`/health`), contrato WSDL y directorio UDDI:
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get
Invoke-WebRequest -Uri "http://localhost:8080/soap/v1/interoperabilidad?wsdl" -Method Get | Select-Object -ExpandProperty Content
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/uddi/servicios" -Method Get | ConvertTo-Json -Depth 5
```

---

## 4. Guía de Ejecución de Casos de Prueba (`CP-01` a `CP-04`) en PowerShell

### Paso Previo: Autenticación JWT (Obtener Tokens para `MEDICO` y `PERSONAL_BANCO`)
```powershell
$loginMedico = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -ContentType "application/json" -Body '{"email":"medico@bloodlink.gob.pe","password":"Medico123!"}'
$tokenMedico = $loginMedico.accessToken

$loginBanco = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -ContentType "application/json" -Body '{"email":"banco@bloodlink.gob.pe","password":"Banco123!"}'
$tokenBanco = $loginBanco.accessToken
```

### Caso de Prueba `CP-01`: Consulta de Disponibilidad de Stock Regional (`O+`)
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/inventario/stock?tipoSangre=O&factorRh=%2B" `
  -Method Get -Headers @{ Authorization = "Bearer $tokenMedico" } | ConvertTo-Json -Depth 5
```

### Caso de Prueba `CP-02`: Solicitud de Emergencia y Matching Automático con Reserva Atómica
Como `HOSP-MINSA-01` tiene `0` unidades `O+`, el motor salta a `RED_EXTERNA`, calcula distancias con Haversine, selecciona a `HOSP-ESSALUD-04` (~2.49 km, hospital legado SOAP) y reserva atómicamente 2 bolsas `O+`:
```powershell
$bodyEmergencia = @{
    hospitalOrigenId  = "HOSP-MINSA-01"
    coordenadasOrigen = "-5.1945,-80.6328"
    medicoId          = 12
    tipoSangreReq     = "O"
    factorRhReq       = "+"
    cantidadReq       = 2
    urgencia          = "ALTA_VITAL"
    simularFalloMapas = $false
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/emergencias" `
  -Method Post -ContentType "application/json" `
  -Headers @{ Authorization = "Bearer $tokenMedico" } `
  -Body $bodyEmergencia | ConvertTo-Json -Depth 5
```

### Caso de Prueba `CP-03`: Validación SOAP RENIEC y Registro de Donante con Cifrado AES-256-GCM
```powershell
# 1. Validar DNI 76543210 vía Bus SOAP RENIEC
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/donantes/validar-dni" `
  -Method Post -ContentType "application/json" -Body '{"dni":"76543210"}' | ConvertTo-Json

# 2. Registrar donante voluntario (cifra DNI y celular con AES-256-GCM en PostgreSQL)
$bodyDonante = @{
    dni                      = "76543210"
    celular                  = "+51987654321"
    correo                   = "jperez@ciudadano.pe"
    tipoSangre               = "O"
    factorRh                 = "+"
    bancoPreferidoId         = "HOSP-MINSA-01"
    aceptaPoliticaPrivacidad = $true
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/donantes" `
  -Method Post -ContentType "application/json" -Body $bodyDonante | ConvertTo-Json
```

### Validación de Regla `CUN-003 Flujo 2A`: Bloqueo `422 Unprocessable Entity` por Unidad Vencida (> 42 días)
```powershell
$bodyVencida = @{
    bancoId         = "HOSP-MINSA-01"
    tipoSangre      = "O"
    factorRh        = "+"
    volumenMl       = 450
    fechaExtraccion = "2025-01-10"
    origen          = "DONACION_VOLUNTARIA"
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8080/api/v1/inventario/unidades" `
  -Method Post -ContentType "application/json" `
  -Headers @{ Authorization = "Bearer $tokenBanco" } `
  -Body $bodyVencida -SkipHttpErrorCheck
```

### Descarga de Certificado de Trazabilidad en PDF y CSV (`RF-13`)
```powershell
Invoke-WebRequest -Uri "http://localhost:8080/api/v1/reportes/trazabilidad/UND-2026-0401?format=pdf" `
  -Headers @{ Authorization = "Bearer $tokenBanco" } -OutFile "trazabilidad_UND-2026-0401.pdf"

Invoke-WebRequest -Uri "http://localhost:8080/api/v1/reportes/trazabilidad/UND-2026-0401?format=csv" `
  -Headers @{ Authorization = "Bearer $tokenBanco" } -OutFile "trazabilidad_UND-2026-0401.csv"
```

### Caso de Prueba `CP-04` (`RNF-01`): Prueba de Carga con 100 Usuarios Concurrentes
```powershell
cd backend
go run ./cmd/loadtest/main.go -url http://localhost:8080 -c 100 -n 3 -mode mixed
```
