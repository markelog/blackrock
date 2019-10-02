package main

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rekki/blackrock/cmd/jubei/disk"
	"github.com/rekki/blackrock/pkg/depths"
)

type QueryRequest struct {
	Query interface{} `json:"query"`
	Size  int         `json:"size"`
	From  string      `json:"from"`
	To    string      `json:"to"`
}

type QueryResponse struct {
	Total int64 `json:"total"`
	Hits  []Hit `json:"hits"`
}

func (qr *QueryResponse) String(c *gin.Context) {
	c.YAML(200, qr)
}

func (qr *QueryResponse) HTML(c *gin.Context) {
	c.YAML(200, qr)
}

func NewTermQuery(root string, tagKey string, tagValue string, c *disk.CompactIndexCache) Query {
	tagKey = depths.Cleanup(strings.ToLower(tagKey))
	tagValue = depths.Cleanup(strings.ToLower(tagValue))
	s := fmt.Sprintf("%s:%s", tagKey, tagValue)
	return NewTerm(s, c.FindPostingsList(root, tagKey, tagValue))
}
