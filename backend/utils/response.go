package utils

import (
	"net/http"
	"ticketek-backend/domain"

	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": data})
}

func ErrorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"success": false, "error": message})
}

// HandleServiceError traduce un error devuelto por la capa de servicios al
// código de estado HTTP correspondiente. Los errores de "no encontrado" dan
// 404, los de permisos 403, y el resto se consideran errores del cliente (400).
func HandleServiceError(c *gin.Context, err error) {
	switch err {
	case domain.ErrEventoNoEncontrado, domain.ErrEntradaNoEncontrada, domain.ErrUsuarioNoEncontrado:
		ErrorResponse(c, http.StatusNotFound, err.Error())
	case domain.ErrSinPermiso:
		ErrorResponse(c, http.StatusForbidden, err.Error())
	default:
		ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
}
