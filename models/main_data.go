package models

import (
	"DeepWorkload/conf"
	"DeepWorkload/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego/orm"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	uuid "github.com/satori/go.uuid"
	"strconv"
	"strings"
	"time"
)

type MainData struct {
	Id             int             `orm:"column(id);pk;auto"`
	Code           string          `orm:"column(code)"`
	CorpId         string          `orm:"column(corpid)"`
	FormFieldInfo  string          `orm:"column(form_field_info);type(json)"`
	FormFieldInfos []FormFieldInfo `orm:"-"`
	FormCode       string          `orm:"column(form_code)"`
	FormName       string          `orm:"-"`
	AuditState     string          `orm:"column(audit_state);type(json)"`
	AuditStates    []MainDataAudit `orm:"-"`
	CreateUserId   string          `orm:"column(create_userid)"`
	CreateUserName string          `orm:"-"`
	TaskCode       string          `orm:"column(task_code)"`
	TaskName       string          `orm:"-"`
	CreateTime     time.Time       `orm:"auto_now_add;type(datetime);column(create_time)"`
	State          int             `orm:"column(state)"` //0.暂存 1.已提交 2.被驳回,3 .通过  4.审核中, 5.不需要审核, 6.导入数据
	AssignState    int             `orm:"column(assign_state)"`
	FField         []JsonFormField `orm:"-"`
	EnableAssign   bool            `orm:"-"`
}

type AuditState struct {
	UserId string `json:"userid"`
	Name   string `json:"name"`
	State  int    `json:"state"`
}

type MainDataAudit struct {
	UserIds       []string            `json:"userids"`
	UserIdAndName []UserNameAndUserId `json:"user_info"`
	Type          int                 `json:"type"`
	NodeState     int                 `json:"node_state"`
	UserAudit     AuditState          `json:"user_audit"`
}

func (this *MainData) TableName() string {
	return "main_data"
}

func InsertMainDatasByExcel(datas []MainData) error {
	o := orm.NewOrm()
	if len(datas) > 0 {
		_, err := o.InsertMulti(len(datas), datas)
		return err
	}
	return nil
}

func (this *MainData) InsertOne() (error, string) {
	o := orm.NewOrm()
	this.Code = uuid.NewV4().String()
	_, err := o.Insert(this)
	json.Unmarshal([]byte(this.FormFieldInfo), &this.FormFieldInfos)

	for _, each := range this.FormFieldInfos {

		if len(each.AssignData) > 0 {
			go func(data []Assign, field_code, form_code string) {
				as := []AssignData{}
				for _, each_as := range data {
					as = append(as, AssignData{
						DataCode:  this.Code,
						ToUserId:  each_as.ToUserId,
						ToValue:   each_as.ToValue,
						FieldCode: field_code,
						FormCode:  form_code,
						TaskCode:  this.TaskCode,
					})
				}
				InsertMultiAssignData(as)
			}(each.AssignData, each.FieldCode, this.FormCode)
		}
	}

	if err != nil {
		fmt.Println(err.Error())
	}
	return err, this.Code
}

func (this *MainData) AllMainDataByTaskCodeAndUserId() (error, []MainData) {
	o := orm.NewOrm()
	main_info := []MainData{}
	_, err := o.QueryTable("main_data").Filter("task_code", this.TaskCode).
		Filter("create_userid", this.CreateUserId).All(&main_info)
	for index, each := range main_info {
		json.Unmarshal([]byte(each.FormFieldInfo), &each.FormFieldInfos)
		if each.AuditState != "" {
			json.Unmarshal([]byte(each.AuditState), &each.AuditStates)
		}

		for _, each_f := range each.FormFieldInfos {
			if each_f.EnableAssign == true {
				each.EnableAssign = true
				break
			}
		}
		main_info[index] = each
	}
	return err, main_info
}

func getUseridsFromMainData(main_info []MainData) (main_codes []string, userinfos []UserInfo, form_codes, task_codes []string) {
	userids := []string{}
	task_codes = []string{}
	for _, each := range main_info {
		main_codes = append(main_codes, each.Code)
		json.Unmarshal([]byte(each.AuditState), &each.AuditStates)
		for _, each_u := range each.AuditStates {
			userids = append(userids, each_u.UserIds...)
		}
		userids = append(userids, each.CreateUserId)
		form_codes = append(form_codes, each.FormCode)
		task_codes = append(task_codes, each.TaskCode)
	}
	userids = utils.DeleteRepeat(userids)
	form_codes = utils.DeleteRepeat(form_codes)
	if len(userids) > 0 {
		o := orm.NewOrm()
		o.QueryTable("user_info").Filter("userid__in", userids).Filter("corpid", main_info[0].CorpId).All(&userinfos)
	}
	return main_codes, userinfos, form_codes, task_codes
}

func formatMainData(main_info []MainData) []MainData {
	o := orm.NewOrm()
	ass_infos := []AssignData{}
	if len(main_info) == 0 {
		return []MainData{}
	}
	main_codes, userinfos, form_codes, task_codes := getUseridsFromMainData(main_info)
	form_infos := []FormInfo{}
	o.QueryTable("form_info").Filter("code__in", form_codes).All(&form_infos)
	o.QueryTable("assign_data").Filter("data_code__in", main_codes).All(&ass_infos)
	t_infos := []Task{}
	o.QueryTable("task").Filter("code__in", task_codes).All(&t_infos)
	ass_infos = formateAssignDataUserName(ass_infos)
	for index, each := range main_info {
		json.Unmarshal([]byte(each.FormFieldInfo), &each.FormFieldInfos)
		each.FormFieldInfo = ""
		for _, each_u := range userinfos {
			if each.CreateUserId == each_u.UserId {
				each.CreateUserName = each_u.Name
				break
			}
		}
		for _, each_form := range form_infos {
			if each.FormCode == each_form.Code {
				each.FormName = each_form.Title
				break
			}
		}
		for _, each_t := range t_infos {
			if each_t.Code == each.TaskCode {
				each.TaskName = each_t.Title
				break
			}
		}
		for index_f, each_f := range each.FormFieldInfos {
			temp := []Assign{}
			for _, each_ass := range ass_infos {
				if each_f.FieldCode == each_ass.FieldCode && each.Code == each_ass.DataCode {
					temp = append(temp, Assign{
						each_ass.ToUserId,
						each_ass.ToValue,
						each_ass.ToUserName,
					})
				}
			}
			each_f.AssignData = temp
			each.FormFieldInfos[index_f] = each_f
		}
		if each.AuditState != "" {
			each.AuditStates = formatAuditState(each.AuditState, userinfos)
			each.AuditState = ""
		}
		for _, each_f := range each.FormFieldInfos {
			if each_f.EnableAssign == true {
				each.EnableAssign = true
				break
			}
		}
		main_info[index] = each
	}
	return main_info
}

