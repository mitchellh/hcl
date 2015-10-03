package parser

import (
	"bytes"
	"io"
	"io/ioutil"
	"unicode"
)

// eof represents a marker rune for the end of the reader.
const eof = rune(0)

// Lexer defines a lexical scanner
type Scanner struct {
	src      *bytes.Buffer
	srcBytes []byte

	ch          rune // current character
	lastCharLen int  // length of last character in bytes
	pos         Position

	// Token text buffer
	tokBuf bytes.Buffer
	tokPos int // token text tail position (srcBuf index); valid if >= 0
	tokEnd int // token text tail end (srcBuf index)
}

// NewLexer returns a new instance of Lexer. Even though src is an io.Reader,
// we fully consume the content.
func NewLexer(src io.Reader) (*Scanner, error) {
	buf, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(buf)
	return &Scanner{
		src:      b,
		srcBytes: b.Bytes(),
	}, nil
}

// next reads the next rune from the bufferred reader. Returns the rune(0) if
// an error occurs (or io.EOF is returned).
func (s *Scanner) next() rune {
	var err error
	var size int
	s.ch, size, err = s.src.ReadRune()
	if err != nil {
		return eof
	}

	s.lastCharLen = size
	s.pos.Offset += size
	s.pos.Column += size

	if s.ch == '\n' {
		s.pos.Line++
		s.pos.Column = 0
	}

	return s.ch
}

// Scan scans the next token and returns the token and it's literal string.
func (s *Scanner) Scan() (tok Token, lit string) {
	ch := s.next()

	// skip white space
	for isWhitespace(ch) {
		ch = s.next()
	}

	// start the token position
	s.tokBuf.Reset()
	s.tokPos = s.pos.Offset - s.lastCharLen

	// identifier
	if isLetter(ch) {
		s.scanIdentifier()
		tok = IDENT
	}

	if isDigit(ch) {
		// scan for number
	}

	switch ch {
	case eof:
		tok = EOF
	}

	s.tokEnd = s.pos.Offset - s.lastCharLen

	return tok, s.TokenLiteral()
}

func (s *Scanner) scanIdentifier() {
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
}

// TokenLiteral returns the literal string corresponding to the most recently
// scanned token.
func (s *Scanner) TokenLiteral() string {
	if s.tokPos < 0 {
		// no token text
		return ""
	}

	// part of the token text was saved in tokBuf: save the rest in
	// tokBuf as well and return its content
	s.tokBuf.Write(s.srcBytes[s.tokPos:s.tokEnd])
	s.tokPos = s.tokEnd // ensure idempotency of TokenText() call
	return s.tokBuf.String()
}

// Pos returns the position of the character immediately after the character or
// token returned by the last call to Next or Scan.
func (s *Scanner) Pos() Position {
	return Position{}
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= 0x80 && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

// isWhitespace returns true if the rune is a space, tab, newline or carriage return
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}