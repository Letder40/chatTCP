package checkers

import(
	"github.com/Letder40/ChatTCP/v1/global"
	"fmt"
)

func CheckConnection(nickname string) bool {
	for _, element := range global.AllNicknames{
		if(element == nickname){
			return true
		}
	}
	global.Nicknames[nickname] = global.Nicknames_data{}
	return false
}

func CheckNickname(nickname string) bool{
	for _, element := range global.AllNicknames {
		if(element == nickname){
			return false
		}
	}
	return true
}

func List(allNicknames []string) string {
	var list string
	for _, element := range allNicknames {
		list += fmt.Sprintf("-> %s \n", element)
	}
	return list
}
