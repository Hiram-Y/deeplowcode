package models

import (
	"DeepWorkload/utils"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	uuid "github.com/satori/go.uuid"
	"time"
)

type TemplateFormInfo struct {
	Code                string            `orm:"column(code);pk"`
	CorpId              string            `orm:"column(corpid)"`
	Title               string            `orm:"column(title)"`
	Remark              string            `orm:"column(remark)"`
	CreateTime          time.Time         `orm:"column(create_time);auto_now;type(datetime)"`
	CreateUserid        string            `orm:"column(create_userid)"`
	CreateUserName      string            `orm:"-"`
	TypeId              int               `orm:"column(type_id)"`
	TypeContent         string            `orm:"-"`
	LabelWidth          int               `orm:"column(label_width)"`
	LabelPosition       string            `orm:"column(label_position)"`
	Cooperation         string            `orm:"column(cooperation);type(json)"`
	CooperationInfo     []FrontDepartInfo `orm:"-"`
	RealCooperation     string            `orm:"column(real_cooperation);type(json)"`
	RealCooperationInfo UserAndDepart     `orm:"-"`
	UpdateTime          time.Time         `orm:"column(update_time);type(datetime)"`
}

type UserAndDepart struct {
	Userid   []string `json:"userid"`
	DepartId []string `json:"departid"`
}

func (this *TemplateFormInfo) TableName() string {
	return "template_form_info"
}

func CopyToUserId(form_code, userid, corpid string) error {
	o := orm.NewOrm()
	tf_info := TemplateFormInfo{}
	err := o.QueryTable("template_form_info").Filter("code", form_code).One(&tf_info)

	tff_infos := []TemplateFormField{}
	_, err = o.QueryTable("template_form_field").Filter("form_code", form_code).OrderBy("id").All(&tff_infos)
	tf_info.Code = uuid.NewV4().String()
	tf_info.CorpId = corpid
	tf_info.TypeId = 1
	tf_info.CreateUserid = userid
	_, err = o.Insert(&tf_info)
	for index, each := range tff_infos {
		each.FormCode = tf_info.Code
		each.Id = 0
		tff_infos[index] = each
	}
	if len(tff_infos) > 0 {
		_, err = o.InsertMulti(len(tff_infos), tff_infos)
	}
	if err != nil {
		return err
	}
	return nil
}

func (this *TemplateFormInfo) GetAllTemplateFormInfoByUserid() (err error, infos map[int]interface{}) {
	o := orm.NewOrm()
	t_info := []TypeInfo{}
	o.QueryTable("type_info").Filter("corpid", this.CorpId).Filter("create_userid", this.CreateUserid).Filter("type_desc", "规则类型").All(&t_info)
	type_ids := []int{}
	for _, each := range t_info {
		type_ids = append(type_ids, each.Id)
	}
	type_ids = append(type_ids, 1)
	sql := ""
	limit := 10
	f_infos := []TemplateFormInfo{}
	for index, each := range type_ids {
		if index == 0 {
			sql = fmt.Sprintf("( select * from template_form_info where corpid = '%s' and create_userid = '%s'  "+
				"and type_id = %d  order by create_time desc LIMIT %d )", this.CorpId, this.CreateUserid, each, limit)
		} else {
			sql += " UNION ALL "
			sql += fmt.Sprintf(" ( select * from template_form_info where corpid = '%s' and create_userid = '%s'  "+
				"and type_id = %d order by create_time desc LIMIT %d )", this.CorpId, this.CreateUserid, each, limit)
		}
	}
	tf_infos := []TemplateFormInfo{}
	_, err = o.Raw(sql).QueryRows(&tf_infos)
	for index, each_t := range tf_infos {
		json.Unmarshal([]byte(each_t.Cooperation), &each_t.CooperationInfo)
		tf_infos[index] = each_t
	}
	f_infos = append(f_infos, tf_infos...)

	typeMap := TypeInfoById(this.CorpId, this.CreateUserid, "规则类型")
	map_infos := formatForm(f_infos, typeMap)
	for k, v := range typeMap {
		if map_infos[k] == nil {
			map_infos[k] = v.Content
		}
	}
	if map_infos[1] == nil {
		map_infos[1] = "其他"
	}

	return err, map_infos
}

