package controllers

import (
	"DeepWorkload/conf"
	"DeepWorkload/models"
	"DeepWorkload/utils/AbstractAPI"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"strings"
	"time"
)

type MainDataController struct {
	beego.Controller
}

func (this *MainDataController) CountFunction() {
	form_code := this.GetString("form_code")
	rule_file_info := this.GetString("rule_file_info")
	rf_info := &models.FormField{FormCode: form_code}
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

func (this *MainDataController) CountTempFunction() {
	form_code := this.GetString("form_code")
	rule_file_info := this.GetString("rule_file_info")
	rf_info := &models.TemplateFormField{FormCode: form_code}
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

func (this *MainDataController) AddOneMainData() {
	corpid := this.GetString("corpid")
	form_field_info := this.GetString("form_field_info")
	form_code := this.GetString("form_code")
	userid := this.GetString("userid")
	task_code := this.GetString("task_code")
	main_data := &models.MainData{
		CorpId:        corpid,
		FormFieldInfo: form_field_info,
		FormCode:      form_code,
		CreateUserId:  userid,
		TaskCode:      task_code,
	}
	err, code := main_data.InsertOne()

	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		go models.AddOneLog(userid, code, 1, "")
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    code,
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) AdminAddOneMainData() {
	admin_userid := this.GetString("admin_userid")
	corpid := this.GetString("corpid")
	form_field_info := this.GetString("form_field_info")
	form_code := this.GetString("form_code")
	userid := this.GetString("userid")
	task_code := this.GetString("task_code")
	main_data := &models.MainData{
		CorpId:        corpid,
		FormFieldInfo: form_field_info,
		FormCode:      form_code,
		CreateUserId:  userid,
		TaskCode:      task_code,
		State:         3,
	}
	err, code := main_data.InsertOne()

	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		go models.AddOneLog(admin_userid, code, 1, "")
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    code,
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) EditOneMainDataDetail() {
	main_code := this.GetString("main_code")
	form_field_info := this.GetString("form_field_info")

	main_data := &models.MainData{
		Code:          main_code,
		FormFieldInfo: form_field_info,
	}
	err := main_data.EditMainDataFieldInfo()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		userid := this.GetString("userid")
		go models.AddOneLog(userid, main_code, 2, "")
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) GetAllMainDataByFormCodeAndTaskCode() {
	task_code := this.GetString("task_code")
	userid := this.GetString("userid")
	form_code := this.GetString("form_code")
	main_data := &models.MainData{
		TaskCode:     task_code,
		CreateUserId: userid,
		FormCode:     form_code,
	}
	err, info, field_infos := main_data.AllMainDataByTaskCodeAndFormCode()
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

func (this *MainDataController) GetAllMainDataByFormCodeAndTaskCodeAndAdu() {
	task_code := this.GetString("task_code")
	userid := this.GetString("userid")
	form_code := this.GetString("form_code")
	corpid := this.GetString("corpid")

	main_data := &models.MainData{
		TaskCode: task_code,
		FormCode: form_code,
	}
	err, info, field_infos := main_data.AllAuditMainDataByTaskCodeAndFormCode(userid, corpid)
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

func (this *MainDataController) GetAllMainDataByTaskCode() {
	task_code := this.GetString("task_code")
	userid := this.GetString("userid")
	main_data := &models.MainData{
		TaskCode:     task_code,
		CreateUserId: userid,
	}
	err, infos := main_data.AllMainDataByTaskCodeAndUserId()
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

func (this *MainDataController) DelOneMainData() {
	code := this.GetString("code")
	not_check := this.GetString("not_check")
	check := true
	if not_check != "" {
		check = false
	}
	main_data := &models.MainData{
		Code: code,
	}
	err := main_data.DelMainDataByTaskCode(check)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": getErrCode(err),
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

func getErrCode(err error) int {
	if err == nil {
		return 0
	}
	errcode := ErrDataBase
	if strings.Contains(err.Error(), "不能") {
		errcode = ErrMainDataCommit
	}
	if strings.Contains(err.Error(), "已经略去") {
		errcode = ErrMainDataBatchCommit
	}
	return errcode
}

func (this *MainDataController) DelMainDatas() {
	main_codes := this.GetString("main_codes")
	not_check := this.GetString("not_check")
	check := true
	if not_check != "" {
		check = false
	}
	codes := []string{}
	json.Unmarshal([]byte(main_codes), &codes)
	err := models.BatchDelMainData(codes, check)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": getErrCode(err),
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

func (this *MainDataController) GetOneMainDataDetail() {
	main_code := this.GetString("main_code")
	main_data := &models.MainData{
		Code: main_code,
	}
	err, info := main_data.MainDataDetailInfoByCode()
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

func (this *MainDataController) CommitMainDataWithOutAudit() {
	main_code := this.GetString("main_code")
	m_info := &models.MainData{Code: main_code}
	err := m_info.CommitOneMainDataWithOutWorkflow()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": getErrCode(err),
			"errmsg":  err.Error(),
		}
	} else {
		userid := this.GetString("userid")
		go models.AddOneLog(userid, main_code, 8, "")
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) BatchCommitMainDataWithOutAudit() {
	main_codes := this.GetString("main_codes")
	codes := []string{}
	json.Unmarshal([]byte(main_codes), &codes)
	if len(codes) == 0 {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrQurey,
			"errmsg":  fmt.Sprintf("main_codes is %s", main_codes),
		}
	} else {
		err, m_codes := models.BatchCommitMainDataWithOutWorkflow(codes)
		if err != nil && getErrCode(err) == ErrDataBase {
			this.Data["json"] = map[string]interface{}{
				"errcode": getErrCode(err),
				"errmsg":  err.Error(),
			}
		} else {
			this.Data["json"] = map[string]interface{}{
				"errcode": OK,
				"message": "OK",
			}
			if err != nil && getErrCode(err) == ErrMainDataBatchCommit {
				this.Data["json"].(map[string]interface{})["data"] = ErrMainDataBatchCommit
				this.Data["json"].(map[string]interface{})["message"] = err.Error()
			}
			userid := this.GetString("userid")
			for _, each := range m_codes {
				go models.AddOneLog(userid, each, 8, "")
			}
		}
	}

	this.ServeJSON()
}

func (this *MainDataController) CommitMainDataWithNodeCode() {
	main_code := this.GetString("main_code")
	depart_code := this.GetString("depart_code")
	node_code := this.GetString("node_code")
	m_info := &models.MainData{Code: main_code}
	err := m_info.CommitOneMainDataWithNodeCode(depart_code, node_code)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": getErrCode(err),
			"errmsg":  err.Error(),
		}
	} else {
		userid := this.GetString("userid")
		go models.AddOneLog(userid, main_code, 8, "")
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) BatchCommitMainDataWithNodeCode() {
	main_codes := this.GetString("main_codes")
	depart_code := this.GetString("depart_code")
	node_code := this.GetString("node_code")
	codes := []string{}
	json.Unmarshal([]byte(main_codes), &codes)
	if len(codes) == 0 {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrQurey,
			"errmsg":  fmt.Sprintf("main_codes is %s", main_codes),
		}
	} else {
		err, m_codes := models.BatchCommitMainDataWithNodeCode(depart_code, node_code, codes)
		if err != nil && getErrCode(err) == ErrDataBase {
			this.Data["json"] = map[string]interface{}{
				"errcode": getErrCode(err),
				"errmsg":  err.Error(),
			}
		} else {
			this.Data["json"] = map[string]interface{}{
				"errcode": OK,
				"message": "OK",
			}
			if err != nil && getErrCode(err) == ErrMainDataBatchCommit {
				this.Data["json"].(map[string]interface{})["data"] = ErrMainDataBatchCommit
				this.Data["json"].(map[string]interface{})["message"] = err.Error()
			}
			userid := this.GetString("userid")
			for _, each := range m_codes {
				go models.AddOneLog(userid, each, 8, "")
			}
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) CommitMainDataWithTask() {
	main_code := this.GetString("main_code")
	m_info := &models.MainData{
		Code: main_code,
	}
	err := m_info.CommitOneMainDataWithTask()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": getErrCode(err),
			"errmsg":  err.Error(),
		}

	} else {
		userid := this.GetString("userid")
		go models.AddOneLog(userid, main_code, 8, "")
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) BatchCommitMainDataWithTask() {
	main_codes := this.GetString("main_codes")
	codes := []string{}
	json.Unmarshal([]byte(main_codes), &codes)
	if len(codes) == 0 {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrQurey,
			"errmsg":  fmt.Sprintf("main_codes is %s", main_codes),
		}
	} else {
		err, m_codes := models.BatchCommitMainDataWithTask(codes)
		if err != nil && getErrCode(err) == ErrDataBase {
			this.Data["json"] = map[string]interface{}{
				"errcode": getErrCode(err),
				"errmsg":  err.Error(),
			}
		} else {
			this.Data["json"] = map[string]interface{}{
				"errcode": OK,
				"message": "OK",
			}
			if err != nil && getErrCode(err) == ErrMainDataBatchCommit {
				this.Data["json"].(map[string]interface{})["data"] = ErrMainDataBatchCommit
				this.Data["json"].(map[string]interface{})["message"] = err.Error()
			}
			userid := this.GetString("userid")
			for _, each := range m_codes {
				go models.AddOneLog(userid, each, 8, "")
			}
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) CallBackOneMainData() {
	main_code := this.GetString("main_code")
	m_info := &models.MainData{
		Code: main_code,
	}
	err := m_info.CallBackMainData()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": getErrCode(err),
			"errmsg":  err.Error(),
		}

	} else {
		userid := this.GetString("userid")
		go models.AddOneLog(userid, main_code, 9, "")
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) GetAllToAuditedMainData() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	err, infos := models.GetAllAuditedMainDataList(userid, corpid, 1)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": getErrCode(err),
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

func (this *MainDataController) GetAllHaveAuditedMainData() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	err, infos := models.GetAllAuditedMainDataList(userid, corpid, 0)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": getErrCode(err),
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

func (this *MainDataController) ApproveOneMainData() {
	main_code := this.GetString("main_code")
	userid := this.GetString("userid")
	m_info := &models.MainData{
		Code: main_code,
	}
	err := m_info.ApproveMainData(userid)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": getErrCode(err),
			"errmsg":  err.Error(),
		}

	} else {
		go models.AddOneLog(userid, main_code, 6, "")
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) ApproveMainDatas() {
	main_codes := this.GetString("main_codes")
	userid := this.GetString("userid")
	codes := []string{}
	json.Unmarshal([]byte(main_codes), &codes)
	if len(codes) == 0 {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrQurey,
			"errmsg":  fmt.Sprintf("main_codes is %s", main_codes),
		}
	} else {
		err := models.ApproveMainDatas(userid, codes)
		if err != nil {
			this.Data["json"] = map[string]interface{}{
				"errcode": getErrCode(err),
				"errmsg":  err.Error(),
			}

		} else {
			go func() {
				for _, each := range codes {
					models.AddOneLog(userid, each, 6, "")
				}
			}()
			this.Data["json"] = map[string]interface{}{
				"errcode": OK,
				"message": "OK",
			}
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) DenyOneMainData() {
	main_code := this.GetString("main_code")
	m_info := &models.MainData{
		Code: main_code,
	}
	err := m_info.DenyOneMainData()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": getErrCode(err),
			"errmsg":  err.Error(),
		}
	} else {
		userid := this.GetString("userid")
		content := this.GetString("content")
		go models.AddOneLog(userid, main_code, 7, content)
		go sendDenyMessage([]string{main_code}, content)
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) DenyMainDatas() {
	main_codes := this.GetString("main_codes")
	codes := []string{}
	json.Unmarshal([]byte(main_codes), &codes)
	if len(codes) == 0 {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrQurey,
			"errmsg":  fmt.Sprintf("main_codes is %s", main_codes),
		}
	} else {
		err := models.DenyMainDataByCodes(codes)
		if err != nil {
			this.Data["json"] = map[string]interface{}{
				"errcode": getErrCode(err),
				"errmsg":  err.Error(),
			}
		} else {
			userid := this.GetString("userid")
			content := this.GetString("content")
			go func() {
				for _, each := range codes {
					models.AddOneLog(userid, each, 7, content)
				}
				sendDenyMessage(codes, content)
			}()
			this.Data["json"] = map[string]interface{}{
				"errcode": OK,
				"message": "OK",
			}
		}
	}
	this.ServeJSON()
}

func (this *MainDataController) GetAuditList() {
	task_code := this.GetString("task_code")
	err, info := models.GetAllWorkflowList(task_code)
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

func (this *MainDataController) GetAllLogByDataCode() {
	data_code := this.GetString("data_code")
	l_info := &models.Log{DataCode: data_code}
	err, info := l_info.GetAllLogByDataCode()
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

func (this *MainDataController) AssignDataOtherToSelf() {
	task_code := this.GetString("task_code")
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	form_code := this.GetString("form_code")
	err, info := models.AssignDataToUserSelf(task_code, userid, corpid, form_code)
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

func (this *MainDataController) GetCountByTaskCodeAndFormCode() {
	task_code := this.GetString("task_code")
	form_code := this.GetString("form_code")
	m_info := &models.MainData{
		TaskCode: task_code,
		FormCode: form_code,
	}
	count, err := m_info.GetCountByTaskCodeAndFormCode()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    count,
		}
	}
	this.ServeJSON()
}

func sendDenyMessage(codes []string, content string) {
	wx_api := AbstractAPI.NewWXAuthAPI(conf.WxAppId, conf.WxSecret, "")
	infos := models.MainDataCodesToUserids(codes)

	nowstr := time.Now().Format("2006-01-02 15:04:05")
	for _, each := range infos {
		userids := each["userids"].([]string)
		coprid := each["corpid"].(string)
		openids := models.UseridsAndDepartidsToOpenIds(userids, []int{}, coprid)
		content := fmt.Sprintf("您提交的【%s】任务中有数据被驳回。\n原因：%s", each["task_name"], content)
		message_data := map[string]interface{}{
			"first":    map[string]interface{}{"value": content},
			"keyword1": map[string]interface{}{"value": "驳回通知"},
			"keyword2": map[string]interface{}{"value": nowstr},
		}
		for _, each_openid := range openids {
			wx_api.SendMessage(each_openid, conf.WxTemplateMessageId, "", message_data)
		}
	}
}
