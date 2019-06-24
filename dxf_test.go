package dxf

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/gdey/dxf/entity"
)

func TestFromStringData(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/point.dxf")
	if err != nil {
		log.Fatal("could not open file: ", err)
	}
	dr, err := FromStringData(string(data))
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range dr.Entities() {
		if p, ok := e.(*entity.Point); ok {
			fmt.Println(p.Coord)
		}
	}
}
func TestFromFile(t *testing.T) {
	dr, err := FromFile("testdata/point.dxf")
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range dr.Entities() {
		if p, ok := e.(*entity.Point); ok {
			fmt.Println(p.Coord)
		}
	}
}
