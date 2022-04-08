package controllers

import (
	"DeepWorkload/models"
	"github.com/astaxie/beego"
	"strings"
)

type MarketController struct {
	beego.Controller
}

func (this *MarketController) GetAllType() {
	err, info := models.GetAllTypeInfo()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    info,
		}
	}
	this.ServeJSON()
}

func (this *MarketController) AddOneType() {
	parent_id, _ := this.GetInt("parent_id")
	content := this.GetString("content")
	t := &models.MarketType{
		ParentId: parent_id,
		Content:  content,
	}
	err := t.AddTypeInfo()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *MarketController) EditOneType() {
	id, _ := this.GetInt("id")
	content := this.GetString("content")
	t := &models.MarketType{
		Id:      id,
		Content: content,
	}
	err := t.UpdateOne()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *MarketController) DelOneType() {
	id, _ := this.GetInt("id")
	mt := &models.MarketType{
		Id: id,
	}
	err := mt.DelOneType()
	if err != nil {
		err_code := ErrDataBase
		if strings.Contains(err.Error(), "存在子类") {
			err_code = ErrMarketTypeDel
		}
		this.Data["json"] = map[string]interface{}{
			"errcode": err_code,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *MarketController) MarketLogin() {
	userid := this.GetString("userid")
	passwd := this.GetString("passwd")
	ma := &models.MarketAdmin{UserId: userid, Passwd: passwd}
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data":    ma.Login(),
	}
	this.ServeJSON()
}

func (this *MarketController) GetAllMarketTemplate() {
	ac := this.GetString("ac")
	f := false
	if ac != "" {
		f = true
	}
	type_id, _ := this.GetInt("type_id")
	err, info := models.GetAllMarketFormByType(type_id, f)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    info,
		}
	}
	this.ServeJSON()
}

func (this *MarketController) AddOneMarketTemplateForm() {
	userid := this.GetString("userid")
	form_info := this.GetString("form_info")
	code := models.AddMarketTemplateFormsByJson(form_info, userid)
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data":    code,
	}
	this.ServeJSON()
}

func (this *MarketController) ActiveFormCode() {
	form_code := this.GetString("form_code")
	active := this.GetString("active")
	is_active := false
	if active != "" {
		is_active = true
	}
	mt := &models.MarketTemplateFormInfo{Code: form_code, IsActive: is_active}
	err := mt.Active()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *MarketController) DelOneTemplate() {
	form_code := this.GetString("form_code")
	mtf := &models.MarketTemplateFormInfo{Code: form_code}
	err := mtf.DelOne()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *MarketController) GetTemplateFormDetail() {
	form_code := this.GetString("form_code")
	tfi := &models.MarketTemplateFormInfo{
		Code: form_code,
	}
	err, infos := tfi.GetTemplateFormDetailInfo()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    infos,
		}
	}
	this.ServeJSON()
}

func (this *MarketController) CountFunction() {
	form_code := this.GetString("form_code")
	rule_file_info := this.GetString("rule_file_info")
	rf_info := &models.MarketTemplateFormField{FormCode: form_code}
	err, info := rf_info.CountFormulaMethod(rule_file_info)

	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrRuleFormula,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    info,
		}
	}
	this.ServeJSON()
}
