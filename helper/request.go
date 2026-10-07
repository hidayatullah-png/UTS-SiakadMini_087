package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
)

func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// ParseListQuery - page/per_page sesuai nama parameter di spek endpoint 3
// (BUKAN "page"/"limit" seperti proyek students lama - perhatikan
// per_page pakai underscore, persis penulisan di PDF).
func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:     c.QueryInt("page", 1),
		PerPage:  c.QueryInt("per_page", 10),
		Search:   strings.TrimSpace(c.Query("search")),
		Prodi:    strings.TrimSpace(c.Query("prodi")),
		Angkatan: c.QueryInt("angkatan", 0),
		Sort:     c.Query("sort", "nama"),
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 {
		q.PerPage = 10
	}
	if q.PerPage > 50 { // batas maks sesuai spek endpoint 3
		q.PerPage = 50
	}
	return q
}