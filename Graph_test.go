package dot

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/wwmoraes/dot/attributes"
	"github.com/wwmoraes/dot/dottest"
)

// TestGraphBehavior tests all components with real use cases
func TestGraphBehavior(t *testing.T) {
	t.Run("default graph", func(t *testing.T) {
		graph, err := NewGraph()
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		graph.Node("n1").Edge(graph.Node("n2")).SetAttributeString(attributes.KeyLabel, "uses")

		expected := `digraph {"n1";"n2";"n1"->"n2"[label="uses"];}`

		if got, want := dottest.MustGetFlattenSerializableString(t, graph), dottest.Flatten(t, expected); got != want {
			t.Errorf("got [\n%v\n]want [\n%v\n]", got, want)
		}
	})
	t.Run("directed graph", func(t *testing.T) {
		graph, err := NewGraph(WithType(GraphTypeDirected))
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		graph.Node("n1").Edge(graph.Node("n2")).SetAttributeString(attributes.KeyLabel, "uses")

		expected := `digraph {"n1";"n2";"n1"->"n2"[label="uses"];}`

		if got, want := dottest.MustGetFlattenSerializableString(t, graph), dottest.Flatten(t, expected); got != want {
			t.Errorf("got [\n%v\n]want [\n%v\n]", got, want)
		}
	})
	t.Run("undirected graph", func(t *testing.T) {
		graph, err := NewGraph(WithType(GraphTypeUndirected))
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		graph.Node("n1").Edge(graph.Node("n2")).SetAttributeString(attributes.KeyLabel, "uses")

		expected := `graph {"n1";"n2";"n1"--"n2"[label="uses"];}`

		if got, want := dottest.MustGetFlattenSerializableString(t, graph), dottest.Flatten(t, expected); got != want {
			t.Errorf("got [\n%v\n]want [\n%v\n]", got, want)
		}
	})
	t.Run("subgraphs", func(t *testing.T) {
		graph, err := NewGraph()
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}
		subGraph, err := graph.Subgraph()
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		subSubGraph, err := subGraph.Subgraph()
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		n1 := graph.Node("n1")
		n1.Edge(subGraph.Node("n2")).Edge(subSubGraph.Node("n3")).Edge(n1)
		subSubGraph.Node("n4").Edge(subSubGraph.Node("n3"))

		expected := `digraph {subgraph {subgraph {"n3";"n4";"n4"->"n3";}"n2";}"n1";"n1"->"n2";"n2"->"n3";"n3"->"n1";}`

		if got, want := dottest.MustGetFlattenSerializableString(t, graph), dottest.Flatten(t, expected); got != want {
			t.Errorf("got [\n%v\n]want [\n%v\n]", got, want)
		}
	})
}

