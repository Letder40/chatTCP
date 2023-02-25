package main

//ChatTCP is a service of chat over TCP based on a custom protocol

import (
	//Local
	"github.com/Letder40/ChatTCP/v1/handler"
	"github.com/Letder40/ChatTCP/v1/writers"

	//Remote packages
	"fmt"
	"net"
)

//Change this to listen in other ip.addr

var lisenning_On = "0.0.0.0"

// -------------------------------------


func main(){
	go writers.CallChannelWriter()
	go writers.PrivateMessageWriter()
	go writers.PrivateResponseWriter()

	server()
}

func server(){
	socket := &net.TCPAddr{
		IP: net.ParseIP(lisenning_On),
		Port: 9701,
	}
	listener, err := net.ListenTCP("tcp", socket)
	if err != nil {
		fmt.Println("Error : ", err)
	}

	fmt.Println("Escuchando conexiones entrantes a chatTCP por el puerto 9701")

	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil{
			fmt.Println("Error en la conexión con el cliente : ", err)
		}

		go handler.ConnectionHandler(connection)
	}
}



//By Letder40