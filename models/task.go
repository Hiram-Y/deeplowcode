package models

import (
	"DeepWorkload/utils"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	uuid "github.com/satori/go.uuid"
	"strconv"
	"time"
)

type Task struct {
	Code                string            `orm:"column(code);pk"`
	CorpId              string            `orm:"column(corpid)"`
	Title               string            `orm:"column(title)"`
	TypeId              int               `orm:"column(type_id)"`
	TypeContent         string            `orm:"-"`
	Scope               string            `orm:"column(scope);type(json)"`
	ScopeInfo           []FrontDepartInfo `orm:"-"`
	Cooperation         string            `orm:"column(cooperation);type(json)"`
	CooperationInfo     []FrontDepartInfo `orm:"-"`
	StartDate           int64             `orm:"column(start_date)"`
	EndDate             int64             `orm:"column(end_date)"`
	State               int               `orm:"column(state)"` //0 进行中 1未开始 2关闭
	CreateUserId        string            `orm:"column(create_userid)"`
	CreateUserName      string            `orm:"-"`
	WorkflowEnable      bool              `orm:"column(workflow_enable)"`
	Icon                string            `orm:"column(icon);type(json)"`
	Icons               Icons             `orm:"-"`
	CreateTime          time.Time         `orm:"auto_now_add;type(datetime);column(create_time)"`
	WorkflowCode        string            `orm:"column(workflow_code)"`
	FormCode            string            `orm:"column(form_code)"`
	FormCodes           []string          `orm:"-"`
	Remark              string            `orm:"column(remark)"`
	RealScope           string            `orm:"column(real_scope);type(json)"`
	RealCooperation     string            `orm:"column(real_cooperation);type(json)"`
	RealScopeInfo       UserAndDepart     `orm:"-"`
	RealCooperationInfo UserAndDepart     `orm:"-"`
}

func (this *Task) TableName() string {
	return "task"
}

type FrontDepartInfo struct {
	Id         string `json:"id"`
	IsDepart   bool   `json:"IsDepart"`
	DepartId   string `json:"depart_id"`
	Department string `json:"Department"`
}

type Icons struct {
	Icon  string `json:"icon"`
	Index int    `json:"index"`
}

func FrontDepartInfoToReal(vue_depart []FrontDepartInfo, corpid string) UserAndDepart {
	departids := []int{}
	userid := []string{}
	for _, each := range vue_depart {
		if each.IsDepart == true {
			t_id, _ := strconv.Atoi(each.Id)
			departids = append(departids, t_id)
		} else {
			userid = append(userid, each.Id)
		}
	}
	real_partids := GetAllSonDepartmentId(corpid, departids)
	real_partids = append(real_partids, departids...)
	real_strdepartid := []string{}
	for _, each := range real_partids {
		real_strdepartid = append(real_strdepartid, strconv.Itoa(each))
	}
	return UserAndDepart{
		DepartId: real_strdepartid,
		Userid:   userid,
	}
}

func strDatetimeToUnix(dateTime string) int64 {
	layout := "2006-01-02T15:04:05.000Z"
	thistime, _ := time.ParseInLocation(layout, dateTime, time.Local)
	return thistime.Unix()
}

func unixToStrDatetime(dateUnix int64) string {
	layout := "2006-01-02T15:04:05.000Z"
	tm := time.Unix(dateUnix, 0)
	return tm.Format(layout)
}

type BaseInfo struct {
	Code           string            `json:"code"`
	Name           string            `json:"name"`
	DepartList     []FrontDepartInfo `json:"departlist"`
	SynergyList    []FrontDepartInfo `json:"synergylist"`
	Icon           Icons             `json:"icon"`
	TypeId         int               `json:"type_id"`
	DateVal        []int64           `json:"dateval"`
	WorkflowEnable bool              `json:"workflow_enable"`
	Remark         string            `json:"remark"`
}

func (this *BaseInfo) BaseInfoToTaskInfo(corpid, userid, workflow_code string, form_codes []string) (task Task) {
	task.Code = this.Code
	task.CorpId = corpid
	task.CreateUserId = userid
	task.Title = this.Name
	task.ScopeInfo = this.DepartList
	sc, _ := json.Marshal(task.ScopeInfo)
	task.Scope = string(sc)
	task.RealScopeInfo = FrontDepartInfoToReal(task.ScopeInfo, task.CorpId)
	rsc, _ := json.Marshal(task.RealScopeInfo)
	task.RealScope = string(rsc)
	icon_json, _ := json.Marshal(this.Icon)
	task.Icon = string(icon_json)
	if len(this.SynergyList) > 0 {
		task.CooperationInfo = this.SynergyList
		ct, _ := json.Marshal(task.CooperationInfo)
		task.Cooperation = string(ct)
		task.RealCooperationInfo = FrontDepartInfoToReal(task.CooperationInfo, task.CorpId)
		rct, _ := json.Marshal(task.RealCooperationInfo)
		task.RealCooperation = string(rct)
	}

	task.WorkflowCode = workflow_code
	if workflow_code != "" {
		task.WorkflowEnable = true
	}
	task.FormCodes = form_codes
	task.FormCode = utils.StringArrayToStr(form_codes)
	task.StartDate = this.DateVal[0] / 1000
	task.EndDate = this.DateVal[1] / 1000
	task.TypeId = this.TypeId
	task.State = 1
	//if task.StartDate < time.Now().Unix(){
	//	task.State = 0
	//}
	task.Remark = this.Remark
	return task
}

