// Copyright 2016 Evans. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gsebleve

import (
	"strings"

	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/registry"
	"github.com/go-creed/sat"
	"github.com/go-ego/gse"
)

var Dicter = sat.DefaultDict()

var (
	TokenName    = "gse"
	SeparateName = "sep"
)

// GseCut gse cut token structure
type GseCut struct {
	seg *gse.Segmenter
	// stop string
	opt  string
	trim string
}

// Separator type separator tokenizer struct
type Separator struct {
	seg  *gse.Segmenter
	sep  string
	trim string
}

// NewSep create a separator tokenizer
func NewSep(sep, trim string) (*Separator, error) {
	var seg gse.Segmenter
	seg.Dict = gse.NewDict()
	seg.Init()
	return &Separator{&seg, sep, trim}, nil
}

// NewGseCut create a gse cut tokenizer
func NewGse(dicts, stop, opt, trim string, alpha bool) (*GseCut, error) {
	var (
		seg gse.Segmenter
		err error
	)

	seg.SkipLog = true
	if alpha {
		seg.AlphaNum = true
	}

	if dicts == "" {
		dicts = "zh"
	}

	if strings.Contains(dicts, "embed") {
		dicts = strings.Replace(dicts, "embed, ", "", 1)
		err = seg.LoadDictEmbed(dicts)
	} else {
		err = seg.LoadDict(dicts)
	}

	if stop != "" {
		if strings.Contains(stop, "embed") {
			stop = strings.Replace(stop, "embed, ", "", 1)
			seg.LoadStopEmbed(stop)
		} else {
			seg.LoadStop(stop)
		}
	}
	return &GseCut{&seg, opt, trim}, err
}

// Trim trim the unused token string
func (c *GseCut) Trim(s []string) []string {
	return Trim(s, c.trim, c.seg)
}

// Trim trim the unused token string
func (c *Separator) Trim(s []string) []string {
	return Trim(s, c.trim, c.seg)
}

// Trim trim the unused token string
func Trim(s []string, trim string, seg *gse.Segmenter) []string {
	if trim == "symbol" {
		return seg.TrimSymbol(s)
	}

	if trim == "punct" {
		return seg.TrimPunct(s)
	}

	if trim == "trim" {
		return seg.Trim(s)
	}

	return s
}

// Cut option the gse cut mode
func (c *GseCut) Cut(text string, opt string) []string {
	if c.trim == "html" {
		return c.seg.CutTrimHtml(text)
	}

	if c.trim == "url" {
		return c.seg.CutUrl(text)
	}

	if opt == "search-hmm" {
		return c.seg.CutSearch(text, true)
	}
	if opt == "search" {
		return c.seg.CutSearch(text)
	}

	if opt == "search-dag" {
		return c.seg.CutSearch(text, false)
	}

	if opt == "all" {
		return c.seg.CutAll(text)
	}

	if opt == "hmm" {
		return c.seg.Cut(text, true)
	}

	if opt == "dag" {
		return c.seg.Cut(text, false)
	}

	return c.seg.Cut(text)
}

// Tokenize cut the text to bleve token stream
func (c *GseCut) Tokenize(text []byte) analysis.TokenStream {
	result := make(analysis.TokenStream, 0)
	t1 := string(text)
	t1 = Dicter.Read(t1)
	cuts := c.Trim(c.Cut(t1, c.opt))

	azs := c.seg.Analyze(cuts, t1)
	for _, az := range azs {
		token := analysis.Token{
			Term:  []byte(az.Text),
			Start: az.Start,
			End:   az.End,

			Position: az.Position,
			Type:     analysis.Ideographic,
		}
		result = append(result, &token)
	}
	return result
}

// Tokenize cut the text to bleve token stream
func (s *Separator) Tokenize(text []byte) analysis.TokenStream {
	result := make(analysis.TokenStream, 0)
	t1 := string(text)
	t1 = Dicter.Read(t1)
	cuts := s.Trim(strings.Split(t1, s.sep))

	azs := s.seg.Analyze(cuts, t1)
	for _, az := range azs {
		token := analysis.Token{
			Term:     []byte(az.Text),
			Start:    az.Start,
			End:      az.End,
			Position: az.Position,
			Type:     analysis.Ideographic,
		}
		result = append(result, &token)
	}
	return result
}

func tokenizerConstructor(config map[string]interface{}, cache *registry.Cache) (analysis.Tokenizer, error) {
	dicts, ok := config["dicts"].(string)
	if !ok {
		dicts = ""
	}
	stop, ok := config["stop"].(string)
	if !ok {
		stop = ""
	}

	opt, ok := config["opt"].(string)
	if !ok || opt == "" {
		opt = ""
	}

	trim, ok := config["trim"].(string)
	if !ok {
		trim = ""
	}

	alpha, ok := config["alpha"].(bool)
	if !ok {
		alpha = false
	}

	return NewGse(dicts, stop, opt, trim, alpha)
}

func tokenizerConstructor2(config map[string]interface{}, cache *registry.Cache) (analysis.Tokenizer, error) {
	sep, ok := config["sep"].(string)
	if !ok {
		sep = " "
	}

	trim, ok := config["trim"].(string)
	if !ok {
		trim = ""
	}

	return NewSep(sep, trim)
}

func init() {
	registry.RegisterTokenizer(TokenName, tokenizerConstructor)
	registry.RegisterTokenizer(SeparateName, tokenizerConstructor2)
}
