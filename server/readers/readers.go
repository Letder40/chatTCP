package readers

import (
	"fmt"
	"net"

	"github.com/Letder40/ChatTCP/v1/autolanguage"
	"github.com/Letder40/ChatTCP/v1/checkers"
	"github.com/Letder40/ChatTCP/v1/global"
)

func ReadCallChannel(nickname string, connection net.Conn) {

	var connected bool

	for {
		
		dataInChannel := <-global.CallChannels[nickname]

		connected = checkers.CheckConnection(nickname)

		var (
			netdata = global.NetData{}
		)

		if !connected {
			return
		}


		var (
			sendedBy  = dataInChannel.SendedBy
			sendedTo  = dataInChannel.SendedTo
			action    = dataInChannel.Action
			privateId = dataInChannel.PrivateId
		)

		println(action)

		fmt.Printf("READ Call Channel => %s to %s by %s | read by %s \n", action, sendedTo, sendedBy, nickname)

		switch action {
		case "CALL":

			global.Nicknames[nickname] = global.Nicknames_data{
				HasCall:   true,
				CalledBy:  sendedBy,
				IncallWith: sendedBy,
				ChannelId: privateId,
				Language:  global.Nicknames[nickname].Language,
			}

			jsonState := netdata.SetState(action, sendedBy)

			connection.Write([]byte(jsonState))

			go ReadPrivateChannel(nickname, connection)

			return

		}

	}
}

func ReadPrivateChannel(nickname string, connection net.Conn) {
	

	var connected bool


	for {
		println("reach!")
		channelId := global.Nicknames[nickname].ChannelId

		dataInChannel := <-global.PrivateChannels[channelId]

		var (
			netdata = global.NetData{}
		)

		connected = checkers.CheckConnection(nickname)

		if !connected {
			return
		}


		var (
			//DATA IN CHANNEL
			sendedBy  = dataInChannel.SendedBy
			sendedTo  = dataInChannel.SendedTo
			action    = dataInChannel.Action
			message   = dataInChannel.Message
			privateId = dataInChannel.PrivateId
		)

		fmt.Printf(" [-] PrivateChannel => %s to %s by %s | readed by %s \n", action, sendedTo, sendedBy, nickname)

		switch action {

		case "ACCEPT":

			if sendedTo != nickname {
				global.MessageQueue = append(global.MessageQueue, dataInChannel)
				continue
			}

			jsonState := netdata.SetState(action, sendedBy)
			connection.Write([]byte(jsonState))

			global.Nicknames[nickname] = global.Nicknames_data{
				IncallWith: sendedBy,
				InCall:     true,
				CallingTo:  "",
				ChannelId:  privateId,
				Language:   global.Nicknames[nickname].Language,
			}

			go ReadPrivateChannel(nickname, connection)
			return

		case "DECLINE":

			if sendedTo != nickname {
				global.MessageQueue = append(global.MessageQueue, dataInChannel)
				continue
			}

			jsonState := netdata.SetState(action, sendedBy)
			connection.Write([]byte(jsonState))

			global.Nicknames[nickname] = global.Nicknames_data{
				IncallWith: "",
				InCall:     false,
				HasCall:    false,
				CallingTo:  "",
				ChannelId:  privateId,
				CalledBy:   "",
				Language:   global.Nicknames[nickname].Language,
			}


		case "END":

			if sendedTo != nickname {
				global.MessageQueue = append(global.MessageQueue, dataInChannel)
				continue
			}

			global.Nicknames[nickname] = global.Nicknames_data{
				HasCall:    false,
				CalledBy:   "",
				CallingTo:  "",
				IncallWith: "",
				ChannelId:  0,
				Language:   global.Nicknames[nickname].Language,
			}

			jsonState := netdata.SetState(action, sendedBy)
			connection.Write([]byte(jsonState))

		case "SEND":

			println(sendedTo)
			println(nickname)

			if sendedTo != nickname {
				global.MessageQueue = append(global.MessageQueue, dataInChannel)
				continue
			}

			if global.Translation_service.Enable {

				var (
					dst_lang = global.Nicknames[nickname].Language
					src_lang = global.Nicknames[sendedBy].Language
				)

				if global.Translation_service.LTranslate {
					message = autolanguage.LTGet_translation(src_lang, dst_lang, dataInChannel.Message)
				} else if global.Translation_service.Deepl {
					message = autolanguage.DLGet_translation(src_lang, dst_lang, dataInChannel.Message)
				}

			} else {
				message = dataInChannel.Message
			}

			var netMessage = global.NetMessage{}
			var message = netMessage.SetMessage(message, sendedBy)
			connection.Write(message)
		}

	}
}
