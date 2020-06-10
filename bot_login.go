package namu

import (
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// Login 함수는 위키 로그인을 처리한다.
// 단순 조회는 로그인이 필요 없으나, 문서의 편집에는 필요하다.
func (bot *Bot) Login(username, password string) error {
	// update bot name
	bot.name = username

	data := make(url.Values)
	data.Set("username", username)
	data.Set("password", password)

	resp, err := bot.client.PostForm(bot.buildUrl("/member/login"), data)
	if err != nil {
		return fmt.Errorf("로그인 - HTTP POST 요청 오류. %s.", err.Error())
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	// PIN 입력 페이지인지 확인한다.
	pinElem := doc.Find("input#pinInput")
	if len(pinElem.Nodes) == 0 {
		bot.cookieJar.Save()
		return nil
	}

	// PIN 입력 필요
	return bot.pin()
}

// PIN 함수는 표준입력으로 PIN 값을 입력받아 위키로 전달한다.
// `trust` 값을 `true`로 전송하여 '이 기기를 항상 신뢰' 하도록 한다.
func (bot *Bot) pin() error {
	fmt.Println("확인되지 않은 기기에서 로그인 하였습니다. 메일로 전송된 PIN을 입력하십시오.")
	fmt.Print("PIN: ")

	var pin string
	n, err := fmt.Scanln(&pin)
	if err != nil {
		return fmt.Errorf("PIN 입력 처리 오류: %s.", err.Error())
	}

	if n != 1 {
		return fmt.Errorf("PIN을 올바르게 입력하십시오.")
	}

	data := make(url.Values)
	data.Set("pin", pin)
	data.Set("trust", "true")

	resp, err := bot.client.PostForm(bot.buildUrl("/member/login/pin"), data)
	if err != nil {
		return fmt.Errorf("PIN 입력 - HTTP POST 요청 오류. %s.", err.Error())
	}
	defer resp.Body.Close()

	// TODO: 여기서 PIN 입력 결과 확인 필요!

	bot.cookieJar.Save()
	return nil
}
