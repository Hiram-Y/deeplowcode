package controllers

import (
	"DeepWorkload/models"
	"fmt"
	"github.com/astaxie/beego"
	"strconv"
)

type TaskController struct {
	beego.Controller
}

func (this *TaskController) CreateOneTask() {
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	base_info := this.GetString("base_info")
	form_info := this.GetString("form_info")
	workflow_info := this.GetString("workflow_info")
	form_codes := models.AddFormsByJson(form_info, corpid, userid)
	workflow_code := ""
	if workflow_info != "" {
		workflow_code = models.WorkflowJsonToDB(workflow_info)
	}
	_, err := models.BaseJsonToDB(base_info, corpid, userid, workflow_code, form_codes)
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

func (this *TaskController) GetAllAuthorizeTaskInfo() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	task := &models.Task{
		CreateUserId: userid,
		CorpId:       corpid,
	}
	err, task_infos := task.AllAuthorizeTaskInfoByUserID()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    task_infos,
		}
	}
	this.ServeJSON()
}

func (this *TaskController) GetAllAuthorizeTaskInfoByTypeID() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	type_id, _ := this.GetInt("type_id")
	page_index, _ := this.GetInt("page_index")
	page_size, _ := this.GetInt("page_size")
	task := &models.Task{
		CreateUserId: userid,
		CorpId:       corpid,
		TypeId:       type_id,
	}
	err, task_infos, count := task.GetAllAuthorizeTaskInfoByTypeID(page_size, page_index)
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
				"tasks": task_infos,
				"count": count,
			},
		}
	}
	this.ServeJSON()
}

func (this *TaskController) GetAllTaskInfoByTypeID() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	type_id, _ := this.GetInt("type_id")
	page_index, _ := this.GetInt("page_index")
	page_size, _ := this.GetInt("page_size")
	task := &models.Task{
		CreateUserId: userid,
		CorpId:       corpid,
		TypeId:       type_id,
	}
	err, task_infos, count := task.GetAllTaskByTypeID(page_size, page_index)
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
				"tasks": task_infos,
				"count": count,
			},
		}
	}
	this.ServeJSON()
}

func (this *TaskController) GetAllClosedTaskInfoByUserId() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	task := &models.Task{
		CreateUserId: userid,
		CorpId:       corpid,
	}
	err, task_infos := task.AllSelfCreateTaskInfoClosed()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    task_infos,
		}
	}
	this.ServeJSON()
}

func (this TaskController) AllTaskInfoByCorpId() {
	corpid := this.GetString("corpid")
	err, info := models.GetAllTaskByCorpId(corpid)
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

func (this *TaskController) GetAllTaskInfoByUserId() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	task := &models.Task{
		CreateUserId: userid,
		CorpId:       corpid,
	}
	err, task_infos := task.AllSelfCreateTaskInfoOnGoing()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    task_infos,
		}
	}
	this.ServeJSON()
}

func (this *TaskController) SearchSelfCreate() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	title := this.GetString("title")
	task := &models.Task{
		CreateUserId: userid,
		CorpId:       corpid,
	}
	err, info := task.AllSelfCreateTaskInfoSearch(title)
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

func (this *TaskController) SearchAuthorize() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	title := this.GetString("title")
	task := &models.Task{
		CreateUserId: userid,
		CorpId:       corpid,
	}
	err, info := task.AllAuthorizeTaskInfoSearch(title)
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

func (this *TaskController) SwitchTaskState() {
	task_code := this.GetString("task_code")
	state, _ := strconv.Atoi(this.GetString("state"))
	task := &models.Task{
		Code:  task_code,
		State: state,
	}
	err := task.UpdateTaskState()
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

func (this *TaskController) DelTask() {
	task_code := this.GetString("task_code")
	task := &models.Task{
		Code: task_code,
	}
	err := task.DelTask()
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

func (this *TaskController) UpdateScope() {
	task_code := this.GetString("task_code")
	scope := this.GetString("scope")
	task := &models.Task{
		Code:  task_code,
		Scope: scope,
	}
	err := task.UpdateTaskInfo()
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

func (this *TaskController) GetTaskFullInfo() {
	task_code := this.GetString("task_code")
	task := &models.Task{
		Code: task_code,
	}
	err, b, f, w := task.GetTaskFullInfoDetailToJson()
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
				"base_info":     b,
				"form_info":     f,
				"workflow_info": w,
			},
		}
	}
	this.ServeJSON()
}

