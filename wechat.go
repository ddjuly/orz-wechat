package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"io/ioutil"
	"net/http"
	"net/url"
	"shuang/controllers/cache"
	"strings"
	"time"
	"unsafe"
)

func PostJson(url string, params map[string]interface{}) map[string]interface{} {
	bytesData, _ := json.Marshal(params)
	println(string(bytesData))

	reader := bytes.NewReader(bytesData)

	request, _ := http.NewRequest("POST", url, reader)
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := http.Client{}
	resp, _ := client.Do(request)

	respBytes, _ := ioutil.ReadAll(resp.Body)

	body := *(*string)(unsafe.Pointer(&respBytes))

	var result map[string]interface{}
	json.Unmarshal([]byte(body), &result)

	return result
}

func parseJson(resp *http.Response) map[string]interface{} {
	if resp.StatusCode != 200 {
		return map[string]interface{}{}
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(body))

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	return result
}

// 从微信拿access_token
func RefreshAccessToken() string {
	wechatConfig, _ := beego.AppConfig.GetSection("wechat")

	// 构造url和参数
	Url, _ := url.Parse("https://api.weixin.qq.com/cgi-bin/token")
	params := url.Values{}
	params.Set("grant_type", "client_credential")
	params.Set("appid", wechatConfig["appid"])
	params.Set("secret", wechatConfig["secret"])
	Url.RawQuery = params.Encode()

	resp, _ := http.Get(Url.String())
	result := parseJson(resp)

	if result["access_token"] != nil {
		return result["access_token"].(string)
	}

	return ""
}

// 从缓存获取access_token
func GetAccessToken() string {
	wechatConfig, _ := beego.AppConfig.GetSection("wechat")

	var access_token string

	cache_token := cache.Redis.Get("[easywechat.common.access_token."+ wechatConfig["appid"] +"][1]")
	if cache_token == nil {
		cacheToken := RefreshAccessToken()
		access_token = fmt.Sprintf("s:%s:\"%s\";", len(cacheToken), access_token)
		cache.Redis.Put("[easywechat.common.access_token."+ wechatConfig["appid"] +"][1]", access_token, 110*time.Minute)
	} else {
		access_token = cache_token.(string)
		access_token = access_token[strings.IndexAny(access_token, "\"")+1:len(access_token)-2]
	}

	return access_token
}

// 创建永久微信二维码
func CreateQrcode(id string) map[string]interface{} {
	url := "https://api.weixin.qq.com/cgi-bin/qrcode/create"

	// {"action_name": "QR_LIMIT_SCENE", "action_info": {"scene": {"scene_id": 123}}}

	scene_id := make(map[string]interface{})
	scene_id["scene_id"] = id

	sence := make(map[string]interface{})
	sence["scene"] = scene_id

	params := make(map[string]interface{})
	params["action_name"] = "QR_LIMIT_SCENE"
	params["action_info"] = sence

	result := PostJson(url, params)

	return result
}

// 创建临时微信二维码
func CreateTempQrcode(id string, expire int) map[string]interface{} {
	url := "https://api.weixin.qq.com/cgi-bin/qrcode/create"

	// {"expire_seconds": 604800, "action_name": "QR_SCENE", "action_info": {"scene": {"scene_id": 123}}}

	scene_id := make(map[string]interface{})
	scene_id["scene_id"] = id

	sence := make(map[string]interface{})
	sence["scene"] = scene_id

	params := make(map[string]interface{})
	params["expire_seconds"] = expire
	params["action_name"] = "QR_SCENE"
	params["action_info"] = sence

	result := PostJson(url, params)

	return result
}