func formatForm(f_infos []TemplateFormInfo, typeMap map[int]TypeInfo) map[int]interface{} {
	map_infos := map[int]interface{}{}
	if len(f_infos) == 0 {
		return map[int]interface{}{}
	}
	userids := []string{}
	for _, each := range f_infos {
		userids = append(userids, each.CreateUserid)
	}
	userids = utils.DeleteRepeat(userids)
	o := orm.NewOrm()
	u_infos := []UserInfo{}
	o.QueryTable("user_info").Filter("userid__in", userids).Filter("corpid", f_infos[0].CorpId).All(&u_infos)
	for index, each := range f_infos {
		for _, each_u := range u_infos {
			if each.CreateUserid == each_u.UserId {
				each.CreateUserName = each_u.Name
				break
			}
		}
		if each.TypeId == 1 {
			each.TypeContent = "其他"
		} else {
			each.TypeContent = typeMap[each.TypeId].Content
		}
		f_infos[index] = each
		if map_infos[each.TypeId] != nil {
			sr := map_infos[each.TypeId].([]interface{})
			cp := make([]interface{}, len(sr))
			copy(cp, sr)
			cp = append(cp, each)
			map_infos[each.TypeId] = cp
		} else {
			map_infos[each.TypeId] = []interface{}{each}
		}
	}
	return map_infos
}

func (this *TemplateFormInfo) GetAllTemplateFormInfoByUseridAndTypeId() (error, []TemplateFormInfo) {
	o := orm.NewOrm()
	tfs := []TemplateFormInfo{}
	_, err := o.QueryTable("template_form_info").Filter("corpid", this.CorpId).Filter("type_id", this.TypeId).
		Filter("create_userid", this.CreateUserid).OrderBy("create_time").All(&tfs)
	for index, each_t := range tfs {
		json.Unmarshal([]byte(each_t.Cooperation), &each_t.CooperationInfo)
		tfs[index] = each_t
	}
	return err, tfs
}

func (this *TemplateFormInfo) AllTemplateFormInfoByAuthorize() (err error, infos map[int]interface{}) {
	o := orm.NewOrm()
	t_info := []TypeInfo{}
	o.QueryTable("type_info").Filter("corpid", this.CorpId).Filter("type_desc", "规则类型").All(&t_info)
	type_ids := []int{}
	for _, each := range t_info {
		type_ids = append(type_ids, each.Id)
	}
	type_ids = append(type_ids, 1)
	u_info := UserInfo{}
	o.QueryTable("user_info").Filter("userid", this.CreateUserid).
		Filter("corpid", this.CorpId).One(&u_info)
	if u_info.UserId == "" {
		return nil, nil
	}
	u_departids := utils.StringValueToIntArray(u_info.DepartId)
	sql := ""
	limit := 10
	f_infos := []TemplateFormInfo{}
	for index, each := range type_ids {
		if index == 0 {
			sql = fmt.Sprintf("( select * FROM template_form_info  WHERE  corpid = '%s' and "+
				"( real_cooperation :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserid)
			for _, each_d := range u_departids {
				sql += fmt.Sprintf(" or real_cooperation :: jsonb -> 'departid' # '%s'", fmt.Sprint(each_d))
			}
			sql += fmt.Sprintf(" ) and  type_id = %d order by create_time LIMIT %d )", each, limit)
		} else {
			sql += " UNION ALL "
			sql += fmt.Sprintf("( select * FROM template_form_info  WHERE  corpid = '%s' and "+
				"( real_cooperation :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserid)
			for _, each_d := range u_departids {
				sql += fmt.Sprintf(" or real_cooperation :: jsonb -> 'departid' # '%s'", fmt.Sprint(each_d))
			}
			sql += fmt.Sprintf(" ) and  type_id = %d order by create_time LIMIT %d )", each, limit)
		}
	}
	_, err = o.Raw(sql).QueryRows(&f_infos)
	typeMap := TypeInfoById(this.CorpId, "", "规则类型")
	map_infos := formatForm(f_infos, typeMap)
	return err, map_infos
}

func (this *TemplateFormInfo) GetAllAuthorizeTemplateFormInfoByTypeId() (error, []TemplateFormInfo) {
	o := orm.NewOrm()
	u_info := UserInfo{}
	o.QueryTable("user_info").Filter("userid", this.CreateUserid).
		Filter("corpid", this.CorpId).One(&u_info)
	u_departids := utils.StringValueToIntArray(u_info.DepartId)
	sql := fmt.Sprintf("select * FROM template_form_info  WHERE  corpid = '%s' and "+
		"( real_cooperation :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserid)

	for _, each_d := range u_departids {
		sql += fmt.Sprintf(" or real_cooperation :: jsonb -> 'departid' # '%s'", fmt.Sprint(each_d))
	}
	sql += fmt.Sprintf(" ) and  type_id = %d order   by create_time", this.TypeId)
	f_infos := []TemplateFormInfo{}
	_, err := o.Raw(sql).QueryRows(&f_infos)
	return err, f_infos
}

type JsonForm struct {
	FormCode      string          `json:"form_code"`
	Title         string          `json:"title"`
	Remark        string          `json:"remark"`
	TypeId        int             `json:"type_id"`
	IsTemplate    bool            `json:"is_template"`
	Fields        []JsonFormField `json:"fields"`
	LabelWidth    int             `json:"labelWidth"`
	LabelPosition string          `json:"labelPosition"`
	URL           string          `json:"urlpath"`
}

func (this *TemplateFormInfo) GetTemplateFormDetailInfo() (err error, infos JsonForm) {
	o := orm.NewOrm()
	form_info := TemplateFormInfo{}
	err = o.QueryTable("template_form_info").Filter("code", this.Code).One(&form_info)
	form_fileds := []TemplateFormField{}
	_, err = o.QueryTable("template_form_field").Filter("form_code", this.Code).OrderBy("id").All(&form_fileds)
	fields := []JsonFormField{}
	for _, each_f := range form_fileds {
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
			EnableAssign:  each_f.EnableAssign,
			DecimalPoint:  each_f.DecimalPoint,
			OutcomeState:  each_f.OutcomeState,
			StrList:       utils.StringValueToStrArray(each_f.StrList),
			StrIdList:     utils.StringValueToStrArray(each_f.StrIdList),
			AllowHalf:     f_conf.AllowHalf,
			Max:           f_conf.Max,
			FiledCode:     each_f.FiledCode,
			Layout:        f_conf.Layout,
			Options:       f_conf.Options,
		})
	}
	jf := JsonForm{
		FormCode:      form_info.Code,
		Title:         form_info.Title,
		Remark:        form_info.Remark,
		TypeId:        form_info.TypeId,
		IsTemplate:    false,
		Fields:        fields,
		LabelWidth:    form_info.LabelWidth,
		LabelPosition: form_info.LabelPosition,
	}
	return err, jf
}

func (this *TemplateFormInfo) EditTitle() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("template_form_info").Filter("code", this.Code).Update(orm.Params{"title": this.Title})
	return err
}

