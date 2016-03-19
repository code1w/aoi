package main

import (
	"fmt"
	"github.com/llgcode/draw2d/draw2dgl"
	//"github.com/llgcode/draw2d/draw2dkit"
	"image/color"
	//"math/rand"
	//"os"
	"runtime"
	//"strconv"
	"unsafe"
)

/*
#cgo CFLAGS: -I ../
#include "stdlib.h"
#include "../navi.h"
*/
import "C"

type Navi struct {
	xmap    *C.struct_inavimap
	mapdesc C.struct_inavimapdesc
	pos     C.struct_ipos
	size    C.struct_isize

	divids C.struct_isize
	cindex int
}

func (a *Navi) Free() {
	C.inavimapdescfreeresource(&a.mapdesc)
	C.inavimapfree(a.xmap)
}

func (a *Navi) Init(width, height float64) {
	factor := 1.0 * 3 / 4
	sizeh := height * factor
	sizew := width * factor
	offsetx := (width - sizew) / 2
	offsety := (height - sizeh) / 2

	// init navi memory system
	C.inavi_mm_init()

	// 从文件读取配置信息
	navimappath := C.CString("./navi.map")
	n := C.inavimapdescreadfromtextfile(&a.mapdesc, navimappath)
	fmt.Println("load map desc result:", n, "mapdesc:", a.mapdesc)
	C.free(unsafe.Pointer(navimappath))

	// 地图居中，偏移量
	a.pos = C.struct_ipos{x: C.ireal(offsetx), y: C.ireal(offsety)}
	a.size = C.struct_isize{w: C.ireal(sizew), h: C.ireal(sizeh)}
	a.divids = C.struct_isize{
		w: a.size.w / (C.ireal)(a.mapdesc.header.width),
		h: a.size.h / (C.ireal)(a.mapdesc.header.height)}
	fmt.Println("values:", width, height, a.pos, a.size, a.divids)

	a.xmap = C.inavimapmake(8)

	// 加载导航图
	C.inavimaploadfromdesc(a.xmap, &a.mapdesc)

	runtime.SetFinalizer(a, (*Navi).Free)
}

func (a *Navi) DrawMap(gc *draw2dgl.GraphicContext) {
	// 先绘制地图本身
	DrawGridC(gc, &a.pos, &a.size, &a.divids)

	// 搞一条线，清晰一点
	DrawLineC(gc, &a.pos, &a.size)
}

func (a *Navi) DrawMapDesc(gc *draw2dgl.GraphicContext) {
}

func (a *Navi) DrawPolygon3dC(gc *draw2dgl.GraphicContext, poly *C.struct_ipolygon3d) {
	n := C.islicelen(poly.pos)
	fmt.Println("DrawPolygon3d", "num", n)
	if n <= 0 {
		return
	}
	fivecolors := []color.RGBA{
		color.RGBA{215, 0, 0, 255},
		color.RGBA{0, 0, 215, 255},
		color.RGBA{215, 215, 0, 255},
		color.RGBA{0, 215, 215, 255},
	}
	//gc.Save()
	//gc.SetLineWidth(4)
	gc.BeginPath()
	p0 := a.TranslatePos((*C.struct_ipos3)(C.isliceat(poly.pos, 0)))
	DrawMoveToC(gc, &p0)
	fmt.Println("DrawPolygon3d", " [", 0, "]", p0)
	i := 1
	for ; i < int(n); i++ {
		p := a.TranslatePos((*C.struct_ipos3)(C.isliceat(poly.pos, C.int(i))))
		DrawLineToC(gc, &p)
		fmt.Println("DrawPolygon3d", " [", i, "]", p)
	}
	//gc.SetStrokeColor(fivecolors[i%5])
	//DrawLineToC(gc, &p0)
	//DrawLineEndC(gc)
	gc.Close()
	gc.SetFillColor(fivecolors[a.cindex%4])
	gc.Fill()

	a.cindex++
}

func (a *Navi) DrawLineC(gc *draw2dgl.GraphicContext, start, end *C.struct_ipos3) {
	p0 := a.TranslatePos(start)
	p1 := a.TranslatePos(end)

	DrawMoveToC(gc, &p0)
	DrawLineToC(gc, &p1)
	DrawLineEndC(gc)
}

/* translate the ipos3 to ipos2 in screen coordinate */
func (a *Navi) TranslatePos(p *C.struct_ipos3) (newp C.struct_ipos) {
	newp.x = p.x*a.divids.w + a.pos.x
	newp.y = p.z*a.divids.h + a.pos.y
	return
}

func (a *Navi) DrawMapCells(gc *draw2dgl.GraphicContext) {
	n := C.iarraylen(a.xmap.polygons)
	fmt.Println("DrawMapCells", n)
	for ; n > 0; n-- {
		poly := *((**C.struct_ipolygon3d)(C.iarrayat(a.xmap.polygons, C.int(n-1))))
		a.DrawPolygon3dC(gc, poly)
	}
}

func (a *Navi) DrawMapConnections(gc *draw2dgl.GraphicContext) {
	n := C.iarraylen(a.xmap.connections)
	fmt.Println("DrawMapConnections", n)

	gc.SetLineWidth(8)
	gc.SetStrokeColor(ColorGreen)
	for ; n > 0; n-- {
		con := *((**C.struct_inavicellconnection)(C.iarrayat(a.xmap.connections, C.int(n-1))))
		a.DrawLineC(gc, &con.start, &con.end)
		fmt.Println("line:", a.TranslatePos(&con.start), a.TranslatePos(&con.end))
	}
}

func (a *Navi) Draw(gc *draw2dgl.GraphicContext) {
	a.DrawMap(gc)
	a.DrawMapDesc(gc)
	a.DrawMapCells(gc)
	a.DrawMapConnections(gc)
}
func (a *Navi) MousePress(x, y float64) {
}
func (a *Navi) MouseMove(x, y float64) {
}
