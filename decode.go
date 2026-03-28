package vdf

import (
	"fmt"
	"io"
)

type options struct {
	ignoreWhitespace    bool
	usesEscapeSequences bool
}

type Option func(*options)

func UseEscapeSequences() Option {
	return func(o *options) {
		o.usesEscapeSequences = true
	}
}

// Decoder reads and decodes VDF from an input stream.
type Decoder struct {
	r    io.Reader
	opts options
}

// NewDecoder returns a new Decoder that reads from r.
func NewDecoder(r io.Reader, opts ...Option) *Decoder {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}
	return &Decoder{r: r, opts: o}
}

// Decode reads VDF from the input and stores the result in v.
func (d *Decoder) Decode(doc *Document) error {
	if doc == nil {
		return fmt.Errorf("doc cannot be nil")
	}

	data, err := io.ReadAll(d.r)
	if err != nil {
		return err
	}

	p := &parser{lexer: newLexer(data, d.opts.usesEscapeSequences)}
	result, err := p.parse()
	if err != nil {
		return err
	}

	*doc = *result
	return nil
}

// Unmarshal parses VDF-encoded data and stores the result in doc.
func Unmarshal(data []byte, doc *Document, opts ...Option) error {
	if doc == nil {
		return fmt.Errorf("doc cannot be nil")
	}

	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	p := &parser{lexer: newLexer(data, o.usesEscapeSequences)}
	result, err := p.parse()
	if err != nil {
		return err
	}

	*doc = *result
	return nil
}
