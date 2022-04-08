package AbstractAPI

import (
	"DeepWorkload/conf"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const base = "https://api.weixin.qq.com"

func MakeUrl(queryArgs string) string {
	if strings.Index(queryArgs, "/") == 0 {
		return base + queryArgs
	}
	return base + "/" + queryArgs
}

func HttpGet(url string) ([]byte, error) {
	if conf.DEBUG {
		fmt.Println("httpGet: " + url)
	}
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func HttpPost(url, postData string) ([]byte, error) {
	if conf.DEBUG {
		fmt.Println("httpPost: " + url)
		fmt.Println("postData: " + postData)
	}

	postData = strings.Replace(postData, "\\u003c", "<", -1)
	postData = strings.Replace(postData, "\\u003e", ">", -1)
	postData = strings.Replace(postData, "\\u0026", "&", -1)

	res, err := http.Post(url, "application/json", strings.NewReader(postData))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
