package wkb

import (
	"github.com/ctessum/geom"
	"reflect"
	"testing"
)

func TestWKB(t *testing.T) {

	var testCases = []struct {
		g   geom.Geom
		xdr []byte
		ndr []byte
	}{
		{
			g:   geom.Point{X: 1, Y: 2},
			xdr: []byte("\x00\x00\x00\x00\x01?\xf0\x00\x00\x00\x00\x00\x00@\x00\x00\x00\x00\x00\x00\x00"),
			ndr: []byte("\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\xf0?\x00\x00\x00\x00\x00\x00\x00@"),
		},
		{
			g:   geom.LineString([]geom.Point{{1, 2}, {3, 4}}),
			xdr: []byte("\x00\x00\x00\x00\x02\x00\x00\x00\x02?\xf0\x00\x00\x00\x00\x00\x00@\x00\x00\x00\x00\x00\x00\x00@\x08\x00\x00\x00\x00\x00\x00@\x10\x00\x00\x00\x00\x00\x00"),
			ndr: []byte("\x01\x02\x00\x00\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\x00\xf0?\x00\x00\x00\x00\x00\x00\x00@\x00\x00\x00\x00\x00\x00\x08@\x00\x00\x00\x00\x00\x00\x10@"),
		},
		{
			g:   geom.Polygon([]geom.Path{{{1, 2}, {3, 4}, {5, 6}, {1, 2}}}),
			xdr: []byte("\x00\x00\x00\x00\x03\x00\x00\x00\x01\x00\x00\x00\x04?\xf0\x00\x00\x00\x00\x00\x00@\x00\x00\x00\x00\x00\x00\x00@\x08\x00\x00\x00\x00\x00\x00@\x10\x00\x00\x00\x00\x00\x00@\x14\x00\x00\x00\x00\x00\x00@\x18\x00\x00\x00\x00\x00\x00?\xf0\x00\x00\x00\x00\x00\x00@\x00\x00\x00\x00\x00\x00\x00"),
			ndr: []byte("\x01\x03\x00\x00\x00\x01\x00\x00\x00\x04\x00\x00\x00\x00\x00\x00\x00\x00\x00\xf0?\x00\x00\x00\x00\x00\x00\x00@\x00\x00\x00\x00\x00\x00\x08@\x00\x00\x00\x00\x00\x00\x10@\x00\x00\x00\x00\x00\x00\x14@\x00\x00\x00\x00\x00\x00\x18@\x00\x00\x00\x00\x00\x00\xf0?\x00\x00\x00\x00\x00\x00\x00@"),
		},
		{
			g:   geom.MultiPoint([]geom.Point{{1, 2}, {3, 4}}),
			xdr: []byte("\x00\x00\x00\x00\x04\x00\x00\x00\x02\x00\x00\x00\x00\x01?\xf0\x00\x00\x00\x00\x00\x00@\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01@\x08\x00\x00\x00\x00\x00\x00@\x10\x00\x00\x00\x00\x00\x00"),
			ndr: []byte("\x01\x04\x00\x00\x00\x02\x00\x00\x00\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\xf0?\x00\x00\x00\x00\x00\x00\x00@\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x08@\x00\x00\x00\x00\x00\x00\x10@"),
		},
	}

	for _, tc := range testCases {

		// test XDR decoding
		if got, err := Decode(tc.xdr); err != nil || !reflect.DeepEqual(got, tc.g) {
			t.Errorf("Decode(%#v) == %#v, %s, want %#v, nil", tc.xdr, got, err, tc.g)
		}

		// test XDR encoding
		if got, err := Encode(tc.g, XDR); err != nil || !reflect.DeepEqual(got, tc.xdr) {
			t.Errorf("Encode(%#v, %#v) == %#v, %#v, want %#v, nil", tc.g, XDR, got, err, tc.xdr)
		}

		// test NDR decoding
		if got, err := Decode(tc.ndr); err != nil || !reflect.DeepEqual(got, tc.g) {
			t.Errorf("Decode(%#v) == %#v, %s, want %#v, nil", tc.ndr, got, err, tc.g)
		}

		// test NDR encoding
		if got, err := Encode(tc.g, NDR); err != nil || !reflect.DeepEqual(got, tc.ndr) {
			t.Errorf("Encode(%#v, %#v) == %#v, %#v, want %#v, nil", tc.g, NDR, got, err, tc.ndr)
		}

	}

}
