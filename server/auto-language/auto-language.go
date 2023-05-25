package autolanguage

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

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

func Check_language(language string) (bool) {
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

func Get_translation(language string, message string) (string){
  form := url.Values{}
  form.Add("q", message)
  form.Add("source", "auto")
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