func formatAuditState(audit_state string, u_infos []UserInfo) []MainDataAudit {
	as := []MainDataAudit{}
	json.Unmarshal([]byte(audit_state), &as)
	for index, each := range as {
		uin := []UserNameAndUserId{}
		for _, each_u := range each.UserIds {
			for _, each_ui := range u_infos {
				if each_u == each_ui.UserId {
					uin = append(uin, UserNameAndUserId{
						Uid:  each_u,
						Name: each_ui.Name,
					})
					break
				}
			}
		}
		each.UserIdAndName = uin
		for _, each_u := range u_infos {
			if each_u.UserId == each.UserAudit.UserId {
				each.UserAudit.Name = each_u.Name
				break
			}
		}
		as[index] = each
	}
	return as
}

func (this *MainData) AllAuditMainDataByTaskCodeAndFormCode(userid, corpid string) (error, []MainData, interface{}) {
	o := orm.NewOrm()
	main_info := []MainData{}
	u_info := UserInfo{
		CorpId: corpid,
		UserId: userid,
	}
	u_ids := u_info.GetAllUserInfoByUserId()

	qs := o.QueryTable("main_data").Filter("task_code", this.TaskCode).Filter("form_code", this.FormCode).
		Filter("create_userid__in", u_ids).Filter("corpid", corpid)
	_, err := qs.OrderBy("-create_time").All(&main_info)
	main_info = formatMainData(main_info)
	field_infos := []FormField{}
	o.QueryTable("form_field").Filter("form_code", this.FormCode).Exclude("tag_icon__in", conf.NotExportIcon).OrderBy("id").All(&field_infos)
	main_info = formatMainDataFiledInfo(main_info, field_infos)

	return err, main_info, field_infos
}

type MainDataWithDepart struct {
	MainData
	Departments []string
}

func formatMainDataToDepart(m_infos []MainData) []MainDataWithDepart {
	if len(m_infos) == 0 {
		return []MainDataWithDepart{}
	}
	_, u_info, _, _ := getUseridsFromMainData(m_infos)
	o := orm.NewOrm()
	de_info := []Department{}
	o.QueryTable("department").Filter("corpid", m_infos[0].CorpId).All(&de_info)
	mds := []MainDataWithDepart{}
	for _, each := range m_infos {
		departids := []int{}
		for _, each_u := range u_info {
			if each.CreateUserId == each_u.UserId {
				departids = utils.SqlStringValue(each_u.DepartId)
				break
			}
		}
		departments := []string{}
		for _, each_dd := range departids {
			for _, each_d := range de_info {
				if each_d.DepartmentId == each_dd {
					departments = append(departments, each_d.Department)
					break
				}
			}
		}
		mds = append(mds, MainDataWithDepart{
			each,
			departments,
		})
	}
	return mds
}

func (this *MainData) AllPubMainDataByTaskCodeAndFormCode() (error, []MainDataWithDepart, interface{}) {
	o := orm.NewOrm()
	main_info := []MainData{}
	qs := o.QueryTable("main_data").Filter("task_code", this.TaskCode).
		Filter("form_code", this.FormCode).Filter("state__in", []int{3, 5})
	_, err := qs.OrderBy("-create_time").All(&main_info)
	main_info = formatMainData(main_info)
	field_infos := []FormField{}
	o.QueryTable("form_field").Filter("form_code", this.FormCode).Filter("is_counted", false).
		Exclude("tag_icon__in", conf.NotExportIcon).OrderBy("id").All(&field_infos)
	main_info = formatMainDataFiledInfo(main_info, field_infos)
	md_info := formatMainDataToDepart(main_info)

	return err, md_info, field_infos
}

func (this *MainData) AllMainDataByTaskCodeAndFormCode() (error, []MainData, interface{}) {
	o := orm.NewOrm()
	main_info := []MainData{}
	qs := o.QueryTable("main_data").Filter("task_code", this.TaskCode).Filter("form_code", this.FormCode)
	if this.CreateUserId != "" {
		qs = qs.Filter("create_userid", this.CreateUserId)
	}
	_, err := qs.OrderBy("-create_time").All(&main_info)
	main_info = formatMainData(main_info)
	field_infos := []FormField{}
	o.QueryTable("form_field").Filter("form_code", this.FormCode).Exclude("tag_icon__in", conf.NotExportIcon).OrderBy("id").All(&field_infos)
	main_info = formatMainDataFiledInfo(main_info, field_infos)

	//for index,each := range main_info{
	//	json.Unmarshal([]byte(each.FormFieldInfo),&each.FormFieldInfos)
	//	temp := []FormFieldInfo{}
	//	for _,each_f := range field_infos{
	//		is_in := false
	//		if !(each_f.TagIcon == "divider" || each_f.TagIcon == "describe"){
	//			for _,each_ff := range each.FormFieldInfos{
	//
	//				if each_f.FiledCode == each_ff.FieldCode{
	//					each_ff.StrList= utils.StringValueToStrArray(each_f.StrList)
	//					temp = append(temp, each_ff)
	//					is_in = true
	//					break
	//				}
	//			}
	//			if is_in == false{
	//				temp = append(temp,FormFieldInfo{
	//					FieldLabel:each_f.Label,
	//					FieldCode:each_f.FiledCode,
	//					JoinGather:each_f.JoinGather,
	//					IsCounted:each_f.IsCounted,
	//					EnableAssign:each_f.EnableAssign,
	//					EnableGreater:each_f.EnableGreater,
	//					StrList:utils.StringValueToStrArray(each_f.StrList),
	//				})
	//			}
	//		}
	//	}
	//	if len(temp) > 0{
	//		each.FormFieldInfos = temp
	//		main_info[index] = each
	//	}
	//}
	return err, main_info, field_infos
}

func (this *MainData) GetMainDataByFormCodeAndUserId() {
	o := orm.NewOrm()
	infos := []MainData{}
	o.QueryTable("main_data").Filter("form_code", this.FormCode).
		Filter("task_code", this.TaskCode).Filter("create_userid", this.CreateUserId).All(&infos)
	for index, each := range infos {
		json.Unmarshal([]byte(each.FormFieldInfo), &each.FormFieldInfos)
		infos[index] = each
	}
}

func (this *MainData) DelMainDataByTaskCode(check bool) error {
	o := orm.NewOrm()
	m_info := MainData{}
	err := o.QueryTable("main_data").Filter("code", this.Code).One(&m_info)
	if check {
		err = checkMainData(m_info, "del")
		if err != nil {
			return err
		}
	}

	_, err = o.QueryTable("main_data").Filter("code", this.Code).Delete()
	return err
}

func BatchDelMainData(codes []string, check bool) error {
	o := orm.NewOrm()
	if len(codes) == 0 {
		return nil
	}
	if check {
		m_infos := []MainData{}
		err := o.QueryTable("main_data").Filter("code__in", codes).One(&m_infos)
		for _, each := range m_infos {
			err = checkMainData(each, "del")
			if err != nil {
				return err
			}
		}
	}
	_, err := o.QueryTable("main_data").Filter("code__in", codes).Delete()
	return err
}

