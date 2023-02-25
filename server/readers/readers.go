package readers

import(
	"net"
	"github.com/Letder40/ChatTCP/v1/global"
	"github.com/Letder40/ChatTCP/v1/checkers"
	"fmt"
	"time"
)

func ReadCallChannel(nickname string, connection net.Conn){
	
	var(
		connected bool
		message_conn string
	) 

	for{
		connected = checkers.CheckConnection(nickname)

		var (
			channelId = global.Nicknames[nickname].ChannelId
			incallWith = global.Nicknames[nickname].IncallWith
		)
		
		if !connected {
			return
		}
		if(channelId != 0){
			return
		}
		if incallWith != ""{
			return
		}
		
		select{
			case dataInChannel := <- global.CallChannel: 
				
			var (
				sendedBy = dataInChannel.SendedBy
				sendedTo = dataInChannel.SendedTo
				action = dataInChannel.Action
				privateId = dataInChannel.PrivateId
			)
		
			fmt.Printf("( %d ) READ Call Channel => %s to %s by %s | readed by %s \n", dataInChannel.Id, action, sendedTo, sendedBy, nickname)
			
			if action == "CALL"{
				if(sendedTo == nickname){
					
					if(incallWith != "") {
		
						global.FrameId += 1
						
						dataToSend := global.DataForChannel {
							Id: global.FrameId,
							Action: "DECLINE",
							SendedBy: nickname,
							SendedTo: sendedBy,
							PrivateId: privateId,
						}

						// SEND TO ResponseQueue
						global.ResponseQueue = append(global.ResponseQueue, dataToSend)
						continue
					}

					global.Nicknames[nickname] = global.Nicknames_data{
						HasCall: true, 
						CalledBy: sendedBy,
						ChannelId: privateId,
					}
				
					message_conn = fmt.Sprintf("%s quiere iniciar una conversación, ¿La aceptas? \n[ %s ]=> ", sendedBy, sendedTo)
					connection.Write([]byte(message_conn))
					go ReadPrivateChannel(nickname, connection, privateId)
					return
				
				}else{

					// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
					global.FrameId += 1
					dataInChannel.Id = global.FrameId
					global.CallQueue = append(global.CallQueue, dataInChannel)
				}
			}

			default:

			time.Sleep(time.Millisecond * 50)

		}
			
	}		
}	

func ReadPrivateChannel(nickname string, connection net.Conn, channelId int){
	
	var(
		connected bool
		message_conn string
	) 

	for{

		connected = checkers.CheckConnection(nickname)

		if !connected {
			return
		}
		
		if(channelId == 0){
			return
		}

		select{

		case dataInChannel := <- global.PrivateChannels[channelId]:
		
		var (
			//DATA IN CHANNEL
			sendedBy = dataInChannel.SendedBy
			sendedTo = dataInChannel.SendedTo
			action = dataInChannel.Action
			message = dataInChannel.Message
			privateId = dataInChannel.PrivateId
			
			//DATA IN NICKNAME
			callingTo = global.Nicknames[nickname].CallingTo 
			incallWith = global.Nicknames[nickname].IncallWith
			calledBy = 	global.Nicknames[nickname].CalledBy	
		)

		fmt.Printf(" [-] PrivateChannel => %s to %s by %s | readed by %s \n", action, sendedTo, sendedBy, nickname)
		switch action {
		
			case "ACCEPT":

			if(sendedTo == nickname){
				
				if(callingTo != sendedBy){
					continue
				}

				message_conn = fmt.Sprintf("Llamada aceptada por %s \n( %s <==> %s ) => ", sendedBy, sendedTo, sendedBy)
				connection.Write([]byte(message_conn))
				global.Nicknames[nickname] = global.Nicknames_data{
					IncallWith: sendedBy,
					CallingTo: "",
					ChannelId: privateId,
				}
			}else{
				
				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				global.FrameId += 1
				dataInChannel.Id = global.FrameId
				global.ResponseQueue = append(global.ResponseQueue, dataInChannel)
			}
		
			case "DECLINE":
			
			if(sendedTo == nickname){
				
				if(callingTo != sendedBy){
					print("REACH")
					continue
				}

				global.Nicknames[nickname] = global.Nicknames_data{
					HasCall: false,
					CalledBy: "",
					CallingTo: "",
					IncallWith: "",
					ChangingToPreviousStep: false,
					ChannelId: 0,
				}

				message_conn = fmt.Sprintf("Llamada rechazada por %s\n[ %s ]=> ", sendedBy, sendedTo)
				connection.Write([]byte(message_conn))
				go ReadCallChannel(nickname, connection)
				return
				
			}else{
				
				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				global.FrameId += 1
				dataInChannel.Id = global.FrameId
				global.ResponseQueue = append(global.ResponseQueue, dataInChannel)
			}

			case "SEND":

			if(sendedTo == nickname){
				tag := fmt.Sprintf(" <=[ %s ]", sendedBy)
				dataInChannel.Message = dataInChannel.Message + tag
				message = fmt.Sprintf("%s \n( %s <==> %s ) => ",dataInChannel.Message ,global.Nicknames[nickname].IncallWith ,nickname)
				connection.Write([]byte(message))
			}else{
				
				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				global.FrameId += 1
				dataInChannel.Id = global.FrameId
				global.MessageQueue = append(global.MessageQueue, dataInChannel)
			}

			case "END":
			
			if(sendedTo == nickname){
				
				if(incallWith != sendedBy && calledBy != sendedBy){
					continue
				}

				if(calledBy == sendedBy && incallWith != sendedBy){
					global.Nicknames[nickname] = global.Nicknames_data{
						HasCall: false,
						CalledBy: "",
						CallingTo: "",
						IncallWith: "",
						ChannelId: 0,
					}
	
					message = fmt.Sprintf("[!] %s ha finalizado la llamada\n[ %s ]=> ",sendedTo, sendedBy)
					connection.Write([]byte(message))
					go ReadCallChannel(nickname, connection)
					return
	
				}

				global.Nicknames[nickname] = global.Nicknames_data{
					HasCall: false,
					CalledBy: "",
					CallingTo: "",
					IncallWith: "",
					ChangingToPreviousStep: true,
					ChannelId: 0,
				}

				connection.Write([]byte(fmt.Sprintf("\n[!] Conversasión Terminada por %s\n[ %s ]=> ", incallWith, nickname)))
				
				go ReadCallChannel(nickname, connection)
				return

			}else{
				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				global.FrameId += 1
				dataInChannel.Id = global.FrameId
				global.MessageQueue = append(global.MessageQueue, dataInChannel)
			}
		}
		
		default:
			time.Sleep(time.Millisecond * time.Duration(50))
		}

	}		
}