func (this *TemplateFormInfo) EditType() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("template_form_info").Filter("code", this.Code).Update(orm.Params{"type_id": this.TypeId})
	return err
}

func (this *TemplateFormInfo) UpdateTemplateCooperation() error {
	o := orm.NewOrm()
	pa := orm.Params{}
	t_info := TemplateFormInfo{}
	o.QueryTable("template_form_info").Filter("code", this.Code).One(&t_info)
	if this.Cooperation != "" {
		pa["cooperation"] = this.Cooperation
		json.Unmarshal([]byte(this.Cooperation), &this.CooperationInfo)
		this.RealCooperationInfo = FrontDepartInfoToReal(this.CooperationInfo, t_info.CorpId)
		real_corp, _ := json.Marshal(this.RealCooperationInfo)
		pa["real_cooperation"] = string(real_corp)
	}
	_, err := o.QueryTable("template_form_info").Filter("code", this.Code).Update(pa)
	return err
}

func AddTemplateFormsByJson(jsonStr string, corpid string, create_userid string) string {
	o := orm.NewOrm()
	form_info := JsonForm{}
	err := json.Unmarshal([]byte(jsonStr), &form_info)
	form_code := uuid.NewV4().String()
	realCooperation := ""
	cooperation := ""
	if form_info.FormCode != "" {
		form_code = form_info.FormCode
	}
	if len(form_info.FormCode) > 0 {
		tf_info := TemplateFormInfo{}
		o.QueryTable("template_form_info").Filter("code", form_code).One(&tf_info)
		realCooperation = tf_info.RealCooperation
		cooperation = tf_info.Cooperation
		o.QueryTable("template_form_info").Filter("code", form_code).Delete()
		o.QueryTable("template_form_field").Filter("form_code", form_code).Delete()
	}

	if err != nil {
		fmt.Println(err.Error())
	}

	form := TemplateFormInfo{
		Code:          form_code,
		CorpId:        corpid,
		Title:         form_info.Title,
		TypeId:        form_info.TypeId,
		Remark:        form_info.Remark,
		CreateUserid:  create_userid,
		LabelPosition: form_info.LabelPosition,
		LabelWidth:    form_info.LabelWidth,
	}
	if realCooperation != "" {
		form.RealCooperation = realCooperation
	}
	if cooperation != "" {
		form.Cooperation = cooperation
	}
	o.Insert(&form)
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, each_f := range form_info.Fields {
		f_conf := FormFieldConf{
			Maxlength:  each_f.Maxlength,
			LabelWidth: each_f.LabelWidth,
			AllowHalf:  each_f.AllowHalf,
			Max:        each_f.Max,
			Layout:     each_f.Layout,
			Options:    each_f.Options,
		}
		conf_str, _ := json.Marshal(f_conf)

		jf := TemplateFormField{
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
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	return form_code
}

func (this *TemplateFormInfo) DelFormInfo() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("template_form_info").Filter("code", this.Code).Delete()
	return err
}

type TempFormTree struct {
	Code       interface{}
	Name       string
	ParentCode int
	ChildTemp  []TempFormTree
	Disable    bool
}

func (this *TemplateFormInfo) GetAllTempFormTree() (error, []TempFormTree) {
	o := orm.NewOrm()
	t_info := []TypeInfo{}
	o.QueryTable("type_info").Filter("corpid", this.CorpId).Filter("type_desc", "规则类型").All(&t_info)
	type_ids := []int{}
	for _, each := range t_info {
		type_ids = append(type_ids, each.Id)
	}
	type_ids = append(type_ids, 1)
	u_info := UserInfo{}
	o.QueryTable("user_info").Filter("userid", this.CreateUserid).
		Filter("corpid", this.CorpId).One(&u_info)
	if u_info.UserId == "" {
		return nil, []TempFormTree{}
	}
	f_infos := []TemplateFormInfo{}
	//sql := ""
	u_departids := utils.StringValueToIntArray(u_info.DepartId)
	for _, each := range type_ids {
		sql := fmt.Sprintf("select * FROM template_form_info  WHERE  corpid = '%s' and "+
			"( real_cooperation :: jsonb -> 'userid' # '%s'", this.CorpId, this.CreateUserid)

		for _, each_d := range u_departids {
			sql += fmt.Sprintf(" or real_cooperation :: jsonb -> 'departid' # '%s'", fmt.Sprint(each_d))
		}
		sql += fmt.Sprintf(" ) and  type_id = %d ", each)
		tf_infos := []TemplateFormInfo{}
		o.Raw(sql).QueryRows(&tf_infos)
		f_infos = append(f_infos, tf_infos...)
	}
	tt_infos := []TemplateFormInfo{}
	o.QueryTable("template_form_info").Filter("corpid", this.CorpId).Filter("create_userid", this.CreateUserid).All(&tt_infos)
	f_infos = append(f_infos, tt_infos...)
	if len(f_infos) == 0 {
		return nil, []TempFormTree{}
	}
	ff_infos := []TemplateFormInfo{}
	m := make(map[string]bool)
	for _, each := range f_infos {
		if _, ok := m[each.Code]; !ok {
			ff_infos = append(ff_infos, each)
			m[each.Code] = true
		}
	}
	type_id := []int{}
	for _, each := range ff_infos {
		type_id = append(type_id, each.TypeId)
	}
	t_ids := utils.RemoveRepeatInt(type_id)
	type_infos := []TypeInfo{}
	_, err := o.QueryTable("type_info").Filter("corpid", this.CorpId).
		Filter("id__in", t_ids).All(&type_infos)

	type_infos = append(type_infos,
		TypeInfo{
			Id:      1,
			Content: "其他",
		},
	)

	tf := []TempFormTree{}
	for _, each := range type_infos {
		tf = append(tf, TempFormTree{
			Code:       each.Id,
			Name:       each.Content,
			ParentCode: 0,
			Disable:    true,
		})
	}
	for _, each := range ff_infos {
		for index, each_f := range tf {
			if each.TypeId == each_f.Code.(int) {
				cp := make([]TempFormTree, len(each_f.ChildTemp))
				copy(cp, each_f.ChildTemp)
				cp = append(cp, TempFormTree{
					Code:       each.Code,
					Name:       each.Title,
					ParentCode: each_f.Code.(int),
				})
				each_f.ChildTemp = cp
				tf[index] = each_f
			}
		}
	}
	return err, tf
}
