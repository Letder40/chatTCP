package main

//ChatTCP is a service of chat over TCP based on a custom protocol

import (
	//local
	"strings"

	"github.com/Letder40/ChatTCP/v1/global"
	"github.com/Letder40/ChatTCP/v1/handler"
	"github.com/Letder40/ChatTCP/v1/writers"

	//Remote packages
	"bufio"
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
  fmt.Print("Do you want to enable Translation functionality?? [!] It requires to install https://github.com/LibreTranslate/LibreTranslate, and be running on 5000 port [Y/n] ")
  reader := bufio.NewReader(os.Stdin) 
  input, _ := reader.ReadString('\n')
  input = strings.Trim(input, "\n")
  if input == "" || strings.ToLower(input) == "y" {
    global.Translation_functionality = true
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
