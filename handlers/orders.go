package handlers

import (
	"bytes"
	"cotizador-productos-eml/models"
	"cotizador-productos-eml/utils"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type OrderHandler struct {
	db *bun.DB
}

func NewOrderHandler(db *bun.DB) *OrderHandler {
	return &OrderHandler{db: db}
}

// CreateOrderRequest estructura para recibir la solicitud de creación de pedido
type CreateOrderRequest struct {
	CiudadDestino    string               `json:"ciudad_destino" binding:"required"`
	DireccionDestino string               `json:"direccion_destino" binding:"required"`
	RutDestinatario  string               `json:"rut_destinatario" binding:"required"`
	Company          string               `json:"company"`
	TipoEnvio        string               `json:"tipo_envio" binding:"required"`
	MetodoPago       string               `json:"metodo_pago" binding:"required"`
	TipoDocumento    string               `json:"tipo_documento" binding:"required"`
	Items            []CreateOrderItemDTO `json:"items" binding:"required,dive"`
}

// CreateOrderItemDTO estructura para los items del pedido
type CreateOrderItemDTO struct {
	ProductoID int `json:"producto_id" binding:"required"`
	Cantidad   int `json:"cantidad" binding:"required,min=1"`
}

// PedidoResponse estructura para la respuesta JSON sin incluir el campo Usuario
type PedidoResponse struct {
	ID               int        `json:"id"`
	UsuarioId        int        `json:"usuario_id"`
	Total            int        `json:"total"`
	Estado           string     `json:"estado"`
	FechaEnvio       *time.Time `json:"fecha_envio,omitempty"`
	CiudadDestino    string     `json:"ciudad_destino"`
	DireccionDestino string     `json:"direccion_destino"`
	RutDestinatario  string     `json:"rut_destinatario"`
	Company          string     `json:"company"`
	TipoEnvio        string     `json:"tipo_envio"`
	MetodoPago       string     `json:"metodo_pago"`
	TipoDocumento    string     `json:"tipo_documento"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// CreateOrder maneja la creación de un nuevo pedido
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	// Obtener token de las cookies
	accessToken, err := c.Cookie("access_token")
	if err != nil || accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "No autenticado",
		})
		return
	}

	// Verificar y extraer datos del token
	claims, err := utils.ParseJWT(accessToken, os.Getenv("JWT_SECRET"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Token inválido o expirado",
		})
		return
	}

	// Extraer ID de usuario del token
	userID, ok := claims["sub"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Token inválido (ID de usuario no encontrado)",
		})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Iniciar transacción
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al iniciar transacción"})
		return
	}
	defer tx.Rollback()

	// Calcular total inicial
	var total int = 0

	// Procesar los items para calcular el total
	detalles := make([]*models.DetallePedido, 0, len(req.Items))
	for _, item := range req.Items {
		// Obtener producto
		producto := new(models.Producto)
		err := tx.NewSelect().Model(producto).Where("id = ?", item.ProductoID).Scan(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Producto no encontrado: " + err.Error()})
			return
		}

		// Calcular subtotal
		subtotal := producto.PrecioVenta * item.Cantidad
		total += subtotal

		// Crear detalle
		detalle := &models.DetallePedido{
			ProductoID:     item.ProductoID,
			Cantidad:       item.Cantidad,
			PrecioUnitario: producto.PrecioVenta,
			PrecioTotal:    subtotal,
		}
		detalles = append(detalles, detalle)
	}

	// Obtener la hora actual para los campos de fecha
	now := time.Now()

	// Crear pedido
	pedido := &models.Pedido{
		UsuarioId:        int(userID), // Usar el ID obtenido del token
		Total:            total,
		Estado:           "pendiente",
		CiudadDestino:    req.CiudadDestino,
		DireccionDestino: req.DireccionDestino,
		RutDestinatario:  req.RutDestinatario,
		Company:          req.Company,
		TipoEnvio:        req.TipoEnvio,
		MetodoPago:       req.MetodoPago,
		TipoDocumento:    req.TipoDocumento,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Insertar pedido
	_, err = tx.NewInsert().Model(pedido).Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear pedido: " + err.Error()})
		return
	}

	// Asignar PedidoID a los detalles
	for _, detalle := range detalles {
		detalle.PedidoID = pedido.ID
	}

	// Insertar detalles
	_, err = tx.NewInsert().Model(&detalles).Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear detalles: " + err.Error()})
		return
	}

	// Confirmar transacción
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al confirmar transacción"})
		return
	}

	// Crear respuesta sin incluir el campo Usuario
	respuesta := PedidoResponse{
		ID:               pedido.ID,
		UsuarioId:        pedido.UsuarioId,
		Total:            pedido.Total,
		Estado:           pedido.Estado,
		FechaEnvio:       pedido.FechaEnvio,
		CiudadDestino:    pedido.CiudadDestino,
		DireccionDestino: pedido.DireccionDestino,
		RutDestinatario:  pedido.RutDestinatario,
		Company:          pedido.Company,
		TipoEnvio:        pedido.TipoEnvio,
		MetodoPago:       pedido.MetodoPago,
		TipoDocumento:    pedido.TipoDocumento,
		CreatedAt:        pedido.CreatedAt,
		UpdatedAt:        pedido.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"mensaje": "Pedido creado correctamente",
		"pedido":  respuesta,
	})
}

// GetOrders devuelve la lista de pedidos con paginación
func (h *OrderHandler) GetOrders(c *gin.Context) {

	// Obtener parámetros de paginación
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	// Definir tamaño de página fijo a 7 elementos
	pageSize := 7
	offset := (page - 1) * pageSize

	// Crear query base para contar
	countQuery := h.db.NewSelect().Model((*models.Pedido)(nil))
	// Crear query base para obtener datos
	query := h.db.NewSelect().Model((*models.Pedido)(nil))

	// Contar el total de pedidos para la paginación
	count, err := countQuery.Count(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al contar pedidos: " + err.Error(),
		})
		return
	}

	// Obtener los pedidos paginados ordenados por fecha de creación descendente
	var pedidos []models.Pedido
	err = query.
		OrderExpr("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(c, &pedidos)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al obtener pedidos: " + err.Error(),
		})
		return
	}

	// Convertir a respuesta sin el campo Usuario
	respuesta := make([]PedidoResponse, 0, len(pedidos))
	for _, pedido := range pedidos {
		respuesta = append(respuesta, PedidoResponse{
			ID:               pedido.ID,
			UsuarioId:        pedido.UsuarioId,
			Total:            pedido.Total,
			Estado:           pedido.Estado,
			FechaEnvio:       pedido.FechaEnvio,
			CiudadDestino:    pedido.CiudadDestino,
			DireccionDestino: pedido.DireccionDestino,
			RutDestinatario:  pedido.RutDestinatario,
			Company:          pedido.Company,
			TipoEnvio:        pedido.TipoEnvio,
			MetodoPago:       pedido.MetodoPago,
			TipoDocumento:    pedido.TipoDocumento,
			CreatedAt:        pedido.CreatedAt,
			UpdatedAt:        pedido.UpdatedAt,
		})
	}

	// Calcular datos de paginación
	totalPages := (count + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"pedidos": respuesta,
			"pagination": gin.H{
				"total":       count,
				"page":        page,
				"page_size":   pageSize,
				"total_pages": totalPages,
				"has_next":    hasNext,
				"has_prev":    hasPrev,
			},
		},
	})
}

// GetUserOrders devuelve todos los pedidos del usuario autenticado
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	// Obtener ID del usuario del contexto (establecido por AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "No se encontró información del usuario",
		})
		return
	}

	// Obtener todos los pedidos del usuario ordenados por fecha de creación descendente
	var pedidos []models.Pedido
	err := h.db.NewSelect().
		Model(&pedidos).
		Where("usuario_id = ?", userID).
		OrderExpr("created_at DESC").
		Scan(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al obtener pedidos: " + err.Error(),
		})
		return
	}

	// Convertir a respuesta sin el campo Usuario
	respuesta := make([]PedidoResponse, 0, len(pedidos))
	for _, pedido := range pedidos {
		respuesta = append(respuesta, PedidoResponse{
			ID:               pedido.ID,
			UsuarioId:        pedido.UsuarioId,
			Total:            pedido.Total,
			Estado:           pedido.Estado,
			FechaEnvio:       pedido.FechaEnvio,
			CiudadDestino:    pedido.CiudadDestino,
			DireccionDestino: pedido.DireccionDestino,
			RutDestinatario:  pedido.RutDestinatario,
			Company:          pedido.Company,
			TipoEnvio:        pedido.TipoEnvio,
			MetodoPago:       pedido.MetodoPago,
			TipoDocumento:    pedido.TipoDocumento,
			CreatedAt:        pedido.CreatedAt,
			UpdatedAt:        pedido.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    respuesta,
	})
}

// DetallePedidoResponse estructura para el detalle de un pedido
type DetallePedidoResponse struct {
	ID             int    `json:"id"`
	ProductoID     int    `json:"producto_id"`
	Nombre         string `json:"nombre_producto"`
	Cantidad       int    `json:"cantidad"`
	PrecioUnitario int    `json:"precio_unitario"`
	PrecioTotal    int    `json:"precio_total"`
}

// PedidoDetalladoResponse estructura para la respuesta con los detalles completos
type PedidoDetalladoResponse struct {
	Pedido   PedidoResponse          `json:"pedido"`
	Detalles []DetallePedidoResponse `json:"detalles"`
}

// GetOrderDetail devuelve los detalles de un pedido específico
func (h *OrderHandler) GetOrderDetail(c *gin.Context) {
	// Obtener ID del pedido de los parámetros de consulta
	pedidoIDStr := c.Query("id")
	pedidoID, err := strconv.Atoi(pedidoIDStr)
	if err != nil || pedidoIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID de pedido inválido o no proporcionado",
		})
		return
	}

	// Obtener ID del usuario del contexto (establecido por AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "No se encontró información del usuario",
		})
		return
	}

	// Obtener el rol del usuario
	rol, exists := c.Get("rol")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "No se encontró información del rol",
		})
		return
	}

	// Obtener el pedido
	pedido := new(models.Pedido)
	err = h.db.NewSelect().
		Model(pedido).
		Where("id = ?", pedidoID).
		Scan(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Pedido no encontrado",
		})
		return
	}

	// Verificar que el usuario tenga permiso para ver este pedido
	// Si no es admin, solo puede ver sus propios pedidos
	if rol != "admin" && pedido.UsuarioId != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "No tienes permiso para ver este pedido",
		})
		return
	}

	// Obtener los detalles del pedido
	var detalles []models.DetallePedido
	err = h.db.NewSelect().
		Model(&detalles).
		Relation("Producto").
		Where("pedido_id = ?", pedidoID).
		Scan(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al obtener detalles del pedido: " + err.Error(),
		})
		return
	}

	// Crear respuesta solo con detalles
	detallesResponse := make([]DetallePedidoResponse, 0, len(detalles))
	for _, detalle := range detalles {
		nombreProducto := ""
		if detalle.Producto != nil {
			nombreProducto = detalle.Producto.Nombre
		}

		detallesResponse = append(detallesResponse, DetallePedidoResponse{
			ID:             detalle.ID,
			ProductoID:     detalle.ProductoID,
			Nombre:         nombreProducto,
			Cantidad:       detalle.Cantidad,
			PrecioUnitario: detalle.PrecioUnitario,
			PrecioTotal:    detalle.PrecioTotal,
		})
	}

	// Respuesta final solo con los detalles
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    detallesResponse,
	})
}

// estructura para recibir la solicitud de actualización de pedido por un admin
type UpdateOrderAdminRequest struct {
	Estado           string                   `json:"estado"`
	FechaEnvio       *time.Time               `json:"fecha_envio"`
	CiudadDestino    string                   `json:"ciudad_destino"`
	DireccionDestino string                   `json:"direccion_destino"`
	RutDestinatario  string                   `json:"rut_destinatario"`
	Company          string                   `json:"company"`
	TipoEnvio        string                   `json:"tipo_envio"`
	MetodoPago       string                   `json:"metodo_pago"`
	TipoDocumento    string                   `json:"tipo_documento"`
	Items            []UpdateOrderItemRequest `json:"items"`
}

// estructura para recibir la solicitud de actualización de productos por un cliente
type UpdateOrderClientRequest struct {
	Items []UpdateOrderItemRequest `json:"items" binding:"required,dive"`
}

// estructura para los items del pedido a actualizar
type UpdateOrderItemRequest struct {
	ID           *int `json:"id,omitempty"` // ID del detalle si ya existe, omitir si es nuevo
	ProductoID   int  `json:"producto_id" binding:"required"`
	Cantidad     int  `json:"cantidad" binding:"required,min=1"`
	EliminarItem bool `json:"eliminar_item,omitempty"` // Si es true, se eliminará el item (solo si ID no es nulo)
}

// UpdateOrderAdmin permite a un administrador actualizar cualquier pedido y sus detalles
func (h *OrderHandler) UpdateOrderAdmin(c *gin.Context) {
	// Verificar que el usuario sea admin
	rol, exists := c.Get("rol")
	if !exists || rol != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "No tienes permisos para realizar esta acción",
		})
		return
	}

	// Obtener ID del pedido de los parámetros de consulta
	pedidoIDStr := c.Query("id")
	pedidoID, err := strconv.Atoi(pedidoIDStr)
	if err != nil || pedidoIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID de pedido inválido o no proporcionado",
		})
		return
	}

	// Parsear la solicitud
	var req UpdateOrderAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Datos inválidos: " + err.Error(),
		})
		return
	}

	// Iniciar transacción
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al iniciar transacción",
		})
		return
	}
	defer tx.Rollback()

	// Obtener el pedido actual
	pedido := new(models.Pedido)
	err = tx.NewSelect().
		Model(pedido).
		Where("id = ?", pedidoID).
		Scan(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Pedido no encontrado",
		})
		return
	}

	// Preparar actualización del pedido
	update := tx.NewUpdate().Model(pedido).WherePK()
	fieldsToUpdate := []string{"updated_at"} // Siempre actualizamos updated_at
	pedido.UpdatedAt = time.Now()

	// Verificar si se actualizará el estado
	if req.Estado != "" {
		pedido.Estado = req.Estado
		fieldsToUpdate = append(fieldsToUpdate, "estado")
	}

	// Verificar si se actualizará la fecha de envío
	if req.FechaEnvio != nil {
		pedido.FechaEnvio = req.FechaEnvio
		fieldsToUpdate = append(fieldsToUpdate, "fecha_envio")
	}

	// Verificar si se actualizará la ciudad de destino
	if req.CiudadDestino != "" {
		pedido.CiudadDestino = req.CiudadDestino
		fieldsToUpdate = append(fieldsToUpdate, "ciudad_destino")
	}

	// Verificar si se actualizará la dirección de destino
	if req.DireccionDestino != "" {
		pedido.DireccionDestino = req.DireccionDestino
		fieldsToUpdate = append(fieldsToUpdate, "direccion_destino")
	}

	// Verificar si se actualizará el RUT del destinatario
	if req.RutDestinatario != "" {
		pedido.RutDestinatario = req.RutDestinatario
		fieldsToUpdate = append(fieldsToUpdate, "rut_destinatario")
	}

	// Verificar si se actualizará la compañía
	if req.Company != "" {
		pedido.Company = req.Company
		fieldsToUpdate = append(fieldsToUpdate, "company")
	}

	// Verificar si se actualizará la compañía
	// Nota: El campo Company puede ser vacío intencionalmente
	if c.Request.Method == "PUT" && c.ContentType() == "application/json" {
		// Verificamos si el campo está presente en el JSON
		bodyBytes, _ := c.GetRawData()
		var jsonData map[string]interface{}
		if json.Unmarshal(bodyBytes, &jsonData) == nil {
			if _, exists := jsonData["company"]; exists {
				pedido.Company = req.Company
				fieldsToUpdate = append(fieldsToUpdate, "company")
			}
		}
		// Restauramos el body para que esté disponible nuevamente
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Verificar si se actualizará el tipo de envío
	if req.TipoEnvio != "" {
		pedido.TipoEnvio = req.TipoEnvio
		fieldsToUpdate = append(fieldsToUpdate, "tipo_envio")
	}

	// Verificar si se actualizará el método de pago
	if req.MetodoPago != "" {
		pedido.MetodoPago = req.MetodoPago
		fieldsToUpdate = append(fieldsToUpdate, "metodo_pago")
	}

	// Verificar si se actualizará el tipo de documento
	if req.TipoDocumento != "" {
		pedido.TipoDocumento = req.TipoDocumento
		fieldsToUpdate = append(fieldsToUpdate, "tipo_documento")
	}

	// Actualizar el pedido en la base de datos solo con los campos especificados
	_, err = update.Column(fieldsToUpdate...).Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al actualizar pedido: " + err.Error(),
		})
		return
	}

	// Solo procesamos los items si se incluyen en la solicitud
	var totalNuevo = pedido.Total // Mantenemos el total actual si no hay cambios en los items

	if len(req.Items) > 0 {
		// Procesar los items del pedido
		// 1. Primero obtenemos los detalles actuales
		var detallesActuales []models.DetallePedido
		err = tx.NewSelect().
			Model(&detallesActuales).
			Where("pedido_id = ?", pedidoID).
			Scan(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Error al obtener detalles actuales: " + err.Error(),
			})
			return
		}

		// Mapear detalles actuales por ID para fácil acceso
		mapaDetallesActuales := make(map[int]*models.DetallePedido)
		for i := range detallesActuales {
			mapaDetallesActuales[detallesActuales[i].ID] = &detallesActuales[i]
		}

		// 2. Procesamos los items de la solicitud
		var detallesNuevos []*models.DetallePedido       // Para insertar
		var detallesActualizar []*models.DetallePedido   // Para actualizar
		var idsDetallesEliminar []int                    // IDs a eliminar
		var idsDetallesActualizados = make(map[int]bool) // Para seguimiento de actualizados

		totalNuevo = 0 // Reiniciar el total para recalcularlo

		for _, item := range req.Items {
			// Si es un item existente (viene con ID)
			if item.ID != nil {
				detalleID := *item.ID

				// Verificar si existe
				detalle, existe := mapaDetallesActuales[detalleID]
				if !existe {
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"error":   "Detalle con ID " + strconv.Itoa(detalleID) + " no existe en este pedido",
					})
					return
				}

				// Si se solicita eliminar
				if item.EliminarItem {
					idsDetallesEliminar = append(idsDetallesEliminar, detalleID)
					continue
				}

				// Obtener el producto actual
				producto := new(models.Producto)
				err := tx.NewSelect().Model(producto).Where("id = ?", item.ProductoID).Scan(c)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"error":   "Producto no encontrado: " + err.Error(),
					})
					return
				}

				// Calcular nuevo precio total
				precioTotal := producto.PrecioVenta * item.Cantidad

				// Actualizar el detalle
				detalle.ProductoID = item.ProductoID
				detalle.Cantidad = item.Cantidad
				detalle.PrecioUnitario = producto.PrecioVenta
				detalle.PrecioTotal = precioTotal
				detalle.UpdatedAt = time.Now()

				detallesActualizar = append(detallesActualizar, detalle)
				idsDetallesActualizados[detalleID] = true
				totalNuevo += precioTotal
			} else {
				// Es un item nuevo
				// Obtener producto
				producto := new(models.Producto)
				err := tx.NewSelect().Model(producto).Where("id = ?", item.ProductoID).Scan(c)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"error":   "Producto no encontrado: " + err.Error(),
					})
					return
				}

				// Calcular precio total
				precioTotal := producto.PrecioVenta * item.Cantidad

				// Crear nuevo detalle
				detalle := &models.DetallePedido{
					PedidoID:       pedidoID,
					ProductoID:     item.ProductoID,
					Cantidad:       item.Cantidad,
					PrecioUnitario: producto.PrecioVenta,
					PrecioTotal:    precioTotal,
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}

				detallesNuevos = append(detallesNuevos, detalle)
				totalNuevo += precioTotal
			}
		}

		// 3. Añadir los detalles que no fueron modificados al total
		for id, detalle := range mapaDetallesActuales {
			if !idsDetallesActualizados[id] && !contains(idsDetallesEliminar, id) {
				totalNuevo += detalle.PrecioTotal
			}
		}

		// 4. Ejecutar operaciones en la base de datos
		// 4.1 Eliminar detalles marcados para eliminación
		if len(idsDetallesEliminar) > 0 {
			_, err = tx.NewDelete().
				Model((*models.DetallePedido)(nil)).
				Where("id IN (?)", bun.In(idsDetallesEliminar)).
				Exec(c)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Error al eliminar detalles: " + err.Error(),
				})
				return
			}
		}

		// 4.2 Actualizar detalles existentes
		for _, detalle := range detallesActualizar {
			_, err = tx.NewUpdate().
				Model(detalle).
				WherePK().
				Exec(c)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Error al actualizar detalle: " + err.Error(),
				})
				return
			}
		}

		// 4.3 Insertar nuevos detalles
		if len(detallesNuevos) > 0 {
			_, err = tx.NewInsert().
				Model(&detallesNuevos).
				Exec(c)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Error al insertar nuevos detalles: " + err.Error(),
				})
				return
			}
		}

		// Actualizar el total del pedido si se modificaron los items
		pedido.Total = totalNuevo
		_, err = tx.NewUpdate().
			Model(pedido).
			Column("total").
			WherePK().
			Exec(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Error al actualizar total del pedido: " + err.Error(),
			})
			return
		}
	}

	// Confirmar transacción
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al confirmar transacción",
		})
		return
	}

	// Obtener el pedido actualizado para la respuesta
	pedidoActualizado := new(models.Pedido)
	err = h.db.NewSelect().
		Model(pedidoActualizado).
		Where("id = ?", pedidoID).
		Scan(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al obtener pedido actualizado",
		})
		return
	}

	// Crear respuesta
	respuesta := PedidoResponse{
		ID:               pedidoActualizado.ID,
		UsuarioId:        pedidoActualizado.UsuarioId,
		Total:            pedidoActualizado.Total,
		Estado:           pedidoActualizado.Estado,
		FechaEnvio:       pedidoActualizado.FechaEnvio,
		CiudadDestino:    pedidoActualizado.CiudadDestino,
		DireccionDestino: pedidoActualizado.DireccionDestino,
		RutDestinatario:  pedidoActualizado.RutDestinatario,
		Company:          pedidoActualizado.Company,
		TipoEnvio:        pedidoActualizado.TipoEnvio,
		MetodoPago:       pedidoActualizado.MetodoPago,
		TipoDocumento:    pedidoActualizado.TipoDocumento,
		CreatedAt:        pedidoActualizado.CreatedAt,
		UpdatedAt:        pedidoActualizado.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"mensaje": "Pedido actualizado correctamente",
		"pedido":  respuesta,
	})
}

// UpdateOrderClient permite a un cliente actualizar solo los productos de su propio pedido
func (h *OrderHandler) UpdateOrderClient(c *gin.Context) {
	// Obtener ID del usuario del contexto (establecido por AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "No se encontró información del usuario",
		})
		return
	}

	// Obtener ID del pedido de los parámetros de consulta
	pedidoIDStr := c.Query("id")
	pedidoID, err := strconv.Atoi(pedidoIDStr)
	if err != nil || pedidoIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID de pedido inválido o no proporcionado",
		})
		return
	}

	// Parsear la solicitud
	var req UpdateOrderClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Datos inválidos: " + err.Error(),
		})
		return
	}

	// Iniciar transacción
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al iniciar transacción",
		})
		return
	}
	defer tx.Rollback()

	// Obtener el pedido actual y verificar que pertenezca al usuario
	pedido := new(models.Pedido)
	err = tx.NewSelect().
		Model(pedido).
		Where("id = ?", pedidoID).
		Scan(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Pedido no encontrado",
		})
		return
	}

	// Verificar que el pedido pertenezca al usuario
	if pedido.UsuarioId != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "No tienes permiso para modificar este pedido",
		})
		return
	}

	// Verificar que el pedido esté en estado "pendiente"
	if pedido.Estado != "pendiente" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Solo se pueden modificar pedidos en estado pendiente",
		})
		return
	}

	// Procesar los items del pedido
	// 1. Primero obtenemos los detalles actuales
	var detallesActuales []models.DetallePedido
	err = tx.NewSelect().
		Model(&detallesActuales).
		Where("pedido_id = ?", pedidoID).
		Scan(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al obtener detalles actuales: " + err.Error(),
		})
		return
	}

	// Mapear detalles actuales por ID para fácil acceso
	mapaDetallesActuales := make(map[int]*models.DetallePedido)
	for i := range detallesActuales {
		mapaDetallesActuales[detallesActuales[i].ID] = &detallesActuales[i]
	}

	// 2. Procesamos los items de la solicitud
	var detallesNuevos []*models.DetallePedido       // Para insertar
	var detallesActualizar []*models.DetallePedido   // Para actualizar
	var idsDetallesEliminar []int                    // IDs a eliminar
	var idsDetallesActualizados = make(map[int]bool) // Para seguimiento de actualizados

	var totalNuevo int = 0 // Para recalcular el total del pedido

	for _, item := range req.Items {
		// Si es un item existente (viene con ID)
		if item.ID != nil {
			detalleID := *item.ID

			// Verificar si existe
			detalle, existe := mapaDetallesActuales[detalleID]
			if !existe {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "Detalle con ID " + strconv.Itoa(detalleID) + " no existe en este pedido",
				})
				return
			}

			// Si se solicita eliminar
			if item.EliminarItem {
				idsDetallesEliminar = append(idsDetallesEliminar, detalleID)
				continue
			}

			// Obtener el producto actual
			producto := new(models.Producto)
			err := tx.NewSelect().Model(producto).Where("id = ?", item.ProductoID).Scan(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "Producto no encontrado: " + err.Error(),
				})
				return
			}

			// Calcular nuevo precio total
			precioTotal := producto.PrecioVenta * item.Cantidad

			// Actualizar el detalle
			detalle.ProductoID = item.ProductoID
			detalle.Cantidad = item.Cantidad
			detalle.PrecioUnitario = producto.PrecioVenta
			detalle.PrecioTotal = precioTotal
			detalle.UpdatedAt = time.Now()

			detallesActualizar = append(detallesActualizar, detalle)
			idsDetallesActualizados[detalleID] = true
			totalNuevo += precioTotal
		} else {
			// Es un item nuevo
			// Obtener producto
			producto := new(models.Producto)
			err := tx.NewSelect().Model(producto).Where("id = ?", item.ProductoID).Scan(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "Producto no encontrado: " + err.Error(),
				})
				return
			}

			// Verificar que el producto esté disponible
			if !producto.Disponible {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "El producto " + producto.Nombre + " no está disponible",
				})
				return
			}

			// Calcular precio total
			precioTotal := producto.PrecioVenta * item.Cantidad

			// Crear nuevo detalle
			detalle := &models.DetallePedido{
				PedidoID:       pedidoID,
				ProductoID:     item.ProductoID,
				Cantidad:       item.Cantidad,
				PrecioUnitario: producto.PrecioVenta,
				PrecioTotal:    precioTotal,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			detallesNuevos = append(detallesNuevos, detalle)
			totalNuevo += precioTotal
		}
	}

	// 3. Añadir los detalles que no fueron modificados al total
	for id, detalle := range mapaDetallesActuales {
		if !idsDetallesActualizados[id] && !contains(idsDetallesEliminar, id) {
			totalNuevo += detalle.PrecioTotal
		}
	}

	// 4. Ejecutar operaciones en la base de datos
	// 4.1 Eliminar detalles marcados para eliminación
	if len(idsDetallesEliminar) > 0 {
		_, err = tx.NewDelete().
			Model((*models.DetallePedido)(nil)).
			Where("id IN (?)", bun.In(idsDetallesEliminar)).
			Exec(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Error al eliminar detalles: " + err.Error(),
			})
			return
		}
	}

	// 4.2 Actualizar detalles existentes
	for _, detalle := range detallesActualizar {
		_, err = tx.NewUpdate().
			Model(detalle).
			WherePK().
			Exec(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Error al actualizar detalle: " + err.Error(),
			})
			return
		}
	}

	// 4.3 Insertar nuevos detalles
	if len(detallesNuevos) > 0 {
		_, err = tx.NewInsert().
			Model(&detallesNuevos).
			Exec(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Error al insertar nuevos detalles: " + err.Error(),
			})
			return
		}
	}

	// 5. Actualizar el total del pedido
	pedido.Total = totalNuevo
	pedido.UpdatedAt = time.Now()
	_, err = tx.NewUpdate().
		Model(pedido).
		Column("total").
		Column("updated_at").
		WherePK().
		Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al actualizar total del pedido: " + err.Error(),
		})
		return
	}

	// Confirmar transacción
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al confirmar transacción",
		})
		return
	}

	// Obtener los detalles actualizados para la respuesta
	var detallesActualizados []models.DetallePedido
	err = h.db.NewSelect().
		Model(&detallesActualizados).
		Relation("Producto").
		Where("pedido_id = ?", pedidoID).
		Scan(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al obtener detalles actualizados",
		})
		return
	}

	// Convertir a respuesta
	detallesResponse := make([]DetallePedidoResponse, 0, len(detallesActualizados))
	for _, detalle := range detallesActualizados {
		nombreProducto := ""
		if detalle.Producto != nil {
			nombreProducto = detalle.Producto.Nombre
		}

		detallesResponse = append(detallesResponse, DetallePedidoResponse{
			ID:             detalle.ID,
			ProductoID:     detalle.ProductoID,
			Nombre:         nombreProducto,
			Cantidad:       detalle.Cantidad,
			PrecioUnitario: detalle.PrecioUnitario,
			PrecioTotal:    detalle.PrecioTotal,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"mensaje": "Productos del pedido actualizados correctamente",
		"total":   totalNuevo,
		"items":   detallesResponse,
	})
}

// Función utilitaria para verificar si un slice contiene un valor
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
