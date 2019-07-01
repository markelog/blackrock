package main

import (
	"math"
)

const (
	NO_MORE   = int64(math.MaxInt64)
	NOT_READY = int64(-1)
)

type Query interface {
	advance(int64) int64
	Next() int64
	GetDocId() int64
	Score() float32
}

type QueryBase struct {
	docId int64
}

func (q *QueryBase) GetDocId() int64 {
	return q.docId
}

type Term struct {
	cursor   int
	postings []int64
	QueryBase
}

func NewTerm(postings []int64) *Term {
	return &Term{
		cursor:    -1,
		postings:  postings,
		QueryBase: QueryBase{NOT_READY},
	}

}
func (t *Term) Score() float32 {
	return float32(1)
}
func (t *Term) advance(target int64) int64 {
	if t.docId == NO_MORE || t.docId == target || target == NO_MORE {
		t.docId = target
		return t.docId
	}
	if t.cursor < 0 {
		t.cursor = 0
	}

	start := t.cursor
	end := len(t.postings)

	for start < end {
		mid := start + ((end - start) / 2)
		current := t.postings[mid]
		if current == target {
			t.cursor = mid
			t.docId = target
			return target
		}

		if current < target {
			start = mid + 1
		} else {
			end = mid
		}
	}
	if start >= len(t.postings) {
		t.docId = NO_MORE
		return NO_MORE
	}
	t.cursor = start
	t.docId = t.postings[start]
	return t.docId
}

func (t *Term) Next() int64 {
	t.cursor++
	if t.cursor >= len(t.postings) {
		t.docId = NO_MORE
	} else {
		t.docId = t.postings[t.cursor]
	}
	return t.docId
}

type BoolQueryBase struct {
	queries []Query
}

func (q *BoolQueryBase) AddSubQuery(sub Query) {
	q.queries = append(q.queries, sub)
}

type BoolOrQuery struct {
	BoolQueryBase
	QueryBase
}

func NewBoolOrQuery(queries ...Query) *BoolOrQuery {
	return &BoolOrQuery{
		BoolQueryBase: BoolQueryBase{queries},
		QueryBase:     QueryBase{NOT_READY},
	}
}

func (q *BoolOrQuery) Score() float32 {
	score := 0
	n := len(q.queries)
	for i := 0; i < n; i++ {
		sub_query := q.queries[i]
		if sub_query.GetDocId() == q.docId {
			score++
		}
	}
	return float32(score)
}

func (q *BoolOrQuery) advance(target int64) int64 {
	new_doc := NO_MORE
	n := len(q.queries)
	for i := 0; i < n; i++ {
		sub_query := q.queries[i]
		cur_doc := sub_query.GetDocId()
		if cur_doc < target {
			cur_doc = sub_query.advance(target)
		}

		if cur_doc < new_doc {
			new_doc = cur_doc
		}
	}
	q.docId = new_doc
	return q.docId
}

func (q *BoolOrQuery) Next() int64 {
	new_doc := NO_MORE
	n := len(q.queries)
	for i := 0; i < n; i++ {
		sub_query := q.queries[i]
		cur_doc := sub_query.GetDocId()
		if cur_doc == q.docId {
			cur_doc = sub_query.Next()
		}

		if cur_doc < new_doc {
			new_doc = cur_doc
		}
	}
	q.docId = new_doc
	return new_doc
}

type BoolAndQuery struct {
	BoolQueryBase
	QueryBase
}

func NewBoolAndQuery(queries ...Query) *BoolAndQuery {
	return &BoolAndQuery{
		BoolQueryBase: BoolQueryBase{queries},
		QueryBase:     QueryBase{NOT_READY},
	}
}
func (q *BoolAndQuery) Score() float32 {
	return float32(len(q.queries))
}

func (q *BoolAndQuery) nextAndedDoc(target int64) int64 {
	// initial iteration skips queries[0]
	n := len(q.queries)
	for i := 1; i < n; i++ {
		sub_query := q.queries[i]
		if sub_query.GetDocId() < target {
			sub_query.advance(target)
		}

		if sub_query.GetDocId() == target {
			continue
		}

		target = q.queries[0].advance(sub_query.GetDocId())
		i = 0 //restart the loop from the first query
	}
	q.docId = target
	return q.docId
}

func (q *BoolAndQuery) advance(target int64) int64 {
	if len(q.queries) == 0 {
		q.docId = NO_MORE
		return NO_MORE
	}

	return q.nextAndedDoc(q.queries[0].advance(target))
}

func (q *BoolAndQuery) Next() int64 {
	if len(q.queries) == 0 {
		q.docId = NO_MORE
		return NO_MORE
	}

	// XXX: pick cheapest leading query
	return q.nextAndedDoc(q.queries[0].Next())
}
