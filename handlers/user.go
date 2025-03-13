package handlers

import (
	"cotizador-productos-eml/models"
	"net/http"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	db *bun.DB
}

func NewUserHandler(db *bun.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "No Autorizado"})
		return
	}
	var usuario models.Usuario
	err := h.db.NewSelect().
		Model(&usuario).
		Where("id = ?", userID).
		Scan(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo usuario"})
		return
	}

	// Responder con los datos del usuario (sin incluir la contraseña)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":       usuario.ID,
			"email":    usuario.Email,
			"nombre":   usuario.Nombre,
			"apellido": usuario.Apellido,
			"celular":  usuario.Celular,
			"ciudad":   usuario.Ciudad,
		},
	})

}

func (h *UserHandler) CreateUser(c *gin.Context) {
	userRole, exists := c.Get("rol")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Acceso denegado: Se requieren privilegios de administrador",
		})
		c.Abort()
		return
	}
	var input struct {
		Nombre   string `json:"nombre"`
		Apellido string `json:"apellido"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Ciudad   string `json:"ciudad"`
		Celular  string `json:"celular"`
		Rol      string `json:"rol"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"error":  "Datos de registro invalidos: " + err.Error(),
		})
		return

	}
	// Verificar si ya existe un usuario con el mismo email
	existingUser := new(models.Usuario)
	err := h.db.NewSelect().Model(existingUser).Where("email = ?", input.Email).Scan(c)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "El email ya está registrado",
		})
		return
	}

	// Encriptar la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al encriptar la contraseña",
		})
		return
	}

	// Crear nuevo usuario con el email ya verificado
	newUser := models.Usuario{
		Nombre:     input.Nombre,
		Apellido:   input.Apellido,
		Email:      input.Email,
		Password:   string(hashedPassword),
		Ciudad:     input.Ciudad,
		Celular:    input.Celular,
		Rol:        input.Rol,
		Verificado: true, // Email ya verificado
	}

	// Insertar usuario en la base de datos
	_, err = h.db.NewInsert().Model(&newUser).Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al crear el usuario",
		})
		return
	}

	// Responder con éxito
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Usuario creado exitosamente",
	})

}
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Verificar si el usuario autenticado tiene rol "admin"
	userRole, exists := c.Get("rol")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Acceso denegado: se requiere rol de admin",
		})
		c.Abort()
		return
	}

	// Obtener el ID del usuario a modificar desde los parámetros
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Se requiere un ID de usuario",
		})
		return
	}

	// Definir la estructura de entrada
	var input struct {
		Nombre   *string `json:"nombre"`
		Apellido *string `json:"apellido"`
		Email    *string `json:"email"`
		Password *string `json:"password"`
		Ciudad   *string `json:"ciudad"`
		Celular  *string `json:"celular"`
		Rol      *string `json:"rol"` // Comentado temporalmente
	}

	// Validar el JSON de entrada
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Datos de actualización inválidos: " + err.Error(),
		})
		return
	}

	// Buscar al usuario en la base de datos
	user := new(models.Usuario)
	err := h.db.NewSelect().Model(user).Where("id = ?", userID).Scan(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Usuario no encontrado",
		})
		return
	}

	// Actualizar solo los campos que se enviaron en la solicitud
	if input.Nombre != nil {
		user.Nombre = *input.Nombre
	}
	if input.Apellido != nil {
		user.Apellido = *input.Apellido
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Error al encriptar la contraseña",
			})
			return
		}
		user.Password = string(hashedPassword)
	}
	if input.Ciudad != nil {
		user.Ciudad = *input.Ciudad
	}
	if input.Celular != nil {
		user.Celular = *input.Celular
	}

	if input.Rol != nil {
		user.Rol = *input.Rol
	}

	// Guardar cambios en la base de datos
	_, err = h.db.NewUpdate().Model(user).Where("id = ?", userID).Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al actualizar el usuario",
		})
		return
	}

	// Responder con éxito
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Usuario actualizado correctamente",
	})
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	userRole, exists := c.Get("rol")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Acceso denegado: Se requieren privilegios de administrador",
		})
		c.Abort()
		return
	}
	var usuarios []struct {
		ID       uint   `json:"id"`
		Email    string `json:"email"`
		Nombre   string `json:"nombre"`
		Apellido string `json:"apellido"`
	}
	err := h.db.NewSelect().
		Model((*models.Usuario)(nil)).
		Column("id", "email", "nombre", "apellido").
		Scan(c, &usuarios)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al obtener los usuarios",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    usuarios,
	})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	userRole, exists := c.Get("rol")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Acceso denegado: Se requieren privilegios de administrador",
		})
		c.Abort()
		return
	}
	userID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Se requiere un ID de usuario",
		})
		return
	}
	// Convertir userID a uint
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID de usuario inválido",
		})
		return
	}
	var usuario struct {
		ID         uint   `json:"id"`
		Email      string `json:"email"`
		Nombre     string `json:"nombre"`
		Apellido   string `json:"apellido"`
		Ciudad     string `json:"ciudad"`
		Celular    string `json:"celular"`
		Rol        string `json:"rol"`
		Verificado bool   `json:"verificado"`
	}
	err = h.db.NewSelect().
		Model((*models.Usuario)(nil)).
		Column("id", "email", "nombre", "apellido", "ciudad", "celular", "rol", "verificado").
		Where("id = ?", id).
		Scan(c, &usuario)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al obtener el usuario: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    usuario,
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userRole, exists := c.Get("rol")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Acceso denegado: Se requieren privilegios de administrador",
		})
		c.Abort()
		return
	}
	userID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Se requiere un ID de usuario",
		})
		return
	}
	// Convertir userID a uint
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID de usuario inválido",
		})
		return
	}

	_, err = h.db.NewDelete().Model((*models.Usuario)(nil)).Where("id = ?", id).Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al eliminar el usuario: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Usuario eliminado correctamente",
	})
}
