package word2vec

import (
	"reflect"
	"testing"

	"github.com/unixpickle/anydiff"
	"github.com/unixpickle/anydiff/anydifftest"
	"github.com/unixpickle/anyvec"
	"github.com/unixpickle/anyvec/anyvec32"
	"github.com/unixpickle/serializer"
)

func TestNetStep(t *testing.T) {
	res := randomNetRes()
	checker := anydifftest.ResChecker{
		F: func() anydiff.Res {
			return res
		},
		V:     []*anydiff.Var{res.Net.Encoder, res.Net.Decoder},
		Delta: 1e-2,
		Prec:  1e-3,
	}
	checker.FullCheck(t)
}

func TestNetSerialize(t *testing.T) {
	c := anyvec32.DefaultCreator{}
	net := NewNet(c, 15, 20, 10)
	data, err := serializer.SerializeAny(net)
	if err != nil {
		t.Fatal(err)
	}
	var net1 *Net
	if err := serializer.DeserializeAny(data, &net1); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(net, net1) {
		t.Error("invalid result")
	}
}

type netRes struct {
	In  map[int]anyvec.Numeric
	Out [][]int
	Net *Net
}

func randomNetRes() *netRes {
	return &netRes{
		In:  map[int]anyvec.Numeric{1: float32(2), 4: float32(1)},
		Out: [][]int{{1, -2, 3}, {1, -3, 5}, {1, 2, -4}},
		Net: NewNet(anyvec32.CurrentCreator(), 5, 3, 10),
	}
}

func (n *netRes) Output() anyvec.Vector {
	cost := n.Net.Step(n.In, n.Out, float32(0)).(float32)
	return anyvec32.MakeVectorData([]float32{cost})
}

func (n *netRes) Vars() anydiff.VarSet {
	res := anydiff.VarSet{}
	res.Add(n.Net.Encoder)
	res.Add(n.Net.Decoder)
	return res
}

func (n *netRes) Propagate(u anyvec.Vector, g anydiff.Grad) {
	netCopy := &Net{
		In:     n.Net.In,
		Out:    n.Net.Out,
		Hidden: n.Net.Hidden,

		Encoder: anydiff.NewVar(n.Net.Encoder.Vector.Copy()),
		Decoder: anydiff.NewVar(n.Net.Decoder.Vector.Copy()),
	}
	netCopy.Step(n.In, n.Out, float32(1))
	netCopy.Encoder.Vector.Sub(n.Net.Encoder.Vector)
	netCopy.Decoder.Vector.Sub(n.Net.Decoder.Vector)
	uScaler := anyvec.Sum(u)
	netCopy.Encoder.Vector.Scale(uScaler)
	netCopy.Decoder.Vector.Scale(uScaler)

	if vec, ok := g[n.Net.Encoder]; ok {
		vec.Add(netCopy.Encoder.Vector)
	}
	if vec, ok := g[n.Net.Decoder]; ok {
		vec.Add(netCopy.Decoder.Vector)
	}
}
