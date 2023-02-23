package main

//By Letder40
import (
	"bytes"
	"math/rand"
	"fmt"
	"net"
	"strings"
	"time"
)

var allNicknames []string 
var nickname string

var channelId = 0
var FrameId = 0

type nicknames_data struct{
	has_call bool
	calling_to string
	called_by string
	incall_with string
	changingTo_PreviousStep bool
	channelId int
}

type dataForChannel struct{
	id int
	Action string
	SendedBy string
	SendedTo string
	message string
	forwardedBy []string
	PrivateId int
}

//Channel Inicialization
var callChannel = make(chan dataForChannel)
var ACKChannels = make(map[int] chan dataForChannel)
var privateChannels = make(map[int] chan dataForChannel)

//Nicknames's data map
var nicknames = make(map[string]nicknames_data)

//Queues of WRITER MASTER
var MessageQueue = make([]dataForChannel, 0, 1024)
var CallQueue = make([]dataForChannel, 0, 1024)
var ResponseQueue = make([]dataForChannel, 0, 1024)


func server(){
	socket := &net.TCPAddr{
		IP: net.ParseIP("192.168.1.12"),
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

		//If user disconnects
		if(err != nil){
			nicknames[nickname] = nicknames_data{
				has_call: false,
				called_by: "",
				calling_to: "",
				incall_with: "",
				changingTo_PreviousStep: false,
				channelId: 0,
			}
			
			var new_allNicknames []string
			for _, element := range allNicknames {
				if(element != nickname){
					new_allNicknames = append(new_allNicknames, element)
				}
			}
			fmt.Printf("%s se ha desconectado\n", nickname)
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

			buffer = make([]byte, 1024)
			connection.Write([]byte("Introduce un nickname: "))
			step += 1
		
		case 2:

			bufferLen := len(strings.Split(text_in_buffer, " "))
			if(text_in_buffer != ""){
				nickname = text_in_buffer
				
				if !checkNickname(nickname, allNicknames){
					buffer = make([]byte, 1024)
					connection.Write([]byte("\n[!] Ese nickname ya esta en uso porfavor selecciona otro\n\nIntroduce un nickname: "))
					continue
				}
				if bufferLen != 1 {
					buffer = make([]byte, 1024)
					connection.Write([]byte("\n[!] Ese nickname contiene espacios por favor selecciona uno sin espacios\n\nIntroduce un nickname: "))
					continue
				}

				allNicknames = append(allNicknames, nickname)
				
				nicknames[nickname] = nicknames_data{
					has_call: false,
					called_by: "",
					calling_to: "",
					incall_with: "",
					changingTo_PreviousStep: false,
					channelId: 0,
				}


				prompt = fmt.Sprintf("[ %s ]=> ", nickname)
				message = fmt.Sprintf("Bienvenido %s \n\n[?] Ayuda para usar ChatTCP:\nlist -> te devuelve todos los nisknames de los clientes conectrados\ncall nickname -> llama a alguien a traves de su nickname para empezar una conversación\naccept nickaneme_del_usuario-> Aceptar una llamada\ndecline nickname_del_usuario -> rechazar una llamada\n\n%s", nickname, prompt)
				connection.Write([]byte(message))
				buffer = make([]byte, 1024)

				go ReadCallChannel(nickname, connection)

				step += 1
			}

		case 3:
			
			var (
				bufferLen = len(strings.Split(text_in_buffer, " "))
				Action = strings.Split(text_in_buffer, " ")[0]
			)
			
			//LIST
			if(Action == "list" && bufferLen == 1){
				
				message = fmt.Sprintf("\n%s\n %s", list(allNicknames), prompt)
				connection.Write([]byte(message))
				buffer = make([]byte, 1024)
	
			//CALL
			}else if(Action == "call" && bufferLen == 2){
				
				var(
					text_in_buffer_splited = strings.Split(text_in_buffer, " ")
					nickname_toCall = text_in_buffer_splited[1]
					Sendto_Isincall_with = nicknames[nickname_toCall].incall_with
				)

				if(Sendto_Isincall_with != ""){
					message = fmt.Sprintf("[!] %s ya esta en una llamada\n%s", nickname_toCall, prompt)
					connection.Write([]byte(message))
					buffer = make([]byte, 1024)
					continue
				}

				if(nicknames[nickname].calling_to != "" || nicknames[nickname].has_call == true){
					message = fmt.Sprintf("[!] Ya estas en una llamada\n%s", prompt)
					connection.Write([]byte(message))
					buffer = make([]byte, 1024)
					continue
				}
				
				if(!checkNickname(nickname_toCall, allNicknames)){
					
					if(nickname_toCall == nickname){
						message = fmt.Sprintf("[!] No te puedes llamar a ti mismo\n%s", prompt)
						connection.Write([]byte(message))
						continue
					}
					
					// Making the call

					FrameId += 1
					channelId += 1

					dataToSend := dataForChannel{
						id: FrameId,
						Action: "CALL",
						SendedBy: nickname,
						SendedTo: nickname_toCall,
						PrivateId: channelId,
					}
					
					nicknames[nickname] = nicknames_data {
						calling_to: nickname_toCall,
						channelId: channelId,
						incall_with: nickname_toCall,
					}

					nicknames[nickname_toCall] = nicknames_data {
						has_call: true,
						called_by: nickname,
					}


					message = fmt.Sprintf("[?] Escribe bye para terminar llamada\n( Esperando.... )=[ %s ]=> ",nickname)
					connection.Write([]byte(message))
					buffer = make([]byte, 1024)

					//CREATING DYNAMIC PRIVATE CHANNEL
					privateChannels[channelId] = make(chan dataForChannel)

					// SEND TO CALLQUEUE
					CallQueue = append(CallQueue, dataToSend)

					go ReadPrivateChannel(nickname, connection, channelId)

					step += 1

				}else{
					message = fmt.Sprintf("\n[!] The nickname: %s, is not connected\n\n%s", nickname_toCall, prompt)
					connection.Write([]byte(message))
					buffer = make([]byte, 1024)
				}
				
			
			//ACCEPT
			}else if(Action == "accept" &&  bufferLen == 2){
				
				var(
					dataSplited = strings.Split(text_in_buffer, " ")
					SendedBy = nickname
					SendedTo = dataSplited[1]
				)
				
				if(nicknames[nickname].has_call == false){
					message = fmt.Sprintf("[!] Nadie te esta llamando\n%s", prompt)
					connection.Write([]byte(message))
					buffer = make([]byte, 1024)
					continue
				}

				if(nicknames[SendedTo].calling_to != nickname){
					message = fmt.Sprintf("[!] %s no te esta llamando\n%s", SendedTo, prompt)
					connection.Write([]byte(message))
					buffer = make([]byte, 1024)
					continue
				}	
				
				privateId := nicknames[nickname].channelId
				
				nicknames[nickname] = nicknames_data {
					called_by: "",
					incall_with: SendedTo,
					channelId: privateId,
				}

				FrameId += 1
				dataToSend := dataForChannel{
					id: FrameId,
					Action: "ACCEPT",
					SendedBy: SendedBy,
					SendedTo: SendedTo,
					PrivateId: privateId,
				}
				
				fmt.Printf("ID ==> %d)\n",privateId)
				// SEND TO ResponseQueue
				ResponseQueue = append(ResponseQueue, dataToSend)
				
				prompt = fmt.Sprintf("( %s <==> %s ) => ", nicknames[nickname].incall_with, nickname)
				connection.Write([]byte(prompt))
				buffer = make([]byte, 1024)
				
				step += 1

			//DECLINE
			}else if(Action == "decline" && bufferLen == 2){
				
				var(
					dataSplited = strings.Split(text_in_buffer, " ")
					SendedBy = nickname
					SendedTo = dataSplited[1]
					privateId = nicknames[nickname].channelId
					called_by = nicknames[nickname].called_by
				)

				if( called_by == SendedTo ){

					nicknames[nickname] = nicknames_data{
						has_call: false,
						called_by: "",
						calling_to: "",
						incall_with: "",
						changingTo_PreviousStep: false,
						channelId: 0,
					}
	
					FrameId += 1
					dataToSend := dataForChannel {
						id: FrameId,
						Action: "DECLINE",
						SendedBy: SendedBy,
						SendedTo: SendedTo,
						PrivateId: privateId,
					}
	
					// SEND TO ResponseQueue
					ResponseQueue = append(ResponseQueue, dataToSend)
					go ReadCallChannel(nickname, connection)
	
					buffer = make([]byte, 1024)
					connection.Write([]byte(prompt))
				
				}else{
					connection.Write([]byte("[!] Ese usuario no te esta llamando"))
					connection.Write([]byte(prompt))
					buffer = make([]byte, 1024)

					continue
				}

			}else if(Action == "bye" && bufferLen == 1){
				connection.Write([]byte("[!] No estas llamando o en conversación"))

			}else{
				message = fmt.Sprintf("\n[!] Not a command\n\n%s", prompt)
				connection.Write([]byte(message))
				buffer = make([]byte, 1024)

			}

		case 4:
			var (
				bufferLen = len(strings.Split(text_in_buffer, " "))
				Action = strings.Split(text_in_buffer, " ")[0]
				incall_with = nicknames[nickname].incall_with
				calling_to = nicknames[nickname].calling_to
				privateId = nicknames[nickname].channelId
			)

			if(incall_with == "" && calling_to == ""){
				step = 3
				prompt = fmt.Sprintf("[ %s ]=> ", nickname)
				continue
			}

			if(Action == "bye" && bufferLen == 1){
				prompt = fmt.Sprintf("[ %s ]=> ", nickname)

				FrameId += 1
				dataToSend := dataForChannel {
					id: FrameId,
					Action: "END",
					SendedBy: nickname,
					SendedTo: incall_with,
					PrivateId: privateId,
				}

				if(calling_to != ""){
					
					FrameId += 1
					dataToSend = dataForChannel {
						id: FrameId,
						Action: "END",
						SendedBy: nickname,
						SendedTo: calling_to,
						PrivateId: privateId,
					}

					nicknames[nickname] = nicknames_data{
						has_call: false,
						called_by: "",
						calling_to: "",
						incall_with: "",
						changingTo_PreviousStep: false,
						channelId: 0,
					}
				}
				
				// SEND TO MessageQueue
				MessageQueue = append(MessageQueue, dataToSend)

				nicknames[nickname] = nicknames_data{
					has_call: false,
					called_by: "",
					calling_to: "",
					incall_with: "",
					changingTo_PreviousStep: false,
					channelId: 0,
				}
				
				connection.Write([]byte(prompt))
				buffer = make([]byte, 1024)
				step = 3
				go ReadCallChannel(nickname, connection)
				continue
			}

			dataToSend := dataForChannel {
				Action: "SEND",
				SendedBy: nickname,
				SendedTo: nicknames[nickname].incall_with,
				message: text_in_buffer,
				PrivateId: privateId,
			}
			
			MessageQueue = append(MessageQueue, dataToSend)

			prompt = fmt.Sprintf("( %s <==> %s ) => ", nicknames[nickname].incall_with, nickname)
			buffer = make([]byte, 1024)
			connection.Write([]byte(prompt))

		}
	}
}

