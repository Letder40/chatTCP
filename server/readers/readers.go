package readers

import (
	"fmt"
	"net"

	autolanguage "github.com/Letder40/ChatTCP/v1/auto-language"
	"github.com/Letder40/ChatTCP/v1/checkers"
	"github.com/Letder40/ChatTCP/v1/global"
)

func ReadCallChannel(nickname string, connection net.Conn) {

	var (
		connected    bool
		message_conn string
	)

	for {
		connected = checkers.CheckConnection(nickname)

		var (
			channelId  = global.Nicknames[nickname].ChannelId
			incallWith = global.Nicknames[nickname].IncallWith
		)

		if !connected {
			return
		}
		if channelId != 0 {
			return
		}
		if incallWith != "" {
			return
		}

		dataInChannel := <-global.CallChannel

		var (
			sendedBy  = dataInChannel.SendedBy
			sendedTo  = dataInChannel.SendedTo
			action    = dataInChannel.Action
			privateId = dataInChannel.PrivateId
		)

		fmt.Printf("( %d ) READ Call Channel => %s to %s by %s | readed by %s \n", dataInChannel.Id, action, sendedTo, sendedBy, nickname)

		if action == "CALL" {
			if sendedTo == nickname {

				if incallWith != "" {

					global.FrameId += 1

					dataToSend := global.DataForChannel{
						Id:        global.FrameId,
						Action:    "DECLINE",
						SendedBy:  nickname,
						SendedTo:  sendedBy,
						PrivateId: privateId,
					}
					// SEND TO ResponseQueue
					global.ResponseQueue = append(global.ResponseQueue, dataToSend)
					continue
				}
				global.Nicknames[nickname] = global.Nicknames_data{
					HasCall:   true,
					CalledBy:  sendedBy,
					ChannelId: privateId,
					Language:  global.Nicknames[nickname].Language,
				}

				message_conn = fmt.Sprintf("%s is calling you, accept/decline \n[ %s ]=> ", sendedBy, sendedTo)
				connection.Write([]byte(message_conn))
				go ReadPrivateChannel(nickname, connection, privateId)
				return

			} else {
				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				global.FrameId += 1
				dataInChannel.Id = global.FrameId
				global.CallQueue = append(global.CallQueue, dataInChannel)
			}
		}

	}
}

func ReadPrivateChannel(nickname string, connection net.Conn, channelId int) {

	var (
		connected    bool
		message_conn string
	)

	for {

		connected = checkers.CheckConnection(nickname)

		if !connected {
			return
		}

		if channelId == 0 {
			return
		}

		dataInChannel := <-global.PrivateChannels[channelId]

		var (
			//DATA IN CHANNEL
			sendedBy  = dataInChannel.SendedBy
			sendedTo  = dataInChannel.SendedTo
			action    = dataInChannel.Action
			message   = dataInChannel.Message
			privateId = dataInChannel.PrivateId

			//DATA IN NICKNAME
			callingTo  = global.Nicknames[nickname].CallingTo
			incallWith = global.Nicknames[nickname].IncallWith
			calledBy   = global.Nicknames[nickname].CalledBy
		)

		fmt.Printf(" [-] PrivateChannel => %s to %s by %s | readed by %s \n", action, sendedTo, sendedBy, nickname)
		switch action {

		case "ACCEPT":

			if sendedTo == nickname {

				if callingTo != sendedBy {
					continue
				}

				message_conn = fmt.Sprintf("Call accepted by %s \n( %s <==> %s ) => ", sendedBy, sendedTo, sendedBy)
				connection.Write([]byte(message_conn))
				global.Nicknames[nickname] = global.Nicknames_data{
					IncallWith: sendedBy,
					CallingTo:  "",
					ChannelId:  privateId,
					Language:   global.Nicknames[nickname].Language,
				}
			} else {

				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				global.FrameId += 1
				dataInChannel.Id = global.FrameId
				global.ResponseQueue = append(global.ResponseQueue, dataInChannel)
			}

		case "DECLINE":

			if sendedTo == nickname {

				if callingTo != sendedBy {
					print("REACH")
					continue
				}

				global.Nicknames[nickname] = global.Nicknames_data{
					HasCall:                false,
					CalledBy:               "",
					CallingTo:              "",
					IncallWith:             "",
					ChangingToPreviousStep: false,
					ChannelId:              0,
					Language:               global.Nicknames[nickname].Language,
				}

				message_conn = fmt.Sprintf("Call declined by %s\n[ %s ]=> ", sendedBy, sendedTo)
				connection.Write([]byte(message_conn))
				go ReadCallChannel(nickname, connection)
				return

			} else {

				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				global.FrameId += 1
				dataInChannel.Id = global.FrameId
				global.ResponseQueue = append(global.ResponseQueue, dataInChannel)
			}

		case "SEND":

			if sendedTo == nickname {
				tag := fmt.Sprintf(" <=[ %s ]", sendedBy)
				language := global.Nicknames[nickname].Language
				println(language)
				if language != "" {
					message = autolanguage.Get_translation(language, dataInChannel.Message)
				} else {
					message = dataInChannel.Message
				}
				message = message + tag
				message = fmt.Sprintf("%s \n( %s <==> %s ) => ", message, global.Nicknames[nickname].IncallWith, nickname)
				connection.Write([]byte(message))
			} else {
				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				global.FrameId += 1
				dataInChannel.Id = global.FrameId
				global.MessageQueue = append(global.MessageQueue, dataInChannel)
			}

		case "END":

			if sendedTo == nickname {

				if incallWith != sendedBy && calledBy != sendedBy {
					continue
				}

				if calledBy == sendedBy && incallWith != sendedBy {
					global.Nicknames[nickname] = global.Nicknames_data{
						HasCall:    false,
						CalledBy:   "",
						CallingTo:  "",
						IncallWith: "",
						ChannelId:  0,
						Language:   global.Nicknames[nickname].Language,
					}

					message = fmt.Sprintf("[!] %s finished the call\n[ %s ]=> ", sendedTo, sendedBy)
					connection.Write([]byte(message))
					go ReadCallChannel(nickname, connection)
					return

				}

				global.Nicknames[nickname] = global.Nicknames_data{
					HasCall:                false,
					CalledBy:               "",
					CallingTo:              "",
					IncallWith:             "",
					ChangingToPreviousStep: true,
					ChannelId:              0,
					Language:               global.Nicknames[nickname].Language,
				}

				connection.Write([]byte(fmt.Sprintf("\n[!] Call finished by %s\n[ %s ]=> ", incallWith, nickname)))

				go ReadCallChannel(nickname, connection)
				return

			} else {
				// FORWARDING THE DATA THAT HASN'T REACH TO THE DESTINATION.
				global.FrameId += 1
				dataInChannel.Id = global.FrameId
				global.MessageQueue = append(global.MessageQueue, dataInChannel)
			}

		}

	}
}
