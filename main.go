package main

import (
	"encoding/json"
	"strconv"
	"syscall/js"

	"github.com/akhenakh/oureadb/index/geodata"
	"github.com/akhenakh/oureadb/s2tools"
	"github.com/golang/geo/s2"
	"github.com/twpayne/go-geom/encoding/geojson"
)

var document js.Value

func init() {
	document = js.Global().Get("document")
}

func getCoverParams() (minLevel, maxLevel, maxCells int) {
	minS := document.Call("getElementById", "minRange").Get("value").String()
	minLevel, err := strconv.Atoi(minS)
	if err != nil {
		println(err.Error())
		return
	}

	maxS := document.Call("getElementById", "maxRange").Get("value").String()
	maxLevel, err = strconv.Atoi(maxS)
	if err != nil {
		println(err.Error())
		return
	}

	maxCS := document.Call("getElementById", "maxCellsRange").Get("value").String()
	maxLevel, err = strconv.Atoi(maxCS)
	if err != nil {
		println(err.Error())
		return
	}

	return minLevel, maxLevel, maxCells
}
func geoFeaturesJSONToCells(i []js.Value) {
	var fc geojson.FeatureCollection
	b := js.ValueOf(i[0]).String()
	err := json.Unmarshal([]byte(b), &fc)
	if err != nil {
		println(err.Error())
		return
	}
	var res s2.CellUnion
	for _, f := range fc.Features {
		cu := computeFeatureCells(f)
		res = append(res, cu...)
	}

	jsonb := s2tools.CellUnionToGeoJSON(res)
	js.Global().Set("data", string(jsonb))
}

func computeFeatureCells(f *geojson.Feature) s2.CellUnion {
	gd := &geodata.GeoData{}
	err := geodata.GeoJSONFeatureToGeoData(f, gd)
	if err != nil {
		println(err)
		return nil
	}

	minLevel, maxLevel, maxCells := getCoverParams()
	coverer := &s2.RegionCoverer{MinLevel: minLevel, MaxLevel: maxLevel, MaxCells: maxCells}

	cu, err := geodata.GeoDataToCellUnion(gd, coverer)
	if err != nil {
		println(err)
		return nil
	}
	return cu
}

func geoJSONToCells(i []js.Value) {
	var f geojson.Feature
	b := js.ValueOf(i[0]).String()
	err := json.Unmarshal([]byte(b), &f)
	if err != nil {
		println(err.Error())
		return
	}
	cu := computeFeatureCells(&f)
	jsonb := s2tools.CellUnionToGeoJSON(cu)
	js.Global().Set("data", string(jsonb))
}

func drawCells(i []js.Value) {
	var cells s2.CellUnion
	for _, cs := range i {
		c := s2.CellIDFromToken(js.ValueOf(cs).String())
		cells = append(cells, c)
	}
	b := s2tools.CellUnionToGeoJSON(cells)
	js.Global().Set("data", string(b))
}

func registerCallbacks() {
	js.Global().Set("drawcells", js.NewCallback(drawCells))
	js.Global().Set("geocell", js.NewCallback(geoJSONToCells))
	js.Global().Set("geofeaturescell", js.NewCallback(geoFeaturesJSONToCells))
}

func main() {
	c := make(chan struct{}, 0)
	println("Wasm ready")
	registerCallbacks()
	<-c
}
