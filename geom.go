// Copyright (c) 2011 Mateusz Czapliński (Go port)
// Copyright (c) 2011 Mahir Iqbal (as3 version)
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// based on http://code.google.com/p/as3polyclip/ (MIT licensed)
// and code by Martínez et al: http://wwwdi.ujaen.es/~fmartin/bool_op.html (public domain)

// Package geomop provides implementation of algorithms for geometry operations.
// For further details, consult the description of Polygon.Construct method.
package geomop

import (
	"math"
	"reflect"

	"github.com/twpayne/gogeom/geom"
)

// Equals returns true if both p1 and p2 describe the same point within
// a tolerance limit.
func PointEquals(p1, p2 geom.Point) bool {
	//	return (p1.X == p2.X && p1.Y == p2.Y)
	return (p1.X == p2.X && p1.Y == p2.Y) ||
		(math.Abs(p1.X-p2.X)/math.Abs(p1.X+p2.X) < tolerance &&
			math.Abs(p1.Y-p2.Y)/math.Abs(p1.Y+p2.Y) < tolerance)
}

func floatEquals(f1, f2 float64) bool {
	return (f1 == f2) ||
		(math.Abs(f1-f2)/math.Abs(f1+f2) < tolerance)
}

func pointSubtract(p1, p2 geom.Point) geom.Point {
	return geom.Point{p1.X - p2.X, p1.Y - p2.Y}
}

// Length returns distance from p to point (0, 0).
func lengthToOrigin(p geom.Point) float64 {
	return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

// Used to represent an edge of a polygon.
type segment struct {
	start, end geom.Point
}

// Contour represents a sequence of vertices connected by line segments, forming a closed shape.
type contour []geom.Point

func (c contour) segment(index int) segment {
	if index == len(c)-1 {
		return segment{c[len(c)-1], c[0]}
	}
	return segment{c[index], c[index+1]}
	// if out-of-bounds, we expect panic detected by runtime
}

// Clone returns a copy of a contour.
func (c contour) Clone() contour {
	return append([]geom.Point{}, c...)
}

// numVertices returns total number of all vertices of all contours of a polygon.
func numVertices(p geom.Polygon) int {
	num := 0
	for _, c := range p.Rings {
		num += len(c)
	}
	return num
}

// Clone returns a duplicate of a polygon.
func Clone(p geom.Polygon) geom.Polygon {
	var r geom.Polygon
	r.Rings = make([][]geom.Point, len(p.Rings))
	for i, rr := range p.Rings {
		r.Rings[i] = make([]geom.Point, len(rr))
		for j, pp := range p.Rings[i] {
			r.Rings[i][j] = pp
		}
	}
	return r
}

// Op describes an operation which can be performed on two polygons.
type Op int

const (
	UNION Op = iota
	INTERSECTION
	DIFFERENCE
	XOR
)

// Function Construct computes a 2D polygon, which is a result of performing
// specified Boolean operation on the provided pair of polygons (p <Op> clipping).
// It uses an algorithm described by F. Martínez, A. J. Rueda, F. R. Feito
// in "A new algorithm for computing Boolean operations on polygons"
// - see: http://wwwdi.ujaen.es/~fmartin/bool_op.html
// The paper describes the algorithm as performing in time O((n+k) log n),
// where n is number of all edges of all polygons in operation, and
// k is number of intersections of all polygon edges.
// "subject" and "clipping" can both be of type geom.Polygon,
// geom.MultiPolygon, geom.LineString, or geom.MultiLineString.
func Construct(subject, clipping geom.T, operation Op) (geom.T, error) {
	if subject == nil && clipping == nil {
		return nil, nil
	} else if subject == nil {
		if operation == INTERSECTION || operation == DIFFERENCE {
			return nil, nil
		} else {
			return clipping, nil
		}
	} else if clipping == nil {
		if operation == INTERSECTION {
			return nil, nil
		} else {
			return subject, nil
		}
	}
	// Prepare the input shapes
	var c clipper
	switch clipping.(type) {
	case geom.Polygon, geom.MultiPolygon:
		c.subject = convertToPolygon(subject)
		c.clipping = convertToPolygon(clipping)
		switch subject.(type) {
		case geom.Polygon, geom.MultiPolygon:
			c.outType = outputPolygons
		case geom.LineString, geom.MultiLineString:
			c.outType = outputLines
		default:
			return nil, NewError(subject)
		}
	case geom.LineString, geom.MultiLineString:
		switch subject.(type) {
		case geom.Polygon, geom.MultiPolygon:
			// swap clipping and subject
			c.subject = convertToPolygon(clipping)
			c.clipping = convertToPolygon(subject)
			c.outType = outputLines
		case geom.LineString, geom.MultiLineString:
			c.subject = convertToPolygon(subject)
			c.clipping = convertToPolygon(clipping)
			c.outType = outputPoints
		default:
			return nil, NewError(subject)
		}
	default:
		return nil, NewError(clipping)
	}
	// Run the clipper
	return c.compute(operation), nil
}

// convert input shapes to polygon to make internal processing easier
func convertToPolygon(g geom.T) geom.Polygon {
	var out geom.Polygon
	switch g.(type) {
	case geom.Polygon:
		out = g.(geom.Polygon)
	case geom.MultiPolygon:
		out.Rings = make([][]geom.Point, 0)
		for _, p := range g.(geom.MultiPolygon).Polygons {
			for _, r := range p.Rings {
				out.Rings = append(out.Rings, r)
			}
		}
	case geom.LineString:
		g2 := g.(geom.LineString)
		out.Rings = make([][]geom.Point, 1)
		out.Rings[0] = make([]geom.Point, len(g2.Points))
		for j, p := range g2.Points {
			out.Rings[0][j] = p
		}
	case geom.MultiLineString:
		g2 := g.(geom.MultiLineString)
		out.Rings = make([][]geom.Point, len(g2.LineStrings))
		for i, ls := range g2.LineStrings {
			out.Rings[i] = make([]geom.Point, len(ls.Points))
			for j, p := range ls.Points {
				out.Rings[i][j] = p
			}
		}
	default:
		panic(NewError(g))
	}
	// The clipper doesn't work well if a shape is made up of only two points.
	// To get around this problem, if there are only 2 points, we add a third
	// one a small distance from the second point.
	// However, if there is only 1 point, we just delete the shape.
	for i, r := range out.Rings {
		if len(r) == 0 {
			continue
		} else if len(r) == 1 {
			out.Rings[i] = make([]geom.Point, 0)
		} else if len(r) == 2 {
			const delta = 0.00001
			newpt := geom.Point{r[1].X + (r[1].X-r[0].X)*delta,
				r[1].Y - (r[1].Y-r[0].Y)*delta}
			out.Rings[i] = append(r, newpt)
		}
	}
	return out
}

type UnsupportedGeometryError struct {
	Type reflect.Type
}

func NewError(g geom.T) UnsupportedGeometryError {
	return UnsupportedGeometryError{reflect.TypeOf(g)}
}

func (e UnsupportedGeometryError) Error() string {
	return "Unsupported geometry type: " + e.Type.String()
}
