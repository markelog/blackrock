package main

import (
	"testing"
)

func postingsList(n int) []int64 {
	list := make([]int64, n)
	for i := 0; i < n; i++ {
		list[i] = int64(i) * 3
	}
	return list
}

func query(query Query) []int64 {
	out := []int64{}
	for query.Next() != NO_MORE {
		out = append(out, query.GetDocId())
	}
	return out
}

func eq(t *testing.T, a, b []int64) {
	if len(a) != len(b) {
		t.Logf("len(a) != len(b) ; len(a) = %d, len(b) = %d", len(a), len(b))
		t.FailNow()
	}

	for i, _ := range a {
		if a[i] != b[i] {
			t.Log("a[i] != b[i]")
			t.FailNow()
		}
	}
}

func BenchmarkNext1000(b *testing.B) {
	x := postingsList(1000)

	for n := 0; n < b.N; n++ {
		sum := int64(0)
		q := NewTerm("", x)
		for q.Next() != NO_MORE {
			sum += q.GetDocId()
		}
	}
}

func BenchmarkOr1000(b *testing.B) {
	x := postingsList(1000)
	y := postingsList(1000)

	for n := 0; n < b.N; n++ {
		sum := int64(0)
		q := NewBoolOrQuery(
			NewTerm("x", x),
			NewTerm("y", y),
		)

		for q.Next() != NO_MORE {
			sum += q.GetDocId()
		}
	}
}

func BenchmarkAnd1000(b *testing.B) {
	x := postingsList(1000000)
	y := postingsList(1000)

	for n := 0; n < b.N; n++ {
		sum := int64(0)
		q := NewBoolAndQuery(
			NewTerm("x", x),
			NewTerm("y", y),
		)

		for q.Next() != NO_MORE {
			sum += q.GetDocId()
		}
	}
}

func TestModify(t *testing.T) {
	a := postingsList(100)
	b := postingsList(1000)
	c := postingsList(10000)
	d := postingsList(100000)
	e := postingsList(1000000)

	eq(t, a, query(NewTerm("x", a)))
	eq(t, b, query(NewTerm("x", b)))
	eq(t, c, query(NewTerm("x", c)))
	eq(t, d, query(NewTerm("x", d)))
	eq(t, e, query(NewTerm("x", e)))

	eq(t, b, query(NewBoolOrQuery(
		NewTerm("x", a),
		NewTerm("x", b),
	)))

	eq(t, c, query(NewBoolOrQuery(
		NewTerm("x", a),
		NewTerm("x", b),
		NewTerm("x", c),
	)))

	eq(t, e, query(NewBoolOrQuery(
		NewTerm("x", a),
		NewTerm("x", b),
		NewTerm("x", c),
		NewTerm("x", d),
		NewTerm("x", e),
	)))

	eq(t, a, query(NewBoolAndQuery(
		NewTerm("x", a),
		NewTerm("x", b),
		NewTerm("x", c),
		NewTerm("x", d),
		NewTerm("x", e),
	)))

	eq(t, a, query(NewBoolAndQuery(
		NewTerm("x", a),
		NewTerm("x", b),
		NewTerm("x", c),
		NewTerm("x", d),
		NewTerm("x", e),
	)))

	eq(t, b, query(NewBoolAndQuery(
		NewBoolOrQuery(
			NewTerm("x", a),
			NewTerm("x", b),
		),
		NewTerm("x", b),
		NewTerm("x", c),
		NewTerm("x", d),
		NewTerm("x", e),
	)))

	eq(t, c, query(NewBoolAndQuery(
		NewBoolOrQuery(
			NewTerm("x", a),
			NewTerm("x", b),
			NewBoolAndQuery(
				NewTerm("x", c),
				NewTerm("x", d),
			),
		),
		NewTerm("x", d),
		NewTerm("x", e),
	)))
}
