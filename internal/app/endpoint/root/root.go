package root

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Endpoint struct {
}

func New() *Endpoint {
	return &Endpoint{}
}
func (root *Endpoint) RootHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "online")
}
