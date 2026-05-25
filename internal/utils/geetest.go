package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"sync"

	"net/http"
	"net/url"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/config"
)

var ENDPOINT = sync.OnceValue(func() string {
	return config.C.Captcha.ApiServer + "/validate?captcha_id=" + config.C.Captcha.CaptchaId
})

func ValidateGeetest(lotNumber, captchaOutput, passToken, genTime string) (bool, *string) {

	sign_token := hmac_encode(config.C.Captcha.CaptchaKey, lotNumber)

	form_data := make(url.Values)
	form_data["lot_number"] = []string{lotNumber}
	form_data["captcha_output"] = []string{captchaOutput}
	form_data["pass_token"] = []string{passToken}
	form_data["gen_time"] = []string{genTime}
	form_data["sign_token"] = []string{sign_token}

	cli := http.Client{Timeout: time.Second * 5}
	resp, err := cli.PostForm(ENDPOINT(), form_data)
	if err != nil || resp.StatusCode != 200 {
		slog.Error("请求极验验证接口失败", "error", err, "status_code", resp.StatusCode)
		return true, nil
	}

	res_json, _ := io.ReadAll(resp.Body)
	var res_map map[string]any

	if err = json.Unmarshal(res_json, &res_map); err != nil {
		slog.Error("Json数据解析错误", "error", err)
		return true, nil
	}

	result := res_map["result"]
	if result == "success" {
		return true, nil
	} else {
		reason := res_map["reason"].(string)
		return false, new(reason)
	}
}

func hmac_encode(key string, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
