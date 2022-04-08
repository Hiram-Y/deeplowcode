package controllers

import (
	"DeepWorkload/models"
	"github.com/astaxie/beego"
)

type PubTaskController struct {
	beego.Controller
}

func (this *PubTaskController) GetPublicInfoByCode() {
	task_code := this.GetString("task_code")
	pb_info := &models.TaskPublic{
		Code: task_code,
	}
	err, info := pb_info.GetTaskPubInfo()
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

func (this *PubTaskController) UpdateTaskPublicInfo() {
	task_code := this.GetString("task_code")
	state, _ := this.GetInt("state")
	start_date, _ := this.GetInt("start_date")
	end_date, _ := this.GetInt("end_date")
	pub_scope := this.GetString("pub_scope")
	remark := this.GetString("remark")
	corpid := this.GetString("corpid")
	pb_info := &models.TaskPublic{
		Code:      task_code,
		StartDate: int64(start_date),
		EndDate:   int64(end_date),
		State:     state,
		PubScope:  pub_scope,
		Remark:    remark,
	}
	err := pb_info.UpdatePubInfo(corpid)
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

func (this *PubTaskController) GetAllPubTask() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	page_index, _ := this.GetInt("page_index")
	page_size, _ := this.GetInt("page_size")
	err, info, count := models.GetAllPubTask(corpid, userid, page_size, page_index)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data": map[string]interface{}{
				"tasks": info,
				"count": count,
			},
		}
	}
	this.ServeJSON()
}

func (this *PubTaskController) GetAllPubMainData() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	task_code := this.GetString("task_code")
	form_code := this.GetString("form_code")
	m_info := &models.MainData{
		CreateUserId: userid,
		CorpId:       corpid,
		TaskCode:     task_code,
		FormCode:     form_code,
	}
	err, info, field_infos := m_info.AllPubMainDataByTaskCodeAndFormCode()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode":     OK,
			"message":     "OK",
			"data":        info,
			"filed_infos": field_infos,
		}
	}
	this.ServeJSON()
}
