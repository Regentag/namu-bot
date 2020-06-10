package namu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/yalp/jsonpath"
)

type ArticleEditInfo struct {
	token      string
	baseRev    string
	identifier string
	//sessionHash string // X-Riko
}

// 위키 문서의 Raw data를 가져온다.
func (bot *Bot) Get(title string) (string, error) {
	json, err := bot.getWikiJsonObject("/raw/" + title)
	if err != nil {
		return "", err
	}

	contentFilter, _ := jsonpath.Prepare("$.page.data.text")
	out, err := contentFilter(json)
	if err != nil {
		return "", fmt.Errorf("[[%s]] 문서 본문을 찾을 수 없습니다. %s", title, err.Error())
	}

	return out.(string), nil
}

// IsDiscussProgress 함수는 문서에 진행중인 토론이 있는지 검사한다.
func (bot *Bot) IsDiscussProgress(title string) (bool, error) {
	json, err := bot.getWikiJsonObject("/w/" + title)
	if err != nil {
		return false, err
	}

	dp, err := jsonpath.Read(json, "$.page.data.discuss_progress")
	if err != nil {
		return false, fmt.Errorf("문서 [[%s]] - 토론 상태를 찾을 수 없습니다. %s", title, err.Error())
	}

	return dp.(bool), nil
}

// Edit 함수는 위키 문서를 수정한다.
func (bot *Bot) Edit(title, text, log string) error {
	editInfo, err := bot.getEditBaseInfo(title)
	if err != nil {
		return fmt.Errorf("문서 [[%s]] - 편집에 필요한 기본정보를 획득하지 못했습니다. %s", title, err.Error())
	}

	data := make(url.Values)
	data.Set("token", editInfo.token)
	data.Set("identifier", editInfo.identifier)
	data.Set("baserev", editInfo.baseRev)
	data.Set("text", text)
	data.Set("log", log)
	data.Set("agree", "Y")

	req, err := postMultipartFormData(bot.buildUrl("/internal/edit/"+title), data)
	if err != nil {
		return fmt.Errorf("문서 [[%s]] - multipart/form-data payload error. %s", title, err.Error())
	}

	// send POST
	// 실험 결과 X-Chika 헤더는 반드시 존재해야 하지만 값은 임의의 값이어도 되는 것으로 보이며,
	// X-Riko 헤더는 편집결과 저장에 불필요.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:76.0) Gecko/20100101 Firefox/76.0")
	req.Header.Set("X-Chika", "chika!")
	//req.Header.Set("X-Riko", editInfo.sessionHash)

	// Submit the request
	res, err := bot.client.Do(req)
	if err != nil {
		return fmt.Errorf("문서 [[%s]] - POST 요청 오류. %s", title, err.Error())
	}
	defer res.Body.Close()

	jsonRes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("문서 [[%s]] - POST 요청 결과를 읽지 못함. %s", title, err.Error())
	}

	// === 편집 결과 검사 ===
	// (1) load json object
	var obj interface{}
	if err = json.Unmarshal(jsonRes, &obj); err != nil {
		return fmt.Errorf("문서 [[%s]] - 편집 전송 결과 JSON 값을 해석할 수 없음. %s", title, err.Error())
	}

	// (2) 결과 코드 검사
	respStatus, err := jsonpath.Read(obj, "$.status")
	if err != nil {
		return fmt.Errorf("문서 [[%s]] - 편집 결과 코드를 읽지 못함. %s", title, err.Error())
	}

	//   -> HTTP Status 302: 오류 없이 편집 결과가 저장되었음. Redirection 할 조회 URL 반환됨.
	if respStatus.(float64) == 302 {
		return nil
	}

	// (3) 오류 메시지 읽기
	errMsg, err := jsonpath.Read(obj, "$.data.error.msg")
	if err != nil {
		return fmt.Errorf("문서 [[%s]] - 편집 오류 메시지를 읽지 못함. %s", title, err.Error())
	}

	return fmt.Errorf("문서 [[%s]] - %s", title, errMsg)
}

// getEditBaseInfo 함수는 문서 편집을 위한 기본 정보를 불러온다.
// 편집 token 값, Base revision, 편집자 세션의 identifier 정보를 가져온다.
func (bot *Bot) getEditBaseInfo(title string) (*ArticleEditInfo, error) {
	json, err := bot.getWikiJsonObject("/edit/" + title)
	if err != nil {
		return nil, err
	}

	token, err := jsonpath.Read(json, "$.page.data.token")
	if err != nil {
		return nil, fmt.Errorf("문서 [[%s]] - token 값을 찾을 수 없음. %s.", title, err.Error())
	}

	baseRev, err := jsonpath.Read(json, "$.page.data.body.baserev")
	if err != nil {
		return nil, fmt.Errorf("문서 [[%s]] - baserev 값을 찾을 수 없음. %s.", title, err.Error())
	}

	identifier, err := jsonpath.Read(json, "$.session.identifier")
	if err != nil {
		return nil, fmt.Errorf("문서 [[%s]] - identifier 값을 찾을 수 없음. %s.", title, err.Error())
	}
	/*
		hash, err := jsonpath.Read(json, "$.session.hash")
		if err != nil {
			return nil, fmt.Errorf("Edit [[%s]] - hash not found. %s", title, err.Error())
		}
	*/
	return &ArticleEditInfo{
		token:      token.(string),
		baseRev:    baseRev.(string),
		identifier: identifier.(string),
		//		sessionHash: hash.(string),
	}, nil
}

// multipart/form-data 형식의 HTTP POST 요청을 만든다.
func postMultipartFormData(url string, values url.Values) (*http.Request, error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for key := range values {
		if fw, err := w.CreateFormField(key); err != nil {
			return nil, err
		} else {
			fmt.Fprint(fw, values.Get(key))
		}
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, err
	}

	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	return req, nil
}
