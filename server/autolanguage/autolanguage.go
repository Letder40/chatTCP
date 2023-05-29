package autolanguage

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/Letder40/ChatTCP/v1/global"
)

// Libretranslate

type lan_flags []struct {
	Code    string   `json:"code"`
	Name    string   `json:"name"`
	Targets []string `json:"targets"`
}

type translation struct {
	DetectedLanguage struct {
		Confidence int    `json:"confidence"`
		Language   string `json:"language"`
	} `json:"detectedLanguage"`
	TranslatedText string `json:"translatedText"`
}


func LTCheck_language(language string) (bool) {
  response, err := http.Get("http://127.0.0.1:5000/languages")
  if err != nil {
    log.Fatal(err) 
  }
  //fmt.Println(language)
  response_body, err := io.ReadAll(response.Body) 
  if err != nil {
    log.Fatal(err) 
  }
  var lan_flags lan_flags
  json.Unmarshal(response_body, &lan_flags)
  all_lan_flags := lan_flags[0].Targets
  all_lan_flags = append(all_lan_flags, "en")
  exists := false
  for _,lang := range(all_lan_flags){
    if lang == strings.Trim(language, "\n") {
      exists = true 
    }
  }
  return exists
}

func LTGet_translation(src_lang string, language string, message string) (string){
  form := url.Values{}
  form.Add("q", message)
  if src_lang != "" {
    form.Add("source", "auto")
  }else{
  form.Add("source", "auto")
  }
  form.Add("target", language)
  form.Add("format", "text")


  response, err := http.PostForm("http://127.0.0.1:5000/translate", form)
  if err != nil {
    log.Fatal(err)
  }
  response_body, err := io.ReadAll(response.Body) 
  if err != nil {
    log.Fatal(err) 
  }
  var translation translation
  json.Unmarshal(response_body, &translation)
  return translation.TranslatedText
}

// Deepl

type dltranslation struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

func DLCheck_language (language string) (bool) {
  supported_languages := [8]string{"EN","DE","FR","ES","JA","IT","PL","NL"}
  exists := false
  for _,lang:= range(supported_languages){
    if lang == strings.Trim(language, "\n") {
      exists = true 
    }
  }
  return exists
}

func DLGet_translation(src_lang string, dst_lang string, message string) (string){
  data := url.Values{}
  data.Set("text", message)
  data.Set("target_lang", dst_lang)
  if src_lang != "" {
  data.Set("source_lang", src_lang)

  }
  
  request, err := http.NewRequest("POST", "https://api-free.deepl.com/v2/translate", bytes.NewBufferString(data.Encode()))
  if err != nil {
    log.Fatal(err) 
  }
  
  request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  request.Header.Add("Authorization", "DeepL-Auth-Key " + global.Translation_service.Deepl_api_key)

  /*request_body, err := io.ReadAll(request.Body) 
  if err != nil {
    log.Fatal(err) 
  }

  log.Println(string(request_body))
  */

  client := &http.Client{}

  response,err := client.Do(request)
  if err != nil {
    log.Fatalf("Got error %s", err.Error())
  } 

  response_body, err := io.ReadAll(response.Body) 
  if err != nil {
    log.Fatal(err) 
  }
  log.Println(string(response_body))

  var dltranslation dltranslation
  json.Unmarshal(response_body, &dltranslation)
  return dltranslation.Translations[0].Text
}
