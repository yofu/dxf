package main

import (
	"fmt"
	"log"

	"github.com/yofu/dxf"
	"github.com/yofu/dxf/entity"
)

func main() {
	dr, err := dxf.Open("point.dxf")
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range dr.Entities() {
		if p, ok := e.(*entity.Point); ok {
			fmt.Println(p.Coord)
		}
	}
}
