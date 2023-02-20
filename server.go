package main

import (
	"bytes"
	"fmt"
	"net"
	//"strings"
)

var allNicknames []string 

func server(){
	socket := &net.TCPAddr{
		IP: net.ParseIP("127.0.0.1"),
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

		go connectionHandler(connection)
	}
}

func connectionHandler(connection net.Conn){
	buffer := make([]byte, 1024)
	step := 1
	var message string 
	var nickname string
	var prompt string

	connection.Write([]byte("ChatTCP V1.0\n"))

	for{

		_, err := connection.Read(buffer)

		if(err != nil){
			var new_allNicknames []string
			for _, element := range allNicknames {
				if(element != nickname){
					new_allNicknames = append(new_allNicknames, element)
				}
			}
			allNicknames = new_allNicknames
		}

		buffer = bytes.Trim(buffer, "\x00")
		buffer = bytes.Trim(buffer, "\n")
		text_in_buffer := string(buffer)
		
		switch step{
		case 1:
			if(text_in_buffer != "Go Connect"){
				connection.Write([]byte("Bad usage of the ChatTCP protocol\n"))
				connection.Close()
				
			}
			step = 2
			buffer = make([]byte, 1024)
			connection.Write([]byte("Introduce un nickname: "))
			continue
		case 2: 
			if(text_in_buffer != ""){
				nickname = text_in_buffer
				if !checkNickname(nickname, allNicknames){
					buffer = make([]byte, 1024)
					connection.Write([]byte("\n[!]Ese nickname ya esta en uso porfavor selecciona otro\n\nIntroduce un nickname: "))
					continue
				}
				allNicknames = append(allNicknames, nickname)
				prompt = fmt.Sprintf("[ %s ]=> ", nickname)
				message = fmt.Sprintf("Bienvenido %s \n\nAyuda para usar ChatTCP:\nlist -> te devuelve todos los nisknames de los clientes conectrados\ncall nickname -> llama a alguien a traves de su nickname para empezar una conversación\n\n%s", nickname, prompt)
				connection.Write([]byte(message))
				step = 3
				buffer = make([]byte, 1024)
				continue
			}
		case 3:
			if(text_in_buffer == "list"){
				message = fmt.Sprintf("\n%s\n %s", list(allNicknames), prompt)
				connection.Write([]byte(message))
				buffer = make([]byte, 1024)
			}else if(text_in_buffer == "call"){
				message = fmt.Sprintf("calling...\n%s", prompt)
				connection.Write([]byte(message))
				buffer = make([]byte, 1024)
			}else{
				message = fmt.Sprintf("\n[!] Not a command\n\n%s", prompt)
				connection.Write([]byte(message))
				buffer = make([]byte, 1024)
			}
			
		}	
	}
}

func checkNickname(nickname string, allNicknames []string) bool{
	for _, element := range allNicknames {
		if(element == nickname){
			return false
		}
	}
	return true
}

func list(allNicknames []string) string {
	var list string
	for _, element := range allNicknames {
		list += fmt.Sprintf("-> %s \n", element)
	}
	return list
}

func main(){
	server()
}