func ReadCallChannel(nickname string, connection net.Conn){
	
	var(
		connected bool
		message_conn string
	) 

	for{
		connected = checkConnection(nickname)

		if !connected {
			return
		}
		if(nicknames[nickname].channelId != 0){
			return
		}
		if nicknames[nickname].incall_with != ""{
			return
		}
		
		select{
			case dataIn_Channel := <-callChannel: 
				
			var (
				SendedBy = dataIn_Channel.SendedBy
				SendedTo = dataIn_Channel.SendedTo
				Action = dataIn_Channel.Action
				privateId = dataIn_Channel.PrivateId
				incall_with = nicknames[nickname].incall_with
			)
		
			fmt.Printf("( %d ) READ Call Channel => %s to %s by %s | readed by %s \n", dataIn_Channel.id, Action, SendedTo, SendedBy, nickname)
			
			if Action == "CALL"{
				if(SendedTo == nickname){
					
					if(incall_with != "") {
						FrameId += 1
						dataToSend := dataForChannel {
							id: FrameId,
							Action: "DECLINE",
							SendedBy: nickname,
							SendedTo: SendedBy,
							PrivateId: privateId,
						}
						// SEND TO ResponseQueue
						ResponseQueue = append(ResponseQueue, dataToSend)
						continue
					}

					nicknames[nickname] = nicknames_data{
						has_call: true, 
						called_by: SendedBy,
						channelId: privateId,
					}
				
					message_conn = fmt.Sprintf("%s quiere iniciar una conversación, ¿La aceptas? \n[ %s ]=> ", SendedBy, SendedTo)
					connection.Write([]byte(message_conn))
					go ReadPrivateChannel(nickname, connection, privateId)
					return
				
				}else{
					FrameId += 1
					dataIn_Channel.id = FrameId
					
					// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
					CallQueue = append(CallQueue, dataIn_Channel)
				}
			}

			default:

			rand.Seed(time.Now().UnixNano())
			randInt := rand.Intn(101)
			duration := 150 + randInt
			time.Sleep(time.Millisecond * time.Duration(duration))

		}
			
	}		
}	

