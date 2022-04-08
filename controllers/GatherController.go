package controllers

import (
	"DeepWorkload/models"
	"fmt"
	"github.com/astaxie/beego"
)

type GatherController struct {
	beego.Controller
}

func (this *GatherController) GatherListByTaskCodeAndUserId() {
	corpid := this.GetString("corpid")
	task_code := this.GetString("task_code")
	userid := this.GetString("userid")
	m_info := &models.MainData{
		CorpId:       corpid,
		TaskCode:     task_code,
		CreateUserId: userid,
	}
	gather, form_info := m_info.GatherList()
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data": map[string]interface{}{
			"gather": gather,
			"form":   form_info,
		},
	}
	this.ServeJSON()
}

func (this *GatherController) GatherListByTaskCode() {
	corpid := this.GetString("corpid")
	task_code := this.GetString("task_code")
	m_info := &models.MainData{
		CorpId:   corpid,
		TaskCode: task_code,
	}
	gather, form_info := m_info.GatherList()
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data": map[string]interface{}{
			"gather": gather,
			"form":   form_info,
		},
	}
	this.ServeJSON()
}

func (this *GatherController) ExportExcel() {
	corpid := this.GetString("corpid")
	task_code := this.GetString("task_code")
	m_info := &models.MainData{
		CorpId:   corpid,
		TaskCode: task_code,
	}
	path, file_name, out_name := m_info.Export()
	fmt.Println(out_name, file_name, path)
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data": map[string]interface{}{
			"path": fmt.Sprintf("docs/%s", file_name),
		},
	}
	this.ServeJSON()
}

func (this *GatherController) GatherListByTaskCodeAndAuditUserId() {
	corpid := this.GetString("corpid")
	task_code := this.GetString("task_code")
	userid := this.GetString("userid")
	m_info := &models.MainData{
		CorpId:       corpid,
		TaskCode:     task_code,
		CreateUserId: userid,
	}
	gather, form_info := m_info.GatherListByAuditUserId()
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data": map[string]interface{}{
			"gather": gather,
			"form":   form_info,
		},
	}
	this.ServeJSON()
}

func (this *GatherController) ExportExcelByAuditUserId() {
	corpid := this.GetString("corpid")
	task_code := this.GetString("task_code")
	userid := this.GetString("userid")
	m_info := &models.MainData{
		CorpId:       corpid,
		TaskCode:     task_code,
		CreateUserId: userid,
	}
	path, file_name, out_name := m_info.ExportAuditUserId()
	fmt.Println(out_name, file_name, path)
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data": map[string]interface{}{
			"path": fmt.Sprintf("docs/%s", file_name),
		},
	}
	this.ServeJSON()
}
