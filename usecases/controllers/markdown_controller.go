package controller

import (
	"github.com/labstack/echo/v4"
	router "markdown-enricher/infrastructure/web"
	"markdown-enricher/usecases/interactors"
	"net/http"
	"strings"
)

type MarkdownController struct {
	enricherInteractor *interactors.EnricherInteractor
}

func MakeMarkdownController(enricherInteractor *interactors.EnricherInteractor) router.Controller {
	return &MarkdownController{
		enricherInteractor: enricherInteractor,
	}
}

func (c *MarkdownController) RegisterRoutes(e *echo.Group) {
	group := e.Group("/v1/markdown")

	group.GET("/enrich", c.enrichHandler)
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
