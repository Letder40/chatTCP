package main

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"
)

var allNicknames []string 
var channel = make(chan string)
var nickname string

type nicknames_data struct{
	has_call bool
	calling_to string
	called_by string
	incall_with string
	changingTo_PreviousStep bool
}

var nicknames = make(map[string]nicknames_data)


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
			nicknames[nickname] = nicknames_data{
				has_call: false,
				called_by: "",
				calling_to: "",
				incall_with: "",
				changingTo_PreviousStep: false,
			}
			
			var new_allNicknames []string
			for _, element := range allNicknames {
				if(element != nickname){
					new_allNicknames = append(new_allNicknames, element)
				}
			}
			allNicknames = new_allNicknames
			return
		}

		if(nicknames[nickname].changingTo_PreviousStep == true){
			step = step - 1
			prompt = fmt.Sprintf("[ %s ]=> ", nickname)
			nicknames[nickname] = nicknames_data{
				changingTo_PreviousStep: false,
			}
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
				
				nicknames[nickname] = nicknames_data{
					calling_to: "",
					has_call: false,
				}


				prompt = fmt.Sprintf("[ %s ]=> ", nickname)
				message = fmt.Sprintf("Bienvenido %s \n\n[?] Ayuda para usar ChatTCP:\nlist -> te devuelve todos los nisknames de los clientes conectrados\ncall nickname -> llama a alguien a traves de su nickname para empezar una conversación\naccept nickaneme_del_usuario-> Aceptar una llamada\ndecline nickname_del_usuario -> rechazar una llamada\n\n%s", nickname, prompt)
				connection.Write([]byte(message))
				step = 3
				go ReadChannel(nickname, connection)
				buffer = make([]byte, 1024)
				continue
			}
		case 3:
			Action := strings.Split(text_in_buffer, " ")[0]
			//LIST
			if(text_in_buffer == "list"){
				
				message = fmt.Sprintf("\n%s\n %s", list(allNicknames), prompt)
				connection.Write([]byte(message))
				buffer = make([]byte, 1024)
	
			//CALL
			}else if(Action == "call"){
				if(nicknames[nickname].calling_to != "" || nicknames[nickname].has_call == true){
					message = fmt.Sprintf("[!] Ya estas en una llamada\n%s", prompt)
					connection.Write([]byte(message))
					buffer = make([]byte, 1024)
					continue
				}
				
				text_in_buffer_splited := strings.Split(text_in_buffer, " ")
				nickname_toCall := text_in_buffer_splited[1]

				nicknames[nickname] = nicknames_data {
					calling_to: nickname_toCall,
				}

				message = fmt.Sprintf("[?] Escribe bye para terminar llamada\n( Esperando.... )=[ %s ]=> ",nickname)
				connection.Write([]byte(message))
				buffer = make([]byte, 1024)
				step = 4
				
				if(!checkNickname(nickname_toCall, allNicknames)){
					
					if(nickname_toCall == nickname){
						message = fmt.Sprintf("[!] No te puedes llamar a ti mismo\n%s", prompt)
						connection.Write([]byte(message))
					}else{
						message = fmt.Sprintf("CALL %s By %s",nickname_toCall, nickname)
						channel <- message
						
					}

					buffer = make([]byte, 1024)

				}else{
					message = fmt.Sprintf("\n[!] The nickname: %s, is not connected\n\n%s", nickname_toCall, prompt)
					connection.Write([]byte(message))
					buffer = make([]byte, 1024)
				}
			
			//ACCEPT
			}else if(Action == "accept" && len(strings.Split(text_in_buffer, " ")) == 2){
				if(nicknames[nickname].has_call == false){
					message = fmt.Sprintf("[!] Nadie te esta llamando\n%s", prompt)
					connection.Write([]byte(message))
					buffer = make([]byte, 1024)
					continue
				}
				step = 4
				dataSplited := strings.Split(text_in_buffer, " ")
				SendedBy := nickname
				SendedTo := dataSplited[1]
				
				nicknames[nickname] = nicknames_data {
					called_by: "",
					incall_with: SendedTo,
				}

				message = fmt.Sprintf("ACCEPT %s By %s", SendedTo, SendedBy)
				channel <- message
				
				prompt = fmt.Sprintf("( %s <==> %s ) => ", nicknames[nickname].incall_with, nickname)
				connection.Write([]byte(prompt))
				buffer = make([]byte, 1024)

			}else if(Action == "decline" && len(strings.Split(text_in_buffer, " ")) == 2){
				dataSplited := strings.Split(text_in_buffer, " ")
				SendedBy := nickname
				SendedTo := dataSplited[1]
				message = fmt.Sprintf("DECLINE %s By %s", SendedTo, SendedBy)

				nicknames[nickname] = nicknames_data{
					has_call: false,
				}

				channel <- message
				buffer = make([]byte, 1024)
				connection.Write([]byte(prompt))

			}else{
				message = fmt.Sprintf("\n[!] Not a command\n\n%s", prompt)
				connection.Write([]byte(message))
				buffer = make([]byte, 1024)
			}

			if(Action == "Bye" || Action == "bye"){
				connection.Write([]byte("[!] No estas llamando o en conversación"))
			}

		case 4:
			Action := strings.Split(text_in_buffer, " ")[0]
			if(nicknames[nickname].incall_with == "" && nicknames[nickname].calling_to == ""){
				step = 3

				prompt = fmt.Sprintf("[ %s ]=> ", nickname)
				continue
			}

			if(Action == "Bye" || Action == "bye"){
				prompt = fmt.Sprintf("[ %s ]=> ", nickname)
				message = fmt.Sprintf("END %s By %s", nicknames[nickname].incall_with, nickname)

				if(nicknames[nickname].calling_to != ""){
					message = fmt.Sprintf("END %s By %s", nicknames[nickname].calling_to, nickname)
					nicknames[nickname] = nicknames_data{
						calling_to: "",
					}
				}
				
				channel <- message
				step = 3

				nicknames[nickname] = nicknames_data{
					has_call: false,
					called_by: "",
					calling_to: "",
					incall_with: "",
				}
				
				connection.Write([]byte(prompt))
				buffer = make([]byte, 1024)
				continue
			}

			message = fmt.Sprintf("SEND %s By %s %s", nicknames[nickname].incall_with, nickname, text_in_buffer)
			prompt = fmt.Sprintf("( %s <==> %s ) => ", nicknames[nickname].incall_with, nickname)
			channel<-message
			buffer = make([]byte, 1024)
			connection.Write([]byte(prompt))

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

