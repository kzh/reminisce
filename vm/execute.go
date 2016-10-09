package vm

import (
    "log"
    "strings"
)

func (p *Process) Simulate(ast []Node) {
    for _, node := range ast {
        switch t := node.(type) {
        case *SectionTextNode:
            p.text = t
            p.handleSectionText(t)
        }
    }
}

func (p *Process) handleSectionText(t *SectionTextNode) {
    for p.currentInstruction < len(t.Instructions) {
        instruction := t.Instructions[p.currentInstruction]

        log.Println(instruction)
        switch instruction.Name {
        case "mov":
            p.handleMovInstruction(instruction)
        case "push":
            p.handlePush(instruction)
        case "pop":
            p.handlePop(instruction)
        case "add":
            p.handleAdd(instruction)
        case "sub":
            p.handleSub(instruction)
        case "jmp":
            p.handleJump(instruction)
        case "cmp":
            p.handleCmp(instruction)
        case "je", "jne", "jl", "jg", "jle", "jge":
            p.handleLogicJump(instruction)
        case "syscall":
            p.handleSyscall(instruction)
        case "call":
            p.handleCall(instruction)
        case "ret":
            p.handleRet(instruction)
        }

        if !p.cmpProtect {
            p.cmpFlags = [7]bool {false,false,false,false,false,false, false}
        } else {
            p.cmpProtect = false;
        }

        p.currentInstruction++
    }
}

//TODO
func (p *Process) handleMovInstruction(instruction *InstructionNode) {
    var source uint64

    switch t := instruction.Arguments[1].(type) {
    case *AccessNode:
        source = uint64(p.evalAccessNode(t.Expression))

        source, _ = p.RetrieveMemory(int(source), 8)
    case *RegisterNode:
        source = p.RetrieveRegister(t.Register)
    case *ConstantNode:
        source = uint64(t.Value)
    }

    switch t := instruction.Arguments[0].(type) {
    case *AccessNode:
        dest := p.evalAccessNode(t.Expression)
        p.InsertMemory(dest, int(source), t.Size)
    case *RegisterNode:
        p.InsertRegister(t.Register, int(source))
    }
}

func (p *Process) evalAccessNode(node Node) (ret int) {
    switch t := node.(type) {
    case *RegisterNode:
        ret = int(p.RetrieveRegister(t.Register))
    case *BinopExprNode:
        ret = p.evalAccessNode(t.Left)
        right := p.evalAccessNode(t.Right)

        if t.Operator == "+" {
            ret += right
        } else if t.Operator == "-" {
            ret -= right
        } else if t.Operator == "*" {
            ret *= right
        }
    case *ConstantNode:
        ret = int(t.Value)
        log.Println(ret)
    }

    return
}

func (p *Process) handlePush(instruction *InstructionNode) {
    push := instruction.Arguments[0].(*RegisterNode).Register;
    rsp := p.RetrieveRegister("rsp")
    val := p.RetrieveRegister(push)

    p.InsertMemory(int(rsp - 8), int(val), 8)
    p.InsertRegister("rsp", int(rsp) - 8)
}

func (p *Process) handlePop(instruction *InstructionNode) {
    pop := instruction.Arguments[0].(*RegisterNode).Register;

    rsp := p.RetrieveRegister("rsp")
    val, _ := p.RetrieveMemory(int(rsp), 8)

    p.InsertRegister(pop, int(val))
    p.InsertRegister("rsp", int(rsp) + 8)
}

func (p *Process) handleAdd(instruction *InstructionNode) {
    var val uint64

    switch t := instruction.Arguments[0].(type) {
    case *RegisterNode:
        val = p.RetrieveRegister(t.Register)
    case *AccessNode:
        loc := p.evalAccessNode(t.Expression)
        val, _ = p.RetrieveMemory(loc, t.Size)
    }

    switch t := instruction.Arguments[1].(type) {
    case *RegisterNode:
        val += p.RetrieveRegister(t.Register)
    case *AccessNode:
        loc := p.evalAccessNode(t.Expression)
        sum, _ := p.RetrieveMemory(loc, t.Size)
        val += sum
    case *ConstantNode:
        val += uint64(t.Value)
    }

    switch t := instruction.Arguments[0].(type) {
    case *RegisterNode:
        p.InsertRegister(t.Register, int(val))
    case *AccessNode:
        loc := p.evalAccessNode(t.Expression)
        p.InsertMemory(loc, int(val), t.Size)
    }
}

func (p *Process) handleSub(instruction *InstructionNode) {
    var val uint64

    switch t := instruction.Arguments[0].(type) {
    case *RegisterNode:
        val = p.RetrieveRegister(t.Register)
    case *AccessNode:
        loc := p.evalAccessNode(t.Expression)
        val, _ = p.RetrieveMemory(loc, t.Size)
    }

    switch t := instruction.Arguments[1].(type) {
    case *RegisterNode:
        val -= p.RetrieveRegister(t.Register)
    case *AccessNode:
        loc := p.evalAccessNode(t.Expression)
        minus, _ := p.RetrieveMemory(loc, t.Size)
        val -= minus
    case *ConstantNode:
        val -= uint64(t.Value)
    }

    switch t := instruction.Arguments[0].(type) {
    case *RegisterNode:
        p.InsertRegister(t.Register, int(val))
    case *AccessNode:
        loc := p.evalAccessNode(t.Expression)
        p.InsertMemory(loc, int(val), t.Size)
    }
}