func checkMainData(m_info MainData, types string) error {
	switch types {
	case "batch_commit":
		if m_info.State == 1 {
			return errors.New("已经略去已经提交的项")
		}
		if m_info.AssignState == 1 {
			return errors.New("已经略去分配未完成的项，请分配完成后提交")
		}
		if m_info.State == 3 {
			return errors.New("已经略去已经通过的项")
		}
		if m_info.State == 6 {
			return errors.New("已经略去导入的数据，请把导入的数据编辑确认后再提交")
		}
	case "commit":
		if m_info.State == 1 {
			return errors.New("已经提交，不能提交")
		}
		if m_info.AssignState == 1 {
			return errors.New("分配未完成，不能提交")
		}
		if m_info.State == 3 {
			return errors.New("已经通过，不能再次提交")
		}
		if m_info.State == 6 {
			return errors.New("导入的数据，不能直接提交，请编辑后提交")
		}
	case "del":
		if m_info.State == 3 {
			return errors.New("已经通过，不能删除")
		}
	case "callback":
		if m_info.State == 3 {
			return nil
		}
	}
	return nil
}

func (this *MainData) EditMainDataFieldInfo() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("main_data").Filter("code", this.Code).Update(
		orm.Params{"form_field_info": this.FormFieldInfo})
	m_info := MainData{}
	o.QueryTable("main_data").Filter("code", this.Code).One(&m_info)
	if m_info.State == 6 {
		o.QueryTable("main_data").Filter("code", this.Code).Update(
			orm.Params{"state": 0})
	}
	o.QueryTable("assign_data").Filter("data_code", this.Code).Delete()
	json.Unmarshal([]byte(this.FormFieldInfo), &this.FormFieldInfos)
	is_ass := false
	for _, each := range this.FormFieldInfos {
		if len(each.AssignData) > 0 {
			is_ass = true
			go func(ass []Assign, field_code string, m MainData) {
				as := []AssignData{}
				for _, each_as := range ass {
					as = append(as, AssignData{
						DataCode:  this.Code,
						ToUserId:  each_as.ToUserId,
						ToValue:   each_as.ToValue,
						FieldCode: field_code,
						FormCode:  m.FormCode,
						TaskCode:  m.TaskCode,
					})
				}
				InsertMultiAssignData(as)
			}(each.AssignData, each.FieldCode, m_info)
		}
	}
	if is_ass == false && m_info.AssignState != 0 {
		o.QueryTable("main_data").Filter("code", this.Code).Update(
			orm.Params{"assign_state": 0})
	}
	return err
}

func (this *MainData) MainDataDetailInfoByCode() (error, MainData) {
	o := orm.NewOrm()
	formf := []FormField{}
	md := MainData{}
	err := o.QueryTable("main_data").Filter("code", this.Code).One(&md)
	mds := formatMainData([]MainData{md})
	md = mds[0]
	assign_data := []AssignData{}
	o.QueryTable("assign_data").Filter("data_code", this.Code).All(&assign_data)
	assign_data = formateAssignDataUserName(assign_data)
	_, err = o.QueryTable("form_field").Filter("form_code", md.FormCode).OrderBy("id").All(&formf)
	json.Unmarshal([]byte(md.FormFieldInfo), &md.FormFieldInfos)
	if md.AuditState != "" {
		json.Unmarshal([]byte(md.AuditState), &md.AuditStates)
	}
	fields := FormFieldToJsonFormField(formf)
	for index, each := range fields {
		for _, eachf := range md.FormFieldInfos {
			if each.FiledCode == eachf.FieldCode {
				fields[index].Value = eachf.Value
				break
			}
		}
		temp := []AssignData{}
		for _, each_a := range assign_data {
			if each.FiledCode == each_a.FieldCode {
				temp = append(temp, each_a)
			}
		}
		fields[index].AssignData = temp
	}
	md.FField = fields
	for _, each := range md.FormFieldInfos {
		if each.EnableAssign == true {
			md.EnableAssign = true
			break
		}
	}
	return err, md
}

//0.暂存 1.已提交 2.被驳回,3 .通过  4.审核中, 5.不需要审核
func (this *MainData) CommitOneMainDataWithTask() error {
	o := orm.NewOrm()
	m_info := MainData{}
	err := o.QueryTable("main_data").Filter("code", this.Code).One(&m_info)
	err = checkMainData(m_info, "commit")
	if err != nil {
		return err
	}
	u_info := UserInfo{}
	err = o.QueryTable("user_info").Filter("userid", m_info.CreateUserId).Filter("corpid", m_info.CorpId).One(&u_info)
	if u_info.UserId == "" {
		return errors.New("no such userid")
	}
	t_info := Task{}
	err = o.QueryTable("task").Filter("code", m_info.TaskCode).One(&t_info)
	u_info.DepartIds = utils.StringValueToIntArray(u_info.DepartId)

	ma := GetWorkflowUserList(u_info, t_info.WorkflowCode)
	if len(ma) == 0 {
		_, err = o.QueryTable("main_data").Filter("code", this.Code).
			Update(orm.Params{"state": 3})
	} else {
		audit_state, _ := json.Marshal(ma)
		_, err = o.QueryTable("main_data").Filter("code", this.Code).
			Update(orm.Params{"state": 1, "audit_state": string(audit_state)})
	}
	return err
}

func BatchCommitMainDataWithTask(codes []string) (error, []string) {
	o := orm.NewOrm()
	m_infos := []MainData{}
	if len(codes) == 0 {
		return nil, []string{}
	}
	o.QueryTable("main_data").Filter("code__in", codes).All(&m_infos)
	commit_codes := []string{}
	var err error
	for _, each := range m_infos {
		cerr := checkMainData(each, "batch_commit")
		if cerr == nil {
			commit_codes = append(commit_codes, each.Code)
		} else {
			err = cerr
		}
	}
	if len(commit_codes) == 0 {
		return err, []string{}
	}
	u_info := UserInfo{}
	o.QueryTable("user_info").Filter("userid", m_infos[0].CreateUserId).Filter("corpid", m_infos[0].CorpId).One(&u_info)
	t_info := Task{}
	o.QueryTable("task").Filter("code", m_infos[0].TaskCode).One(&t_info)
	u_info.DepartIds = utils.StringValueToIntArray(u_info.DepartId)
	ma := GetWorkflowUserList(u_info, t_info.WorkflowCode)
	audit_state, _ := json.Marshal(ma)
	o.QueryTable("main_data").Filter("code__in", commit_codes).
		Update(orm.Params{"state": 1, "audit_state": string(audit_state)})
	return err, commit_codes
}

func (this *MainData) CommitOneMainDataWithOutWorkflow() error {
	o := orm.NewOrm()
	m_info := MainData{}
	err := o.QueryTable("main_data").Filter("code", this.Code).One(&m_info)
	err = checkMainData(m_info, "commit")
	if err != nil {
		return err
	}
	_, err = o.QueryTable("main_data").Filter("code", this.Code).
		Update(orm.Params{"state": 5})
	return err
}

