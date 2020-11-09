package dot

import (
	"fmt"
	"testing"
)

func TestEdgeStyleHelpers(t *testing.T) {

	type test struct {
		input string
		want  string
	}

	tests := []test{
		{input: "solid", want: `digraph "" {"%[1]s"[label="A"];"%[2]s"[label="B"];"%[1]s"->"%[2]s"[style="solid"];}`},
		{input: "bold", want: `digraph "" {"%[1]s"[label="A"];"%[2]s"[label="B"];"%[1]s"->"%[2]s"[style="bold"];}`},
		{input: "dashed", want: `digraph "" {"%[1]s"[label="A"];"%[2]s"[label="B"];"%[1]s"->"%[2]s"[style="dashed"];}`},
		{input: "dotted", want: `digraph "" {"%[1]s"[label="A"];"%[2]s"[label="B"];"%[1]s"->"%[2]s"[style="dotted"];}`},
	}

	for _, tc := range tests {

		di := NewGraph(nil)
		n1 := di.Node("A")
		n2 := di.Node("B")

		switch tc.input {
		case "solid":
			di.Edge(n1, n2).Solid()
		case "bold":
			di.Edge(n1, n2).Bold()
		case "dashed":
			di.Edge(n1, n2).Dashed()
		case "dotted":
			di.Edge(n1, n2).Dotted()
		}

		if got, want := flatten(di.String()), fmt.Sprintf(tc.want, n1.ID(), n2.ID()); got != want {
			t.Errorf("got [%v] want [%v]", got, want)
		}
	}
}