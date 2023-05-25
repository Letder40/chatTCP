package main

//ChatTCP is a service of chat over TCP based on a custom protocol

import (
	//local
	"github.com/Letder40/ChatTCP/v1/global"
	"github.com/Letder40/ChatTCP/v1/handler"
	"github.com/Letder40/ChatTCP/v1/writers"

	//Remote packages
	"fmt"
	"net"
	"os"
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
  if len(os.Args) == 3 || len(os.Args) == 2{
	option := os.Args[1]
	switch option {
	case "--libre-translate", "-lt":
		global.Translation_service.Enable = true
		global.Translation_service.LTranslate = true
	
	case "--deepl", "-dl":
		global.Translation_service.Enable = true
		global.Translation_service.Deepl = true
		global.Translation_service.Deepl_api_key = os.Args[2]
	
	default :
		fmt.Println("incorrect option, valid arguments --libre-translate or --depl [api-key]")
		os.Exit(1)
	}
  }

  socket := &net.TCPAddr{
		IP: net.ParseIP(lisenning_On),
		Port: 9701,
	}
	listener, err := net.ListenTCP("tcp", socket)
	if err != nil {
		fmt.Println("Error : ", err)
	}

	fmt.Println("Listenning connections to chatTCP on 9701/tcp port")

	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil{
      fmt.Println("Error in the client-server connection : ", err)
		}

		go handler.ConnectionHandler(connection)
	}
}



//By Letder40
