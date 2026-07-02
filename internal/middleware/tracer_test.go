package middleware

import "testing"

func TestNestedTraces(t *testing.T) {
	rt := NewRequestTracer("req-1")

	rt.StartTrace("outer", "middleware", nil)
	rt.StartTrace("inner", "handler", nil)
	rt.StartTrace("innermost", "custom", nil)
	rt.EndTrace(nil)
	rt.EndTrace(nil)
	rt.EndTrace(nil)

	traces := rt.GetTraces()
	if len(traces) != 1 {
		t.Fatalf("expected 1 root trace, got %d", len(traces))
	}

	outer := traces[0]
	if outer.Name != "outer" || outer.Status != "completed" {
		t.Fatalf("root trace = %q (%s), want outer (completed)", outer.Name, outer.Status)
	}
	if len(outer.Children) != 1 || outer.Children[0].Name != "inner" {
		t.Fatalf("outer children = %+v, want [inner]", outer.Children)
	}

	inner := outer.Children[0]
	if inner.Status != "completed" {
		t.Fatalf("inner status = %s, want completed", inner.Status)
	}
	if len(inner.Children) != 1 || inner.Children[0].Name != "innermost" {
		t.Fatalf("inner children = %+v, want [innermost]", inner.Children)
	}
}

func TestNestedTraceSiblings(t *testing.T) {
	rt := NewRequestTracer("req-2")

	rt.StartTrace("root", "handler", nil)
	rt.StartTrace("step 1", "custom", nil)
	rt.EndTrace(nil)
	rt.StartTrace("step 2", "custom", nil)
	rt.EndTrace(nil)
	rt.EndTrace(nil)

	traces := rt.GetTraces()
	if len(traces) != 1 {
		t.Fatalf("expected 1 root trace, got %d", len(traces))
	}
	root := traces[0]
	if root.Status != "completed" {
		t.Fatalf("root status = %s, want completed", root.Status)
	}
	if len(root.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(root.Children))
	}
	if root.Children[0].Name != "step 1" || root.Children[1].Name != "step 2" {
		t.Fatalf("children = [%s, %s], want [step 1, step 2]", root.Children[0].Name, root.Children[1].Name)
	}
}
