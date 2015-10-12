package parser

import (
	"errors"
	"fmt"

	"github.com/fatih/hcl/scanner"
)

type Parser struct {
	sc *scanner.Scanner

	tok     scanner.Token // last read token
	prevTok scanner.Token // previous read token

	enableTrace bool
	indent      int
	n           int // buffer size (max = 1)
}

func New(src []byte) *Parser {
	return &Parser{
		sc: scanner.New(src),
	}
}

// Parse returns the fully parsed source and returns the abstract syntax tree.
func (p *Parser) Parse() (Node, error) {
	defer un(trace(p, "ParseSource"))
	node := &Source{}

	for {
		n, err := p.parseNode()
		if err != nil {
			return nil, err
		}

		node.add(n)

		// break if we hit the end
		if p.tok.Type == scanner.EOF {
			break
		}
	}

	return node, nil
}

func (p *Parser) parseNode() (Node, error) {
	defer un(trace(p, "ParseNode"))

	tok := p.scan()

	fmt.Println(tok) // debug

	if tok.Type.IsLiteral() {
		if p.prevTok.Type.IsLiteral() {
			return p.parseObjectType()
		}

		tok := p.scan()
		if tok.Type == scanner.ASSIGN {
			return p.parseAssignment()
		}

		p.unscan()
		return p.parseIdent()
	}

	return nil, errors.New("not yet implemented")
}

func (p *Parser) parseAssignment() (*AssignStatement, error) {
	defer un(trace(p, "ParseAssignment"))
	a := &AssignStatement{
		lhs: &Ident{
			token: p.prevTok,
		},
		assign: p.tok.Pos,
	}

	n, err := p.parseNode()
	if err != nil {
		return nil, err
	}

	a.rhs = n
	return a, nil
}

func (p *Parser) parseIdent() (*Ident, error) {
	defer un(trace(p, "ParseIdent"))

	if !p.tok.Type.IsLiteral() {
		return nil, errors.New("can't parse non literal token")
	}

	return &Ident{
		token: p.tok,
	}, nil
}

func (p *Parser) parseObjectType() (*ObjectStatement, error) {
	return nil, errors.New("ObjectStatement is not implemented yet")
}

func (p *Parser) parseListType() (*ListType, error) {
	return nil, errors.New("ListStatement is not implemented yet")
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() scanner.Token {
	// If we have a token on the buffer, then return it.
	if p.n != 0 {
		p.n = 0
		return p.tok
	}

	// store previous token
	p.prevTok = p.tok

	// Otherwise read the next token from the scanner and Save it to the buffer
	// in case we unscan later.
	p.tok = p.sc.Scan()
	return p.tok
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() {
	p.n = 1
	p.tok = p.prevTok
}

// ----------------------------------------------------------------------------
// Parsing support

func (p *Parser) printTrace(a ...interface{}) {
	if !p.enableTrace {
		return
	}

	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = len(dots)
	fmt.Printf("%5d:%3d: ", p.tok.Pos.Line, p.tok.Pos.Column)

	i := 2 * p.indent
	for i > n {
		fmt.Print(dots)
		i -= n
	}
	// i <= n
	fmt.Print(dots[0:i])
	fmt.Println(a...)
}

func trace(p *Parser, msg string) *Parser {
	p.printTrace(msg, "(")
	p.indent++
	return p
}

// Usage pattern: defer un(trace(p, "..."))
func un(p *Parser) {
	p.indent--
	p.printTrace(")")
}
