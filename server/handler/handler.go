package handler

import (
	// Local modules
	"github.com/Letder40/ChatTCP/v1/global"
	"github.com/Letder40/ChatTCP/v1/checkers"
	"github.com/Letder40/ChatTCP/v1/readers"
	
	// Remote modules
	"strings"
	"net"
	"bytes"
	"fmt"
	
)

func ConnectionHandler(connection net.Conn){

	var (
		step = 1
		message string
		nickname string
		prompt string
	) 

	connection.Write([]byte("ChatTCP V1.0\n"))

	for{

		buffer := make([]byte, 1024 * 5)
		_, err := connection.Read(buffer)
		
		buffer = bytes.Trim(buffer, "\x00")
		buffer = bytes.Trim(buffer, "\n")

		var (
			callingTo = global.Nicknames[nickname].CallingTo
			changingToPreviousStep = global.Nicknames[nickname].ChangingToPreviousStep
			channelId = global.Nicknames[nickname].ChannelId
			hasCall = global.Nicknames[nickname].HasCall
			incallWith = global.Nicknames[nickname].IncallWith
		)

		var (
			textInBuffer = string(buffer)
			bufferLen = len(strings.Split(textInBuffer, " "))
			textInBufferSplited = strings.Split(textInBuffer, " ")
			Action = textInBufferSplited[0]
		)

		//If user disconnects
		if(err != nil){
			
			global.Nicknames[nickname] = global.Nicknames_data{
				HasCall: false,
				CalledBy: "",
				CallingTo: "",
				IncallWith: "",
				ChangingToPreviousStep: false,
				ChannelId: 0,
			}
			
			var new_allNicknames []string
			for _, element := range global.AllNicknames {
				if(element != nickname){
					new_allNicknames = append(new_allNicknames, element)
				}
			}
			fmt.Printf("%s se ha desconectado\n", nickname)
			global.AllNicknames = new_allNicknames
			return
		}

		if( changingToPreviousStep ){
			step = 3
			prompt = fmt.Sprintf("[ %s ]=> ", nickname)
			global.Nicknames[nickname] = global.Nicknames_data{
				ChangingToPreviousStep: false,
			}
		}
		
		switch step{
		case 1:

			if(textInBuffer != "Go Connect"){
				connection.Write([]byte("Bad usage of the ChatTCP protocol\n"))
				connection.Close()				
			}

			connection.Write([]byte("Introduce un nickname: "))
			step += 1
		
		case 2:

			if(textInBuffer != ""){
				nickname = textInBuffer
				
				if !checkers.CheckNickname(nickname){
					connection.Write([]byte("\n[!] Ese nickname ya esta en uso porfavor selecciona otro\n\nIntroduce un nickname: "))
					continue
				}

				if bufferLen != 1 {
					connection.Write([]byte("\n[!] Ese nickname contiene espacios por favor selecciona uno sin espacios\n\nIntroduce un nickname: "))
					continue
				}

				global.AllNicknames = append(global.AllNicknames, nickname)
				
				global.Nicknames[nickname] = global.Nicknames_data{
					HasCall: false,
					CalledBy: "",
					CallingTo: "",
					IncallWith: "",
					ChangingToPreviousStep: false,
					ChannelId: 0,
				}

				prompt = fmt.Sprintf("[ %s ]=> ", nickname)
				message = fmt.Sprintf("Bienvenido %s \n\n[?] Ayuda para usar ChatTCP:\nlist -> te devuelve todos los nisknames de los clientes conectrados\ncall nickname -> llama a alguien a traves de su nickname para empezar una conversación\naccept nickaneme_del_usuario-> Aceptar una llamada\ndecline nickname_del_usuario -> rechazar una llamada\n\n%s", nickname, prompt)
				connection.Write([]byte(message))

				go readers.ReadCallChannel(nickname, connection)

				step += 1
			}

		case 3:
						
			//LIST
			if(Action == "list" && bufferLen == 1){				
				
				message = fmt.Sprintf("\n%s\n %s", checkers.List(global.AllNicknames), prompt)
				connection.Write([]byte(message))
	
			//CALL
			}else if(Action == "call" && bufferLen == 2){
				
				var(
					nicknameToCall = textInBufferSplited[1]
					SendtoIsincallwith = global.Nicknames[nicknameToCall].IncallWith
				)

				if(SendtoIsincallwith != ""){
					message = fmt.Sprintf("[!] %s ya esta en una llamada\n%s", nicknameToCall, prompt)
					connection.Write([]byte(message))
					continue
				}

				if(callingTo != "" || hasCall ){
					message = fmt.Sprintf("[!] Ya estas en una llamada\n%s", prompt)
					connection.Write([]byte(message))
					continue
				}
				
				if(!checkers.CheckNickname(nicknameToCall)){
					
					if(nicknameToCall == nickname){
						message = fmt.Sprintf("[!] No te puedes llamar a ti mismo\n%s", prompt)
						connection.Write([]byte(message))
						continue
					}
					
					// Making the call

					global.ChannelId += 1
					global.FrameId += 1
					
					dataToSend := global.DataForChannel{
						Id: global.FrameId,
						Action: "CALL",
						SendedBy: nickname,
						SendedTo: nicknameToCall,
						PrivateId: global.ChannelId,
					}
					
					global.Nicknames[nickname] = global.Nicknames_data {
						CallingTo: nicknameToCall,
						ChannelId: global.ChannelId,
						IncallWith: nicknameToCall,
					}

					global.Nicknames[nicknameToCall] = global.Nicknames_data {
						HasCall: true,
						CalledBy: nickname,
					}


					message = fmt.Sprintf("[?] Escribe bye para terminar llamada\n( Esperando.... )=[ %s ]=> ",nickname)
					connection.Write([]byte(message))

					//CREATING DYNAMIC PRIVATE CHANNEL
					global.PrivateChannels[global.ChannelId] = make(chan global.DataForChannel)

					// SEND TO CALLQUEUE
					global.CallQueue = append(global.CallQueue, dataToSend)

					go readers.ReadPrivateChannel(nickname, connection, global.ChannelId)

					step += 1

				}else{
					message = fmt.Sprintf("\n[!] The nickname: %s, is not connected\n\n%s", nicknameToCall, prompt)
					connection.Write([]byte(message))
				}
				
			
			//ACCEPT
			}else if(Action == "accept" &&  bufferLen == 2){
				
				var(
					dataSplited = strings.Split(textInBuffer, " ")
					SendedBy = nickname
					SendedTo = dataSplited[1]
				)
				
				if(!global.Nicknames[nickname].HasCall){
					message = fmt.Sprintf("[!] Nadie te esta llamando\n%s", prompt)
					connection.Write([]byte(message))
					continue
				}

				if(global.Nicknames[SendedTo].CallingTo != nickname){
					message = fmt.Sprintf("[!] %s no te esta llamando\n%s", SendedTo, prompt)
					connection.Write([]byte(message))
					continue
				}	
				
				privateId := global.Nicknames[nickname].ChannelId
				
				global.Nicknames[nickname] = global.Nicknames_data {
					CalledBy: "",
					IncallWith: SendedTo,
					ChannelId: privateId,
				}

				global.FrameId += 1

				dataToSend := global.DataForChannel{
					Id: global.FrameId,
					Action: "ACCEPT",
					SendedBy: SendedBy,
					SendedTo: SendedTo,
					PrivateId: privateId,
				}
				
				fmt.Printf("ID ==> %d)\n",privateId)
				// SEND TO ResponseQueue
				global.ResponseQueue = append(global.ResponseQueue, dataToSend)
				
				prompt = fmt.Sprintf("( %s <==> %s ) => ", global.Nicknames[nickname].IncallWith, nickname)
				connection.Write([]byte(prompt))

				step += 1

			//DECLINE
			}else if(Action == "decline" && bufferLen == 2){
				
				var(
					dataSplited = strings.Split(textInBuffer, " ")
					SendedBy = nickname
					SendedTo = dataSplited[1]
					privateId = global.Nicknames[nickname].ChannelId
					called_by = global.Nicknames[nickname].CalledBy
				)

				if( called_by == SendedTo ){

					global.Nicknames[nickname] = global.Nicknames_data{
						HasCall: false,
						CalledBy: "",
						CallingTo: "",
						IncallWith: "",
						ChangingToPreviousStep: false,
						ChannelId: 0,
					}
	
					global.FrameId += 1

					dataToSend := global.DataForChannel {
						Id: global.FrameId,
						Action: "DECLINE",
						SendedBy: SendedBy,
						SendedTo: SendedTo,
						PrivateId: privateId,
					}
	
					// SEND TO ResponseQueue
					global.ResponseQueue = append(global.ResponseQueue, dataToSend)
					go readers.ReadCallChannel(nickname, connection)
	
					connection.Write([]byte(prompt))
				
				}else{
					connection.Write([]byte("[!] Ese usuario no te esta llamando"))
					connection.Write([]byte(prompt))

					continue
				}

			}else if(Action == "user_info" && bufferLen == 1){
				connection.Write([]byte(fmt.Sprintf("\n\nnickname => %s\nchannel id => %d\nIn call with => %s\n%s", nickname, channelId, incallWith, prompt)))

			}else if(Action == "bye" && bufferLen == 1){
				connection.Write([]byte("[!] No estas llamando o en conversación"))

			}else{
				message = fmt.Sprintf("\n[!] Not a command\n\n%s", prompt)
				connection.Write([]byte(message))

			}

		case 4:

			if(incallWith == "" && callingTo == ""){
				step = 3
				prompt = fmt.Sprintf("[ %s ]=> ", nickname)
				continue
			}

			if(Action == "bye" && bufferLen == 1){
				prompt = fmt.Sprintf("[ %s ]=> ", nickname)

				global.FrameId += 1

				dataToSend := global.DataForChannel {
					Id: global.FrameId,
					Action: "END",
					SendedBy: nickname,
					SendedTo: incallWith,
					PrivateId: channelId,
				}

				if(callingTo != ""){
					
					global.FrameId += 1
					
					dataToSend = global.DataForChannel {
						Id: global.FrameId,
						Action: "END",
						SendedBy: nickname,
						SendedTo: callingTo,
						PrivateId: channelId,
					}

				}
				
				// SEND TO MessageQueue
				global.MessageQueue = append(global.MessageQueue, dataToSend)

				global.Nicknames[nickname] = global.Nicknames_data{
					HasCall: false,
					CalledBy: "",
					CallingTo: "",
					IncallWith: "",
					ChangingToPreviousStep: false,
					ChannelId: 0,
				}
				
				connection.Write([]byte(prompt))
				step = 3
				go readers.ReadCallChannel(nickname, connection)
				continue
			
			}else if(Action == "user_info" && bufferLen == 1){
				connection.Write([]byte(fmt.Sprintf("\n\nnickname => %s\nchannel id => %d\nIn call with => %s\n%s", nickname, channelId, incallWith, prompt)))
				continue
			}

			dataToSend := global.DataForChannel {
				Action: "SEND",
				SendedBy: nickname,
				SendedTo: incallWith,
				Message: textInBuffer,
				PrivateId: channelId,
			}
			
			global.MessageQueue = append(global.MessageQueue, dataToSend)

			prompt = fmt.Sprintf("( %s <==> %s ) => ", incallWith, nickname)
			connection.Write([]byte(prompt))

		}
	}
}
