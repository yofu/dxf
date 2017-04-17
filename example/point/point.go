package main

import (
	"fmt"
	"log"

	"github.com/yofu/dxf"
	"github.com/yofu/dxf/entity"
)

func main() {
	dxf, err := dxf.Open("point.dxf")
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range dxf.Entities() {
		if p, ok := e.(*entity.Point); ok {
			fmt.Println(p.Coord)
		}
	}
}
