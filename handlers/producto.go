package handlers

import (
	"cotizador-productos-eml/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type ProductoHandler struct {
	db *bun.DB
}

func NewProductoHandler(db *bun.DB) *ProductoHandler {
	return &ProductoHandler{db: db}
}

func (h *ProductoHandler) CreateProducto(c *gin.Context) {

	var input struct {
		Nombre             string `json:"nombre"`
		PrecioVenta        int    `json:"precio_venta"`
		PrecioCompra       int    `json:"precio_compra"`
		UltimaVezIngresado string `json:"ultima_vez_ingresado"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Parsear la fecha
	fecha, err := time.Parse("02/01/2006", input.UltimaVezIngresado)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha inválido"})
		return
	}
	nuevoProducto := models.Producto{
		Nombre:             input.Nombre,
		PrecioVenta:        input.PrecioVenta,
		PrecioCompra:       input.PrecioCompra,
		UltimaVezIngresado: fecha,
		Disponible:         true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Insertar en la base de datos
	_, err = h.db.NewInsert().Model(&nuevoProducto).Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"producto": nuevoProducto,
	})

}

func (h *ProductoHandler) GetAllProductos(c *gin.Context) {
	var productos []models.Producto

	// Consulta para obtener todos los productos
	err := h.db.NewSelect().Model(&productos).Order("nombre ASC").Scan(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron obtener los productos"})
		return
	}

	// Formatear las fechas de los productos antes de enviarlas
	var productosConFechasFormateadas []map[string]interface{}
	for _, producto := range productos {
		productoFormateado := map[string]interface{}{
			"id":                   producto.ID,
			"nombre":               producto.Nombre,
			"precio_venta":         producto.PrecioVenta,
			"precio_compra":        producto.PrecioCompra,
			"disponible":           producto.Disponible,
			"ultima_vez_ingresado": producto.UltimaVezIngresado.Format("02/01/2006"),
			"created_at":           producto.CreatedAt.Format("02/01/2006"),
			"updated_at":           producto.UpdatedAt.Format("02/01/2006"),
		}
		productosConFechasFormateadas = append(productosConFechasFormateadas, productoFormateado)
	}

	// Devolver los productos con fechas formateadas
	c.JSON(http.StatusOK, gin.H{
		"data": productosConFechasFormateadas,
	})
}

func (h *ProductoHandler) DeleteProducto(c *gin.Context) {
	id := c.Param("id")

	// Convertir el ID a entero
	productID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	// Verificar si el producto existe en la base de datos
	var producto models.Producto
	err = h.db.NewSelect().Model(&producto).Where("id = ?", productID).Scan(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Producto no encontrado"})
		return
	}

	// Eliminar el producto de la base de datos
	_, err = h.db.NewDelete().Model(&models.Producto{}).Where("id = ?", productID).Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo eliminar el producto"})
		return
	}

	// Devolver una respuesta exitosa
	c.JSON(http.StatusOK, gin.H{
		"message": "Producto eliminado correctamente",
	})

}

func (h *ProductoHandler) UpdateProducto(c *gin.Context) {
	productID := c.Param("id") // El ID que se pasa en la URL

	// Convertir el ID de la URL a un entero
	productIDInt, err := strconv.Atoi(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Buscar el producto en la base de datos
	var producto models.Producto
	err = h.db.NewSelect().Model(&producto).Where("id = ?", productIDInt).Scan(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Producto no encontrado"})
		return
	}

	// Estructura para capturar los datos que se envían en el body
	var input struct {
		Nombre             string `json:"nombre" binding:"required"`
		PrecioVenta        int    `json:"precio_venta" binding:"required"`
		PrecioCompra       int    `json:"precio_compra" binding:"required"`
		UltimaVezIngresado string `json:"ultima_vez_ingresado" binding:"required"`
		Disponible         string `json:"disponible" binding:"required"`
	}

	// Verificar que todos los datos requeridos estén presentes
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Transformar el campo "Disponible" de string a booleano
	var disponibleBool bool
	if input.Disponible == "true" {
		disponibleBool = true
	} else if input.Disponible == "false" {
		disponibleBool = false
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El campo 'disponible' debe ser 'true' o 'false'"})
		return
	}

	// Actualizar los campos del producto con los datos enviados
	producto.Nombre = input.Nombre
	producto.PrecioVenta = input.PrecioVenta
	producto.PrecioCompra = input.PrecioCompra
	fecha, err := time.Parse("02/01/2006", input.UltimaVezIngresado)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha inválido"})
		return
	}
	producto.UltimaVezIngresado = fecha
	producto.Disponible = disponibleBool
	producto.UpdatedAt = time.Now() // Fecha de actualización

	// Actualizar el producto en la base de datos
	_, err = h.db.NewUpdate().Model(&producto).Where("id = ?", productIDInt).Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Responder con el producto actualizado
	c.JSON(http.StatusOK, gin.H{
		"data": producto,
	})
}

func (h *ProductoHandler) GetProductoById(c *gin.Context) {
	id := c.Param("id")

	productIDInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var producto models.Producto
	err = h.db.NewSelect().Model(&producto).Where("id = ?", productIDInt).Scan(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Producto no encontrado"})
		return
	}

	// Formatear las fechas del producto
	productoFormateado := map[string]interface{}{
		"id":                   producto.ID,
		"nombre":               producto.Nombre,
		"precio_venta":         producto.PrecioVenta,
		"precio_compra":        producto.PrecioCompra,
		"disponible":           producto.Disponible,
		"ultima_vez_ingresado": producto.UltimaVezIngresado.Format("02/01/2006"),
		"created_at":           producto.CreatedAt.Format("02/01/2006"),
		"updated_at":           producto.UpdatedAt.Format("02/01/2006"),
	}

	c.JSON(http.StatusOK, gin.H{
		"data": productoFormateado,
	})
}
