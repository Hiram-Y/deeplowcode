package AbstractAPI

import (
	"DeepWorkload/lib/redisgo"
	"fmt"
)

var WX_API_TYPE = map[string][]string{
	"GET_ACCESS_TOKEN": {"/cgi-bin/token", "GET"},
	"GET_OPEN_ID":      {"sns/oauth2/access_token", "GET"},
	//"USER_INFO":{"/sns/userinfo?access_token=ACCESS_TOKEN","GET"},
	"SEND_MESSAGE":        {"/cgi-bin/message/template/send?access_token=ACCESS_TOKEN", "POST"},
	"SEND_SIMPLE_MESSAGE": {"/cgi-bin/message/mass/send?access_token=ACCESS_TOKEN", "POST"},
}

type WXAuthAPI struct {
	API
	AppID       string
	Secret      string
	Code        string
	AccessToken string
}

func (this *WXAuthAPI) GetAccessToken() string {
	if this.AccessToken == "" {
		this.RefreshAccessToken()
	}
	return this.AccessToken
}

func (this *WXAuthAPI) RefreshAccessToken() {
	con, _ := redisgo.New(redisgo.Options{})
	AccessTokenKey := this.AppID + "AuthAccessTokenKey"
	AccessToken, err := con.GetString(AccessTokenKey)
	if err == nil {
		this.AccessToken = AccessToken
	} else {
		request_body := map[string]interface{}{
			"grant_type": "client_credential",
			"appid":      this.AppID,
			"secret":     this.Secret,
			//"code" : this.Code,
		}
		response, _ := this.HttpCall(
			WX_API_TYPE["GET_ACCESS_TOKEN"],
			request_body)
		this.AccessToken = response["access_token"].(string)
		expires_in_seconds := int64(response["expires_in"].(float64))
		_ = con.Set(AccessTokenKey, this.AccessToken, expires_in_seconds)
	}
}

func NewWXAuthAPI(appid, secrect, code string) *WXAuthAPI {
	s := &WXAuthAPI{
		AppID:  appid,
		Secret: secrect,
		Code:   code,
	}
	s.API.AccessAPI = s
	return s
}

func (this *WXAuthAPI) GetOpenid() map[string]interface{} {
	request_body := map[string]interface{}{
		"grant_type": "authorization_code",
		"appid":      this.AppID,
		"secret":     this.Secret,
		"code":       this.Code,
	}
	response, _ := this.HttpCall(
		WX_API_TYPE["GET_OPEN_ID"],
		request_body)
	return response
}

func (this *WXAuthAPI) SendMessage(to_user, template_id, url string, content map[string]interface{}) map[string]interface{} {
	message_info := map[string]interface{}{
		"touser":      to_user,
		"template_id": template_id,
		"data":        content,
	}
	if url != "" {
		message_info["url"] = url
	}
	response, _ := this.HttpCall(
		WX_API_TYPE["SEND_MESSAGE"],
		message_info,
	)
	fmt.Println(response)
	return response
}

func (this *WXAuthAPI) SendSimpleMessage(openids []string, content string) map[string]interface{} {
	request_body := map[string]interface{}{
		"to_user": openids,
		"msgtype": "text",
		"text":    map[string]string{"content": content},
	}
	response, _ := this.HttpCall(
		WX_API_TYPE["SEND_SIMPLE_MESSAGE"],
		request_body)
	return response

}
