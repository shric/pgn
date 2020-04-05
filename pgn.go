package pgn

import (
	"fmt"
	"io"
	"strings"
	"text/scanner"
)

type PGNScanner struct {
	s scanner.Scanner
}

type Game struct {
	Moves []Move
	Tags  map[string]string
}

type Move struct {
	From    Position
	To      Position
	Promote Piece
}

func (m Move) String() string {
	if m.Promote == NoPiece {
		return fmt.Sprintf("%v%v", m.From, m.To)
	}
	return fmt.Sprintf("%v%v%v", m.From, m.To, m.Promote)
}

var (
	NilMove Move = Move{From: NoPosition, To: NoPosition}
)

func ParseGame(s *scanner.Scanner) (*Game, error) {
	g := Game{Tags: map[string]string{}, Moves: []Move{}}
	err := ParseTags(s, &g)
	if err != nil {
		return nil, err
	}
	err = ParseMoves(s, &g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func ParseTags(s *scanner.Scanner, g *Game) error {
	//fmt.Println("starting tags parse")
	run := s.Peek()
	for run != scanner.EOF {
		switch run {
		case '[', ']', '\n', '\r', ' ':
			run = s.Next()
		case '0':
			return nil
		case '1':
			return nil
		default:
			s.Scan()
			tag := s.TokenText()
			s.Scan()
			val := s.TokenText()
			//fmt.Println("tag:", tag, "; val:", val)
			g.Tags[tag] = strings.Trim(val, "\"")
		}
		run = s.Peek()
	}
	return nil
}

func isEnd(str string) bool {
	if str == "1/2-1/2" {
		return true
	}
	if str == "0-1" {
		return true
	}
	if str == "1-0" {
		return true
	}
	if str == "*" {
		return true
	}
	return false
}

func ParseMoves(s *scanner.Scanner, g *Game) error {
	//fmt.Println("starting moves parse")
	s.Mode = scanner.ScanIdents | scanner.ScanChars | scanner.ScanInts | scanner.ScanStrings
	run := s.Peek()
	board := NewBoard()
	var err error
	if len(g.Tags["FEN"]) > 0 {
		board, err = NewBoardFEN(g.Tags["FEN"])
		if err != nil {
			return err
		}
	}
	num := ""
	white := ""
	black := ""
	for run != scanner.EOF {
		switch run {
		case '(':
			for run != ')' && run != scanner.EOF {
				run = s.Next()
			}
		case '{':
			for run != '}' && run != scanner.EOF {
				run = s.Next()
			}
		case '#', '.', '+', '!', '?', '\n', '\r':
			run = s.Next()
			run = s.Peek()
		default:
			s.Scan()
			if s.TokenText() == "{" {
				run = '{'
				continue
			}
			if num == "" {
				num = s.TokenText()
				for s.Peek() == '-' {
					s.Scan()
					num += s.TokenText()
					s.Scan()
					num += s.TokenText()
				}
				for s.Peek() == '/' {
					s.Scan()
					num += s.TokenText()
					s.Scan()
					num += s.TokenText()
					s.Scan()
					num += s.TokenText()
					s.Scan()
					num += s.TokenText()
					s.Scan()
					num += s.TokenText()
					s.Scan()
					num += s.TokenText()
				}
				if isEnd(num) {
					return nil
				}
			} else if white == "" {
				white = ParseMoveOrResult(s)
				if isEnd(white) {
					return nil
				}
				for s.Peek() == ' ' {
					s.Next()
				}
				if s.Peek() == '{' {
					for s.Peek() != '}' {
						s.Next()
					}
					s.Next()
				}
				move, err := board.MoveFromAlgebraic(white, White)
				if err != nil {
					fmt.Println(board)
					return err
				}
				g.Moves = append(g.Moves, move)
				board.MakeMove(move)

				// Sometimes whites move is followed by the move number and '...'
				// eg. `1. e4 {long comment} 1... e5
				for s.Peek() == ' ' {
					s.Next()
				}
				if s.Peek() >= '0' && s.Peek() <= '9' {
					s.Scan()
					numReply := ParseMoveOrResult(s)
					if isEnd(numReply) {
						return nil
					}
				}
			} else if black == "" {
				black = ParseMoveOrResult(s)
				if isEnd(black) {
					return nil
				}
				move, err := board.MoveFromAlgebraic(black, Black)
				if err != nil {
					fmt.Println(board)
					return err
				}
				g.Moves = append(g.Moves, move)
				board.MakeMove(move)
				num = ""
				white = ""
				black = ""
			}
			run = s.Peek()
		}
	}
	return nil
}

func ParseMoveOrResult(s *scanner.Scanner) string {
	result := s.TokenText()
	for s.Peek() == '-' {
		s.Scan()
		result += s.TokenText()
		s.Scan()
		result += s.TokenText()
	}
	for s.Peek() == '/' {
		s.Scan()
		result += s.TokenText()
		s.Scan()
		result += s.TokenText()
		s.Scan()
		result += s.TokenText()
		s.Scan()
		result += s.TokenText()
		s.Scan()
		result += s.TokenText()
		s.Scan()
		result += s.TokenText()
	}
	if s.Peek() == '=' {
		s.Scan()
		result += s.TokenText()
		s.Scan()
		result += s.TokenText()
	}
	for {
		switch s.Peek() {
		case '#', '.', '+', '!', '?', '\n', '\r':
			s.Next()
		default:
			return result
		}
	}
}

func NewPGNScanner(r io.Reader) *PGNScanner {
	s := scanner.Scanner{}
	s.Init(r)
	return &PGNScanner{s: s}
}

func (ps *PGNScanner) Next() bool {
	if ps.s.Peek() == scanner.EOF {
		return false
	}
	return true
}

func (ps *PGNScanner) Scan() (*Game, error) {
	game, err := ParseGame(&ps.s)
	//fmt.Println(game)
	return game, err
}
