package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackdoe/blackrock/depths"
	"github.com/jackdoe/blackrock/jubei/disk"
	log "github.com/sirupsen/logrus"
)

type ExpQueryRequest struct {
	Exp           string         `json:"exp"`
	Cohort        map[string]int `json:"cohort"`
	Query         interface{}    `json:"query"`
	Variants      int            `json:"variants"`
	ExperimentKey string         `json:"key"`
	From          string         `json:"from"`
	To            string         `json:"to"`
}

type QueryRequest struct {
	Query          interface{} `json:"query"`
	Size           int         `json:"size"`
	DecodeMetadata bool        `json:"decode_metadata"`
	From           string      `json:"from"`
	To             string      `json:"to"`
}

type QueryResponse struct {
	Total int64 `json:"total"`
	Hits  []Hit `json:"hits"`
}

func (qr *QueryResponse) String(c *gin.Context) {
	c.YAML(200, qr)
}

func (qr *QueryResponse) VW(c *gin.Context) {
	labels := map[string]int{}

	for i := -1; i < 10; i++ {
		l := c.Query(fmt.Sprintf("label_%d", i))
		if l != "" {
			labels[l] = i
		}
	}
	if len(labels) == 0 {
		c.JSON(400, gin.H{"error": "no labels found, use ?label_1=some_event_type"})
		return
	}
	w := c.Writer
	for _, hit := range qr.Hits {
		m := hit.Metadata
		if m == nil {
			continue
		}
		label, ok := labels[m.EventType]
		if !ok {
			continue
		}

		w.Write([]byte(fmt.Sprintf("%d |%s %s ", label, hit.ForeignType, depths.CleanupVW(hit.ForeignId))))
		for _, kv := range m.Search {
			w.Write([]byte(fmt.Sprintf("|%s %s ", kv.Key, depths.CleanupVW(kv.Value))))
		}
		for _, kv := range m.Count {
			w.Write([]byte(fmt.Sprintf("|%s %s ", kv.Key, depths.CleanupVW(kv.Value))))
		}
		for _, ctx := range hit.Context {
			for _, kv := range ctx.Properties {
				w.Write([]byte(fmt.Sprintf("|%s_%s %s ", ctx.ForeignType, kv.Key, depths.CleanupVW(kv.Value))))
			}
		}
		w.Write([]byte{'\n'})
	}
}

func (qr *QueryResponse) HTML(c *gin.Context) {
	c.YAML(200, qr)
}

func NewTermQuery(dates []time.Time, inverted *disk.InvertedWriter, dictionary *disk.PersistedDictionary, tagKey string, tagValue string) Query {
	s := fmt.Sprintf("%s:%s", tagKey, tagValue)
	tk, ok := dictionary.Resolve(tagKey)
	if !ok {
		log.Warnf("error reading key for %s", tagKey)
		return NewTerm(s, []int64{})
	}
	queries := []Query{}

	for _, d := range dates {
		seg := depths.SegmentFromNs(d.UnixNano())
		s := fmt.Sprintf("%s:%s:%s", seg, tagKey, tagValue)
		queries = append(queries, NewTerm(s, inverted.Read(seg, tk, tagValue)))
	}

	if len(queries) == 1 {
		return queries[1]
	}
	return NewBoolOrQuery(queries...)
}