func AddFormsByJson(jsonStr string, corpid string, create_userid string) []string {
	o := orm.NewOrm()
	form_infos := []JsonForm{}
	err := json.Unmarshal([]byte(jsonStr), &form_infos)
	form_codes := []string{}
	for _, each := range form_infos {
		if each.FormCode != "" {
			form_codes = append(form_codes, each.FormCode)
		}
	}
	if len(form_codes) > 0 {
		o.QueryTable("form_info").Filter("code__in", form_codes).Delete()
		o.QueryTable("form_field").Filter("form_code__in", form_codes).Delete()
	}

	if err != nil {
		fmt.Println(err.Error())
	}
	for _, each := range form_infos {
		form_code := uuid.NewV4().String()
		if each.FormCode != "" {
			form_code = each.FormCode
		}
		form := FormInfo{
			Code:          form_code,
			CorpId:        corpid,
			Title:         each.Title,
			TypeId:        each.TypeId,
			Remark:        each.Remark,
			CreateUserid:  create_userid,
			LabelPosition: each.LabelPosition,
			LabelWidth:    each.LabelWidth,
		}
		o.Insert(&form)
		temp_form_code := ""
		if each.IsTemplate == true {
			tf := formToTempForm(form)
			temp_form_code = tf.Code
			go o.Insert(&tf)
		}
		if err != nil {
			fmt.Println(err.Error())
		}
		for _, each_f := range each.Fields {
			f_conf := FormFieldConf{
				Maxlength:  each_f.Maxlength,
				LabelWidth: each_f.LabelWidth,
				AllowHalf:  each_f.AllowHalf,
				Max:        each_f.Max,
				Layout:     each_f.Layout,
				Options:    each_f.Options,
			}
			conf_str, _ := json.Marshal(f_conf)

			jf := FormField{
				FormCode:      form_code,
				CmpType:       each_f.CmpType,
				Label:         each_f.Label,
				Tag:           each_f.Tag,
				TagIcon:       each_f.TagIcon,
				Placeholder:   each_f.Placeholder,
				Clearable:     each_f.Clearable,
				Required:      each_f.Required,
				Idx:           each_f.Idx,
				RenderKey:     each_f.RenderKey,
				JoinGather:    each_f.JoinGather,
				IsCounted:     each_f.IsCounted,
				EnableGreater: each_f.EnableGreater,
				EnableAssign:  each_f.EnableAssign,
				DecimalPoint:  each_f.DecimalPoint,
				OutcomeState:  each_f.OutcomeState,
				StrLists:      each_f.StrIdList,
				StrList:       utils.StringArrayToStr(each_f.StrList),
				StrIdLists:    each_f.StrIdList,
				StrIdList:     utils.StringArrayToStr(each_f.StrIdList),
				FiledCode:     utils.GetLetterByIdx(each_f.Idx),
				Conf:          string(conf_str),
			}
			_, err := o.Insert(&jf)
			if each.IsTemplate == true {
				tjf := formfiledToTemplateFormField(jf)
				tjf.FormCode = temp_form_code
				go o.Insert(&tjf)
			}
			if err != nil {
				fmt.Println(err.Error())
			}

			//if is_edit{
			//	go func(jfd FormField) {
			//		m_infos := []MainData{}
			//		o.QueryTable("main_data").Filter("form_code",form_code).All(&m_infos)
			//		for _,each_m := range m_infos{
			//			json.Unmarshal([]byte(each_m.FormFieldInfo),&each_m.FormFieldInfos)
			//			ori := each_m.FormFieldInfo
			//			for e_index,each_ffi := range each_m.FormFieldInfos{
			//				if each_ffi.FieldCode == jfd.FiledCode{
			//					each_ffi.EnableAssign = jfd.EnableAssign
			//					each_ffi.JoinGather = jfd.JoinGather
			//					each_m.FormFieldInfos[e_index] = each_ffi
			//				}
			//			}
			//			fi,_ := json.Marshal(each_m.FormFieldInfos)
			//			if ori != string(fi){
			//				each_m.FormFieldInfo = string(fi)
			//				o.Update(&each_m,"form_field_info")
			//			}
			//		}
			//	}(jf)
			//}
		}
		if each.FormCode == "" {
			form_codes = append(form_codes, form_code)
		}
	}
	return form_codes
}

func formToTempForm(form FormInfo) TemplateFormInfo {
	return TemplateFormInfo{
		Code:          uuid.NewV4().String(),
		CorpId:        form.CorpId,
		Title:         form.Title,
		Remark:        form.Remark,
		CreateUserid:  form.CreateUserid,
		LabelPosition: form.LabelPosition,
		LabelWidth:    form.LabelWidth,
		TypeId:        form.TypeId,
	}
}

func formfiledToTemplateFormField(form_filed FormField) TemplateFormField {
	return TemplateFormField{
		FormCode:      form_filed.FormCode,
		CmpType:       form_filed.CmpType,
		Label:         form_filed.Label,
		Tag:           form_filed.Tag,
		TagIcon:       form_filed.TagIcon,
		Placeholder:   form_filed.Placeholder,
		Clearable:     form_filed.Clearable,
		Required:      form_filed.Required,
		Idx:           form_filed.Idx,
		RenderKey:     form_filed.RenderKey,
		JoinGather:    form_filed.JoinGather,
		IsCounted:     form_filed.IsCounted,
		EnableGreater: form_filed.EnableGreater,
		EnableAssign:  form_filed.EnableAssign,
		DecimalPoint:  form_filed.DecimalPoint,
		OutcomeState:  form_filed.OutcomeState,
		StrLists:      form_filed.StrLists,
		StrList:       form_filed.StrList,
		StrIdLists:    form_filed.StrIdLists,
		StrIdList:     form_filed.StrIdList,
		FiledCode:     form_filed.FiledCode,
		Conf:          form_filed.Conf,
	}
}

