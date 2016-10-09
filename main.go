package main

import (
    "log"
    "net/http"
    "encoding/json"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/furryfaust/reminisce/vm"
)

func main() {
    r := gin.Default()

    r.GET("/ws", func(c *gin.Context) {
        handleSocket(c.Writer, c.Request)
    })
    r.StaticFile("/", "public/index.html")
    r.StaticFS("/assets", http.Dir("public/assets/"))

    r.Run()
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

type Init struct {
    Method string
    Code   string
}

func handleSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)

    if err != nil {
        return
    }

    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            continue
        }

        var message map[string]interface{}
        if err := json.Unmarshal(msg, &message); err != nil {
            panic(err)
        }


        toks := vm.Lex(message["Code"].(string))
        log.Println("Finished lexing")
        ast  := vm.Parse(toks)
        log.Println("Finished parsing")
        proc := vm.NewProcess(conn, []string {})

        proc.Simulate(ast)
        log.Println("Simulated")
    }
}
