package controllers

import (
	"DeepWorkload/conf"
	"DeepWorkload/lib/redisgo"
	"DeepWorkload/models"
	"DeepWorkload/utils"
	"DeepWorkload/utils/AbstractAPI"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/qiniu/api.v7/v7/auth"
	"github.com/qiniu/api.v7/v7/sms"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"strings"
	"time"
)

type LoginController struct {
	beego.Controller
}

func createRandomGUID(start, end float64) map[string]int64 {
	rand.Seed(time.Now().UnixNano())
	randf := rand.Float64()
	if randf == 0 {
		return createRandomGUID(start, end)
	}
	uid := uuid.NewV4().String()
	con, _ := redisgo.New(redisgo.Options{})
	randI := randf*(end-start) + start
	con.Set(uid, int64(randI), 60*2)
	return map[string]int64{uid: int64(randI)}
}

func (this *LoginController) GetGUID() {
	start, _ := this.GetFloat("start")
	end, _ := this.GetFloat("end")
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data":    createRandomGUID(start, end),
	}
	this.ServeJSON()
}

func (this *LoginController) CheckYanZM() {
	ssid := this.GetString("ssid")
	randf, _ := this.GetInt("randf")
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data":    checkYanZM(ssid, randf),
	}
	this.ServeJSON()
}

func checkYanZM(uid string, randf int) bool {
	con, _ := redisgo.New(redisgo.Options{})
	f, err := con.GetInt64(uid)
	if err != nil {
		return false
	}
	if float64(randf) >= float64(f)*1.1 || float64(randf) <= float64(f)*0.9 {
		return false
	}
	return true
}

func checkPhoneYanZM(mobile, rands string) bool {
	con, _ := redisgo.New(redisgo.Options{})
	mobile_verify_key := fmt.Sprintf("%s-verify-code", mobile)
	f, err := con.GetString(mobile_verify_key)
	if err != nil {
		return false
	}
	if f != rands {
		return false
	}
	con.Del(mobile_verify_key)
	return true
}

func (this *LoginController) CheckMobile() {
	mobile := this.GetString("mobile")
	isExist := models.SignUpCheck(mobile)
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data":    isExist,
	}
	this.ServeJSON()
}

func (this *LoginController) SignUp() {
	corp_name := this.GetString("corp_name")
	userid := this.GetString("userid")
	user_name := this.GetString("user_name")
	mobile := this.GetString("mobile")
	passwd := this.GetString("passwd")
	verify_code := this.GetString("verify_code")
	if !checkPhoneYanZM(mobile, verify_code) {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrLoginVerifyCode,
			"message": "手机验证码错误",
		}
	} else {
		au := &models.AuthCorpInfo{
			CorpName: corp_name,
		}
		err := au.SignUpAuthCorpInfo(userid, user_name, passwd, mobile)
		if err != nil {
			errcode := ErrDataBase
			if strings.Contains(err.Error(), "手机号") {
				errcode = ErrLoginMobileExit
			}
			this.Data["json"] = map[string]interface{}{
				"errcode": errcode,
				"message": err.Error(),
			}
		} else {
			this.Data["json"] = map[string]interface{}{
				"errcode": OK,
				"message": "OK",
			}
		}
	}
	this.ServeJSON()
}

func (this *LoginController) GetOpenId() {
	code := this.GetString("auth_code")
	wx := AbstractAPI.NewWXAuthAPI(conf.WxAppId, conf.WxSecret, code)
	re_info := wx.GetOpenid()
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"errmsg":  "OK",
		"data":    re_info,
	}

	fmt.Println(re_info)

	this.ServeJSON()
}

func (this *LoginController) Login() {
	mobile := this.GetString("mobile")
	passwd := this.GetString("passwd")
	open_id := this.GetString("openid")
	u_info := &models.UserInfo{
		Mobile: mobile,
		Passwd: passwd,
	}
	state, au := u_info.CheckPasswd(this.Ctx.Request.RemoteAddr)
	if open_id != "" && state == 0 {
		go models.BindUserInfo(au, open_id)
	}
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data": map[string]interface{}{
			"state": state,
			"corps": au,
		},
	}
	this.ServeJSON()
}

