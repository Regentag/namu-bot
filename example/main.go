package main

import (
	"fmt"

	"github.com/Regentag/namu-bot"
)

func TestSeed() {
	bot := namu.New("https://theseed.io")
	err := bot.Login("VictoriaBot", "==SECRET==")
	if err != nil {
		fmt.Println("로그인 오류: ", err)
		return
	}

	fmt.Println("로그인 되었습니다.")

	art, err := bot.Get("test")
	if err != nil {
		fmt.Println("오류: ", err.Error())
	} else {
		fmt.Println(art)
	}

	err = bot.Edit("test", "== TEST ==", "Bot을 활용한 내용 수정 테스트.")
	if err != nil {
		fmt.Println("오류: ", err.Error())
	}
}

func TestNamu() {
	bot := namu.New("https://namu.wiki")
	err := bot.Login("rainy_bot", "==SECRET==")
	if err != nil {
		fmt.Println("로그인 오류: ", err)
		return
	}

	fmt.Println("로그인 되었습니다.")

	art, err := bot.Get("뉴트리아")
	if err != nil {
		fmt.Println("오류: ", err.Error())
	} else {
		fmt.Println(art)
	}

	stop, err := bot.IsDiscussProgress("쇠말뚝")
	if err != nil {
		fmt.Println("오류: ", err.Error())
	}
	fmt.Println("DiscussProgress: ", stop)
}

func main() {
	TestSeed()
}