func BaseJsonToDB(base_info, corpid, userid, workflow_code string, form_codes []string) (string, error) {
	bs := BaseInfo{}
	err := json.Unmarshal([]byte(base_info), &bs)
	if err != nil {
		fmt.Println(err.Error())
	}
	task := bs.BaseInfoToTaskInfo(corpid, userid, workflow_code, form_codes)
	if task.Code != "" {
		old_task := Task{}
		sql := fmt.Sprintf("select * from task where code = '%s' ", task.Code)
		o := orm.NewOrm()
		o.Raw(sql).QueryRow(&old_task)
		task.DelTaskWithOutDelMainData()
		task.CreateUserId = old_task.CreateUserId
		task.State = old_task.State
		old_task.FormCodes = utils.StringValueToStrArray(old_task.FormCode)
		diffs := []string{}
		for _, each := range old_task.FormCodes {
			is_in := false
			for _, each_f := range form_codes {
				if each == each_f {
					is_in = true
					break
				}
			}
			if is_in == false {
				diffs = append(diffs, each)
			}
		}
		if len(diffs) > 0 {
			o.QueryTable("main_data").Filter("form_code__in", diffs).Delete()
		}
	}
	code, err := task.InsertTaskInfo()
	if IsNoLastInsertIdError(err) {
		return code, nil
	}
	return "", err
}

func (this *Task) DelTaskWithOutDelMainData() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("task").Filter("code", this.Code).Delete()
	return err
}

func (this *Task) DelTask() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("task").Filter("code", this.Code).Delete()
	if err == nil {
		go o.QueryTable("main_data").Filter("task_code", this.Code).Delete()
	}
	return err
}

func (this *Task) InsertTaskInfo() (string, error) {
	o := orm.NewOrm()
	if this.Code == "" {
		this.Code = uuid.NewV4().String()
	}
	_, err := o.Insert(this)
	return this.Code, err
}

func (this *Task) TaskInfoToBaseInfo() (base_info BaseInfo) {
	base_info.Code = this.Code
	base_info.Name = this.Title
	base_info.TypeId = this.TypeId
	json.Unmarshal([]byte(this.Icon), &this.Icons)
	base_info.Icon = this.Icons
	if this.Cooperation != "" {
		json.Unmarshal([]byte(this.Cooperation), &this.CooperationInfo)
		base_info.SynergyList = this.CooperationInfo
	}

	json.Unmarshal([]byte(this.Scope), &this.ScopeInfo)
	base_info.DepartList = this.ScopeInfo
	base_info.WorkflowEnable = this.WorkflowEnable
	base_info.Remark = this.Remark
	base_info.DateVal = []int64{this.StartDate, this.EndDate}
	return base_info
}

func TaskInfoByTaskCode(code string) Task {
	o := orm.NewOrm()
	t_info := Task{}
	o.QueryTable("task").Filter("code", code).One(&t_info)
	json.Unmarshal([]byte(t_info.RealScope), &t_info.RealScopeInfo)
	return t_info
}

type TaskWithDenyCount struct {
	Task
	DenyCount int
}

func formatDenyCount(tasks []Task, userid string) []TaskWithDenyCount {
	o := orm.NewOrm()
	task_codes := []string{}
	for _, each := range tasks {
		task_codes = append(task_codes, each.Code)
	}
	if len(task_codes) == 0 {
		return nil
	}
	m_info := []MainData{}
	twd := []TaskWithDenyCount{}
	o.QueryTable("main_data").Filter("task_code__in", task_codes).Filter("create_userid", userid).Filter("state", 2).All(&m_info)
	for _, each := range tasks {
		count := 0
		for _, each_m := range m_info {
			if each.Code == each_m.TaskCode {
				count += 1
			}
		}
		twd = append(twd, TaskWithDenyCount{each, count})
	}
	return twd
}

func formatFullTask(infos []Task) []Task {
	userids := []string{}
	if len(infos) == 0 {
		return []Task{}
	}
	for index, each := range infos {
		userids = append(userids, each.CreateUserId)
		json.Unmarshal([]byte(each.RealScope), &each.RealScopeInfo)
		json.Unmarshal([]byte(each.Cooperation), &each.CooperationInfo)
		json.Unmarshal([]byte(each.Scope), &each.ScopeInfo)
		json.Unmarshal([]byte(each.Icon), &each.Icons)
		if each.FormCode != "" {
			each.FormCodes = utils.StringValueToStrArray(each.FormCode)
		}
		infos[index] = each
	}
	t := TypeInfoById(infos[0].CorpId, "", "任务类型")
	userids = utils.RemoveRepByMap(userids)
	uMap := GetUserNameMapByUseridAndCorpId(userids, infos[0].CorpId)
	for index, each := range infos {
		each.CreateUserName = uMap[each.CreateUserId]
		each.TypeContent = t[each.TypeId].Content
		if each.TypeId == 2 {
			each.TypeContent = "其他"
		}
		infos[index] = each
	}
	return infos
}

func (this *Task) AllSelfCreateTaskInfoSearch(title string) (err error, infos []Task) {
	o := orm.NewOrm()
	_, err = o.QueryTable("task").Filter("corpid", this.CorpId).
		Filter("create_userid", this.CreateUserId).Filter("title__icontains", title).OrderBy("-create_time").All(&infos)
	infos = formatFullTask(infos)
	return err, infos
}