func (this *LoginController) CheckAdmin() {
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	ad := &models.Admin{
		CorpId: corpid,
		UserId: userid,
	}
	isExist := ad.CheckAdmin()
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"info": map[string]interface{}{
			"state":  isExist,
			"userid": userid,
			"corpid": corpid,
		},
	}
	this.ServeJSON()
}

func (this *LoginController) SendVerifySMS() {
	mobile := this.GetString("mobile")
	if checkMobileCount(mobile) {
		sendSMSToMobile(mobile)
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrLoginSendMessageCount,
			"message": "一分钟内请求次数过多",
		}
	}
	this.ServeJSON()
}

func (this *LoginController) LoginByMobile() {
	mobile := this.GetString("mobile")
	verify_code := this.GetString("verify_code")
	if !checkPhoneYanZM(mobile, verify_code) {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrLoginVerifyCode,
			"message": "手机验证码错误",
		}
	} else {
		state, au := models.GetLoginUserInfoByMobile(mobile)
		open_id := this.GetString("openid")
		if open_id != "" && state == 0 {
			go models.BindUserInfo(au, open_id)
		}
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data": map[string]interface{}{
				"state": state,
				"corps": au,
			},
		}
	}
	this.ServeJSON()
}

func (this *LoginController) UnbindUser() {
	openid := this.GetString("openid")
	userid := this.GetString("userid")
	corpid := this.GetString("coprid")
	wx_userinfo := &models.UserInfo{
		OpenId: openid,
		UserId: userid,
		CorpId: corpid,
	}
	if wx_userinfo.CheckOpenid() {
		err := wx_userinfo.UnBindUserInfo()
		if err != nil {
			this.Data["json"] = map[string]interface{}{
				"errcode": ErrDataBase,
				"errmsg":  err.Error(),
			}
		} else {
			this.Data["json"] = map[string]interface{}{
				"errcode": OK,
				"errmsg":  "OK",
			}
		}

	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrOpenId,
			"message": "err openid",
		}
	}
	this.ServeJSON()
}

func (this *LoginController) OpenidToUserid() {
	openid := this.GetString("openid")
	state, infos := models.GetLoginUserInfoByOpenId(openid)
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"errmsg":  "OK",
		"data": map[string]interface{}{
			"state": state,
			"corps": infos,
		},
	}
	this.ServeJSON()
}

func (this *LoginController) SetPasswd() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	passwd := this.GetString("passwd")
	u_info := &models.UserInfo{
		UserId: userid,
		CorpId: corpid,
		Passwd: passwd,
	}
	err := u_info.SetPasswd()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func sendSMSToMobile(mobile string) {
	verify_code := utils.GenValidateCode(6)
	mobile_verify_key := fmt.Sprintf("%s-verify-code", mobile)
	con, _ := redisgo.New(redisgo.Options{})
	con.Set(mobile_verify_key, verify_code, 60*5)
	args := sms.MessagesRequest{
		SignatureID: conf.QN_SMS_SignatureID,
		TemplateID:  conf.QN_SMS_TemplateID,
		Mobiles:     []string{mobile},
		Parameters: map[string]interface{}{
			"code": verify_code,
		},
	}

	mac := auth.New(conf.QN_AK, conf.QN_SK)
	manager := sms.NewManager(mac)
	manager.SendMessage(args)
}

func checkMobileCount(mobile string) bool {
	mobile_key := fmt.Sprintf("%s-count", mobile)
	con, _ := redisgo.New(redisgo.Options{})
	count, err := con.GetInt(mobile_key)
	if err != nil {
		con.Set(mobile_key, 1, 60)
		return true
	} else {
		count += 1
		con.Set(mobile_key, count, 60)
		if count > 4 {
			return false
		}
		return true
	}
}

func (this *LoginController) CheckMobileVerify() {
	corpid := this.GetString("corpid")
	mobile := this.GetString("mobile")
	verify_code := this.GetString("verify_code")
	u_info := &models.UserInfo{
		CorpId: corpid,
		Mobile: mobile,
	}
	isExist := u_info.CheckMobileExistInThisCorp()
	if isExist {
		this.Data["json"] = map[string]interface{}{
			"errcode": 80001,
			"message": "手机号已经存在",
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    checkPhoneYanZM(mobile, verify_code),
		}
	}

	this.ServeJSON()
}
