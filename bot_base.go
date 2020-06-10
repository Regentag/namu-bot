package namu

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	cj "github.com/juju/persistent-cookiejar"
)

type Bot struct {
	name string

	wikiUrl   string
	cookieJar *cj.Jar
	client    *http.Client
}

// Create new bot instance.
func New(wikiUrl string) *Bot {
	cookieJar, _ := cj.New(&cj.Options{
		Filename: "./cookiejar.txt",
	})

	return &Bot{
		wikiUrl:   wikiUrl,
		cookieJar: cookieJar,
		client: &http.Client{
			Timeout: time.Second * 10,
			Jar:     cookieJar,
		},
	}
}

func (bot *Bot) buildUrl(uri string) string {
	return bot.wikiUrl + uri
}

// getWikiJsonObject 함수는 위키 페이지 내에 포함된 JSON 데이터를 가져온다.
// 문서 조회(/w), 편집(/edit), 역사(/history) 페이지가 JSON 데이터를 포함하고 있다.
func (bot *Bot) getWikiJsonObject(uri string) (interface{}, error) {
	// get html content
	resp, err := bot.client.Get(bot.buildUrl(uri))
	if err != nil {
		return nil, fmt.Errorf("HTTP 요청 오류. URL: %s, 오류메시지: %s.", uri, err.Error())
	}

	defer resp.Body.Close()

	// find json string
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTTP 응답 결과를 읽지 못함. %s", err.Error())
	}

	// 나무위키(The Seed 엔진)는 UTF-8을 사용한다.
	// 별도의 인코딩 처리는 불필요.
	htmlStr := string(html)

	// JSON 데이터의 위치를 찾는다.
	beg := strings.Index(htmlStr, "INITIAL_STATE=") + 14
	end := strings.Index(htmlStr, "}}</script>") + 2

	// load json object
	var obj interface{}
	if err = json.Unmarshal([]byte(htmlStr[beg:end]), &obj); err != nil {
		return nil, fmt.Errorf("JSON 값을 해석할 수 없음. %s", err.Error())
	}

	return obj, nil
}