func ReadPrivateChannel(nickname string, connection net.Conn, channelId int){
	
	var(
		connected bool
		message_conn string
	) 

	for{

		connected = checkConnection(nickname)

		if !connected {
			return
		}
		
		if(channelId == 0){
			return
		}

		select{

		case dataIn_Channel := <-privateChannels[channelId]:
		
		var (
			//DATA IN CHANNEL
			SendedBy = dataIn_Channel.SendedBy
			SendedTo = dataIn_Channel.SendedTo
			Action = dataIn_Channel.Action
			message = dataIn_Channel.message
			privateId = dataIn_Channel.PrivateId
			//DATA IN NICKNAME
			calling_to = nicknames[nickname].calling_to 
		)

		fmt.Printf(" [-] PrivateChannel => %s to %s by %s | readed by %s \n", Action, SendedTo, SendedBy, nickname)
		switch Action {
		
			case "ACCEPT":

			if(SendedTo == nickname){
				
				if(calling_to != SendedBy){
					continue
				}

				message_conn = fmt.Sprintf("Llamada aceptada por %s \n( %s <==> %s ) => ", SendedBy, SendedTo, SendedBy)
				connection.Write([]byte(message_conn))
				nicknames[nickname] = nicknames_data{
					incall_with: SendedBy,
					calling_to: "",
					channelId: privateId,
				}
			}else{
				FrameId += 1
				dataIn_Channel.id = FrameId
				privateChannels[dataIn_Channel.PrivateId]<-dataIn_Channel

				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				ResponseQueue = append(ResponseQueue, dataIn_Channel)
			}
		
			case "DECLINE":
			
			if(SendedTo == nickname){
				
				if(calling_to != SendedBy){
					print("REACH")
					continue
				}

				nicknames[nickname] = nicknames_data{
					has_call: false,
					called_by: "",
					calling_to: "",
					incall_with: "",
					changingTo_PreviousStep: false,
					channelId: 0,
				}

				message_conn = fmt.Sprintf("Llamada rechazada por %s\n[ %s ]=> ", SendedBy, SendedTo)
				connection.Write([]byte(message_conn))
				go ReadCallChannel(nickname, connection)
				return
				
			}else{
				FrameId += 1
				dataIn_Channel.id = FrameId

				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				ResponseQueue = append(ResponseQueue, dataIn_Channel)
			}

			case "SEND":

			if(SendedTo == nickname){
				tag := fmt.Sprintf(" <=[ %s ]", SendedBy)
				dataIn_Channel.message = dataIn_Channel.message + tag
				message = fmt.Sprintf("%s \n( %s <==> %s ) => ",dataIn_Channel.message ,nicknames[nickname].incall_with ,nickname)
				connection.Write([]byte(message))
			}else{
				FrameId += 1
				dataIn_Channel.id = FrameId

				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				MessageQueue = append(MessageQueue, dataIn_Channel)
			}

			case "END":
			
			if(SendedTo == nickname){
				
				if(nicknames[nickname].incall_with != SendedBy && nicknames[nickname].called_by != SendedBy){
					continue
				}

				if(nicknames[nickname].called_by == SendedBy && nicknames[nickname].incall_with != SendedBy){
					nicknames[nickname] = nicknames_data{
						has_call: false,
						called_by: "",
						calling_to: "",
						incall_with: "",
						channelId: 0,
					}
	
					message = fmt.Sprintf("[!] %s ha finalizado la llamada\n[ %s ]=> ",SendedTo, SendedBy)
					connection.Write([]byte(message))
					go ReadCallChannel(nickname, connection)
					return
	
				}

				connection.Write([]byte(fmt.Sprintf("\n[!] Conversasión Terminada por %s\n[ %s ]=> ", nicknames[nickname].incall_with, nickname)))

				nicknames[nickname] = nicknames_data{
					has_call: false,
					called_by: "",
					calling_to: "",
					incall_with: "",
					changingTo_PreviousStep: true,
					channelId: 0,
				}
				go ReadCallChannel(nickname, connection)
				return

			}else{
				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				MessageQueue = append(MessageQueue, dataIn_Channel)
			}
		}
		
		default:
			rand.Seed(time.Now().UnixNano())
			randInt := rand.Intn(101)
			duration := 150 + randInt
			time.Sleep(time.Millisecond * time.Duration(duration))
		}

	}		
}