func ReadChannel(nickname string, connection net.Conn){
	var(
		connected bool
		//id int
		message string
	) 

	for{
		for _, element := range allNicknames{
			if(element == nickname){
				connected = true
			}
		}

		if !connected {
			return
		}
		select {
		case dataIn_Channel := <-channel:
			fmt.Println(dataIn_Channel)
			
			//READING THE CHANNEL
			dataSplited := strings.Split(dataIn_Channel, " ")
			SendedBy := dataSplited[3]
			SendedTo := dataSplited[1]
			Action := dataSplited[0]
			if(Action == "SESSION_START"){
				//SESSION_START Letder40 By Letder401 With randint
				//id, _ = strconv.Atoi(dataSplited[5])
			}
			switch Action{	
			case "CALL":
				if(SendedTo == nickname){
					if(nicknames[nickname].incall_with != "") {
						message = fmt.Sprintf("DECLINE %s By %s", SendedBy, SendedTo)
						channel<-message
					}

					nicknames[nickname] = nicknames_data{
						has_call: true, 
						called_by: SendedBy,
					}
			
					message = fmt.Sprintf("%s quiere iniciar una conversación, ¿La aceptas? \n[ %s ]=> ", SendedBy, SendedTo)
					connection.Write([]byte(message))
				}else{
					channel<-dataIn_Channel
				}
			
			case "ACCEPT":
				if(SendedTo == nickname){
					if(nicknames[nickname].calling_to != SendedBy){
						continue
					}
					message = fmt.Sprintf("Llamada aceptada por %s \n( %s <==> %s ) => ", SendedBy, SendedTo, SendedBy)
					connection.Write([]byte(message))
					
					nicknames[nickname] = nicknames_data{
						incall_with: SendedBy, 
						calling_to: "",
					}

				}else{
					channel<-dataIn_Channel
				}
			case "DECLINE":
				if(SendedTo == nickname){
					if( nicknames[nickname].calling_to == SendedBy){
						message = fmt.Sprintf("Llamada rechazada por %s", SendedBy)
						connection.Write([]byte(message))
					}
				}else{
					channel<-dataIn_Channel
				}
			case "SEND":
				if(SendedTo == nickname){
					tag := fmt.Sprintf(" <=[ %s ]", SendedBy)
					dataSplited = append(dataSplited, tag)
					new_slice := dataSplited[4:]
					message = fmt.Sprintf("%s \n( %s <==> %s ) => ",strings.Join(new_slice, " "),nicknames[nickname].incall_with ,nickname)

					connection.Write([]byte(message))
				}else{
					channel<-dataIn_Channel
				}
			case "END":
				if(SendedTo == nickname){
					if(nicknames[nickname].incall_with == SendedBy){
						connection.Write([]byte(fmt.Sprintf("\n[!] Conversasión Terminada por %s\n[ %s ]=> ", nicknames[nickname].incall_with, nickname)))
						nicknames[nickname] = nicknames_data{
							has_call: false,
							called_by: "",
							calling_to: "",
							incall_with: "",
							changingTo_PreviousStep: true,
						}
					}else if(nicknames[nickname].called_by == SendedBy){
						nicknames[nickname] = nicknames_data{
							has_call: false,
							called_by: "",
						}
						message = fmt.Sprintf("[!] %s ha finalizado la llamada\n[ %s ]=> ",SendedTo, SendedBy)
						connection.Write([]byte(message))
					}
				}else{
					channel<-dataIn_Channel
				}
			}
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}		
}	


func main(){
	server()
}