func BatchCommitMainDataWithOutWorkflow(codes []string) (error, []string) {
	o := orm.NewOrm()
	m_infos := []MainData{}
	if len(codes) == 0 {
		return nil, []string{}
	}
	o.QueryTable("main_data").Filter("code__in", codes).All(&m_infos)
	var err error
	commit_codes := []string{}
	for _, each := range m_infos {
		cerr := checkMainData(each, "batch_commit")
		if cerr == nil {
			commit_codes = append(commit_codes, each.Code)
		} else {
			err = cerr
		}
	}

	if len(commit_codes) == 0 {
		return err, []string{}
	}
	o.QueryTable("main_data").Filter("code__in", commit_codes).Update(orm.Params{"state": 5})
	return err, commit_codes
}

func (this *MainData) CommitOneMainDataWithNodeCode(depart_code, node_code string) error {
	o := orm.NewOrm()
	m_info := MainData{}
	err := o.QueryTable("main_data").Filter("code", this.Code).One(&m_info)
	err = checkMainData(m_info, "commit")
	if err != nil {
		return err
	}
	t_info := Task{}
	err = o.QueryTable("task").Filter("code", m_info.TaskCode).One(&t_info)
	ma := GetWorkflowDepartAndCondition(depart_code, node_code, t_info.WorkflowCode)
	audit_state, _ := json.Marshal(ma)
	_, err = o.QueryTable("main_data").Filter("code", this.Code).
		Update(orm.Params{"state": 1, "audit_state": string(audit_state)})
	return err
}

func BatchCommitMainDataWithNodeCode(depart_code, node_code string, codes []string) (error, []string) {
	o := orm.NewOrm()
	if len(codes) == 0 {
		return nil, []string{}
	}
	m_infos := []MainData{}
	o.QueryTable("main_data").Filter("code__in", codes).All(&m_infos)
	commit_codes := []string{}
	var err error
	for _, each := range m_infos {
		cerr := checkMainData(each, "batch_commit")
		if cerr == nil {
			commit_codes = append(commit_codes, each.Code)
		} else {
			err = cerr
		}
	}
	if len(commit_codes) == 0 {
		return err, []string{}
	}
	t_info := Task{}
	o.QueryTable("task").Filter("code", m_infos[0].TaskCode).One(&t_info)
	ma := GetWorkflowDepartAndCondition(depart_code, node_code, t_info.WorkflowCode)
	audit_state, _ := json.Marshal(ma)
	o.QueryTable("main_data").Filter("code__in", commit_codes).
		Update(orm.Params{"state": 1, "audit_state": string(audit_state)})
	return err, commit_codes
}

func (this *MainData) CallBackMainData() error {
	o := orm.NewOrm()
	m_info := MainData{}
	err := o.QueryTable("main_data").Filter("code", this.Code).One(&m_info)
	err = checkMainData(m_info, "callback")
	if err != nil {
		return err
	}
	_, err = o.QueryTable("main_data").Filter("code", this.Code).
		Update(orm.Params{"state": 0, "audit_state": nil})
	return err
}

type AuditMainData struct {
	MainData
	TaskName string
}

func formatMainDataFiledInfo(m_infos []MainData, field_infos []FormField) []MainData {
	for index, each := range m_infos {
		json.Unmarshal([]byte(each.FormFieldInfo), &each.FormFieldInfos)
		temp := []FormFieldInfo{}
		for _, each_f := range field_infos {
			if each.FormCode == each_f.FormCode {
				is_in := false
				if !utils.IsExistStr(conf.NotExportIcon, each_f.TagIcon) {
					for _, each_ff := range each.FormFieldInfos {
						if each_f.FiledCode == each_ff.FieldCode {
							each_ff.StrList = utils.StringValueToStrArray(each_f.StrList)
							temp = append(temp, each_ff)
							is_in = true
							break
						}
					}
					if is_in == false {
						temp = append(temp, FormFieldInfo{
							FieldLabel:    each_f.Label,
							FieldCode:     each_f.FiledCode,
							JoinGather:    each_f.JoinGather,
							IsCounted:     each_f.IsCounted,
							EnableAssign:  each_f.EnableAssign,
							EnableGreater: each_f.EnableGreater,
							StrList:       utils.StringValueToStrArray(each_f.StrList),
						})
					}
				}
			}
		}
		if len(temp) > 0 {
			each.FormFieldInfos = temp
			m_infos[index] = each
		}
	}
	return m_infos
}

func GetAllAuditedMainDataList(userid, corpid string, nodes_state int) (error, []AuditMainData) {
	o := orm.NewOrm()
	m_infos := []MainData{}
	rem_infos := []MainData{}
	task_codes := []string{}
	active_task_codes := []string{}
	t_info := []Task{}
	o.QueryTable("task").Filter("corpid", corpid).Filter("state", 0).All(&t_info)
	if len(t_info) == 0 {
		return nil, []AuditMainData{}

	}
	for _, each := range t_info {
		active_task_codes = append(active_task_codes, each.Code)
	}

	_, err := o.QueryTable("main_data").Filter("state__in", []int{1, 3, 4}).Filter("task_code__in", active_task_codes).All(&m_infos)
	if len(m_infos) == 0 {
		return nil, []AuditMainData{}
	}
	m_infos = formatMainData(m_infos)
	form_codes := []string{}
	for _, each := range m_infos {
		form_codes = append(form_codes, each.FormCode)
	}
	f_codes := utils.DeleteRepeat(form_codes)
	field_infos := []FormField{}
	o.QueryTable("form_field").Filter("form_code__in", f_codes).OrderBy("id").All(&field_infos)
	m_infos = formatMainDataFiledInfo(m_infos, field_infos)

	for _, each := range m_infos {
		task_codes = append(task_codes, each.TaskCode)
		is_u := false
		for _, each_a := range each.AuditStates {
			if utils.StrInArray(userid, each_a.UserIds) && (each_a.NodeState == nodes_state) {
				is_u = true
				break
			}
		}
		if is_u == true {
			rem_infos = append(rem_infos, each)
		}
	}
	task_codes = utils.DeleteRepeat(task_codes)
	if len(task_codes) == 0 {
		return nil, []AuditMainData{}

	}
	t_infos := []Task{}
	taskm := map[string]string{}
	_, err = o.QueryTable("task").Filter("code__in", task_codes).All(&t_infos)
	for _, each := range t_infos {
		taskm[each.Code] = each.Title
	}
	re_data := []AuditMainData{}
	for _, each := range rem_infos {
		re_data = append(re_data, AuditMainData{
			each,
			taskm[each.TaskCode],
		})
	}
	return err, re_data
}

func mainDataToUserid(task_code string, m_infos []MainData) []string {
	userid := []string{}
	for _, each := range m_infos {
		if each.TaskCode == task_code {
			userid = append(userid, each.CreateUserId)
		}
	}
	return userid
}

func (this *MainData) DenyOneMainData() error {
	o := orm.NewOrm()
	//m_info := MainData{}
	//err:=o.QueryTable("main_data").Filter("code",this.Code).One(&m_info)
	_, err := o.QueryTable("main_data").Filter("code", this.Code).
		Update(orm.Params{"state": 2, "audit_state": nil})
	return err
}

