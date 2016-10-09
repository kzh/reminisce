package vm

import (
//    "log"
    "strconv"
)

type Parser struct {
    index int

    currentLabel string

    tokens []*Token
}

type Register struct {
    Location, Size int
}

var registers map[string]Register = map[string]Register{
    "rax": {0, 8}, "rbx": {8, 8}, "rcx": {16, 8}, "rdx": {24, 8}, "rsi": {32, 8}, "rdi": {40, 8}, "rbp": {48, 8}, "rsp": {56, 8}, "r8": {64, 8}, "r9":  {72, 8}, "r10": {80, 8}, "r11": {88, 8}, "r12": {96, 8}, "r13": {104, 8}, "r14": {112, 8}, "r15": {120, 8},
    "eax": {4, 4}, "ebx": {12, 4}, "ecx": {20, 4}, "edx": {28, 4}, "esi": {36, 4}, "edi": {44, 4}, "ebp": {52, 4}, "esp": {60, 4}, "r8d": {68, 4}, "r9d": {76, 4}, "r10d": {84, 4}, "r11d": {92, 4}, "r12d": {100, 4}, "r13d": {108, 4}, "r14d": {116, 4}, "r15d": {124, 4},
    "ax": {6, 2}, "bx": {14, 2}, "cx": {22, 2}, "dx": {30, 2}, "si": {38, 2}, "di": {46, 2}, "bp": {54, 2}, "sp": {62, 2}, "r8w": {70, 2}, "r9w": {78, 2}, "r10w": {86, 2}, "r11w": {94, 2}, "r12w": {102, 2}, "r13w": {110, 2}, "r14w": {118, 2}, "r15w": {126, 2},
    "al": {7, 1}, "bl": {15, 1}, "cl": {23, 1}, "dl": {31, 1}, "sil": {39, 1}, "dil": {47, 1}, "bpl": {55, 1}, "spl": {63, 1}, "r8b": {71, 1}, "r9b": {79, 1}, "r10b": {87, 1}, "r11b": {95, 1}, "r12b": {103, 1}, "r13b": {111, 1}, "r14b": {119, 1}, "r15b": {127, 1},
    "ah": {5, 1}, "bh": {13, 1}, "ch": {21, 1}, "dh": {29, 1},
}

func (p *Parser) peek(ahead int) *Token{
    if p.index >= len(p.tokens) {
        return nil
    } else {
        return p.tokens[p.index]
    }
}

func (p *Parser) consume(amount int) {
    p.index += amount
}

func Parse(tokens []*Token) (nodes []Node) {
    parser := &Parser{
        index: 0,
        tokens: tokens,
    }

    for parser.index < len(tokens) {
        if parser.peek(0).Content == "section" {
            parser.consume(1)

            if parser.peek(0).Content == ".text" {
                nodes = append(nodes, parser.parseSectionTextNode())
            } else if parser.peek(0).Content == ".data" {

            }
        }
    }

    return nodes
}

func (p *Parser) parseSectionTextNode() (n *SectionTextNode) {
    p.consume(2)

    n = &SectionTextNode{
        Instructions: []*InstructionNode{},
        Labels: make(map[string]int),
    }

    for p.index < len(p.tokens) {
        for p.index < len(p.tokens) && p.peek(0).Content == "\n" {
            p.consume(1)
        }

        if p.index >= len(p.tokens) {
            break
        }

        name := p.peek(0).Content
        line := p.peek(0).Line
        p.consume(1)

        if p.peek(0) != nil && p.peek(0).Content == ":" {
            if name[0] == '.' {
                name = p.currentLabel + name
            } else {
                p.currentLabel = name
            }

            n.Labels[name] =  len(n.Instructions);
            p.consume(2);

//            log.Println("Found label decl")
        } else {
            instruc := &InstructionNode{
                Name: name,
            }
            instruc.SetLocation(line + 1)

            for p.peek(0) != nil && p.peek(0).Content != "\n" {
                if n := p.parseArgument(); n != nil {
                    instruc.Arguments = append(instruc.Arguments, n)
                }

                if p.peek(0) == nil {
                    break
                }

                if p.peek(0).Content == "," {
                    p.consume(1)
                }

                if p.peek(0).Content == "\n" {
                    p.consume(1)
                    break;
                }
            }

            n.Instructions = append(n.Instructions, instruc)
        }
    }

    return
}

func (p *Parser) parseArgument() (n Node) {
    if IsSize(p.peek(0).Content) {
        n = &AccessNode{
            Size: GetSize(p.peek(0).Content),
        }
        p.consume(1)
        n.(*AccessNode).Expression = p.parseArgument().(*AccessNode).Expression
//        log.Println("Found expression arg")
    } else if p.peek(0).Content == "[" {
        p.consume(1)
        if n == nil {
            n = &AccessNode{Size: 8}
        }

        expr := p.parseExpression()
        if binop := p.parseBinopExpression(expr, 0); binop != nil {
            expr = binop
        }

        n.(*AccessNode).Expression = expr
        p.consume(1)
//        log.Println("Found expression arg")
    } else if c := p.parseConstants(); c != nil {
//        log.Println("Found constant arg")
        n = c
    } else if r := p.parseRegister(); r != nil {
//        log.Println("Found register arg")
        n = r
    } else {
//        log.Println("Found label arg" + p.peek(0).Content)
        labelName := p.peek(0).Content
        if labelName[0] == '.' {
            labelName = p.currentLabel + labelName
        }

        n = &LabelNode{Name: labelName}
        p.consume(1)
    }

    return
}

var precedence map[string]int = map[string]int {
    "+": 0, "-": 0, "*": 1,
}

func (p *Parser) parseExpression() (n Node) {
    var left Node
    if c := p.parseConstants(); c != nil {
        left = c
    } else if r := p.parseRegister(); r != nil {
        left = r
    } else {
        return nil
    }

    return left
}

func (p *Parser) parseBinopExpression(left Node, high int) (n Node) {
    operator := p.peek(0).Content
    if operator != "+" && operator != "-" && operator != "*" {
        return nil
    }
    p.consume(1)


    right := p.parseExpression()
    if right == nil {
        return nil
    }

    n = &BinopExprNode{Operator: operator, Left: left, Right:right}
    return
}

func (p *Parser) parseRegister() (n *RegisterNode) {
    if _, ok := registers[p.peek(0).Content]; ok {

        n = &RegisterNode{Register: p.peek(0).Content}
        p.consume(1)
    }

    return
}

func (p *Parser) parseConstants() (n *ConstantNode) {
    if (p.peek(0).Type == TOKEN_DEC) {
        i, _ := strconv.Atoi(p.peek(0).Content)
        n = &ConstantNode{Value: int64(i)}
        p.consume(1)
    } else if (p.peek(0).Type == TOKEN_HEX) {
        i, _ := strconv.ParseInt(p.peek(0).Content, 0, 64)
        n = &ConstantNode{Value: i}
        p.consume(1)
    }

    return
}

var sizes map[string]int = map[string]int {
    "byte": 1, "word": 2, "dword": 4, "qword": 8,
}

func IsSize(s string) bool {
     _, ok := sizes[s]
    return ok
}

func GetSize(s string) int {
    return sizes[s];
}
