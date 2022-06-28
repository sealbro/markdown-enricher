package controller

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	router "markdown-enricher/infrastructure/web"
	websocket2 "markdown-enricher/infrastructure/websocket"
	"markdown-enricher/usecases/interactors"
	"net/http"
	"strings"
)

type MarkdownController struct {
	enricherInteractor *interactors.EnricherInteractor
	ws                 *websocket2.WebSocketProvider
}

func MakeMarkdownController(enricherInteractor *interactors.EnricherInteractor, provider *websocket2.WebSocketProvider) router.Controller {
	return &MarkdownController{
		enricherInteractor: enricherInteractor,
		ws:                 provider,
	}
}

func (c *MarkdownController) RegisterRoutes(e *echo.Group) {
	group := e.Group("/v1/markdown")

	group.GET("/enrich", c.enrichHandler)
	group.GET("/status", c.status)
}

var (
	upgrader = websocket.Upgrader{}
)

func (c *MarkdownController) status(e echo.Context) error {
	ws, err := upgrader.Upgrade(e.Response(), e.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	client := c.ws.AddClient()

	go func() {
		for {
			// Read
			_, _, err := ws.ReadMessage()
			if err != nil {
				if !errors.Is(err, websocket.ErrCloseSent) {
					e.Logger().Error(err)
				}

				client.Close()
				break
			}
		}
	}()

	for {
		// Write
		bytes := client.Get()
		err := ws.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			if !errors.Is(err, websocket.ErrCloseSent) {
				e.Logger().Error(err)
			}
			break
		}
	}

	return nil
}

// Get enriched markdown elements
// @Tags state
// @Accept json
// @Produce json
// @Description Get enriched markdown elements
// @Param md_file_url query string true "md file url (https://raw.githubusercontent.com/avelino/awesome-go/main/README.md)"
// @Success 200 {object} model.MarkdownEnriched "markdown enriched"
// @Router /v1/markdown/enrich [get]
func (c *MarkdownController) enrichHandler(e echo.Context) error {
	ctx := e.Request().Context()
	mdFileUrl := strings.TrimSpace(e.QueryParam("md_file_url"))

	markdown, err := c.enricherInteractor.Markdown(ctx, mdFileUrl)
	if err != nil {
		return err
	}

	return e.JSON(http.StatusOK, markdown)
}
