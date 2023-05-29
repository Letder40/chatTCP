package handler

import (
	// Local modules
	"github.com/Letder40/ChatTCP/v1/autolanguage"
	"github.com/Letder40/ChatTCP/v1/checkers"
	"github.com/Letder40/ChatTCP/v1/global"
	"github.com/Letder40/ChatTCP/v1/readers"

	// Remote modules
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

func ConnectionHandler(connection net.Conn) {

	connection.Write([]byte("ChatTCP V2.0\n"))
	var nickname string

	for {

		buffer := make([]byte, 256)
		_, err := connection.Read(buffer)
		buffer = bytes.Trim(buffer, "\x00")
		buffer = bytes.Trim(buffer, "\n")
		var textInBuffer = string(buffer)
		var netError = global.NetError{}

		//If user disconnects
		if err != nil {
			return
		}

		//nickname selection

		//valid characters
		charlist := []string{
			"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
			"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
			".", "_", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0",
		}

		var nickanme_lower = strings.ToLower(textInBuffer)
		var nickanme = strings.Split(nickanme_lower, "")
		lenofNickname := len(nickanme)
		var counter = 0

		for _, char := range nickanme {
			for _, char2compare := range charlist {
				if char == char2compare {
					counter += 1
				}
			}
		}

		if counter != lenofNickname {
			jsonErr := netError.SetError("Banned chars")
			connection.Write(jsonErr)
			continue
		}

		if !checkers.CheckNickname(textInBuffer) {
			jsonErr := netError.SetError("Nickname used")
			connection.Write(jsonErr)
			continue
		}

		nickname = textInBuffer

		global.AllNicknames = append(global.AllNicknames, nickname)
		callbox := make(chan global.DataForChannel)
		global.CallChannels[nickname] = callbox
		go readers.ReadCallChannel(nickname, connection)

		global.Nicknames[nickname] = global.Nicknames_data{
			InCall:     false,
			CalledBy:   "",
			CallingTo:  "",
			IncallWith: "",
			ChannelId:  0,
		}

		break
	}

	for {

		var (
			callingTo  = global.Nicknames[nickname].CallingTo
			channelId  = global.Nicknames[nickname].ChannelId
			inCall     = global.Nicknames[nickname].InCall
			incallWith = global.Nicknames[nickname].IncallWith
		)

		//if inCall || callingTo != "" {
		//	go readers.ReadPrivateChannel(nickname, connection)
		//} else {
		//	go readers.ReadCallChannel(nickname, connection)
		//}

		buffer := make([]byte, 256)
		_, err := connection.Read(buffer)

		if err != nil {

			delete(global.Nicknames, nickname)
			delete(global.CallChannels, nickname)

			var newAllNicknames []string
			for _, element := range global.AllNicknames {
				if element != nickname {
					newAllNicknames = append(newAllNicknames, element)
				}
			}

			fmt.Printf("%s has disconnected\n", nickname)
			global.AllNicknames = newAllNicknames
			return
		}

		buffer = bytes.Trim(buffer, "\x00")
		buffer = bytes.Trim(buffer, "\n")

		var (
			textInBuffer        = string(buffer)
			bufferLen           = len(strings.Split(textInBuffer, " "))
			textInBufferSplited = strings.Split(textInBuffer, " ")
			Action              = textInBufferSplited[0]
			netError            = global.NetError{}
		)

		switch Action {

		case "language":

			if bufferLen != 2 || !global.Translation_service.Enable {
				continue
			}

			language := textInBufferSplited[1]

			if global.Translation_service.LTranslate {
				language = strings.ToLower(language)
				exists := autolanguage.LTCheck_language(language)

				if exists {
					global.Nicknames[nickname] = global.Nicknames_data{
						Language: language,
					}
					continue
				} else {
					jsonErr := netError.SetError("Invalid language")
					connection.Write(jsonErr)
					continue
				}

			} else if global.Translation_service.Deepl {
				language = strings.ToUpper(language)
				if autolanguage.DLCheck_language(language) {
					global.Nicknames[nickname] = global.Nicknames_data{
						Language: language,
					}
					continue
				} else {
					jsonErr := netError.SetError("Invalid language")
					connection.Write(jsonErr)
					continue
				}
			}
		case "list":

			if bufferLen != 1 {
				continue
			}

			jsonList, err := json.Marshal(global.AllNicknames)
			if err != nil {
				jsonErr := netError.SetError("Deserialization error")
				connection.Write(jsonErr)
				continue
			} else {
				connection.Write([]byte(jsonList))

			}

		case "call":

			if bufferLen != 2 {
				continue
			}

			var (
				nicknameToCall     = textInBufferSplited[1]
				SendtoIsincallwith = global.Nicknames[nicknameToCall].IncallWith
			)

			println(nicknameToCall)

			if SendtoIsincallwith != "" {
				jsonErr := netError.SetError("User is in call")
				connection.Write(jsonErr)
				continue
			}

			if callingTo != "" || inCall {
				jsonErr := netError.SetError("You are in call")
				connection.Write(jsonErr)
				continue
			}

			if checkers.CheckNickname(nicknameToCall) {
				jsonErr := netError.SetError("Not connected")
				connection.Write(jsonErr)
				continue
			}

			// Call initialization and changing state of users

			global.ChannelId += 1
			
			dataToSend := global.DataForChannel{
				Action:    "CALL",
				SendedBy:  nickname,
				SendedTo:  nicknameToCall,
				PrivateId: global.ChannelId,
			}

			global.Nicknames[nickname] = global.Nicknames_data{
				CallingTo: nicknameToCall,
				IncallWith: nicknameToCall,
				ChannelId: global.ChannelId,
				Language:  global.Nicknames[nickname].Language,
				InCall: true,
			}

			global.Nicknames[nicknameToCall] = global.Nicknames_data{
				IncallWith: nickname,
				HasCall:  true,
				CalledBy: nickname,
				Language: global.Nicknames[nicknameToCall].Language,
			}

			//CREATING DYNAMIC PRIVATE CHANNEL
			global.PrivateChannels[global.ChannelId] = make(chan global.DataForChannel)

			// SEND TO CALLQUEUE
			global.CallQueue = append(global.CallQueue, dataToSend)

			go readers.ReadPrivateChannel(nickname, connection)


		case "accept":

			var (
				dataSplited = strings.Split(textInBuffer, " ")
				SendedBy    = nickname
				SendedTo    = dataSplited[1]
			)

			if bufferLen != 2 {
				continue
			}

			if !global.Nicknames[nickname].HasCall {
				jsonErr := netError.SetError("Not called")
				connection.Write(jsonErr)
				continue
			}

			if global.Nicknames[SendedTo].CallingTo != nickname {
				jsonErr := netError.SetError("Not calling you")
				connection.Write(jsonErr)
				continue
			}

			privateId := global.Nicknames[nickname].ChannelId

			global.Nicknames[nickname] = global.Nicknames_data{
				CalledBy:   "",
				IncallWith: SendedTo,
				ChannelId:  privateId,
				HasCall:    false,
				InCall:     true,
				Language:   global.Nicknames[nickname].Language,
			}

			dataToSend := global.DataForChannel{
				Action:    "ACCEPT",
				SendedBy:  SendedBy,
				SendedTo:  SendedTo,
				PrivateId: privateId,
			}

			// SEND TO ResponseQueue
			global.MessageQueue = append(global.MessageQueue, dataToSend)

		case "decline":

			var (
				dataSplited = strings.Split(textInBuffer, " ")
				SendedBy    = nickname
				SendedTo    = dataSplited[1]
			)

			if bufferLen != 2 {
				continue
			}

			if !global.Nicknames[nickname].HasCall {
				jsonErr := netError.SetError("Not called")
				connection.Write(jsonErr)
				continue
			}

			if global.Nicknames[SendedTo].CallingTo != nickname {
				jsonErr := netError.SetError("Not calling you")
				connection.Write(jsonErr)
				continue
			}

			privateId := global.Nicknames[nickname].ChannelId

			global.Nicknames[nickname] = global.Nicknames_data{
				CalledBy:   "",
				IncallWith: SendedTo,
				ChannelId:  0,
				HasCall:    false,
				InCall:     false,
				Language:   global.Nicknames[nickname].Language,
			}

			dataToSend := global.DataForChannel{
				Action:    "DECLINE",
				SendedBy:  SendedBy,
				SendedTo:  SendedTo,
				PrivateId: privateId,
			}

			global.MessageQueue = append(global.MessageQueue, dataToSend)

		case "bye":

			if !inCall {
				jsonErr := netError.SetError("Not in call")
				connection.Write(jsonErr)
				continue
			}

			var dataToSend = global.DataForChannel{}

			if callingTo == "" {
				dataToSend = global.DataForChannel{
					Action:    "END",
					SendedBy:  nickname,
					SendedTo:  incallWith,
					PrivateId: channelId,
				}
			} else {
				dataToSend = global.DataForChannel{
					Action:    "END",
					SendedBy:  nickname,
					SendedTo:  incallWith,
					PrivateId: channelId,
				}
			}

			go readers.ReadCallChannel(nickname, connection)

			// SEND TO MessageQueue
			global.MessageQueue = append(global.MessageQueue, dataToSend)

			// changing user state
			global.Nicknames[nickname] = global.Nicknames_data{
				HasCall:    false,
				InCall:     false,
				CalledBy:   "",
				CallingTo:  "",
				IncallWith: "",
				ChannelId:  0,
				Language:   global.Nicknames[nickname].Language,
			}

		case "send":
			
			if !inCall {
				jsonErr := netError.SetError("Not in call")
				connection.Write(jsonErr)
				continue
			}

			message := strings.Split(textInBuffer, " ")
			message = message[1:]

			dataToSend := global.DataForChannel{
				Action:    "SEND",
				SendedBy:  nickname,
				SendedTo:  incallWith,
				Message:   strings.Join(message, ""),
				PrivateId: channelId,
			}

			global.MessageQueue = append(global.MessageQueue, dataToSend)

		default:
			jsonErr := netError.SetError("Invalid action")
			connection.Write(jsonErr)
			continue

		}
	}
}
