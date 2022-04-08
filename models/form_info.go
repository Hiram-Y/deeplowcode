package models

import (
	"DeepWorkload/utils"
	"encoding/json"
	"github.com/astaxie/beego/orm"
	uuid "github.com/satori/go.uuid"
	"time"
)

type FormInfo struct {
	Code          string    `orm:"column(code);pk"`
	CorpId        string    `orm:"column(corpid)"`
	Title         string    `orm:"column(title)"`
	Remark        string    `orm:"column(remark)"`
	CreateTime    time.Time `orm:"column(create_time);auto_now;type(datetime)"`
	CreateUserid  string    `orm:"column(create_userid)"`
	TypeId        int       `orm:"column(type_id)"`
	TypeContent   string    `orm:"-"`
	LabelWidth    int       `orm:"column(label_width)"`
	LabelPosition string    `orm:"column(label_position)"`
}

func (this *FormInfo) TableName() string {
	return "form_info"
}

func (this *FormInfo) AddOneForm() (string, error) {
	o := orm.NewOrm()
	this.Code = uuid.NewV4().String()
	_, err := o.Insert(this)
	if IsNoLastInsertIdError(err) {
		return this.Code, nil
	}
	return "", err
}

func (this *FormInfo) EditOneFormInfo() error {
	o := orm.NewOrm()
	pa := orm.Params{}
	if this.Title != "" {
		pa["title"] = this.Title
	}
	if this.Remark != "" {
		pa["remark"] = this.Remark
	}
	if this.TypeId != 0 {
		pa["type_id"] = this.TypeId
	}
	if this.LabelWidth != 0 {
		pa["label_width"] = this.LabelWidth
	}
	if this.LabelPosition != "" {
		pa["label_position"] = this.LabelPosition
	}
	_, err := o.QueryTable("form_info").Filter("code", this.Code).Update(pa)
	return err
}

func FormFieldToJsonFormField(form_fileds []FormField) []JsonFormField {
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
			AssignData:    []string{},
		})
	}
	return fields
}

func (this *FormInfo) GetFormDetailInfo() (err error, infos JsonForm) {
	o := orm.NewOrm()
	form_info := FormInfo{}
	err = o.QueryTable("form_info").Filter("code", this.Code).One(&form_info)
	form_fileds := []FormField{}
	_, err = o.QueryTable("form_field").Filter("form_code", this.Code).OrderBy("id").All(&form_fileds)
	fields := FormFieldToJsonFormField(form_fileds)
	jf := JsonForm{
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