// TestNewGraph tests NewGraph factory function with static outcome
func TestNewGraph(t *testing.T) {
	type args struct {
		options []GraphOptionFn
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty graph",
			args: args{
				options: nil,
			},
			want: `digraph {}`,
		},
		{
			name: "empty named graph",
			args: args{
				options: []GraphOptionFn{
					WithID("test"),
				},
			},
			want: `digraph "test" {}`,
		},
		{
			name: "empty directed graph",
			args: args{
				options: []GraphOptionFn{
					WithType(GraphTypeDirected),
				},
			},
			want: `digraph {}`,
		},
		{
			name: "empty named directed graph",
			args: args{
				options: []GraphOptionFn{
					WithID("test"),
					WithType(GraphTypeDirected),
				},
			},
			want: `digraph "test" {}`,
		},
		{
			name: "empty undirected graph",
			args: args{
				options: []GraphOptionFn{
					WithType(GraphTypeUndirected),
				},
			},
			want: `graph {}`,
		},
		{
			name: "empty named undirected graph",
			args: args{
				options: []GraphOptionFn{
					WithID("test"),
					WithType(GraphTypeUndirected),
				},
			},
			want: `graph "test" {}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := NewGraph(tt.args.options...)
			if err != nil {
				t.Fatal("graph is nil, expected a valid instance")
			}

			if got := dottest.MustGetFlattenSerializableString(t, graph); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGraph() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("graph with random ID", func(t *testing.T) {
		graph, err := NewGraph(WithID("-"))
		if err != nil {
			t.Fatalf("unexpected error: %++v", err)
		}
		if graph == nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		if graph.ID() == "-" {
			t.Fatal("graph id was not generated")
		}
	})
}

func TestNewGraph_invalid(t *testing.T) {
	tests := []struct {
		name    string
		options []GraphOptionFn
	}{
		{
			name: "root subgraph",
			options: []GraphOptionFn{
				WithType(GraphTypeSub),
			},
		},
		{
			name: "non-root digraph",
			options: []GraphOptionFn{
				func() GraphOptionFn {
					graph, _ := NewGraph()
					return WithParent(graph)
				}(),
				WithType(GraphTypeDirected),
			},
		},
		{
			name: "non-root graph",
			options: []GraphOptionFn{
				func() GraphOptionFn {
					graph, _ := NewGraph()
					return WithParent(graph)
				}(),
				WithType(GraphTypeUndirected),
			},
		},
		{
			name: "graph without generator",
			options: []GraphOptionFn{
				WithGenerator(nil),
			},
		},
		{
			name: "graph without generator and dash id",
			options: []GraphOptionFn{
				WithGenerator(nil),
				WithID("-"),
			},
		},
		{
			name: "root cluster graph",
			options: []GraphOptionFn{
				WithCluster(),
			},
		},
	}

	for _, tt := range tests {
		graph, err := NewGraph(tt.options...)
		if graph != nil {
			t.Error("graph is not nil, a valid instance wasn't expected")
		}
		if err == nil {
			t.Error("error is nil, an error was expected")
		}
	}
}

func TestGraph_Initializers(t *testing.T) {
	t.Run("graph with node initializer", func(t *testing.T) {
		graph, err := NewGraph(
			WithNodeInitializer(func(nodeInstance Node) {
				nodeInstance.SetAttributeString(attributes.KeyClass, "test-class")
			}),
		)
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		graph.Node("n1")

		if got, want := dottest.MustGetFlattenSerializableString(t, graph), `digraph {"n1"[class="test-class"];}`; got != want {
			t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
		}
	})
	t.Run("graph with edge initializer", func(t *testing.T) {
		graph, err := NewGraph(
			WithEdgeInitializer(func(edgeInstance StyledEdge) {
				edgeInstance.SetAttributeString(attributes.KeyClass, "test-class")
			}),
		)
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		graph.Node("n1").Edge(graph.Node("n2"))

		if got, want := dottest.MustGetFlattenSerializableString(t, graph), `digraph {"n1";"n2";"n1"->"n2"[class="test-class"];}`; got != want {
			t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
		}
	})
}

func TestGraph_FindSubgraph(t *testing.T) {
	t.Run("find existing subgraph from another subgraph", func(t *testing.T) {
		graph, err := NewGraph(WithID("root-graph"))
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}
		sub1, err := graph.Subgraph(WithID("subgraph-one"))
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		sub2, err := graph.Subgraph(WithID("subgraph-two"))
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		got, found := sub1.FindSubgraph(sub2.ID())

		if !found {
			t.Error("subgraph not found as expected")
		}

		if want := sub2; !reflect.DeepEqual(got, want) {
			t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
		}
	})
	t.Run("find no un-existent subgraph from another subgraph", func(t *testing.T) {
		graph, err := NewGraph(WithID("root-graph"))
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		sub1, err := graph.Subgraph(WithID("subgraph-one"))
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		got, found := sub1.FindSubgraph("subgraph-two")

		if found {
			t.Error("subgraph was found, it wasn't expected")
		}

		if got != nil {
			t.Errorf("got [%v] want [%v]", got, nil)
		}
	})
	t.Run("fail to create invalid subgraph type", func(t *testing.T) {
		graph, err := NewGraph(WithID("root-graph"))
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		_, err = graph.Subgraph(
			WithID("subgraph-one"),
			WithType(GraphTypeDirected),
		)
		if err == nil {
			t.Fatal("graph error is nil, expected an error")
		}

		got, found := graph.FindSubgraph("subgraph-one")

		if found {
			t.Error("subgraph was found, it wasn't expected")
		}

		if got != nil {
			t.Errorf("got [%v] want [%v]", got, nil)
		}
	})
}

// TestNewGraph_generatedID tests NewGraph factory option to generate unique IDs
func TestNewGraph_generatedID(t *testing.T) {
	tests := []struct {
		name    string
		options []GraphOptionFn
		want    func(graph Graph) string
	}{
		{
			name:    "empty randomly-named graph",
			options: nil,
			want: func(graph Graph) string {
				return fmt.Sprintf(`digraph "%s" {}`, graph.ID())
			},
		},
		{
			name: "empty randomly-named directed graph",
			options: []GraphOptionFn{
				WithType(GraphTypeDirected),
			},
			want: func(graph Graph) string {
				return fmt.Sprintf(`digraph "%s" {}`, graph.ID())
			},
		},
		{
			name: "empty randomly-named undirected graph",
			options: []GraphOptionFn{
				WithType(GraphTypeUndirected),
			},
			want: func(graph Graph) string {
				return fmt.Sprintf(`graph "%s" {}`, graph.ID())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := make([]GraphOptionFn, 0, len(tt.options)+1)
			options = append(options, WithID("-"))
			options = append(options, tt.options...)

			graph, err := NewGraph(options...)
			if err != nil {
				t.Fatal("graph is nil, expected a valid instance")
			}

			if graph.ID() == "" {
				t.Error("got empty ID instead of a random one")
			}

			if graph.ID() == "-" {
				t.Error("got dash ID instead of a random one")
			}

			want := tt.want(graph)

			if got := dottest.MustGetFlattenSerializableString(t, graph); !reflect.DeepEqual(got, want) {
				t.Errorf("NewGraph() = %v, want %v", got, want)
			}
		})
	}
}

// TestGraph_Subgraph tests Graph.Subgraph factory
func TestGraph_Subgraph(t *testing.T) {
	tests := []struct {
		name    string
		options []GraphOptionFn
		want    string
	}{
		{
			name:    "empty anonymous subgraph",
			options: nil,
			want:    `digraph {subgraph {}}`,
		},
		{
			name: "empty anonymous subgraph with enforced type",
			options: []GraphOptionFn{
				WithType(GraphTypeSub),
			},
			want: `digraph {subgraph {}}`,
		},
		{
			name: "empty named subgraph",
			options: []GraphOptionFn{
				WithID("test-sub"),
			},
			want: `digraph {subgraph "test-sub" {}}`,
		},
		{
			name: "remove cluster preffix subgraph",
			options: []GraphOptionFn{
				WithID("cluster_test-sub"),
			},
			want: `digraph {subgraph "test-sub" {}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := NewGraph()
			if err != nil {
				t.Fatal("graph is nil, expected a valid instance")
			}

			_, err = graph.Subgraph(tt.options...)
			if err != nil {
				t.Fatal("graph is nil, expected a valid instance")
			}

			if got := dottest.MustGetFlattenSerializableString(t, graph); got != tt.want {
				t.Errorf("got [%v] want [%v]", got, tt.want)
			}
		})
	}

	t.Run("empty randomly-named subgraph", func(t *testing.T) {
		graph, err := NewGraph()
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		subGraph, err := graph.Subgraph(WithID("-"))
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		if subGraph.ID() == "" {
			t.Error("got empty ID instead of a random one")
		}

		if subGraph.ID() == "-" {
			t.Error("got dash ID instead of a random one")
		}

		if got, want := dottest.MustGetFlattenSerializableString(t, graph), fmt.Sprintf(`digraph {subgraph "%s" {}}`, subGraph.ID()); got != want {
			t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
		}
	})
}