func DenyMainDataByCodes(codes []string) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("main_data").Filter("code__in", codes).
		Update(orm.Params{"state": 2, "audit_state": nil})
	return err
}

func (this *MainData) ApproveMainData(userid string) error {
	o := orm.NewOrm()
	m_info := MainData{}
	err := o.QueryTable("main_data").Filter("code", this.Code).One(&m_info)
	json.Unmarshal([]byte(m_info.AuditState), &m_info.AuditStates)
	for index, each := range m_info.AuditStates {
		if utils.StrInArray(userid, each.UserIds) && (each.NodeState == 1) {
			each.NodeState = 0
			each.UserAudit = AuditState{UserId: userid}
			m_info.AuditStates[index] = each
			break
		}
	}
	audit, _ := json.Marshal(m_info.AuditStates)
	is_last := false
	if m_info.AuditStates[len(m_info.AuditStates)-1].NodeState == 0 {
		is_last = true
	}
	state := 4
	if is_last {
		state = 3
	}
	_, err = o.QueryTable("main_data").Filter("code", this.Code).
		Update(orm.Params{"state": state, "audit_state": string(audit)})
	return err
}

func ApproveMainDatas(userid string, codes []string) error {
	o := orm.NewOrm()
	m_infos := []MainData{}
	_, err := o.QueryTable("main_data").Filter("code__in", codes).All(&m_infos)
	for _, each_m := range m_infos {
		json.Unmarshal([]byte(each_m.AuditState), &each_m.AuditStates)
		for index, each := range each_m.AuditStates {
			if utils.StrInArray(userid, each.UserIds) && (each.NodeState == 1) {
				each.NodeState = 0
				each.UserAudit = AuditState{UserId: userid}
				each_m.AuditStates[index] = each
				break
			}
		}
		audit, _ := json.Marshal(each_m.AuditStates)
		is_last := false
		if each_m.AuditStates[len(each_m.AuditStates)-1].NodeState == 0 {
			is_last = true
		}
		state := 4
		if is_last {
			state = 3
		}
		_, err = o.QueryTable("main_data").Filter("code", each_m.Code).
			Update(orm.Params{"state": state, "audit_state": string(audit)})
	}
	return err
}

func (this *MainData) GatherListByAuditUserId() ([]map[string]interface{}, map[string]interface{}) {
	o := orm.NewOrm()
	u_info := UserInfo{
		CorpId: this.CorpId,
		UserId: this.CreateUserId,
	}
	u_ids := u_info.GetAllUserInfoByUserId()
	if len(u_ids) == 0 {
		return []map[string]interface{}{}, nil
	}
	infos := []MainData{}
	_, err := o.QueryTable("main_data").Filter("task_code", this.TaskCode).
		Filter("state__in", []int{3, 5}).Filter("corpid", this.CorpId).Filter("create_userid__in", u_ids).All(&infos)

	form_codes := []string{}
	t_info := Task{}
	err = o.QueryTable("task").Filter("code", this.TaskCode).One(&t_info)
	form_codes = utils.StringValueToStrArray(t_info.FormCode)
	as_datas := []AssignData{}
	sql := fmt.Sprintf("select assign_data.id, assign_data.to_userid, assign_data.create_time, "+
		"assign_data.data_code, assign_data.to_departid, assign_data.to_value, assign_data.field_code, "+
		"assign_data.form_code, assign_data.task_code "+
		" from assign_data,main_data where assign_data.task_code = '%s' "+
		"and main_data.state in (3,5) and main_data.code = assign_data.data_code", this.TaskCode)
	sql += fmt.Sprintf(" and assign_data.to_userid in (%s)", utils.StringArrayToINArray(u_ids))
	_, err = o.Raw(sql).QueryRows(&as_datas)
	if err != nil {
		fmt.Println(err.Error())
	}
	f_map, ff_info := formatFormMap(form_codes, true)
	ga_data := gatherData(as_datas, infos, ff_info)
	return formatGatherData(ga_data, this.CorpId), f_map
}

func (this *MainData) ExportAuditUserId() (path string, file_name string, out_name string) {
	o := orm.NewOrm()
	u_info := UserInfo{
		CorpId: this.CorpId,
		UserId: this.CreateUserId,
	}
	u_ids := u_info.GetAllUserInfoByUserId()
	all_infos := []MainData{}
	_, err := o.QueryTable("main_data").Filter("task_code", this.TaskCode).
		Filter("state__in", []int{3, 5}).Filter("corpid", this.CorpId).Filter("create_userid__in", u_ids).All(&all_infos)

	form_codes := []string{}
	t_info := Task{}
	err = o.QueryTable("task").Filter("code", this.TaskCode).One(&t_info)
	form_codes = utils.StringValueToStrArray(t_info.FormCode)
	as_datas := []AssignData{}
	sql := fmt.Sprintf("select assign_data.id, assign_data.to_userid, assign_data.create_time, "+
		"assign_data.data_code, assign_data.to_departid, assign_data.to_value, assign_data.field_code, "+
		"assign_data.form_code, assign_data.task_code "+
		" from assign_data,main_data where assign_data.task_code = '%s' "+
		"and main_data.state in (3,5) and main_data.code = assign_data.data_code", this.TaskCode)
	sql += fmt.Sprintf(" and assign_data.to_userid in (%s)", utils.StringArrayToINArray(u_ids))
	_, err = o.Raw(sql).QueryRows(&as_datas)
	if err != nil {
		fmt.Println(err.Error())
	}
	gather_map, ff_infos := formatFormMap(form_codes, true)
	all_map, _ := formatFormMap(form_codes, false)
	ga_data := gatherData(as_datas, all_infos, ff_infos)
	gaf_data := formatGatherData(ga_data, this.CorpId)
	userdepart := getUserInfoMap(this.CorpId)
	task_info := Task{}
	o.QueryTable("task").Filter("code", this.TaskCode).One(&task_info)
	path, file_name = createExcel(task_info.Title, gather_map, all_map, gaf_data, userdepart, all_infos, as_datas)
	return path, file_name, task_info.Title
}

func (this *MainData) GatherList() ([]map[string]interface{}, map[string]interface{}) {
	o := orm.NewOrm()
	infos := []MainData{}
	qs := o.QueryTable("main_data").Filter("task_code", this.TaskCode).
		Filter("state__in", []int{3, 5}).Filter("corpid", this.CorpId)
	if this.CreateUserId != "" {
		qs = qs.Filter("create_userid", this.CreateUserId)
	}
	_, err := qs.All(&infos)
	form_codes := []string{}
	t_info := Task{}
	err = o.QueryTable("task").Filter("code", this.TaskCode).One(&t_info)
	form_codes = utils.StringValueToStrArray(t_info.FormCode)
	as_datas := []AssignData{}
	sql := fmt.Sprintf("select assign_data.id, assign_data.to_userid, assign_data.create_time, "+
		"assign_data.data_code, assign_data.to_departid, assign_data.to_value, assign_data.field_code, "+
		"assign_data.form_code, assign_data.task_code "+
		" from assign_data,main_data where assign_data.task_code = '%s' "+
		"and main_data.state in (3,5) and main_data.code = assign_data.data_code", this.TaskCode)

	if this.CreateUserId != "" {
		sql += fmt.Sprintf(" and to_userid = '%s'", this.CreateUserId)
	}
	_, err = o.Raw(sql).QueryRows(&as_datas)
	if err != nil {
		fmt.Println(err.Error())
	}
	f_map, ff_info := formatFormMap(form_codes, true)
	ga_data := gatherData(as_datas, infos, ff_info)
	return formatGatherData(ga_data, this.CorpId), f_map
}

