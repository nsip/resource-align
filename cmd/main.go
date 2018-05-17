package main

import (
	"log"

	"github.com/labstack/echo"
	cu "github.com/nsip/curriculum-align"
	re "github.com/nsip/resource-align"
)

func main() {
	cu.Init()
	re.Init()
	log.Println("Editor: localhost:1576")
	e := echo.New()
	e.GET("/align", re.Align)
	e.GET("/curricalign", cu.Align)
	log.Println("Editor: localhost:1576")
	e.Logger.Fatal(e.Start(":1576"))
}
