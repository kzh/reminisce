package vm

import (
)

type TokenType int

const (
    TOKEN_IDENTIFIER TokenType = iota

    TOKEN_HEX
    TOKEN_DEC
    TOKEN_STRING
    TOKEN_CHARACTER

    TOKEN_OPERATOR
    TOKEN_NEWLINE
)

type Token struct {
    Type    TokenType
    Line    int
    Content string
}

func Lex(s string) (t []*Token) {
    line := 0

    for i := 0; i != len(s); i++ {
        if i < len(s) && s[i] == ' ' {
            continue
        } else if i < len(s) && s[i] == ';' {
            for s[i] != '\n' {
                i++
            }
            i--
        } else if i < len(s) && IsLetter(s[i]) || s[i] == '.' {
            content := []byte{s[i]}
            i++

            for i < len(s) && (IsLetter(s[i]) || IsDecimal(s[i])) {
                content = append(content, s[i])
                i++
            }

            i--

            t = append(t, &Token{
                Type: TOKEN_IDENTIFIER,
                Line: line,
                Content: string(content),
            })
        } else if i < len(s) && s[i] == '"' {
            content := []byte{}
            i++

            for s[i] != '"' {
                content = append(content, s[i])
                i++
            }

            i--

            t = append(t, &Token{
                Type: TOKEN_STRING,
                Line: line,
                Content: string(content),
            })
        } else if i < len(s) && s[i] == '\'' {
            content := []byte{}
            i++

            for s[i] != '\'' {
                content = append(content, s[i])
                i++
            }

            i--

            t = append(t, &Token{
                Type: TOKEN_CHARACTER,
                Line: line,
                Content: string(content),
            })
        } else if i < len(s) && (s[i] == ':' || s[i] == '[' || s[i] == ']' || s[i] == '+' || s[i] == '-' || s[i] == ',' || s[i] == '*')  {
            t = append(t, &Token{
                Type: TOKEN_OPERATOR,
                Line: line,
                Content: string(s[i]),
            })
        } else if i < len(s) && s[i] == '\n' {
            line++
            t = append(t, &Token{
                Type: TOKEN_NEWLINE,
                Line: line,
                Content: string('\n'),
            })
        } else if i < len(s) && s[i] == '0' && s[i + 1] == 'x' {
            content := []byte{'0', 'x'}
            i += 2;

            for i < len(s) && (IsDecimal(s[i]) || IsLetter(s[i])) {
                content = append(content, s[i])
                i++
            }

            i--

            t = append(t, &Token{
                Type: TOKEN_HEX,
                Line: line,
                Content: string(content),
            })
        } else if i < len(s) && IsDecimal(s[i]) {
            content := []byte{s[i]}
            i++

            for IsDecimal(s[i]) {
                content = append(content, s[i])
                i++
            }

            i--

            t = append(t, &Token{
                Type: TOKEN_DEC,
                Line: line,
                Content: string(content),
            })
        }

    }

    return
}

func IsLetter(r byte) bool {
    return (r >= 'a' && r <= 'z') || (r >= 'A' && r <='Z') || r == '_'
}

func IsDecimal(r byte) bool {
    return r >= '0' && r <= '9'
}