func formatGatherData(redata map[string]map[string][]FormFieldInfo, corpid string) []map[string]interface{} {
	userids := []string{}
	for key, _ := range redata {
		userids = append(userids, key)
	}
	if len(userids) == 0 {
		return nil
	}

	o := orm.NewOrm()
	u_infos := []UserInfo{}
	o.QueryTable("user_info").Filter("corpid", corpid).Filter("userid__in", userids).All(&u_infos)
	u_name := map[string]string{}
	for _, each := range u_infos {
		u_name[each.UserId] = each.Name
	}

	re_data := []map[string]interface{}{}
	for key, value := range redata {
		temp_map := map[string]interface{}{}
		temp_map["userid"] = key
		temp_map["user_name"] = u_name[key]
		for form_code, each := range value {
			for _, each_f := range each {
				temp_map[each_f.FieldCode+"||"+form_code] = each_f.Value
			}
		}
		re_data = append(re_data, temp_map)

	}
	return re_data
}

func formatFormMap(form_codes []string, join_gather bool) (map[string]interface{}, []FormField) {
	o := orm.NewOrm()
	f_info := []FormInfo{}
	if len(form_codes) > 0 {
		o.QueryTable("form_info").Filter("code__in", form_codes).All(&f_info)
	}
	form_map := map[string]interface{}{}
	for _, each := range f_info {
		temp := map[string]interface{}{"name": each.Title}
		temp["data"] = []map[string]string{}
		form_map[each.Code] = temp
	}
	ff_infos := []FormField{}
	if len(form_codes) > 0 {
		qs := o.QueryTable("form_field").Filter("form_code__in", form_codes)
		if join_gather == true {
			qs = qs.Filter("join_gather", join_gather)
		}
		qs.OrderBy("id").All(&ff_infos)
	}
	for _, each := range ff_infos {
		if !utils.IsExistStr(conf.NotExportIcon, each.TagIcon) {
			temp := form_map[each.FormCode].(map[string]interface{})
			data := temp["data"].([]map[string]string)
			data = append(data, map[string]string{
				"filed_code": each.FiledCode + "||" + each.FormCode,
				"filed_name": each.Label,
			})
			temp["data"] = data
			form_map[each.FormCode] = temp
		}
	}
	return form_map, ff_infos
}

func gatherData(assign_data []AssignData, all_data []MainData, ff_infos []FormField) map[string]map[string][]FormFieldInfo {
	gather := map[string]map[string][]FormFieldInfo{}
	for _, each := range all_data {
		json.Unmarshal([]byte(each.FormFieldInfo), &each.FormFieldInfos)
		if gather[each.CreateUserId] == nil || gather[each.CreateUserId][each.FormCode] == nil {
			rf := []FormFieldInfo{}
			for _, each_f := range ff_infos {
				if each_f.JoinGather == true && each_f.FormCode == each.FormCode {
					rf = append(rf, FormFieldInfo{
						FieldCode:  each_f.FiledCode,
						FieldLabel: each_f.Label,
						JoinGather: each_f.JoinGather,
						Value:      0.00,
					})
				}
			}
			if gather[each.CreateUserId] == nil {
				temp := map[string][]FormFieldInfo{
					each.FormCode: rf,
				}
				gather[each.CreateUserId] = temp
			} else if gather[each.CreateUserId][each.FormCode] == nil {
				gather[each.CreateUserId][each.FormCode] = rf
			}
		}

		fileds := gather[each.CreateUserId][each.FormCode]
		cp_fileds := make([]FormFieldInfo, len(fileds))
		copy(cp_fileds, fileds)
		for index, each_f := range cp_fileds {
			for _, each_c := range each.FormFieldInfos {
				if (each_f.FieldCode == each_c.FieldCode) && (each_c.JoinGather == true) && (len(each_c.AssignData) == 0) {
					if each_c.Value == nil {
						each_c.Value = float64(0)
					}
					each_f.Value = round(each_f.Value.(float64)+each_c.Value.(float64), 2)
				}
			}
			cp_fileds[index] = each_f
		}
		gather[each.CreateUserId][each.FormCode] = cp_fileds

	}

	for _, each_as := range assign_data {
		tempF := FormFieldInfo{}
		for _, each := range ff_infos {
			if each_as.FormCode == each.FormCode && each_as.FieldCode == each.FiledCode {
				tempF.IsCounted = each.IsCounted
				tempF.Value = each_as.ToValue
				tempF.JoinGather = each.JoinGather
				tempF.FieldCode = each_as.FieldCode
				tempF.FieldLabel = each.Label
				break
			}
		}
		if gather[each_as.ToUserId] != nil {
			if gather[each_as.ToUserId][each_as.FormCode] != nil {
				fileds := gather[each_as.ToUserId][each_as.FormCode]
				cp_fileds := make([]FormFieldInfo, len(fileds))
				copy(cp_fileds, fileds)
				for index, each_f := range cp_fileds {
					if tempF.JoinGather == true {
						if each_f.FieldCode == each_as.FieldCode {
							each_f.Value = each_f.Value.(float64) + tempF.Value.(float64)
						}
						cp_fileds[index] = each_f
					}
				}
				gather[each_as.ToUserId][each_as.FormCode] = cp_fileds
			} else {
				if tempF.JoinGather == true {
					gather[each_as.ToUserId][each_as.FormCode] = []FormFieldInfo{tempF}
				}
			}
		} else {
			if tempF.JoinGather == true {
				rf := []FormFieldInfo{tempF}
				temp := map[string][]FormFieldInfo{
					each_as.FormCode: rf,
				}
				gather[each_as.ToUserId] = temp
			}
		}
	}
	return gather
}

type UDInfo struct {
	UserId      string
	UserName    string
	Departments []string
}

func getUserInfoMap(corpid string) map[string]UDInfo {
	o := orm.NewOrm()
	u_info := []UserInfo{}
	o.QueryTable("user_info").Filter("corpid", corpid).All(&u_info)
	d_info := []Department{}
	o.QueryTable("department").Filter("corpid", corpid).All(&d_info)

	u_map := map[string]UDInfo{}
	for _, each_u := range u_info {
		ud := UDInfo{
			UserId:   each_u.UserId,
			UserName: each_u.Name,
		}
		departids := utils.StringValueToIntArray(each_u.DepartId)
		for _, each_d := range departids {
			for _, each_de := range d_info {
				if each_d == each_de.DepartmentId {
					ud.Departments = append(ud.Departments, each_de.Department)
					break
				}
			}
		}
		u_map[each_u.UserId] = ud
	}
	return u_map
}