func (this *TaskController) GetReportTaskListByUserId() {
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	task := &models.Task{
		CreateUserId: userid,
		CorpId:       corpid,
	}
	err, info := task.GetToReportTaskByUserid()
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

func (this *TaskController) GetReportedTaskListByUserId() {
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	task := &models.Task{
		CreateUserId: userid,
		CorpId:       corpid,
	}
	err, info := task.GetReportedTaskByUserID()
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

func (this *TaskController) GetReportedTaskInfoByTypeID() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	type_id, _ := this.GetInt("type_id")
	page_index, _ := this.GetInt("page_index")
	page_size, _ := this.GetInt("page_size")
	task := &models.Task{
		CreateUserId: userid,
		CorpId:       corpid,
		TypeId:       type_id,
	}
	err, task_infos, count := task.GetReportedTaskByTypeId(page_size, page_index)
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
				"tasks": task_infos,
				"count": count,
			},
		}
	}
	this.ServeJSON()
}

func (this *TaskController) GetSelfTaskFormList() {
	task_code := this.GetString("task_code")
	userid := this.GetString("userid")
	task := &models.Task{Code: task_code, CreateUserId: userid}
	err, info := task.GetFormListAndCount()
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

func (this *TaskController) GetFormList() {
	task_code := this.GetString("task_code")
	task := &models.Task{Code: task_code}
	err, info := task.GetFormListAndCount()
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

func (this *TaskController) GetAuditFormList() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	task_code := this.GetString("task_code")
	task := &models.Task{Code: task_code, CorpId: corpid, CreateUserId: userid}
	err, info := task.GetFormListAndCountByAuditUserId()
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

func (this *TaskController) CheckTaskAuditType() {
	task_code := this.GetString("task_code")
	err, info := models.CheckWorkflowTypeByTaskCode(task_code)
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

func (this *TaskController) GetFormDetail() {
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

func (this *TaskController) CheckAuditMissionByUserId() {
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	info := models.CheckUserIsAudit(userid, corpid)
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data":    info,
	}
	this.ServeJSON()
}

func (this *TaskController) GetAllAuditTaskByUserId() {
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	page_index, _ := this.GetInt("page_index")
	page_size, _ := this.GetInt("page_size")
	err, task_infos, count := models.GetAllTaskByAuditUserId(userid, corpid, page_size, page_index)
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
				"tasks": task_infos,
				"count": count,
			},
		}
	}
	this.ServeJSON()
}

func (this *TaskController) DownLoadExcelTemplate() {
	form_code := this.GetString("form_code")
	err, path := models.CreateFormUploadExcel(form_code)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    path,
		}
	}
	this.ServeJSON()
}

func (this *TaskController) AddMainDatasByExcel() {
	doc_code := this.GetString("doc_code")
	err, docInfo := models.GetDocinfo(doc_code)
	fmt.Println(err)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": "没有该文件的Code",
		}
	} else {
		userid := this.GetString("userid")
		corpid := this.GetString("corpid")
		task_code := this.GetString("task_code")
		form_code := this.GetString("form_code")
		u_infos, m_infos := models.GetMainDatasByExcel(docInfo.Path, corpid, task_code, form_code)
		err := models.InsertMainDatasByExcel(m_infos)
		if err != nil {
			this.Data["json"] = map[string]interface{}{
				"errcode": ErrDataBase,
				"message": err.Error(),
			}
		} else {
			this.Data["json"] = map[string]interface{}{
				"errcode": OK,
				"message": "OK",
				"data": map[string]interface{}{
					"success": len(m_infos),
					"f_data":  u_infos,
				},
			}
			go func() {
				for _, each := range m_infos {
					models.AddOneLog(userid, each.Code, 11, "")
				}
			}()
		}
	}
	this.ServeJSON()
}