func CallChannelWriter(){
	
	var lenOfCallQueue int

	for{

		lenOfCallQueue = len(CallQueue) 
		
		if lenOfCallQueue != 0 {
			index := lenOfCallQueue - 1
			callChannel <- CallQueue[index]
			fmt.Printf("( %d )WRITE Call Channel => %s to %s by %s\n", CallQueue[index].id, CallQueue[index].Action, CallQueue[index].SendedTo, CallQueue[index].SendedBy)
			CallQueue = append(CallQueue[:index], CallQueue[index+1:]...)
		}

		time.Sleep(time.Millisecond * 100)		

	}
}

func PrivateResponseWriter(){
	
	var lenOfResponseQueue int
	
	for { 

		lenOfResponseQueue = len(ResponseQueue) 

		if lenOfResponseQueue != 0 {
			index := lenOfResponseQueue - 1
			channelId := ResponseQueue[index].PrivateId
			fmt.Printf("( %d ) WRITE [-]PRIVATE Channel => %s to %s by %s\n", ResponseQueue[index].id, ResponseQueue[index].Action, ResponseQueue[index].SendedTo, ResponseQueue[index].SendedBy)
			privateChannels[channelId] <- ResponseQueue[index]
			ResponseQueue = append(ResponseQueue[:index], ResponseQueue[index+1:]...)
		}

		time.Sleep(time.Millisecond * 100)	
	}
}

func PrivateMessageWriter(){
	
	var lenOfMessageQueue int
	
	for {

		lenOfMessageQueue = len(MessageQueue) 

		if lenOfMessageQueue != 0 {
			lenOfMessageQueue = len(MessageQueue) 
			index := lenOfMessageQueue - 1
			channelId := MessageQueue[index].PrivateId
			fmt.Printf("( %d ) WRITE [-]PRIVATE Channel => %s to %s by %s\n",MessageQueue[index].id, MessageQueue[index].Action, MessageQueue[index].SendedTo, MessageQueue[index].SendedBy)
			privateChannels[channelId] <- MessageQueue[index]
			MessageQueue = append(MessageQueue[:index], MessageQueue[index+1:]...)
		}

		time.Sleep(time.Millisecond * 100)	
	}
}

func checkConnection(nickname string) bool {
	for _, element := range allNicknames{
		if(element == nickname){
			return true
		}
	}
	nicknames[nickname] = nicknames_data{}
	return false
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
	go CallChannelWriter()
	go PrivateMessageWriter()
	go PrivateResponseWriter()

	server()
}


//By Letder40