func (this *MainData) Export() (path string, file_name string, out_name string) {
	o := orm.NewOrm()
	all_infos := []MainData{}
	qs := o.QueryTable("main_data").Filter("task_code", this.TaskCode).
		Filter("state__in", []int{3, 5}).Filter("corpid", this.CorpId)
	if this.CreateUserId != "" {
		qs = qs.Filter("create_userid", this.CreateUserId)
	}
	_, err := qs.All(&all_infos)
	form_codes := []string{}
	t_info := Task{}
	err = o.QueryTable("task").Filter("code", this.TaskCode).One(&t_info)
	form_codes = utils.StringValueToStrArray(t_info.FormCode)
	as_datas := []AssignData{}
	sql := fmt.Sprintf("select assign_data.id, assign_data.to_userid, assign_data.create_time, "+
		"assign_data.data_code, assign_data.to_departid, assign_data.to_value, assign_data.field_code, "+
		"assign_data.form_code, assign_data.task_code "+
		" from assign_data,main_data where assign_data.task_code = '%s' "+
		"and main_data.state in (3,5) and main_data.code = assign_data.data_code", this.TaskCode)

	if this.CreateUserId != "" {
		sql += fmt.Sprintf(" and to_userid = '%s'", this.CreateUserId)
	}
	_, err = o.Raw(sql).QueryRows(&as_datas)
	if err != nil {
		fmt.Println(err.Error())
	}
	gather_map, ff_infos := formatFormMap(form_codes, true)
	all_map, _ := formatFormMap(form_codes, false)
	ga_data := gatherData(as_datas, all_infos, ff_infos)
	gaf_data := formatGatherData(ga_data, this.CorpId)
	userdepart := getUserInfoMap(this.CorpId)
	task_info := Task{}
	o.QueryTable("task").Filter("code", this.TaskCode).One(&task_info)
	path, file_name = createExcel(task_info.Title, gather_map, all_map, gaf_data, userdepart, all_infos, as_datas)
	return path, file_name, task_info.Title
}

func createExcel(mission_title string, gather_map, full_map map[string]interface{},
	format_gather []map[string]interface{}, userdepart map[string]UDInfo,
	all_main_data []MainData, assign_data []AssignData) (string, string) {

	f := excelize.NewFile()
	sheetName := fmt.Sprintf("%s汇总", mission_title)
	index := f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(index)
	titles := []string{"工号", "姓名", "部门"}
	for index, each := range titles {
		file_index := fmt.Sprintf("%s%d", string(65+index), 2)
		f.SetCellStr(sheetName, file_index, each)
	}
	start := 68
	filedCodeMap := map[string]string{}
	for _, value := range gather_map {
		each := value.(map[string]interface{})["data"].([]map[string]string)
		title := value.(map[string]interface{})["name"].(string)
		start_index := fmt.Sprintf("%s%d", string(start), 1)
		end_index := fmt.Sprintf("%s%d", string(start+len(each)-1), 1)
		f.SetCellStr(sheetName, start_index, title)
		f.MergeCell(sheetName, start_index, end_index)
		for index, each_f := range each {
			file_index := fmt.Sprintf("%s%d", string(start+index), 2)
			f.SetCellStr(sheetName, file_index, each_f["filed_name"])
			filedCodeMap[each_f["filed_code"]] = string(start + index)
		}
		start += len(each)
	}
	for index, each := range format_gather {
		for k, v := range each {
			if k == "user_name" {
				continue
			}
			if k == "userid" {
				fillueseridanddepartid(f, index+3, sheetName, v.(string), userdepart)
			} else {
				file_index := fmt.Sprintf("%s%d", filedCodeMap[k], index+3)
				f.SetCellValue(sheetName, file_index, v)
			}
		}
	}
	for key, each := range full_map {
		filedCodeMap := map[string]string{}
		filedNameMap := map[string]string{}
		temp_sheet_name := each.(map[string]interface{})["name"].(string)
		f.NewSheet(temp_sheet_name)
		datas := each.(map[string]interface{})["data"].([]map[string]string)
		for index, each_t := range titles {
			file_index := fmt.Sprintf("%s%d", string(65+index), 1)
			f.SetCellStr(temp_sheet_name, file_index, each_t)
		}
		var end_file int
		for index, each_f := range datas {
			file_index := fmt.Sprintf("%s%d", string(68+index), 1)
			f.SetCellStr(temp_sheet_name, file_index, each_f["filed_name"])
			filedCodeMap[each_f["filed_code"]] = string(68 + index)
			filedNameMap[each_f["filed_code"]] = each_f["filed_name"]
			end_file = 68 + index
		}
		data_index := 2
		for _, each_m := range all_main_data {
			if each_m.FormCode != key {
				continue
			}
			fillueseridanddepartid(f, data_index, temp_sheet_name, each_m.CreateUserId, userdepart)
			json.Unmarshal([]byte(each_m.FormFieldInfo), &each_m.FormFieldInfos)
			for _, each_rf := range each_m.FormFieldInfos {
				if filedCodeMap[each_rf.FieldCode+"||"+each_m.FormCode] == "" {
					continue
				}
				file_index := fmt.Sprintf("%s%d", filedCodeMap[each_rf.FieldCode+"||"+each_m.FormCode], data_index)
				excel_value := each_rf.Value
				is_link := false
				switch each_rf.Value.(type) {
				case []interface{}:
					if len(each_rf.Value.([]interface{})) == 1 {
						is_link = true
						if each_rf.Value.([]interface{})[0].(map[string]interface{})["url"] != nil {
							excel_value = each_rf.Value.([]interface{})[0].(map[string]interface{})["url"].(string)
						} else {
							excel_value = ""
						}
					} else {
						ex_vs := []string{}
						for _, each_v := range each_rf.Value.([]interface{}) {
							if each_rf.Value.([]interface{})[0].(map[string]interface{})["url"] != nil {
								ex_vs = append(ex_vs, each_v.(map[string]interface{})["url"].(string))
							}
						}
						excel_value = strings.Join(ex_vs, "，")
					}
				}
				if is_link {
					if excel_value != nil {
						f.SetCellHyperLink(temp_sheet_name, file_index, excel_value.(string), "External")
					}
				}
				if excel_value != nil {
					if each_rf.OptLabel != "" {
						excel_value = cellValue(excel_value)
						excel_value = fmt.Sprintf("%s,%s", each_rf.OptLabel, excel_value)
					}
					f.SetCellValue(temp_sheet_name, file_index, excel_value)
				}
			}
			if each_m.AssignState == 2 {
				for _, each_as := range assign_data {
					if each_as.DataCode == each_m.Code {
						ass_index_c1 := filedCodeMap[each_as.FieldCode+"||"+each_m.FormCode+"||"+"Ass"]
						if ass_index_c1 == "" {
							filedCodeMap[each_as.FieldCode+"||"+each_m.FormCode+"||"+"Ass"] = string(end_file + 1)
							ass_index_c1 = string(end_file + 1)
							filed_name := filedNameMap[each_as.FieldCode+"||"+each_m.FormCode]
							ass_index1 := fmt.Sprintf("%s%d", string(end_file+1), 1)
							ass_index2 := fmt.Sprintf("%s%d", string(end_file+2), 1)
							ass_index3 := fmt.Sprintf("%s%d", string(end_file+3), 1)
							f.SetCellValue(temp_sheet_name, ass_index1, fmt.Sprintf("%s（被分配人工号）", filed_name))
							f.SetCellValue(temp_sheet_name, ass_index2, fmt.Sprintf("%s（被分配人名称）", filed_name))
							f.SetCellValue(temp_sheet_name, ass_index3, fmt.Sprintf("%s（被分配值）", filed_name))
							end_file += 3
						}
						ass_index1 := fmt.Sprintf("%s%d", ass_index_c1, data_index)
						ass_index_c2 := getNextLetters(ass_index_c1)
						ass_index2 := fmt.Sprintf("%s%d", ass_index_c2, data_index)
						ass_index3 := fmt.Sprintf("%s%d", getNextLetters(ass_index_c2), data_index)
						f.SetCellValue(temp_sheet_name, ass_index1, each_as.ToUserId)
						f.SetCellValue(temp_sheet_name, ass_index2, userdepart[each_as.ToUserId].UserName)
						f.SetCellValue(temp_sheet_name, ass_index3, each_as.ToValue)
						data_index += 1
					}
				}
				continue
			}
			data_index += 1
		}
	}

	file_name := mission_title + "_汇总" + strconv.FormatInt(time.Now().Unix(), 10) + ".xlsx"
	path := "uploads/docs/" + file_name
	err := f.SaveAs(path)
	if err != nil {
		fmt.Println(err)
	}
	return path, file_name
}

