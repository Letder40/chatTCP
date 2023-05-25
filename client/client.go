package main

import (
	"bufio"
	"fmt"
	"net"
	"os"

)

func main() {
    // Dirección IP y puerto del servidor
    if len(os.Args) != 2{
      fmt.Println("usage: ./client ip-addr")
      os.Exit(1)
    } 
    serverAddr := os.Args[1] + ":9701"

    // Conectar al servidor
    conn, err := net.Dial("tcp", serverAddr)
    if err != nil {
        fmt.Println("Error al conectar al servidor:", err)
        return
    }
    defer conn.Close()

    // Leer datos del servidor y mostrarlos en pantalla
    go func() {
        serverBuffer := make([]byte, 1024)
        conn.Write([]byte("Go Connect"))
        for {
            _,err = conn.Read(serverBuffer)
            
            if err != nil {
                fmt.Printf("\n[!] El servidor de ChatTCP ha finalizado la sesión\n\n")
                os.Exit(1)
            }
            textInBuffer := string(serverBuffer)
            print(textInBuffer)
        }
    }()

    // Leer entrada del usuario y enviarla al servidor
    for {
        reader := bufio.NewReader(os.Stdin)
        input, _ := reader.ReadString('\n')
        
        conn.Write([]byte(input))
    }
}
