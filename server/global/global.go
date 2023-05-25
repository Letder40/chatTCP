package global

// Global vars

var(

	//Channel Inicialization
	CallChannel = make(chan DataForChannel)
	PrivateChannels = make(map[int] chan DataForChannel)

	//Nicknames's data map
	Nicknames = make(map[string]Nicknames_data)

	//Queues of WRITER MASTER
	MessageQueue = make([]DataForChannel, 0, 1024)
	CallQueue = make([]DataForChannel, 0, 1024)
	ResponseQueue = make([]DataForChannel, 0, 1024)

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
	HasCall bool
  	Language string
	CallingTo string
	CalledBy string
	IncallWith string
	ChangingToPreviousStep bool
	ChannelId int
}

type DataForChannel struct{
	Id int
	Action string
	SendedBy string
	SendedTo string
	Message string
	PrivateId int
}
