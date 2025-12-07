package http

import (
	"callers-go/application"
	"callers-go/domain"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	ALL = "ALL"
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

	output := make([]map[string]any, len(callers))
	for i, caller := range callers {
		data := make(map[string]any)
		data["id"] = caller.DeviceId
		data["name"] = caller.DeviceName
		data["location"] = fmt.Sprintf("🧭 %s 🏡 %s 🛏️ %s", caller.Location.Zone, caller.Location.Room, caller.Location.Bed)
		data["isCalling"] = caller.DeviceStatus
		data["startedAt"] = time.Unix(0, 0).UTC().UnixMilli()

		if caller.DeviceStatus {
			data["startedAt"] = time.Now().UnixMilli()
		}

		output[i] = data
	}

	c.JSON(200, output)
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
