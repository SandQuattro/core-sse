package headerchecker

import (
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go/log"
	"net/http"
	"sse-demo-core/internal/app/interfaces/services"
)

const AUTHORIZATION = "Authorization"

func HeaderCheck(service services.JwtService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			span := jaegertracing.CreateChildSpan(ctx, "header checker middleware")
			defer span.Finish()

			logger := logdoc.GetLogger()
			logger.Debug(">> Header check Middleware started...")

			token := ctx.Request().Header.Get(AUTHORIZATION)
			if token == "" {
				cookie, err := ctx.Cookie("sse_demoToken")
				if cookie != nil || err == nil {
					token = cookie.Value
				} else {
					span.LogFields(log.String("auth cookie", "not available"))
				}
			}
			if token != "" {
				claims, isValid, err := service.ValidateToken(token)
				if !isValid && err != nil {
					logger.Errorf("token validation error: %s", err.Error())
					return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
				}

				// Кладем нужные нам данные в контекст и используем далее в хендлерах
				ctx.Set("claims", claims)

				err = next(ctx)
				if err != nil {
					logger.Error("error executing handler from mv")
					return err
				}

				logger.Debug("<< Header check Middleware done")
				return nil
			}
			return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
		}
	}
}
