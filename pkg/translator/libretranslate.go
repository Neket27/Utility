package translator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Response struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
	} `json:"responseData"`
}

func Translate(text, from, to string) (string, error) {
	baseURL := "https://api.mymemory.translated.net/get"
	params := url.Values{}
	params.Add("q", text)
	params.Add("langpair", from+"|"+to)

	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return "", fmt.Errorf("ошибка запроса: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("статус ответа: %d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	return result.ResponseData.TranslatedText, nil
}
