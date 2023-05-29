package writers

import (
	"github.com/Letder40/ChatTCP/v1/global"
	"time"
)

func CallChannelWriter() {

	var lenOfCallQueue int

	for {
		lenOfCallQueue = len(global.CallQueue)

		if lenOfCallQueue != 0 {
			index := lenOfCallQueue - 1

			nicknametoSend := global.CallQueue[index].SendedTo
			global.CallChannels[nicknametoSend] <- global.CallQueue[index]
			global.CallQueue = append(global.CallQueue[:index], global.CallQueue[index+1:]...)
		}

		time.Sleep(time.Millisecond * 100)

	}
}


func PrivateMessageWriter() {

	var lenOfMessageQueue int

	for {

		lenOfMessageQueue = len(global.MessageQueue)

		if lenOfMessageQueue != 0 {
			lenOfMessageQueue = len(global.MessageQueue)
			index := lenOfMessageQueue - 1

			channelId := global.MessageQueue[index].PrivateId
			println(channelId)
			global.PrivateChannels[channelId] <- global.MessageQueue[index]

			global.MessageQueue = append(global.MessageQueue[:index], global.MessageQueue[index+1:]...)
		}

		time.Sleep(time.Millisecond * 100)
	}
}
