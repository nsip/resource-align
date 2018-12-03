package main

import (
	"github.com/labstack/echo"
	cu "github.com/nsip/curriculum-align"
	re "github.com/nsip/resource-align"
	"log"
	"net/http"
)

func main() {
	cu.Init()
	re.Init("1576")
	log.Println("Editor: localhost:1576")
	e := echo.New()
	e.GET("/align", re.Align)
	e.GET("/curricalign", cu.Align)
	e.GET("/index", func(c echo.Context) error {
		query := c.QueryParam("search")
		ret, err := cu.Search(query)
		if err != nil {
			return err
		} else {
			return c.String(http.StatusOK, string(ret))
		}
	})
	log.Println("Editor: localhost:1576")
	e.Logger.Fatal(e.Start(":1576"))
}
