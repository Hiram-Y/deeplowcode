package controllers

import (
	"DeepWorkload/models"
	"github.com/astaxie/beego"
	"strconv"
	"strings"
)

type TypeController struct {
	beego.Controller
}

func (this *TypeController) GetAllTypeByTypeDesc() {
	typedesc := this.GetString("typedesc")
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	t_info := &models.TypeInfo{
		TypeDesc:     typedesc,
		CorpId:       corpid,
		CreateUserId: userid,
	}
	err, infos := t_info.GetAllTypeInfoByTypeDesc()
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

func (this *TypeController) DelTypeById() {
	corpid := this.GetString("corpid")
	id, _ := strconv.Atoi(this.GetString("id"))
	userid := this.GetString("userid")
	t_info := &models.TypeInfo{
		Id:           id,
		CorpId:       corpid,
		CreateUserId: userid,
	}
	err := t_info.DelOneTypeInfo()
	if err != nil {
		code := ErrDataBase
		if strings.Contains(err.Error(), "the type is holding") {
			code = TypeIsHolding
		}
		this.Data["json"] = map[string]interface{}{
			"errcode": code,
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

func (this *TypeController) AddOneType() {
	corpid := this.GetString("corpid")
	type_desc := this.GetString("type_desc")
	content := this.GetString("content")
	userid := this.GetString("userid")
	t_info := &models.TypeInfo{
		CorpId:       corpid,
		TypeDesc:     type_desc,
		Content:      content,
		CreateUserId: userid,
	}
	err, id := t_info.AddOneType()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    id,
		}
	}
	this.ServeJSON()

}

func (this *TypeController) EditOneTypeName() {
	corpid := this.GetString("corpid")
	id, _ := strconv.Atoi(this.GetString("id"))
	content := this.GetString("content")
	t_info := &models.TypeInfo{
		Id:      id,
		CorpId:  corpid,
		Content: content,
	}
	err := t_info.EditTypeInfo()
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
