package AbstractAPI

import (
	"DeepWorkload/conf"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type AccessAPI interface {
	GetAccessToken() string
	RefreshAccessToken()
}

type ComponentAccessAPI interface {
	GetComponentAccess() string
	RefreshComponentAccess()
}

type AuthorizeAccessAPI interface {
	GetAuthorizeAccess() string
	RefreshAuthorizeAccess()
}

type API struct {
	AccessAPI
	ComponentAccessAPI
	AuthorizeAccessAPI
}

func (a *API) HttpCall(urlType []string, args map[string]interface{}) (map[string]interface{}, error) {
	shortUrl := urlType[0]
	method := urlType[1]
	retryCnt := 0
	data := make(map[string]interface{}, 0)
	var response []byte
	var err error
	for retryCnt < 3 {
		switch method {
		case "POST":
			url := MakeUrl(shortUrl)
			url, err := a.appendUrlArgs(url, args)
			if err != nil {
				return nil, err
			}
			var postData string
			if args != nil {
				jsonStr, err := json.Marshal(args)
				if err != nil {
					return nil, err
				}
				postData = string(jsonStr)
			}
			response, err = a.httpPost(url, postData)

		case "GET":
			url := MakeUrl(shortUrl)
			url, err = a.appendArgs(url, args)
			if err != nil {
				return nil, err
			}
			response, err = a.httpGet(url)
		default:
			return nil, errors.New("unknown method type")
		}

		if conf.DEBUG {
			fmt.Println(string(response))
		}

		err = json.Unmarshal(response, &data)
		if err != nil {
			retryCnt++
			continue
		}
		errCode, _ := data["errcode"].(float64)
		//if !ok {
		//	err = errors.New("response didn't contains errcode")
		//	retryCnt++
		//	continue
		//}

		if a.tokenExpired(errCode) {
			a.refreshToken(shortUrl)
			retryCnt++
			continue
		}
		break
	}
	if err != nil {
		return nil, err
	}
	return a.checkResponse(data)
}

func (a *API) appendUrlArgs(url string, args map[string]interface{}) (string, error) {
	if args == nil {
		return url, nil
	}
	var err error
	if args["urlArgs"] != nil {
		url, err = a.appendArgs(url, args["urlArgs"].(map[string]interface{}))
		if err != nil {
			return url, err
		}
		delete(args, "urlArgs")
	}
	return url, nil
}

func (a *API) appendArgs(url string, args map[string]interface{}) (string, error) {
	if args == nil {
		return url, nil
	}
	for key, value := range args {
		if strings.Index(url, "?") > -1 {
			url += "&" + key + "=" + value.(string)
		} else {
			url += "?" + key + "=" + value.(string)
		}
	}
	return url, nil
}

func (a *API) appendToken(url string) string {

	if strings.Contains(url, "COMPONENT_ACCESS_TOKEN") {
		return strings.Replace(url, "COMPONENT_ACCESS_TOKEN", a.GetComponentAccess(), 1)
	}
	if strings.Contains(url, "AUTHORIZE_ACCESS_TOKEN") {
		return strings.Replace(url, "AUTHORIZE_ACCESS_TOKEN", a.GetAuthorizeAccess(), 1)
	}
	if strings.Contains(url, "ACCESS_TOKEN") {
		return strings.Replace(url, "ACCESS_TOKEN", a.GetAccessToken(), -1)
	}
	return url
}

func (a *API) httpPost(url, args string) ([]byte, error) {
	realUrl := a.appendToken(url)
	return HttpPost(realUrl, args)
}

func (a *API) httpGet(url string) ([]byte, error) {
	realUrl := a.appendToken(url)

	return HttpGet(realUrl)
}

func (a *API) checkResponse(data map[string]interface{}) (map[string]interface{}, error) {
	errCode, err := data["errcode"].(float64)
	errMsg, _ := data["errmsg"].(string)
	if errCode == 0 || (err == false) {
		return data, nil
	}
	return data, errors.New(errMsg)
}

func (a *API) tokenExpired(errCode float64) bool {
	if errCode == 40014 || errCode == 42001 || errCode == 42007 || errCode == 42009 {
		return true
	}
	return false
}

func (a *API) refreshToken(url string) {
	if strings.Contains(url, "COMPONENT_ACCESS_TOKEN") {
		a.RefreshComponentAccess()
	}
	if strings.Contains(url, "AUTHORIZE_ACCESS_TOKEN") {
		a.RefreshAuthorizeAccess()
	}
	if strings.Contains(url, "ACCESS_TOKEN") {
		a.RefreshAccessToken()
	}

}