func (this *Task) AllAuthorizeTaskInfoSearch(title string) (err error, t_infos []Task) {
	o := orm.NewOrm()
	u_info := UserInfo{}
	err = o.QueryTable("user_info").Filter("userid", this.CreateUserId).
		Filter("corpid", this.CorpId).One(&u_info)
	if u_info.UserId == "" {
		return err, []Task{}
	}
	u_departids := utils.StringValueToIntArray(u_info.DepartId)
	sql := fmt.Sprintf("select * FROM task  WHERE  corpid = '%s' and "+
		"( real_cooperation :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserId)
	for _, each := range u_departids {
		sql += fmt.Sprintf(" or real_cooperation :: jsonb -> 'departid' # '%s'", fmt.Sprint(each))
	}
	sql += fmt.Sprintf(" ) ")
	sql += fmt.Sprintf(" and title like '%%%s%%'", title)
	sql += fmt.Sprintf(" order by create_time desc")
	_, err = o.Raw(sql).QueryRows(&t_infos)
	t_infos = formatFullTask(t_infos)
	return err, t_infos
}

func (this *Task) AllSelfCreateTaskInfoClosed() (err error, infos map[int]interface{}) {
	o := orm.NewOrm()
	t_infos := []Task{}
	typeMap := TypeInfoById(this.CorpId, this.CreateUserId, "任务类型")
	type_ids := []int{}
	for k, _ := range typeMap {
		type_ids = append(type_ids, k)
	}
	type_ids = append(type_ids, 2)
	sql := ""
	limit := 10
	for index, each := range type_ids {
		if index == 0 {
			sql = fmt.Sprintf("( select * from task where corpid = '%s' and create_userid = '%s'  "+
				"and type_id = %d and state = 2  order by create_time desc LIMIT %d )", this.CorpId, this.CreateUserId, each, limit)
		} else {
			sql += " UNION ALL "
			sql += fmt.Sprintf(" ( select * from task where corpid = '%s' and create_userid = '%s'  "+
				"and type_id = %d and state = 2 order by create_time desc LIMIT %d )", this.CorpId, this.CreateUserId, each, limit)
		}
	}
	_, err = o.Raw(sql).QueryRows(&t_infos)
	t_infos = formatFullTask(t_infos)
	mapTask := formatTask(t_infos, typeMap)
	for k, v := range typeMap {
		if mapTask[k] == nil {
			mapTask[k] = v.Content
		}
	}
	if mapTask[2] == nil {
		mapTask[2] = "其他"
	}
	return err, mapTask
}

func (this *Task) AllSelfCreateTaskInfoOnGoing() (err error, infos map[int]interface{}) {
	o := orm.NewOrm()
	t_infos := []Task{}
	typeMap := TypeInfoById(this.CorpId, this.CreateUserId, "任务类型")
	type_ids := []int{}
	for k, _ := range typeMap {
		type_ids = append(type_ids, k)
	}
	type_ids = append(type_ids, 2)
	sql := ""
	limit := 10
	for index, each := range type_ids {
		if index == 0 {
			sql = fmt.Sprintf("( select * from task where corpid = '%s' and create_userid = '%s'  "+
				"and type_id = %d and (state = 1 or state = 0)  order by create_time desc LIMIT %d )", this.CorpId, this.CreateUserId, each, limit)
		} else {
			sql += " UNION ALL "
			sql += fmt.Sprintf(" ( select * from task where corpid = '%s' and create_userid = '%s'  "+
				"and type_id = %d and (state = 1 or state = 0) order by create_time desc LIMIT %d )", this.CorpId, this.CreateUserId, each, limit)
		}
	}
	_, err = o.Raw(sql).QueryRows(&t_infos)
	t_infos = formatFullTask(t_infos)
	mapTask := formatTask(t_infos, typeMap)
	for k, v := range typeMap {
		if mapTask[k] == nil {
			mapTask[k] = v.Content
		}
	}
	if mapTask[2] == nil {
		mapTask[2] = "其他"
	}
	return err, mapTask
}

func GetAllTaskByCorpId(corpid string) (error, []Task) {
	o := orm.NewOrm()
	t_info := []Task{}
	_, err := o.QueryTable("task").Filter("corpid", corpid).All(&t_info)
	//t_info = formatFullTask(t_info)
	return err, t_info
}

func (this *Task) GetAllTaskByTypeID(page_size, page_index int) (error, []Task, int64) {
	o := orm.NewOrm()
	t_infos := []Task{}
	count, err := o.QueryTable("task").Filter("create_userid", this.CreateUserId).
		Filter("corpid", this.CorpId).Filter("type_id", this.TypeId).Count()
	_, err = o.QueryTable("task").Filter("create_userid", this.CreateUserId).
		Filter("corpid", this.CorpId).Filter("type_id", this.TypeId).OrderBy("-create_time").Limit(page_size).Offset((page_index - 1) * page_size).All(&t_infos)
	t_infos = formatFullTask(t_infos)
	return err, t_infos, count
}

func (this *Task) GetAllAuthorizeTaskInfoByTypeID(page_size, page_index int) (error, []Task, int) {
	o := orm.NewOrm()
	u_info := UserInfo{}
	err := o.QueryTable("user_info").Filter("userid", this.CreateUserId).
		Filter("corpid", this.CorpId).One(&u_info)
	if u_info.UserId == "" {
		return err, []Task{}, 0
	}
	u_departids := utils.StringValueToIntArray(u_info.DepartId)

	sql := fmt.Sprintf(" select * FROM task  WHERE  corpid = '%s' and "+
		"( real_cooperation :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserId)
	for _, each := range u_departids {
		sql += fmt.Sprintf(" or real_cooperation :: jsonb -> 'departid' # '%s'", fmt.Sprint(each))
	}
	sql += fmt.Sprintf(" ) and type_id = %d", this.TypeId)
	sql += fmt.Sprintf(" order by state, create_time desc ")
	if page_size != 0 {
		sql += fmt.Sprintf(" limit %d offset %d", page_size, (page_index-1)*page_size)
	}
	t_infos := []Task{}
	_, err = o.Raw(sql).QueryRows(&t_infos)
	t_infos = formatFullTask(t_infos)
	count_sql := fmt.Sprintf(" select count(*) as count FROM task  WHERE  corpid = '%s' and "+
		"( real_cooperation :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserId)
	for _, each := range u_departids {
		count_sql += fmt.Sprintf(" or real_cooperation :: jsonb -> 'departid' # '%s'", fmt.Sprint(each))
	}
	count_sql += fmt.Sprintf(" ) and type_id = %d ", this.TypeId)
	c := Counts{}
	o.Raw(count_sql).QueryRow(&c)
	return err, t_infos, c.Count
}

func (this *Task) AllAuthorizeTaskInfoByUserID() (err error, infos map[int]interface{}) {
	o := orm.NewOrm()
	u_info := UserInfo{}
	o.QueryTable("user_info").Filter("userid", this.CreateUserId).
		Filter("corpid", this.CorpId).One(&u_info)
	if u_info.UserId == "" {
		return nil, nil
	}
	typeMap := TypeInfoById(this.CorpId, "", "任务类型")
	type_ids := []int{}
	for k, _ := range typeMap {
		type_ids = append(type_ids, k)
	}
	type_ids = append(type_ids, 2)
	limit := 10
	sql := ""
	u_departids := utils.StringValueToIntArray(u_info.DepartId)
	for index, each := range type_ids {
		if index == 0 {
			sql = fmt.Sprintf("( select * FROM task  WHERE  corpid = '%s' and "+
				"( real_cooperation :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserId)
			for _, each := range u_departids {
				sql += fmt.Sprintf(" or real_cooperation :: jsonb -> 'departid' # '%s'", fmt.Sprint(each))
			}
			sql += fmt.Sprintf(" ) ")
			sql += fmt.Sprintf(" and type_id = %d", each)

			sql += fmt.Sprintf(" order by state, create_time desc limit %d )", limit)
		} else {
			sql += fmt.Sprintf(" UNION ALL ")
			sql += fmt.Sprintf("( select * FROM task  WHERE  corpid = '%s' and "+
				"( real_cooperation :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserId)
			for _, each := range u_departids {
				sql += fmt.Sprintf(" or real_cooperation :: jsonb -> 'departid' # '%s'", fmt.Sprint(each))
			}
			sql += fmt.Sprintf(" ) ")
			sql += fmt.Sprintf(" and type_id = %d", each)

			sql += fmt.Sprintf(" order by state, create_time desc limit %d )", limit)
		}
	}

	task_infos := []Task{}
	_, err = o.Raw(sql).QueryRows(&task_infos)
	task_infos = formatFullTask(task_infos)
	mapTask := formatTask(task_infos, typeMap)
	return err, mapTask
}

func formatTask(t_infos interface{}, typeMap map[int]TypeInfo) map[int]interface{} {
	switch t_infos.(type) {
	case []Task:
		task_infos := t_infos.([]Task)
		if len(task_infos) == 0 {
			return map[int]interface{}{}
		}
		for index, each := range task_infos {
			if each.TypeId == 2 {
				each.TypeContent = "其他"
			} else {
				each.TypeContent = typeMap[each.TypeId].Content
			}
			task_infos[index] = each
		}
		mapTask := map[int]interface{}{}
		for _, each := range task_infos {
			if mapTask[each.TypeId] != nil {
				temp := mapTask[each.TypeId].([]interface{})
				cp := make([]interface{}, len(temp))
				copy(cp, temp)
				cp = append(cp, each)
				mapTask[each.TypeId] = cp
			} else {
				mapTask[each.TypeId] = []interface{}{each}
			}
		}
		return mapTask
	case []TaskWithDenyCount:
		task_infos := t_infos.([]TaskWithDenyCount)
		if len(task_infos) == 0 {
			return map[int]interface{}{}
		}
		for index, each := range task_infos {
			if each.TypeId == 2 {
				each.TypeContent = "其他"
			} else {
				each.TypeContent = typeMap[each.TypeId].Content
			}
			task_infos[index] = each
		}
		mapTask := map[int]interface{}{}
		for _, each := range task_infos {
			if mapTask[each.TypeId] != nil {
				temp := mapTask[each.TypeId].([]interface{})
				cp := make([]interface{}, len(temp))
				copy(cp, temp)
				cp = append(cp, each)
				mapTask[each.TypeId] = cp
			} else {
				mapTask[each.TypeId] = []interface{}{each}
			}
		}
		return mapTask
	}

	return nil
}

func (this *Task) UpdateTaskState() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("task").Filter("code", this.Code).Update(orm.Params{"state": this.State})
	return err
}

func (this *Task) UpdateTaskInfo() error {
	o := orm.NewOrm()
	pa := orm.Params{}
	t_info := Task{}
	o.QueryTable("task").Filter("code", this.Code).One(&t_info)
	if this.TypeId != 0 {
		pa["type_id"] = this.TypeId
	}
	if this.Scope != "" {
		pa["scope"] = this.Scope
		json.Unmarshal([]byte(this.Scope), &this.ScopeInfo)
		this.RealScopeInfo = FrontDepartInfoToReal(this.ScopeInfo, t_info.CorpId)
		real_scope, _ := json.Marshal(this.RealScopeInfo)
		pa["real_scope"] = string(real_scope)
	}
	if this.Title != "" {
		pa["title"] = this.Title
	}
	if this.Cooperation != "" {
		pa["cooperation"] = this.Cooperation
		json.Unmarshal([]byte(this.Cooperation), &this.CooperationInfo)
		this.RealCooperationInfo = FrontDepartInfoToReal(this.CooperationInfo, t_info.CorpId)
		real_corp, _ := json.Marshal(this.RealCooperationInfo)
		pa["real_cooperation"] = string(real_corp)
	}
	_, err := o.QueryTable("task").Filter("code", this.Code).Update(pa)
	return err
}

func (this *Task) GetTaskFullInfoDetailToJson() (error, BaseInfo, []JsonForm, WorkflowJson) {
	o := orm.NewOrm()
	task := Task{}
	err := o.QueryTable("task").Filter("code", this.Code).One(&task)
	base_info := task.TaskInfoToBaseInfo()
	jsf := []JsonForm{}
	if task.Code == "" {
		return nil, BaseInfo{}, nil, WorkflowJson{}
	}

	this.FormCodes = utils.StringValueToStrArray(task.FormCode)
	if len(this.FormCodes) > 0 {
		form_infos := []FormInfo{}
		_, err = o.QueryTable("form_info").Filter("code__in", this.FormCodes).All(&form_infos)
		form_fileds := []FormField{}
		_, err = o.QueryTable("form_field").Filter("form_code__in", this.FormCodes).OrderBy("id").All(&form_fileds)
		for _, each := range form_infos {
			fields := []JsonFormField{}
			for _, each_f := range form_fileds {
				if each_f.FormCode == each.Code {
					f_conf := FormFieldConf{}
					json.Unmarshal([]byte(each_f.Conf), &f_conf)

					fields = append(fields, JsonFormField{
						CmpType:       each_f.CmpType,
						Label:         each_f.Label,
						Tag:           each_f.Tag,
						TagIcon:       each_f.TagIcon,
						Placeholder:   each_f.Placeholder,
						Clearable:     each_f.Clearable,
						Maxlength:     f_conf.Maxlength,
						Required:      each_f.Required,
						LabelWidth:    f_conf.LabelWidth,
						Idx:           each_f.Idx,
						RenderKey:     each_f.RenderKey,
						JoinGather:    each_f.JoinGather,
						IsCounted:     each_f.IsCounted,
						EnableGreater: each_f.EnableGreater,
						DecimalPoint:  each_f.DecimalPoint,
						OutcomeState:  each_f.OutcomeState,
						StrList:       utils.StringValueToStrArray(each_f.StrList),
						StrIdList:     utils.StringValueToStrArray(each_f.StrIdList),
						EnableAssign:  each_f.EnableAssign,
						AllowHalf:     f_conf.AllowHalf,
						Max:           f_conf.Max,
						FiledCode:     each_f.FiledCode,
						Layout:        f_conf.Layout,
						Options:       f_conf.Options,
					})
				}
			}
			jf := JsonForm{
				FormCode:      each.Code,
				Title:         each.Title,
				Remark:        each.Remark,
				TypeId:        each.TypeId,
				IsTemplate:    false,
				Fields:        fields,
				LabelWidth:    each.LabelWidth,
				LabelPosition: each.LabelPosition,
			}
			jsf = append(jsf, jf)
		}
	}
	wfj := WorkflowJson{}
	if task.WorkflowEnable == true {
		wf := Workflow{}
		err = o.QueryTable("workflow").Filter("workflow_code", task.WorkflowCode).One(&wf)
		wfj.WorkflowDef.WorkflowCode = wf.WorkflowCode
		wfj.WorkflowDef.Name = wf.Name
		wfj.WorkflowDef.Type = wf.TypeId
		json.Unmarshal([]byte(wf.FlowPermission), &wfj.FlowPermissions)
		nc := NodeConfig{WorkflowCode: task.WorkflowCode}
		err, tree := nc.GetNodeTree()
		if err != nil {
			fmt.Println("get node tree error", err.Error())
		}
		wfj.NodeConfig = tree
	}
	if err != nil {
		fmt.Println(err.Error())
	}
	return err, base_info, jsf, wfj
}

func (this *Task) GetFormListAndCountByAuditUserId() (error, interface{}) {
	o := orm.NewOrm()
	type selfForm struct {
		FormInfo
		DenyCount   int
		AllCount    int
		AuditCount  int
		CommitCount int
		PassCount   int
	}

	tk := Task{}
	err := o.QueryTable("task").Filter("code", this.Code).One(&tk)
	if tk.FormCode == "" {
		return nil, nil
	}
	tk.FormCodes = utils.StringValueToStrArray(tk.FormCode)
	if err != nil {
		fmt.Println(err.Error())
	}

	f_infos := []FormInfo{}
	_, err = o.QueryTable("form_info").Filter("code__in", tk.FormCodes).All(&f_infos)
	u_info := UserInfo{
		CorpId: this.CorpId,
		UserId: this.CreateUserId,
	}
	u_ids := u_info.GetAllUserInfoByUserId()
	m_infos := []MainData{}
	qs := o.QueryTable("main_data").Filter("form_code__in", tk.FormCodes).Filter("create_userid__in", u_ids).Filter("corpid", this.CorpId)
	_, err = qs.All(&m_infos)
	twd := []selfForm{}
	for _, each := range f_infos {
		deny_count := 0
		all_count := 0
		commit_count := 0
		pass_count := 0
		au_count := 0
		for _, each_m := range m_infos {
			if each.Code == each_m.FormCode {
				all_count += 1
				switch each_m.State {
				case 1:
					commit_count += 1
				case 2:
					deny_count += 1
				case 3:
					pass_count += 1
				case 4:
					au_count += 1
				case 5:
					pass_count += 1
				}
			}
		}
		twd = append(twd, selfForm{
			each,
			deny_count,
			all_count,
			au_count,
			commit_count,
			pass_count},
		)
	}

	return err, twd
}

func (this *Task) GetFormListAndCount() (error, interface{}) {
	o := orm.NewOrm()
	type selfForm struct {
		FormInfo
		DenyCount   int
		AllCount    int
		AuditCount  int
		CommitCount int
		PassCount   int
	}

	tk := Task{}
	err := o.QueryTable("task").Filter("code", this.Code).One(&tk)
	if tk.FormCode == "" {
		return nil, nil
	}
	tk.FormCodes = utils.StringValueToStrArray(tk.FormCode)
	if err != nil {
		fmt.Println(err.Error())
	}

	f_infos := []FormInfo{}
	_, err = o.QueryTable("form_info").Filter("code__in", tk.FormCodes).All(&f_infos)
	m_infos := []MainData{}
	qs := o.QueryTable("main_data").Filter("form_code__in", tk.FormCodes)
	if this.CreateUserId != "" {
		qs = qs.Filter("create_userid", this.CreateUserId)
	}
	_, err = qs.All(&m_infos)
	twd := []selfForm{}
	for _, each := range f_infos {
		deny_count := 0
		all_count := 0
		commit_count := 0
		pass_count := 0
		au_count := 0
		for _, each_m := range m_infos {
			if each.Code == each_m.FormCode {
				all_count += 1
				switch each_m.State {
				case 1:
					commit_count += 1
				case 2:
					deny_count += 1
				case 3:
					pass_count += 1
				case 4:
					au_count += 1
				case 5:
					pass_count += 1
				}
			}
		}
		twd = append(twd, selfForm{
			each,
			deny_count,
			all_count,
			au_count,
			commit_count,
			pass_count},
		)
	}

	return err, twd
}

func (this *Task) GetFormList() (error, []FormInfo) {
	o := orm.NewOrm()
	tk := Task{}
	err := o.QueryTable("task").Filter("code", this.Code).One(&tk)
	if tk.FormCode == "" {
		return nil, []FormInfo{}
	}
	tk.FormCodes = utils.StringValueToStrArray(tk.FormCode)
	if err != nil {
		fmt.Println(err.Error())
	}

	f_infos := []FormInfo{}
	_, err = o.QueryTable("form_info").Filter("code__in", tk.FormCodes).All(&f_infos)
	return err, f_infos
}

func (this *Task) GetReportedTaskByUserID() (err error, infos map[int]interface{}) {
	o := orm.NewOrm()
	u_info := UserInfo{}
	o.QueryTable("user_info").Filter("userid", this.CreateUserId).
		Filter("corpid", this.CorpId).One(&u_info)
	u_departids := utils.StringValueToIntArray(u_info.DepartId)
	typeMap := TypeInfoById(this.CorpId, "", "任务类型")
	type_ids := []int{}
	for k, _ := range typeMap {
		type_ids = append(type_ids, k)
	}
	type_ids = append(type_ids, 2)
	limit := 10
	sql := ""
	for index, each := range type_ids {
		if index == 0 {
			fmt.Println(index)
			sql = fmt.Sprintf("( select * FROM task  WHERE  corpid = '%s' and "+
				"( real_scope :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserId)
			for _, each := range u_departids {
				sql += fmt.Sprintf(" or real_scope :: jsonb -> 'departid' # '%s'", fmt.Sprint(each))
			}
			sql += fmt.Sprintf(" ) ")
			sql += fmt.Sprintf(" and type_id = %d and state = 2", each)

			sql += fmt.Sprintf(" order by state, create_time desc limit %d )", limit)
		} else {

			sql += fmt.Sprintf(" UNION ALL ")
			sql += fmt.Sprintf("( select * FROM task  WHERE  corpid = '%s' and "+
				"( real_scope :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserId)
			for _, each := range u_departids {
				sql += fmt.Sprintf(" or real_scope :: jsonb -> 'departid' # '%s'", fmt.Sprint(each))
			}
			sql += fmt.Sprintf(" ) ")
			sql += fmt.Sprintf(" and type_id = %d  and state =2", each)

			sql += fmt.Sprintf(" order by state, create_time desc limit %d )", limit)
		}
	}

	task_infos := []Task{}
	_, err = o.Raw(sql).QueryRows(&task_infos)
	task_infos = formatFullTask(task_infos)
	mapTask := formatTask(task_infos, typeMap)
	return err, mapTask
}

func (this *Task) GetReportedTaskByTypeId(page_size, page_index int) (error, []Task, int) {
	o := orm.NewOrm()
	u_info := UserInfo{}
	err := o.QueryTable("user_info").Filter("userid", this.CreateUserId).
		Filter("corpid", this.CorpId).One(&u_info)
	if u_info.UserId == "" {
		return err, []Task{}, 0
	}
	u_departids := utils.StringValueToIntArray(u_info.DepartId)

	sql := fmt.Sprintf(" select * FROM task  WHERE  corpid = '%s' and "+
		"( real_scope :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserId)
	for _, each := range u_departids {
		sql += fmt.Sprintf(" or real_scope :: jsonb -> 'departid' # '%s'", fmt.Sprint(each))
	}
	sql += fmt.Sprintf(" ) and type_id = %d and state = 2", this.TypeId)
	sql += fmt.Sprintf(" order by state, create_time desc ")
	if page_size != 0 {
		sql += fmt.Sprintf("limit %d offset %d", page_size, (page_index-1)*page_size)
	}
	t_infos := []Task{}
	_, err = o.Raw(sql).QueryRows(&t_infos)
	t_infos = formatFullTask(t_infos)
	count_sql := fmt.Sprintf(" select count(*) as count FROM task  WHERE  corpid = '%s' and "+
		"( real_scope :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserId)
	for _, each := range u_departids {
		count_sql += fmt.Sprintf(" or real_scope :: jsonb -> 'departid' # '%s'", fmt.Sprint(each))
	}
	count_sql += fmt.Sprintf(" ) and type_id = %d and state = 2", this.TypeId)
	c := Counts{}
	o.Raw(count_sql).QueryRow(&c)
	return err, t_infos, c.Count
}

//func (this *Task) GetPubTaskList(page_size,page_index int)  (error , []Task, int)  {
//	o := orm.NewOrm()
//	u_info := UserInfo{}
//	o.QueryTable("user_info").Filter("userid",this.CreateUserId).
//		Filter("corpid",this.CorpId).One(&u_info)
//	u_departids := utils.StringValueToIntArray(u_info.DepartId)
//	sql := fmt.Sprintf("select * FROM task  WHERE  corpid = '%s' and " +
//		"( real_scope :: jsonb -> 'userid' # '%s'",this.CorpId,this.CreateUserId)
//	for _,each := range u_departids{
//		sql += fmt.Sprintf(" or real_scope :: jsonb -> 'departid' # '%s'",fmt.Sprint(each))
//	}
//	sql += fmt.Sprintf(" )")
//	sql += fmt.Sprintf(" and state = 0  order by state, create_time desc")
//
//
//}

func (this *Task) GetToReportTaskByUserid() (error, map[int]interface{}) {
	o := orm.NewOrm()
	u_info := UserInfo{}
	o.QueryTable("user_info").Filter("userid", this.CreateUserId).
		Filter("corpid", this.CorpId).One(&u_info)
	u_departids := utils.StringValueToIntArray(u_info.DepartId)
	sql := fmt.Sprintf("select * FROM task  WHERE  corpid = '%s' and "+
		"( real_scope :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserId)
	for _, each := range u_departids {
		sql += fmt.Sprintf(" or real_scope :: jsonb -> 'departid' # '%s'", fmt.Sprint(each))
	}
	sql += fmt.Sprintf(" )")

	sql += fmt.Sprintf(" and state = 0  order by state, create_time desc")
	task_infos := []Task{}
	_, err := o.Raw(sql).QueryRows(&task_infos)

	task_infos = formatFullTask(task_infos)
	tdc_infos := formatDenyCount(task_infos, this.CreateUserId)
	typeMap := TypeInfoById(this.CorpId, "", "任务类型")
	mapTask := formatTask(tdc_infos, typeMap)
	return err, mapTask
}

func CheckUserIsAudit(userid, corpid string) bool {
	o := orm.NewOrm()
	codes := []string{}
	task_infos := []Task{}
	sql := fmt.Sprintf("select workflow_code from task where corpid = '%s' and workflow_enable = true", corpid)
	o.Raw(sql).QueryRows(&task_infos)
	for _, each := range task_infos {
		codes = append(codes, each.WorkflowCode)
	}
	if len(codes) == 0 {
		return false
	}
	nodec := []NodeConfig{}
	o.QueryTable("node_config").Filter("workflow_code__in", codes).Filter("node_type", 1).All(&nodec)
	for _, each := range nodec {
		json.Unmarshal([]byte(each.UserList), &each.NodeUserList)
		for _, each_n := range each.NodeUserList {
			t_userid := fmt.Sprint(each_n.TargetId)
			if t_userid == userid {
				return true
			}
		}
	}
	return false
}

func GetAllTaskByAuditUserId(userid, corpid string, page_size, page_index int) (error, []Task, int) {
	o := orm.NewOrm()
	usF := fmt.Sprintf("{\"targetId\":\"%s\"}", userid)
	sql := fmt.Sprintf("select * from task where  task.workflow_code in (SELECT workflow_code  "+
		"FROM node_config WHERE user_list::jsonb @> '[%s]'::jsonb) and task.workflow_enable = TRUE and corpid = '%s'", usF, corpid)
	if page_size != 0 {
		sql += fmt.Sprintf(" limit %d offset %d", page_size, (page_index-1)*page_size)
	}
	t_infos := []Task{}
	_, err := o.Raw(sql).QueryRows(&t_infos)
	count_sql := fmt.Sprintf("select count(*) as count from task where task.workflow_code in (SELECT workflow_code  "+
		"FROM node_config WHERE user_list::jsonb @> '[%s]'::jsonb) and task.workflow_enable = TRUE and corpid = '%s'", usF, corpid)
	c := Counts{}
	o.Raw(count_sql).QueryRow(&c)
	t_infos = formatFullTask(t_infos)
	return err, t_infos, c.Count
}
