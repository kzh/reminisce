package vm

import (
//    "log"
    "errors"
    "encoding/json"
    "encoding/binary"
    "github.com/gorilla/websocket"
)

type Page [4096]byte

const (
    CMP_JE = iota
    CMP_JNE
    CMP_JL
    CMP_JG
    CMP_JLE
    CMP_JGE
)

type Process struct {
    conn *websocket.Conn

    cmpProtect bool
    cmpFlags [7]bool

    topMmap int

    currentInstruction int
    text *SectionTextNode
    registers [128]byte
    pages map[int]*Page
}

func NewPage() *Page {
    return &Page{}
}

func NewProcess(conn *websocket.Conn, arguments []string) (p *Process) {
    p = &Process{
        conn: conn,
        pages: make(map[int]*Page),
        topMmap: 0x40000000,
    }

    p.InsertRegister("rbp", 0xc0000000)
    p.InsertRegister("rsp", 0xc0000000)

    // make room for stack
    p.pages[0xc0001000] = NewPage()
    p.pages[0xc0000000] = NewPage()
    p.pages[0xBFFFF000] = NewPage()
    p.pages[0xBFFFE000] = NewPage()
    return
}

type Response struct {
    Action   string `json:"action"`
    Location int    `json:"location"`
    Register string `json:"register"`
    Change   []int `json:"change"`
    Line     int    `json:"line"`
}

func (p *Process) InsertMemory(location, data, size int) error {
    offset := location % 0x1000

    page, ok := p.pages[location - offset]
    if !ok {
        return errors.New("SEG FAULT")
    }

    bytes := make([]byte, size)
    if size == 1 {
        bytes[0] = byte(data)
    } else if size == 2 {
        binary.BigEndian.PutUint16(bytes, uint16(data))
    } else if size == 4 {
        binary.BigEndian.PutUint32(bytes, uint32(data))
    } else if size == 8 {
        binary.BigEndian.PutUint64(bytes, uint64(data))
    }

    for i := 0; i != len(bytes); i++ {
        page[offset + i] = bytes[i]
    }

    p.send(bytes, "", location)
    return nil
}

func conv(bytes []byte) (i []int) {
    for _, b := range bytes {
        i = append(i, int(b))
    }

    return
}

func (p *Process) RetrieveMemory(location, size int) (uint64, error) {
    offset := location % 0x1000

    page, ok := p.pages[location - offset]
    if !ok {
        return 0, errors.New("SEG FAULT")
    }

    bytes := make([]byte, size)
    for i := 0; i != len(bytes); i++ {
        bytes[i] = page[offset + i]
    }

    var ret uint64
    if size == 1 {
        ret = uint64(bytes[0])
    } else if size == 2 {
        ret = uint64(binary.BigEndian.Uint16(bytes))
    } else if size == 4 {
        ret = uint64(binary.BigEndian.Uint32(bytes))
    } else if size == 8 {
        ret = uint64(binary.BigEndian.Uint64(bytes))
    }

    return ret, nil
}

func (p *Process) InsertRegister(reg string, data int) {
    r := registers[reg]

    bytes := make([]byte, r.Size)

    if r.Size == 1 {
        bytes[0] = byte(data)
    } else if r.Size == 2 {
        binary.BigEndian.PutUint16(bytes, uint16(data))
    } else if r.Size == 4 {
        binary.BigEndian.PutUint32(bytes, uint32(data))
    } else if r.Size == 8 {
        binary.BigEndian.PutUint64(bytes, uint64(data))
    }

    for i := 0; i != len(bytes); i++ {
        p.registers[r.Location + i] = bytes[i]
    }

    if parent := GetParent(reg); parent != "" {
        p.InsertRegister(parent, int(p.RetrieveRegister(parent)))
    } else {
        p.send(bytes, reg, -1)
    }
}

func (p *Process) send(bytes []byte, reg string, loc int) {
    var line int
    if p.text != nil && p.currentInstruction < len(p.text.Instructions) {
        line = p.text.Instructions[p.currentInstruction].Location()
    }

    resp := Response{
        Action:   "change",
        Register: reg,
        Location: loc,
        Line:     line,
        Change:   conv(bytes),
    }

    encode, _ := json.Marshal(resp)
    p.conn.WriteMessage(1, encode)
}



func (p *Process) RetrieveRegister(reg string) uint64 {
    r := registers[reg]

    bytes := p.registers[r.Location:r.Location + r.Size]

    var ret uint64
    if r.Size == 1 {
        ret = uint64(bytes[0])
    } else if r.Size == 2 {
        ret = uint64(binary.BigEndian.Uint16(bytes))
    } else if r.Size == 4 {
        ret = uint64(binary.BigEndian.Uint32(bytes))
    } else if r.Size == 8 {
        ret = uint64(binary.BigEndian.Uint64(bytes))
    }

    return ret
}
