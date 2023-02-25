package writers

import (
	"github.com/Letder40/ChatTCP/v1/global"
	"fmt"
	"time"
)

func CallChannelWriter() {

	var lenOfCallQueue int

	for {
		lenOfCallQueue = len(global.CallQueue)

		if lenOfCallQueue != 0 {
			index := lenOfCallQueue - 1
			global.CallChannel <- global.CallQueue[index]
			fmt.Printf("( %d )WRITE Call Channel => %s to %s by [ %s CID => 0 ] \n", global.CallQueue[index].Id, global.CallQueue[index].Action, global.CallQueue[index].SendedTo, global.CallQueue[index].SendedBy)
			global.CallQueue = append(global.CallQueue[:index], global.CallQueue[index+1:]...)
		}

		time.Sleep(time.Millisecond * 100)

	}
}

func PrivateResponseWriter() {

	var lenOfResponseQueue int

	for {

		lenOfResponseQueue = len(global.ResponseQueue)

		if lenOfResponseQueue != 0 {
			index := lenOfResponseQueue - 1
			channelId := global.ResponseQueue[index].PrivateId
			fmt.Printf("( %d ) WRITE [-]PRIVATE Channel => %s to %s by [ %s CID => %d ]\n", global.ResponseQueue[index].Id, global.ResponseQueue[index].Action, global.ResponseQueue[index].SendedTo, global.ResponseQueue[index].SendedBy, channelId)
			global.PrivateChannels[channelId] <- global.ResponseQueue[index]
			global.ResponseQueue = append(global.ResponseQueue[:index], global.ResponseQueue[index+1:]...)
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
			fmt.Printf("( %d ) WRITE [-]PRIVATE Channel => %s to %s by [ %s CID => %d ]\n", global.MessageQueue[index].Id, global.MessageQueue[index].Action, global.MessageQueue[index].SendedTo, global.MessageQueue[index].SendedBy, channelId)
			global.PrivateChannels[channelId] <- global.MessageQueue[index]
			global.MessageQueue = append(global.MessageQueue[:index], global.MessageQueue[index+1:]...)
		}

		time.Sleep(time.Millisecond * 100)
	}
}