func cellValue(value interface{}) string {
	switch t := value.(type) {
	case float32:
		return strconv.FormatFloat(float64(value.(float32)), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(float64(value.(float64)), 'f', -1, 64)
	case string:
		return t
	case []byte:
		return string(t)
	case time.Duration:
		return strconv.FormatFloat(float64(value.(time.Duration).Seconds()/86400), 'f', -1, 32)
	case nil:
		return ""

	}
	return ""
}

func fillueseridanddepartid(file *excelize.File, index int, sheetname string, userid string, userdepart map[string]UDInfo) {
	u_index := fmt.Sprintf("%s%d", "A", index)
	file.SetCellValue(sheetname, u_index, userid)
	name_index := fmt.Sprintf("%s%d", "B", index)
	name_value := fmt.Sprintf("%s", userdepart[userid].UserName)
	file.SetCellStr(sheetname, name_index, name_value)
	depart_values := strings.Join(userdepart[userid].Departments, ",")
	depart_index := fmt.Sprintf("%s%d", "C", index)
	file.SetCellStr(sheetname, depart_index, depart_values)
}

func getNextLetters(this_letter string) string {
	codeAscaii := []rune(this_letter)
	if len(codeAscaii) == 1 {
		if codeAscaii[0] < 90 {
			return string(codeAscaii[0] + 1)
		} else {
			return "AA"
		}
	} else {
		if codeAscaii[1] == 90 {
			return fmt.Sprintf("%s%s", string(codeAscaii[0]+1), "A")
		} else {
			return fmt.Sprintf("%s%s", string(codeAscaii[0]), string(codeAscaii[1]+1))
		}
	}
}

func (this *MainData) GetCountByTaskCodeAndFormCode() (int64, error) {
	o := orm.NewOrm()
	return o.QueryTable("main_data").Filter("task_code", this.TaskCode).Filter("form_code", this.FormCode).Count()
}

func qiniuUrl(key string) string {
	mac := qbox.NewMac(conf.QN_AK, conf.QN_SK)
	deadline := time.Now().Add(time.Second * 3600 * 2).Unix() //2小时有效期
	privateAccessURL := storage.MakePrivateURL(mac, conf.QN_URL, key, deadline)
	return privateAccessURL
}

func GetAuditUseridCount() map[string]map[string]int {
	o := orm.NewOrm()
	m_infos := []MainData{}
	u_map := map[string]map[string]int{}
	task_code := []string{}
	t_info := []Task{}
	o.QueryTable("task").Filter("state", 0).All(&t_info)
	for _, each := range t_info {
		task_code = append(task_code, each.Code)
	}
	if len(task_code) == 0 {
		return nil
	}
	o.QueryTable("main_data").Filter("task_code__in", task_code).Filter("state__in", []int{1, 4}).All(&m_infos)
	if len(m_infos) == 0 {
		return nil
	}
	corpids := []string{}
	userids := []string{}
	for _, each := range m_infos {
		corpids = append(corpids, each.CorpId)
		if each.AuditState != "" {
			json.Unmarshal([]byte(each.AuditState), &each.AuditStates)
			for _, each_a := range each.AuditStates {
				userids = append(userids, each_a.UserIds...)
				for _, each_u := range each_a.UserIds {
					if u_map[each.CorpId][each_u] == 0 {
						u_map[each.CorpId][each_u] = 1
					} else {
						u_map[each.CorpId][each_u] += 1
					}
				}
			}
		}
	}
	corpids = utils.RemoveRepByMap(corpids)
	userids = utils.RemoveRepByMap(userids)
	u_info := []UserInfo{}
	o.QueryTable("user_info").Filter("corpid__in", corpids).Filter("userid__in", userids).All(&u_info)
	r_map := map[string]map[string]int{}
	for c, um := range u_map {
		for _, each_u := range u_info {
			if each_u.CorpId == c {
				if um[each_u.UserId] != 0 {
					if each_u.OpenId != "" {
						r_map[c][each_u.OpenId] = um[each_u.UserId]
						break
					}
				}
			}
		}
	}
	return r_map
}

func MainDataCodesToUserids(codes []string) map[string]map[string]interface{} {
	if len(codes) == 0 {
		return nil
	}
	o := orm.NewOrm()
	m_infos := []MainData{}
	o.QueryTable("main_data").Filter("code__in", codes).All(&m_infos)
	t_codes := []string{}
	for _, each := range m_infos {
		t_codes = append(t_codes, each.TaskCode)
	}
	t_codes = utils.DeleteRepeat(t_codes)
	task_infos := []Task{}
	o.QueryTable("task").Filter("code__in", t_codes).All(&task_infos)
	t_map := map[string]map[string]interface{}{}
	for _, each := range task_infos {
		t_map[each.Code] = map[string]interface{}{
			"userids":    mainDataToUserid(each.Code, m_infos),
			"task_name":  each.Title,
			"task_state": each.State,
			"corpid":     each.CorpId,
		}
	}
	return t_map
}

func round(f float64, n int) float64 {
	floatStr := fmt.Sprintf("%."+strconv.Itoa(n)+"f", f)
	inst, _ := strconv.ParseFloat(floatStr, 64)
	return inst
}
