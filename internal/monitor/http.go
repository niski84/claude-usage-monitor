package monitor

import (
	"bytes"
	"io/fs"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/niski84/claude-usage-monitor/internal/monitor/views"
	certweb "github.com/niski84/claude-usage-monitor/web"
)

// NewEcho creates and configures the Echo server with all routes.
func NewEcho(svc *Service) (*echo.Echo, error) {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	staticRoot, err := fs.Sub(certweb.FS, "monitor/static")
	if err != nil {
		return nil, err
	}
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", http.FileServer(http.FS(staticRoot)))))

	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"service": "claude-usage-monitor",
			"status":  "ok",
		})
	})

	e.GET("/api/stats", func(c echo.Context) error {
		return c.JSON(http.StatusOK, svc.Stats())
	})

	e.POST("/api/refresh", func(c echo.Context) error {
		go svc.refresh()
		return c.JSON(http.StatusOK, map[string]bool{"success": true})
	})

	e.POST("/api/refresh-limits", func(c echo.Context) error {
		go svc.RefreshRateLimits()
		return c.JSON(http.StatusOK, map[string]string{"status": "fetching"})
	})

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/dashboard")
	})

	e.GET("/dashboard", func(c echo.Context) error {
		return renderHTML(c, http.StatusOK, views.Dashboard(svc.Stats()))
	})

	return e, nil
}

// renderHTML renders a templ component to the response writer.
func renderHTML(c echo.Context, code int, comp templ.Component) error {
	var buf bytes.Buffer
	if err := comp.Render(c.Request().Context(), &buf); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.HTMLBlob(code, buf.Bytes())
}
