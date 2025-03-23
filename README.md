# API Backend - Cotizador Productos EML

Este repositorio contiene el backend para la aplicación de cotización y gestión de pedidos de EML. Desarrollado con Go y Gin Framework, proporciona una API RESTful para gestionar productos, usuarios, autenticación y pedidos.

## Tecnologías Utilizadas

- **Go**: Lenguaje de programación principal
- **Gin Framework**: Framework web para crear APIs
- **Bun**: ORM para interactuar con la base de datos PostgreSQL
- **JWT**: Para autenticación y autorización
- **PostgreSQL**: Base de datos relacional
- **CORS**: Configuración para comunicación con el frontend
- **Godotenv**: Para gestión de variables de entorno

## Estructura del Proyecto

```
backend/
├── db/              # Configuración de conexión a la base de datos
├── handlers/        # Controladores que manejan las peticiones HTTP
├── models/          # Definición de estructuras de datos y modelos
├── routes/          # Definición de rutas de la API
├── utils/           # Utilidades compartidas (JWT, middleware, email)
├── main.go          # Punto de entrada de la aplicación
├── .env             # Archivo de variables de entorno (no incluido en Git)
└── go.mod, go.sum   # Gestión de dependencias
```

## Modelos Principales

- **Usuarios**: Gestión de usuarios con autenticación y roles
- **Productos**: Catálogo de productos disponibles
- **Pedidos**: Órdenes de compra realizadas por los usuarios
- **Detalles de Pedido**: Elementos individuales dentro de un pedido

## Instalación y Ejecución

1. Clona el repositorio:
   ```bash
   git clone https://github.com/tu-usuario/cotizador-productos-EML.git
   cd cotizador-productos-EML/backend
   ```

2. Instala las dependencias:
   ```bash
   go mod download
   ```

3. Configura las variables de entorno:
   - Crea un archivo `.env` basado en el ejemplo proporcionado
   - Configura la conexión a la base de datos y otras variables necesarias

4. Inicia el servidor:
   ```bash
   go run main.go
   ```

5. El servidor estará disponible en `http://localhost:8000`

## Endpoints de la API

### Autenticación

- `POST /auth/register`: Registro de nuevos usuarios
- `GET /auth/verify-email`: Verificación de correo electrónico
- `POST /auth/resend-verification-email`: Reenvío de email de verificación
- `POST /auth/login`: Inicio de sesión
- `POST /auth/refresh-token`: Renovación de token
- `GET /auth/logout`: Cierre de sesión
- `POST /auth/forgot-password`: Solicitud de recuperación de contraseña
- `POST /auth/reset-password`: Restablecimiento de contraseña
- `GET /auth/verify`: Verificación de autenticación

### Usuarios

- Endpoints para gestión de usuarios con diferentes permisos según el rol

### Productos

- Endpoints para gestión del catálogo de productos (CRUD)
- Soporte para búsqueda y filtrado

### Pedidos

- `POST /orders/create-order`: Creación de nuevos pedidos
- `GET /orders/get-user-orders`: Obtener pedidos del usuario autenticado
- `GET /orders/get-order-detail`: Obtener detalle de un pedido específico
- `PATCH /orders/update-order-client`: Actualizar pedido (cliente)
- `GET /orders/get-orders`: Obtener todos los pedidos (admin)
- `PATCH /orders/update-order-admin`: Actualizar cualquier pedido (admin)

## Características Principales

### Sistema de Autenticación Completo

- Registro con verificación de email
- Manejo de sesiones con JWT
- Recuperación de contraseña
- Protección de rutas por rol

### Gestión de Usuarios

- Roles diferenciados (admin, cliente)
- Perfiles de usuario

### Catálogo de Productos

- Gestión completa de productos
- Precios, disponibilidad y categorización

### Sistema de Pedidos

- Creación de pedidos con múltiples productos
- Cálculo automático de totales
- Estados de pedido (pendiente, enviado, entregado, etc.)
- Historial de pedidos por usuario
- Detalles completos de los pedidos

### Seguridad

- Protección CORS configurada
- Hashing seguro de contraseñas
- Validación de datos de entrada
- Manejo de transacciones para integridad de datos

## Configuración de CORS

El backend está configurado para aceptar solicitudes únicamente desde la URL del frontend especificada en las variables de entorno, con soporte completo para cookies y credenciales.

## Contribución

1. Haz un fork del proyecto
2. Crea una nueva rama (`git checkout -b feature/nueva-caracteristica`)
3. Realiza tus cambios y haz commit (`git commit -m 'Añadir nueva característica'`)
4. Sube tus cambios (`git push origin feature/nueva-caracteristica`)
5. Abre un Pull Request 