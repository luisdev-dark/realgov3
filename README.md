# Backend MVP - Transporte

## 📋 Estructura del proyecto

```
backend/
├── main.go              # Bootstrap de la app
├── db/db.go             # Conexión a Postgres (Neon)
├── models/              # Structs Go (Route, Trip, User, etc.)
├── handlers/            # Lógica de endpoints HTTP
├── routes/              # Configuración de rutas chi
├── seed.sql             # Datos de prueba
└── test_connection.go   # Script para probar conexión
```

## 🔧 Pasos para configurar

### 1. Verificar conexión a Neon

Ejecuta el script de prueba:

```bash
go run test_connection.go
```

**Salida esperada:**
```
Conectando a Neon...
DATABASE_URL: postgresql://neondb_owner:...
✅ Conexión exitosa a Neon!
📦 Versión de Postgres: PostgreSQL 16.x...

📋 Tablas en la base de datos:
  - routes
  - route_stops
  - trips
  - users
  Total: 4 tablas

🚀 Rutas en la base de datos:
  - Ruta Centro - Norte: Centro de Lima → Norte de Lima
  - Ruta Sur - Este: Sur de Lima → Este de Lima

✅ Prueba de conexión completada!
```

### 2. Ejecutar seed.sql en Neon

**Opción A: Desde el dashboard de Neon**
1. Ve a https://console.neon.tech
2. Abre tu proyecto
3. Ve a "SQL Editor"
4. Copia el contenido de `seed.sql`
5. Pega y ejecuta (▶️ Run)

**Opción B: Desde terminal (si tienes psql instalado)**
```bash
psql $DATABASE_URL < seed.sql
```

**Opción C: Usando el script de Go**
```bash
# Primero ejecuta el seed en el SQL Editor de Neon
# Luego verifica con:
go run test_connection.go
```

### 3. Ejecutar el servidor

```bash
go run .
# o
.\server.exe
```

El servidor iniciará en `http://localhost:8080`

## 🧪 Probar los endpoints

### 1. Listar rutas
```bash
curl http://localhost:8080/routes
```

### 2. Ver detalle de ruta
```bash
curl http://localhost:8080/routes/11111111-1111-1111-1111-111111111111
```

### 3. Crear un viaje
```bash
curl -X POST http://localhost:8080/trips \
  -H "Content-Type: application/json" \
  -d '{
    "route_id": "11111111-1111-1111-1111-111111111111",
    "pickup_stop_id": "11111111-1111-1111-1111-111111111112",
    "dropoff_stop_id": "11111111-1111-1111-1111-111111111114",
    "payment_method": "cash"
  }'
```

### 4. Ver estado del viaje
```bash
curl http://localhost:8080/trips/{id_del_viaje}
```

## 📊 Datos de prueba

### Usuario
- ID: `00000000-0000-0000-0000-000000000001`
- Nombre: Juan Pérez

### Rutas
1. **Ruta Centro - Norte** (`11111111-1111-1111-1111-111111111111`)
   - Paradas: Plaza de Armas, Parque Kennedy, Estación Central
   - Precio: S/ 5.00

2. **Ruta Sur - Este** (`22222222-2222-2222-2222-222222222222`)
   - Paradas: Mall del Sur, Avenida Benavides, Terminal de Buses
   - Precio: S/ 6.50

## 🔍 Troubleshooting

### Error: "DATABASE_URL no está definida"
- Verifica que el archivo `.env` existe en la raíz del proyecto
- Verifica que contiene `DATABASE_URL=postgresql://...`

### Error: "No hay tablas en el schema 'app'"
- Ejecuta `seed.sql` en el SQL Editor de Neon
- Verifica que las tablas se crearon correctamente

### Error: "Ruta no encontrada"
- Verifica que ejecutaste el seed.sql
- Usa los UUIDs del seed.sql para las pruebas

## 📝 Endpoints del MVP

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/routes` | Lista todas las rutas activas |
| GET | `/routes/{id}` | Detalle de ruta con paradas |
| POST | `/trips` | Crear una reserva |
| GET | `/trips/{id}` | Estado del viaje |
