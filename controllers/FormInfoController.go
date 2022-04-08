package controllers

import (
	"DeepWorkload/conf"
	"DeepWorkload/models"
	"encoding/json"
	"github.com/astaxie/beego"
)

type FormInfoController struct {
	beego.Controller
}

func (this *FormInfoController) GetOpenTemplateFormInfo() {
	tfi := &models.TemplateFormInfo{
		CorpId:       conf.OpenCorpId,
		CreateUserid: conf.OpenUserId,
	}
	err, infos := tfi.GetAllTemplateFormInfoByUserid()
	delete(infos, 1)
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

func (this *FormInfoController) GetAllOpenFormInfoByTypeId() {
	corpid := conf.OpenCorpId
	userid := conf.OpenUserId
	type_id, _ := this.GetInt("type_id")
	tfs := &models.TemplateFormInfo{
		CorpId:       corpid,
		CreateUserid: userid,
		TypeId:       type_id,
	}
	err, info := tfs.GetAllTemplateFormInfoByUseridAndTypeId()
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

func (this *FormInfoController) OpenTemplateInstall() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	form_code := this.GetString("form_code")
	err := models.CopyToUserId(form_code, userid, corpid)
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

func (this *FormInfoController) GetTemplateFormInfo() {
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	tfi := &models.TemplateFormInfo{
		CorpId:       corpid,
		CreateUserid: userid,
	}
	err, infos := tfi.GetAllTemplateFormInfoByUserid()
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

func (this *FormInfoController) GetFormDetail() {
	form_code := this.GetString("form_code")
	tfi := &models.FormInfo{
		Code: form_code,
	}
	err, infos := tfi.GetFormDetailInfo()
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

func (this *FormInfoController) GetTemplateFormDetail() {
	form_code := this.GetString("form_code")
	tfi := &models.TemplateFormInfo{
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

func (this *FormInfoController) DelTemplateForm() {
	form_code := this.GetString("form_code")
	tfi := &models.TemplateFormInfo{
		Code: form_code,
	}
	err := tfi.DelFormInfo()
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

func (this *FormInfoController) AddOneTemplateForm() {
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	form_info := this.GetString("form_info")
	code := models.AddTemplateFormsByJson(form_info, corpid, userid)
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data":    code,
	}
	this.ServeJSON()
}

func (this *FormInfoController) EditName() {
	code := this.GetString("code")
	title := this.GetString("title")
	f := &models.TemplateFormInfo{
		Code:  code,
		Title: title,
	}
	err := f.EditTitle()
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

func (this *FormInfoController) EditType() {
	code := this.GetString("code")
	type_id, _ := this.GetInt("type_id")
	f := &models.TemplateFormInfo{
		Code:   code,
		TypeId: type_id,
	}
	err := f.EditType()
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

func (this *FormInfoController) UpdateTemplateCooperation() {
	code := this.GetString("code")
	cooper := this.GetString("cooper")
	tf := &models.TemplateFormInfo{
		Code:        code,
		Cooperation: cooper,
	}
	err := tf.UpdateTemplateCooperation()
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

func (this *FormInfoController) GetAllFormInfoByTypeId() {
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	type_id, _ := this.GetInt("type_id")
	tfs := &models.TemplateFormInfo{
		CorpId:       corpid,
		CreateUserid: userid,
		TypeId:       type_id,
	}
	err, info := tfs.GetAllTemplateFormInfoByUseridAndTypeId()
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

func (this *FormInfoController) GetAuthorizeTemplateFormInfo() {
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	tfi := &models.TemplateFormInfo{
		CorpId:       corpid,
		CreateUserid: userid,
	}
	err, infos := tfi.AllTemplateFormInfoByAuthorize()
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

func (this *FormInfoController) GetAllAuthorizeFormInfoByTypeId() {
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	type_id, _ := this.GetInt("type_id")
	tfs := &models.TemplateFormInfo{
		CorpId:       corpid,
		CreateUserid: userid,
		TypeId:       type_id,
	}
	err, info := tfs.GetAllAuthorizeTemplateFormInfoByTypeId()
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

func (this *FormInfoController) CheckEnableFormulaMethod() {
	formula_methods := this.GetString("formula_methods")
	fms := []string{}
	json.Unmarshal([]byte(formula_methods), &fms)
	err := models.CheckFormulaMethod(fms)
	if err != nil {

		this.Data["json"] = map[string]interface{}{
			"errcode": ErrRuleFormula,
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

func (this *FormInfoController) TemplateFormTree() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	tfs := &models.TemplateFormInfo{
		CorpId:       corpid,
		CreateUserid: userid,
	}
	err, info := tfs.GetAllTempFormTree()
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
