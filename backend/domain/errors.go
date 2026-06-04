package domain

import "errors"

// Errores de dominio reutilizables. Los servicios los devuelven y los
// controladores los traducen al código de estado HTTP correcto:
//   - NoEncontrado  -> 404
//   - SinPermiso    -> 403
// El resto de los errores de negocio se tratan como 400 (Bad Request).
var (
	ErrEventoNoEncontrado  = errors.New("evento no encontrado")
	ErrEntradaNoEncontrada = errors.New("entrada no encontrada")
	ErrUsuarioNoEncontrado = errors.New("usuario no encontrado")
	ErrSinPermiso          = errors.New("no tenés permiso para realizar esta acción")
)