// TestGraph_Subgraph_generatedID tests Graph.Subgraph factory
func TestGraph_Subgraph_generatedID(t *testing.T) {
	tests := []struct {
		name    string
		options []GraphOptionFn
		want    string
	}{
		{
			name: "cluster then id subgraph",
			options: []GraphOptionFn{
				WithCluster(),
				WithID("-"),
			},
			want: `digraph {subgraph "%s" {}}`,
		},
		{
			name: "id then cluster subgraph",
			options: []GraphOptionFn{
				WithID("-"),
				WithCluster(),
			},
			want: `digraph {subgraph "%s" {}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := NewGraph()
			if err != nil {
				t.Fatal("graph is nil, expected a valid instance")
			}

			subGraph, err := graph.Subgraph(tt.options...)
			if err != nil {
				t.Fatal("graph is nil, expected a valid instance")
			}

			if subGraph.ID() == "-" {
				t.Error("graph has dash id, expected a randomly generated one")
			}

			if got := dottest.MustGetFlattenSerializableString(t, graph); got != fmt.Sprintf(tt.want, subGraph.ID()) {
				t.Errorf("got [%v] want [%v]", got, tt.want)
			}
		})
	}

	t.Run("empty randomly-named subgraph", func(t *testing.T) {
		graph, err := NewGraph()
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		subGraph, err := graph.Subgraph(WithID("-"))
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		if subGraph.ID() == "" {
			t.Error("got empty ID instead of a random one")
		}

		if subGraph.ID() == "-" {
			t.Error("got dash ID instead of a random one")
		}

		if got, want := dottest.MustGetFlattenSerializableString(t, graph), fmt.Sprintf(`digraph {subgraph "%s" {}}`, subGraph.ID()); got != want {
			t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
		}
	})
}

func TestGraph_Subgraph_invalid(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatalf("unexpected error when creating graph: %++v", err)
	}

	tests := []struct {
		name    string
		options []GraphOptionFn
	}{
		{
			name: "remove parent from subgraph",
			options: []GraphOptionFn{
				WithParent(nil),
			},
		},
	}

	for _, tt := range tests {
		subgraph, err := graph.Subgraph(tt.options...)
		if subgraph != nil {
			t.Error("subgraph is not nil, a valid instance wasn't expected")
		}
		if err == nil {
			t.Error("error is nil, an error was expected")
		}
	}
}

func TestEmpty(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	if got, want := dottest.MustGetFlattenSerializableString(t, di), `digraph {}`; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
	di2, err := NewGraph(WithID("test"))
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	if got, want := dottest.MustGetFlattenSerializableString(t, di2), `digraph "test" {}`; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
	di3, err := NewGraph(WithID("-"))
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	if di3.ID() == "-" {
		t.Error("got dash id instead of randomly generated one")
	}
	if got, want := dottest.MustGetFlattenSerializableString(t, di3), fmt.Sprintf(`digraph "%s" {}`, di3.ID()); got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestStrict(t *testing.T) {
	// test strict directed
	{
		graph, err := NewGraph(WithStrict())
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		if got, want := dottest.MustGetFlattenSerializableString(t, graph), `strict digraph {}`; got != want {
			t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
		}
	}
	// test strict undirected
	{
		graph, err := NewGraph(
			WithStrict(),
			WithType(GraphTypeUndirected),
		)
		if err != nil {
			t.Fatal("graph is nil, expected a valid instance")
		}

		if got, want := dottest.MustGetFlattenSerializableString(t, graph), `strict graph {}`; got != want {
			t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
		}
	}
}

func TestEmptyWithIDAndAttributes(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	di.SetAttribute(attributes.KeyStyle, attributes.NewString("filled"))
	di.SetAttribute(attributes.KeyColor, attributes.NewString("lightgrey"))
	if got, want := dottest.MustGetFlattenSerializableString(t, di), `digraph {graph [color="lightgrey",style="filled"];}`; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestEmptyWithHTMLLabel(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	di.SetAttribute(attributes.KeyLabel, attributes.NewHTML("<B>Hi</B>"))
	if got, want := dottest.MustGetFlattenSerializableString(t, di), `digraph {graph [label=<<B>Hi</B>>];}`; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestEmptyWithLiteralValueLabel(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	di.SetAttribute(attributes.KeyLabel, attributes.NewLiteral(`"left-justified text\l"`))
	if got, want := dottest.MustGetFlattenSerializableString(t, di), `digraph {graph [label="left-justified text\l"];}`; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestTwoConnectedNodes(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	n1 := di.Node("A")
	n2 := di.Node("B")
	di.Edge(n1, n2)
	if got, want := dottest.MustGetFlattenSerializableString(t, di), fmt.Sprintf(`digraph {"%[1]s";"%[2]s";"%[1]s"->"%[2]s";}`, n1.ID(), n2.ID()); got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestTwoConnectedNodesAcrossSubgraphs(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	n1 := di.Node("A")
	sub, err := di.Subgraph(WithID("my-sub"))
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	n2 := sub.Node("B")
	edge := di.Edge(n1, n2)
	edge.SetAttributeString(attributes.KeyLabel, "cross-graph")

	// test graph-level edge finding
	{
		want := []Edge{edge}
		got := di.FindEdges(n1, n2)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
		}
	}

	// test finding target edges
	{
		n3 := sub.Node("C")
		newEdge := di.Edge(n2, n3)
		want := []Edge{newEdge}
		got := edge.EdgesTo(n3)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
		}
	}
}

func TestUndirectedTwoConnectedNodes(t *testing.T) {
	di, err := NewGraph(
		WithType(GraphTypeUndirected),
	)
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	n1 := di.Node("A")
	n2 := di.Node("B")
	di.Edge(n1, n2)
	if got, want := dottest.MustGetFlattenSerializableString(t, di), fmt.Sprintf(`graph {"%[1]s";"%[2]s";"%[1]s"--"%[2]s";}`, n1.ID(), n2.ID()); got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestGraph_FindEdges(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	n1 := di.Node("A")
	n2 := di.Node("B")
	want := []Edge{di.Edge(n1, n2)}
	got := di.FindEdges(n1, n2)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestGraph.FindEdges() = %v, want %v", got, want)
	}
}

func TestSubgraph(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	sub, err := di.Subgraph(
		WithID("test-id"),
	)
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	sub.SetAttributeString(attributes.KeyStyle, "filled")
	if got, want := dottest.MustGetFlattenSerializableString(t, di), fmt.Sprintf(`digraph {subgraph "%s" {graph [style="filled"];}}`, sub.ID()); got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
	sub.SetAttributeString(attributes.KeyLabel, "new-label")
	if got, want := dottest.MustGetFlattenSerializableString(t, di), fmt.Sprintf(`digraph {subgraph "%s" {graph [label="new-label",style="filled"];}}`, sub.ID()); got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
	foundGraph, _ := di.FindSubgraph("test-id")
	if got, want := foundGraph, sub; got != want {
		t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
	}
	subsub, err := sub.Subgraph(
		WithID("sub-test-id"),
	)
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	foundGraph, _ = di.FindSubgraph("test-id")
	if got, want := foundGraph, sub; got != want {
		t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
	}

	// test getting an existing subgraph

	gotGraph, found := sub.FindSubgraph(subsub.ID())
	if !found {
		t.Errorf("%s not found, it was expected to be found", subsub.ID())
	}
	if gotGraph != subsub {
		t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", gotGraph, subsub)
	}
}

func TestSubgraphClusterOption(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	sub, err := di.Subgraph(
		WithID("s1"),
		WithCluster(),
	)
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	if got, want := sub.ID(), "cluster_s1"; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestNode(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	node := graph.Node("")
	node.SetAttributesString(attributes.MapString{
		attributes.KeyLabel: "test",
		attributes.KeyShape: "box",
	})

	if node.ID() == "" {
		t.Error("got blank node id, expected a randonly generated")
		return
	}

	if got, want := dottest.MustGetFlattenSerializableString(t, graph), fmt.Sprintf(`digraph {"%s"[label="test",shape="box"];}`, node.ID()); got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}

	// test extra node + finding inexistent edge
	{
		node2 := graph.Node("")
		node.EdgesTo(node2)

		want := []Edge{}
		got := node.EdgesTo(node2)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
		}
	}
}

func TestEdgeLabel(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	n1 := di.Node("e1")
	n2 := di.Node("e2")
	attr := attributes.NewAttributes()
	attr.SetAttributeString(attributes.KeyLabel, "what")
	n1.EdgeWithAttributes(n2, attr)
	if got, want := dottest.MustGetFlattenSerializableString(t, di), `digraph {"e1";"e2";"e1"->"e2"[label="what"];}`; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestSameRank(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	foo1 := di.Node("foo1")
	foo2 := di.Node("foo2")
	bar := di.Node("bar")
	foo1.Edge(foo2)
	foo1.Edge(bar)
	di.AddToSameRank("top-row", foo1, foo2)
	if got, want := dottest.MustGetFlattenSerializableString(t, di), `digraph {"bar";"foo1";"foo2";{rank=same;"foo1";"foo2";}"foo1"->"foo2";"foo1"->"bar";}`; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestCluster(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	outside := di.Node("Outside")
	clusterA, err := di.Subgraph(
		WithID("Cluster A"),
		WithCluster(),
	)
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	clusterA.SetAttributeString(attributes.KeyLabel, "Cluster A")
	insideOne := clusterA.Node("one")
	insideTwo := clusterA.Node("two")
	clusterB, err := di.Subgraph(
		WithID("Cluster B"),
		WithCluster(),
	)
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	clusterB.SetAttributeString(attributes.KeyLabel, "Cluster B")
	insideThree := clusterB.Node("three")
	insideFour := clusterB.Node("four")
	outside.Edge(insideFour).Edge(insideOne).Edge(insideTwo).Edge(insideThree).Edge(outside)
	filePath := path.Join(t.TempDir(), "cluster.dot")
	if err := ioutil.WriteFile(filePath, []byte(dottest.MustGetSerializableString(t, di)), os.ModePerm); err != nil {
		t.Errorf("unable to write dot file: %w", err)
	}
}

func TestDeleteLabel(t *testing.T) {
	g, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	n := g.Node("my-id")
	n.DeleteAttribute(attributes.KeyLabel)
	if got, want := dottest.MustGetFlattenSerializableString(t, g), `digraph {"my-id";}`; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestGraph_FindNodeById_emptyGraph(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	_, found := di.FindNodeByID("F")

	if got, want := found, false; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestGraph_FindNodeById_multiNodeGraph(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	di.Node("A")
	di.Node("B")

	node, found := di.FindNodeByID("A")

	if got, want := node.ID(), "A"; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}

	if got, want := found, true; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestGraph_FindNodeById_multiNodesInSubGraphs(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	di.Node("A")
	di.Node("B")
	sub, err := di.Subgraph(
		WithID("new subgraph"),
	)
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	sub.Node("C")

	node, found := di.FindNodeByID("C")

	if got, want := node.ID(), "C"; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}

	if got, want := found, true; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestGraph_FindNodes_multiNodesInSubGraphs(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	di.Node("A")
	di.Node("B")
	sub, err := di.Subgraph(
		WithID("new subgraph"),
	)
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	sub.Node("C")

	nodes := di.FindNodes()

	if got, want := len(nodes), 3; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestLabelWithEscaping(t *testing.T) {
	di, err := NewGraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	n := di.Node("without-linefeed")
	n.SetAttribute(attributes.KeyLabel, attributes.NewLiteral(`"with \l linefeed"`))
	if got, want := dottest.MustGetFlattenSerializableString(t, di), `digraph {"without-linefeed"[label="with \l linefeed"];}`; got != want {
		t.Errorf("got [\n%v\n] want [\n%v\n]", got, want)
	}
}

func TestGraphNodeInitializer(t *testing.T) {
	di, err := NewGraph(
		WithNodeInitializer(func(n Node) {
			n.SetAttribute(attributes.KeyLabel, attributes.NewString("test"))
		}),
	)
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	n := di.Node("A")
	gotAttr, gotOk := n.GetAttribute(attributes.KeyLabel)
	if !gotOk {
		t.Error("attribute not found")
	}
	if got, want := gotAttr.(*attributes.String), attributes.NewString("test"); !reflect.DeepEqual(got, want) {
		t.Errorf("got [%v[1]:%[1]T] want [%[2]v:%[2]T]", got, want)
	}
}

func TestGraphEdgeInitializer(t *testing.T) {
	di, err := NewGraph(
		WithEdgeInitializer(func(e StyledEdge) {
			e.SetAttribute(attributes.KeyLabel, attributes.NewString("test"))
		}),
	)
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	e := di.Node("A").Edge(di.Node("B"))
	gotAttr, gotOk := e.GetAttribute(attributes.KeyLabel)
	if !gotOk {
		t.Error("attribute not found")
	}
	if got, want := gotAttr.(*attributes.String), attributes.NewString("test"); !reflect.DeepEqual(got, want) {
		t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
	}
}

func TestGraphCreateNodeOnce(t *testing.T) {
	di, err := NewGraph(
		WithType(GraphTypeUndirected),
	)
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	n1 := di.Node("A")
	n2 := di.Node("A")
	if got, want := n1, n2; &n1 == &n2 {
		t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
	}
}

func TestGraph_WriteTo(t *testing.T) {
	tests := []struct {
		name       string
		wantErr    error
		wantString string
	}{
		{
			name:       "zero data written",
			wantErr:    dottest.ErrLimit,
			wantString: "",
		},
		{
			name:       "partially written - strict preffix",
			wantErr:    dottest.ErrLimit,
			wantString: `strict `,
		},
		{
			name:       "partially written - graph type",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph`,
		},
		{
			name:       "partially written - graph id",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph"`,
		},
		{
			name:       "partially written - global graph attributes group",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {`,
		},
		{
			name:       "partially written - global graph attributes group",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph `,
		},
		{
			name:       "partially written - global graph attributes",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"]`,
		},
		{
			name:       "partially written - global graph attributes separator",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];`,
		},
		{
			name:       "partially written - subgraph start",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph`,
		},
		{
			name:       "partially written - subgraph open block",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {`,
		},
		{
			name:       "partially written - subgraph first node ID",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {"n1"`,
		},
		{
			name:       "partially written - subgraph first node separator",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {"n1";`,
		},
		{
			name:       "partially written - subgraph second node ID",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {"n1";"n2"`,
		},
		{
			name:       "partially written - subgraph second node separator",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {"n1";"n2";`,
		},
		{
			name:       "partially written - subgraph anonymous same rank group open",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {"n1";"n2";{`,
		},
		{
			name:       "partially written - subgraph anonymous same rank group attribute",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {"n1";"n2";{rank=same;`,
		},
		{
			name:       "partially written - subgraph anonymous same rank group first node",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {"n1";"n2";{rank=same;"n1"`,
		},
		{
			name:       "partially written - subgraph anonymous same rank group node separator",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {"n1";"n2";{rank=same;"n1";`,
		},
		{
			name:       "partially written - subgraph anonymous same rank group second node",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {"n1";"n2";{rank=same;"n1";"n2"`,
		},
		{
			name:       "partially written - subgraph anonymous same rank group node separator",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {"n1";"n2";{rank=same;"n1";"n2";`,
		},
		{
			name:       "partially written - subgraph anonymous same rank group close",
			wantErr:    dottest.ErrLimit,
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {"n1";"n2";{rank=same;"n1";"n2";}`,
		},
		{
			name:       "fully written",
			wantString: `strict digraph "test-graph" {graph [label="test-graph"];subgraph {"n1";"n2";{rank=same;"n1";"n2";}"n1"->"n2"[label="subgraph-edge"];}"n1"[label="node1"];"n2"[label="node2"];"n1"->"n2"[label="graph-edge"];}`,
		},
	}

	graph, err := NewGraph(
		WithID("test-graph"),
		WithStrict(),
	)
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	graph.SetAttributeString("label", "test-graph")
	subGraph, err := graph.Subgraph()
	if err != nil {
		t.Fatal("graph is nil, expected a valid instance")
	}

	subGraph.AddToSameRank("g1", subGraph.Node("n1"), subGraph.Node("n2"))
	subEdge := subGraph.Edge(subGraph.Node("n1"), subGraph.Node("n2"))
	subEdge.SetAttributeString("label", "subgraph-edge")
	edge := graph.Edge(graph.Node("n1"), graph.Node("n2"))
	graph.Node("n1").SetAttributeString("label", "node1")
	graph.Node("n2").SetAttributeString("label", "node2")
	edge.SetAttributeString("label", "graph-edge")

	for limit, tt := range tests {
		if limit == len(tests)-1 {
			limit = math.MaxInt32
		}
		t.Run(tt.name, func(t *testing.T) {
			wantN := int64(len(tt.wantString))
			dottest.TestByteWrite(t, graph, limit, tt.wantErr, wantN, tt.wantString)
		})
	}
}

func BenchmarkGraph_WriteTo(b *testing.B) {
	b.Run("empty graph to buffer", func(b *testing.B) {
		var err error
		var buf bytes.Buffer

		graph, err := NewGraph()
		if err != nil {
			b.Fatal("graph is nil, expected a valid instance")
		}

		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			_, err = graph.WriteTo(&buf)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("empty graph to ioutil.Discard", func(b *testing.B) {
		var err error

		graph, err := NewGraph()
		if err != nil {
			b.Fatal("graph is nil, expected a valid instance")
		}

		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			_, err = graph.WriteTo(ioutil.Discard)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
