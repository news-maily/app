package middleware

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/news-maily/app/utils/pagination"
)

// Paginate is a middleware that populates the pagination object and sets it to the context.
// If the pagination params are not valid the request is aborted with a 400 bad request error.
func Paginate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var p = new(pagination.Pagination)

		p.Page = 0
		p.PerPage = pagination.DefaultPerPage
		p.Total = math.MaxUint64
		p.Collection = make([]interface{}, 0)

		if len(c.Query("per_page")) > 0 {
			perpage, err := strconv.ParseUint(c.Query("per_page"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "per_page field must be an integer."})
				c.Abort()
				return
			}

			p.PerPage = uint(perpage)

			//Lock on 100 if the user requests more than 100 items per page
			if p.PerPage > 100 {
				p.PerPage = 100
			}
		}
		if len(c.Query("page")) > 0 {
			page, err := strconv.ParseUint(c.Query("page"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "page field must be an integer."})
				c.Abort()
				return
			}
			p.Page = uint(page)
			p.Offset = uint(page * uint64(p.PerPage))
		}

		c.Set("pagination", p)
		c.Next()
	}
}
