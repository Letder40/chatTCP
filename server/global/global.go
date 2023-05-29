package global

import "encoding/json"

// Global vars

var(

	//Channels Inicialization
	CallChannels = make(map[string] chan DataForChannel)
	PrivateChannels = make(map[int] chan DataForChannel)

	//Nicknames's data map
	Nicknames = make(map[string]Nicknames_data)

	//Queues of WRITER MASTER
	MessageQueue = make([]DataForChannel, 0, 1024)
	CallQueue = make([]DataForChannel, 0, 1024)

	//Nicknames
	AllNicknames []string 

	//Global identifiers
	ChannelId = 0
	FrameId = 0

  	//Translation functionality
	Translation_service = Translation {}
)
type Translation struct {
	Enable bool
	Deepl bool
	Deepl_api_key string
	LTranslate bool
}

type Nicknames_data struct{
	InCall bool
	HasCall bool
  	Language string
	CallingTo string
	CalledBy string
	IncallWith string
	ChannelId int
}

type DataForChannel struct{
	Action string
	SendedBy string
	SendedTo string
	Message string
	PrivateId int
}

type NetError struct {
	Error string `json:"error"`
}

func (e *NetError) SetError(jsonError string) []byte {
	e.Error = jsonError
	errorJson, _ := json.Marshal(e)
	return []byte(errorJson)
}

type NetData struct {
	Action string `json:"state"`
	FromUser string `json:"fromUser"`
}

func (e *NetData) SetState(action string, fromUser string) []byte {
	e.Action = action
	e.FromUser = fromUser
	stateJson, _ := json.Marshal(e)
	return []byte(stateJson)
}

type NetMessage struct {
	Message string `json:"Message"`
	FromUser string `json:"fromUser"`
}

func (e *NetMessage) SetMessage(message string, fromUser string) []byte {
	e.Message = message
	e.FromUser = fromUser
	messageJson, _ := json.Marshal(e)
	return []byte(messageJson)
}