package handler

import (
	// Local modules
	autolanguage "github.com/Letder40/ChatTCP/v1/auto-language"
	"github.com/Letder40/ChatTCP/v1/checkers"
	"github.com/Letder40/ChatTCP/v1/global"
	"github.com/Letder40/ChatTCP/v1/readers"

	// Remote modules
	"bytes"
	"fmt"
	"net"
	"strings"
)

func ConnectionHandler(connection net.Conn) {

	var (
		step     = 1
		message  string
		nickname string
		prompt   string
	)

	connection.Write([]byte("ChatTCP V1.0\n"))

	for {

		buffer := make([]byte, 1024*5)
		_, err := connection.Read(buffer)

		buffer = bytes.Trim(buffer, "\x00")
		buffer = bytes.Trim(buffer, "\n")

		var (
			callingTo              = global.Nicknames[nickname].CallingTo
			changingToPreviousStep = global.Nicknames[nickname].ChangingToPreviousStep
			channelId              = global.Nicknames[nickname].ChannelId
			hasCall                = global.Nicknames[nickname].HasCall
			incallWith             = global.Nicknames[nickname].IncallWith
		)

		var (
			textInBuffer        = string(buffer)
			bufferLen           = len(strings.Split(textInBuffer, " "))
			textInBufferSplited = strings.Split(textInBuffer, " ")
			Action              = textInBufferSplited[0]
		)



		//If user disconnects
		if err != nil {

			global.Nicknames[nickname] = global.Nicknames_data{}

			var new_allNicknames []string
			for _, element := range global.AllNicknames {
				if element != nickname {
					new_allNicknames = append(new_allNicknames, element)
				}
			}
			fmt.Printf("%s has disconnected\n", nickname)
			global.AllNicknames = new_allNicknames
			return
		}

		if changingToPreviousStep {
			step = 3
			prompt = fmt.Sprintf("[ %s ]=> ", nickname)
			global.Nicknames[nickname] = global.Nicknames_data{
				ChangingToPreviousStep: false,
				Language:               global.Nicknames[nickname].Language,
			}
		}

		switch step {
		case 1:

			if textInBuffer != "Go Connect" {
				connection.Write([]byte("Bad usage of the ChatTCP protocol\n"))
				connection.Close()
			}

			connection.Write([]byte("Select nickname: "))
			step += 1

		case 2:

			if textInBuffer != "" {
				nickname = textInBuffer

				if !checkers.CheckNickname(nickname) {
					connection.Write([]byte("\n[!] Someone is using that nickname\n\nSelect other nickname: "))
					continue
				}

				if bufferLen != 1 {
					connection.Write([]byte("\n[!] The nickname cannot have whitespaces\n\nSelect other nickname: "))
					continue
				}

				global.AllNicknames = append(global.AllNicknames, nickname)

				global.Nicknames[nickname] = global.Nicknames_data{
					HasCall:                false,
					CalledBy:               "",
					CallingTo:              "",
					IncallWith:             "",
					ChangingToPreviousStep: false,
					ChannelId:              0,
					Language:               global.Nicknames[nickname].Language,
				}

				prompt = fmt.Sprintf("[ %s ]=> ", nickname)
				if global.Translation_functionality {
					message = fmt.Sprintf("Welcome %s \n\n[?] help for ChatTCP:\nlist -> Show the nicknames of all connected users\ncall nickname -> call someone by the user nickname\naccept [nickanme of the other user]-> Accept a call\ndecline [nickanme of the other user] -> decline a call\n\n[ Feature: Auto Translation enabled ] \nlanguage [ two first letters of your language ] -> set a language to automatically translate all received messages into it\nExample:\nlanguage en \n\n%s", nickname, prompt)
				} else {
					message = fmt.Sprintf("Welcome %s \n\n[?] help for ChatTCP:\nlist -> Show the nicknames of all connected users\ncall nickname -> call someone by the user nickname\naccept [nickanme of the other user]-> Accept a call\ndecline [nickanme of the other user] -> decline a call\n%s", nickname, prompt)
				}
				connection.Write([]byte(message))

				go readers.ReadCallChannel(nickname, connection)

				step += 1
			}

		case 3:
			//Translation_functionality
			if Action == "language" && bufferLen == 2 && global.Translation_functionality {
				language := textInBufferSplited[1]
				exists := autolanguage.Check_language(language)

				if exists {
					global.Nicknames[nickname] = global.Nicknames_data{
						Language: language,
					}
					connection.Write([]byte(prompt))
					continue
				} else {
					message = fmt.Sprintf("[!] Invalid language code\n%s", prompt)
					connection.Write([]byte(message))
					continue
				}

				//LIST
			} else if Action == "list" && bufferLen == 1 {

				message = fmt.Sprintf("\n%s\n %s", checkers.List(global.AllNicknames), prompt)
				connection.Write([]byte(message))

				//CALL
			} else if Action == "call" && bufferLen == 2 {

				var (
					nicknameToCall     = textInBufferSplited[1]
					SendtoIsincallwith = global.Nicknames[nicknameToCall].IncallWith
				)

				if SendtoIsincallwith != "" {
					message = fmt.Sprintf("[!] %s is already in a call\n%s", nicknameToCall, prompt)
					connection.Write([]byte(message))
					continue
				}

				if callingTo != "" || hasCall {
					message = fmt.Sprintf("[!] You are already in a call\n%s", prompt)
					connection.Write([]byte(message))
					continue
				}

				if !checkers.CheckNickname(nicknameToCall) {

					if nicknameToCall == nickname {
						message = fmt.Sprintf("[!] You cannot call yourself\n%s", prompt)
						connection.Write([]byte(message))
						continue
					}

					// Making the call

					global.ChannelId += 1
					global.FrameId += 1

					dataToSend := global.DataForChannel{
						Id:        global.FrameId,
						Action:    "CALL",
						SendedBy:  nickname,
						SendedTo:  nicknameToCall,
						PrivateId: global.ChannelId,
					}

					global.Nicknames[nickname] = global.Nicknames_data{
						CallingTo:  nicknameToCall,
						ChannelId:  global.ChannelId,
						IncallWith: nicknameToCall,
						Language:   global.Nicknames[nickname].Language,
					}

					global.Nicknames[nicknameToCall] = global.Nicknames_data{
						HasCall:  true,
						CalledBy: nickname,
						Language: global.Nicknames[nicknameToCall].Language,
					}

					message = fmt.Sprintf("[?] Send bye to stop the call\n( Waitting.... )=[ %s ]=> ", nickname)
					connection.Write([]byte(message))

					//CREATING DYNAMIC PRIVATE CHANNEL
					global.PrivateChannels[global.ChannelId] = make(chan global.DataForChannel)

					// SEND TO CALLQUEUE
					global.CallQueue = append(global.CallQueue, dataToSend)

					go readers.ReadPrivateChannel(nickname, connection, global.ChannelId)

					step += 1

				} else {
					message = fmt.Sprintf("\n[!] The nickname: %s, is not connected\n\n%s", nicknameToCall, prompt)
					connection.Write([]byte(message))
				}

				//ACCEPT
			} else if Action == "accept" && bufferLen == 2 {

				var (
					dataSplited = strings.Split(textInBuffer, " ")
					SendedBy    = nickname
					SendedTo    = dataSplited[1]
				)

				if !global.Nicknames[nickname].HasCall {
					message = fmt.Sprintf("[!] Nobody is calling you\n%s", prompt)
					connection.Write([]byte(message))
					continue
				}

				if global.Nicknames[SendedTo].CallingTo != nickname {
					message = fmt.Sprintf("[!] %s is not calling you\n%s", SendedTo, prompt)
					connection.Write([]byte(message))
					continue
				}

				privateId := global.Nicknames[nickname].ChannelId

				global.Nicknames[nickname] = global.Nicknames_data{
					CalledBy:   "",
					IncallWith: SendedTo,
					ChannelId:  privateId,
					Language:   global.Nicknames[nickname].Language,
				}

				global.FrameId += 1

				dataToSend := global.DataForChannel{
					Id:        global.FrameId,
					Action:    "ACCEPT",
					SendedBy:  SendedBy,
					SendedTo:  SendedTo,
					PrivateId: privateId,
				}

				fmt.Printf("ID ==> %d)\n", privateId)

				// SEND TO ResponseQueue
				global.ResponseQueue = append(global.ResponseQueue, dataToSend)

				prompt = fmt.Sprintf("( %s <==> %s ) => ", global.Nicknames[nickname].IncallWith, nickname)
				connection.Write([]byte(prompt))

				step += 1

				//DECLINE
			} else if Action == "decline" && bufferLen == 2 {

				var (
					dataSplited = strings.Split(textInBuffer, " ")
					SendedBy    = nickname
					SendedTo    = dataSplited[1]
					privateId   = global.Nicknames[nickname].ChannelId
					called_by   = global.Nicknames[nickname].CalledBy
				)

				if called_by == SendedTo {

					global.Nicknames[nickname] = global.Nicknames_data{
						HasCall:                false,
						CalledBy:               "",
						CallingTo:              "",
						IncallWith:             "",
						ChangingToPreviousStep: false,
						ChannelId:              0,
						Language:               global.Nicknames[nickname].Language,
					}

					global.FrameId += 1

					dataToSend := global.DataForChannel{
						Id:        global.FrameId,
						Action:    "DECLINE",
						SendedBy:  SendedBy,
						SendedTo:  SendedTo,
						PrivateId: privateId,
					}

					// SEND TO ResponseQueue
					global.ResponseQueue = append(global.ResponseQueue, dataToSend)
					go readers.ReadCallChannel(nickname, connection)

					connection.Write([]byte(prompt))

				} else {
					connection.Write([]byte("[!] That user is not calling you"))
					connection.Write([]byte(prompt))

					continue
				}

			} else if Action == "user_info" && bufferLen == 1 {
				connection.Write([]byte(fmt.Sprintf("\n\nnickname => %s\nchannel id => %d\nIn call with => %s\nLanguage => %s\n%s", nickname, channelId, incallWith,  global.Nicknames[nickname].Language, prompt)))

			} else if Action == "bye" && bufferLen == 1 {
				connection.Write([]byte("[!] You are not calling or in a conversation"))

			} else {
				message = fmt.Sprintf("\n[!] Not a command\n\n%s", prompt)
				connection.Write([]byte(message))

			}

		case 4:

			if incallWith == "" && callingTo == "" {
				step = 3
				prompt = fmt.Sprintf("[ %s ]=> ", nickname)
				continue
			}

			if Action == "bye" && bufferLen == 1 {
				prompt = fmt.Sprintf("[ %s ]=> ", nickname)

				global.FrameId += 1

				dataToSend := global.DataForChannel{
					Id:        global.FrameId,
					Action:    "END",
					SendedBy:  nickname,
					SendedTo:  incallWith,
					PrivateId: channelId,
				}

				if callingTo != "" {

					global.FrameId += 1

					dataToSend = global.DataForChannel{
						Id:        global.FrameId,
						Action:    "END",
						SendedBy:  nickname,
						SendedTo:  callingTo,
						PrivateId: channelId,
					}

				}

				// SEND TO MessageQueue
				global.MessageQueue = append(global.MessageQueue, dataToSend)

				global.Nicknames[nickname] = global.Nicknames_data{
					HasCall:                false,
					CalledBy:               "",
					CallingTo:              "",
					IncallWith:             "",
					ChangingToPreviousStep: false,
					ChannelId:              0,
					Language:               global.Nicknames[nickname].Language,
				}

				connection.Write([]byte(prompt))
				step = 3
				go readers.ReadCallChannel(nickname, connection)
				continue

			} else if Action == "user_info" && bufferLen == 1 {
				connection.Write([]byte(fmt.Sprintf("\n\nnickname => %s\nchannel id => %d\nIn call with => %s\nLanguage => %s\n%s", nickname, channelId, incallWith,  global.Nicknames[nickname].Language, prompt)))
				continue
			}

			dataToSend := global.DataForChannel{
				Action:    "SEND",
				SendedBy:  nickname,
				SendedTo:  incallWith,
				Message:   textInBuffer,
				PrivateId: channelId,
			}

			global.MessageQueue = append(global.MessageQueue, dataToSend)

			prompt = fmt.Sprintf("( %s <==> %s ) => ", incallWith, nickname)
			connection.Write([]byte(prompt))

		}
	}
}