func (p *Process) handleJump(instruction *InstructionNode) {
    p.currentInstruction, _ = p.text.Labels[instruction.Arguments[0].(*LabelNode).Name]
    p.currentInstruction--
}

func (p *Process) handleCmp(instruction *InstructionNode) {
    var left, right uint64

    switch t := instruction.Arguments[0].(type) {
        case *ConstantNode:
            left = uint64(t.Value)
        case *RegisterNode:
            left = p.RetrieveRegister(t.Register)
        case *AccessNode:
            loc := p.evalAccessNode(t.Expression)
            left, _ = p.RetrieveMemory(loc, t.Size)
    }

    switch t := instruction.Arguments[1].(type) {
        case *ConstantNode:
            right = uint64(t.Value)
        case *RegisterNode:
            right = p.RetrieveRegister(t.Register)
        case *AccessNode:
            loc := p.evalAccessNode(t.Expression)
            right, _ = p.RetrieveMemory(loc, t.Size)
    }

    if left == right {
        p.cmpFlags[CMP_JE]  = true
        p.cmpFlags[CMP_JLE] = true
        p.cmpFlags[CMP_JGE] = true
    } else {
        p.cmpFlags[CMP_JNE] = true
    }

    if left > right {
        p.cmpFlags[CMP_JG] = true;
    } else if left < right {
        p.cmpFlags[CMP_JL] = true;
    }

    p.cmpProtect = true
}

func (p *Process) handleLogicJump(instruction *InstructionNode) {
    jump, _ := p.text.Labels[instruction.Arguments[0].(*LabelNode).Name]

    switch instruction.Name {
    case "je":
        if p.cmpFlags[CMP_JE] {
            p.currentInstruction = jump - 1
            break
        }
    case "jne":
        if p.cmpFlags[CMP_JNE] {
            p.currentInstruction = jump - 1
            break
        }
    case "jl":
        if p.cmpFlags[CMP_JL] {
            p.currentInstruction = jump - 1
            break
        }
    case "jg":
        if p.cmpFlags[CMP_JG] {
            p.currentInstruction = jump - 1
            break
        }
    case "jle":
        if p.cmpFlags[CMP_JLE] {
            p.currentInstruction = jump - 1
            break
        }
    case "jge":
        if p.cmpFlags[CMP_JGE] {
            p.currentInstruction = jump - 1
            break
        }
    }
}

const (
    sys_exit int = 0x2000001
    sys_mmap int = 0x20000C5
)


func (p *Process) handleSyscall(instruction *InstructionNode) {
    rax := int(p.RetrieveRegister("rax"))

    switch rax {
    case sys_exit:
        p.currentInstruction = len(p.text.Instructions)
    case sys_mmap:
        p.InsertRegister("rax", p.topMmap)
        p.pages[p.topMmap] = NewPage()
        p.topMmap += 0x1000
    }
}

func GetParent(reg string) (s string) {
    reg = reg + " "
    if strings.Contains("eax ax al ah", reg) {
        s = "rax"
    } else if strings.Contains("ebx bx bl bh ", reg) {
        s = "rbx"
    } else if strings.Contains("ecx cx cl ch ", reg) {
        s = "rcx"
    } else if strings.Contains("edx dx dl dh ", reg) {
        s = "rdx"
    } else if strings.Contains("esi si sil ", reg) {
        s = "rsi"
    } else if strings.Contains("edi di dil ", reg) {
        s = "rdi"
    } else if strings.Contains("ebp bp bpl ", reg) {
        s = "rbp"
    } else if strings.Contains("esp sp spl ", reg) {
        s = "rsp"
    } else if strings.Contains("r8d r8w r8b ", reg) {
        s = "r8"
    } else if strings.Contains("r9d r9w r9b ", reg) {
        s = "r9"
    } else if strings.Contains("r10d r10w r10b ", reg) {
        s = "r10"
    } else if strings.Contains("r11d r11w r11b ", reg) {
        s = "r11"
    } else if strings.Contains("r12d r12w r12b ", reg) {
        s = "r12"
    } else if strings.Contains("r13d r13w r13b ", reg) {
        s = "r13"
    } else if strings.Contains("r14d r14w r14b ", reg) {
        s = "r14"
    } else if strings.Contains("r15d r15w r15b ", reg) {
        s = "r15"
    }

    return
}

func (p *Process) handleCall(instruction *InstructionNode) {
    rsp := p.RetrieveRegister("rsp")
    val := p.currentInstruction + 1

    p.InsertMemory(int(rsp - 8), val, 8)
    p.InsertRegister("rsp", int(rsp) - 8)

    p.currentInstruction = p.text.Labels[instruction.Arguments[0].(*LabelNode).Name] - 1
}

func (p *Process) handleRet(instruction *InstructionNode) {
    rsp := p.RetrieveRegister("rsp")
    val, _ := p.RetrieveMemory(int(rsp), 8)
    p.InsertRegister("rsp", int(rsp) + 8)

    p.currentInstruction = int(val) - 1
}
