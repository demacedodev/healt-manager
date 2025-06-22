package http

import (
	"callers-go/application"
	"callers-go/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	ALL = "all"
)

type (
	Handlers struct {
		app application.Callers
	}

	Presentation interface {
		GetCallers(c *gin.Context)
		CreateCallers(c *gin.Context)
	}
)

func NewHandlers(app application.Callers) Handlers {
	return Handlers{app: app}
}

func (h *Handlers) GetCallers(c *gin.Context) {
	zone := c.Query("zone")
	if len(zone) == 0 {
		zone = ALL
	}

	callers, err := h.app.GetDevices(c.Request.Context(), zone)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
		return
	}

	c.JSON(200, callers)
}

func (h *Handlers) CreateCallers(c *gin.Context) {
	var devices []domain.Device
	if err := c.Bind(&devices); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
	}

	if err := h.app.CreateDevice(c.Request.Context(), devices); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
	}

	c.JSON(200, gin.H{"devices": devices})
}
