package models

import (
	"DeepWorkload/utils"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	uuid "github.com/satori/go.uuid"
	"time"
)

type MarketTemplateFormInfo struct {
	Code           string    `orm:"column(code);pk"`
	Title          string    `orm:"column(title)"`
	Remark         string    `orm:"column(remark)"`
	CreateTime     time.Time `orm:"column(create_time);auto_now;type(datetime)"`
	CreateUserid   string    `orm:"column(create_userid)"`
	CreateUserName string    `orm:"-"`
	TypeId         int       `orm:"column(type_id)"`
	UrlPath        string    `orm:"column(url_path)"`
	TypeContent    string    `orm:"-"`
	LabelWidth     int       `orm:"column(label_width)"`
	LabelPosition  string    `orm:"column(label_position)"`
	IsActive       bool      `orm:"column(is_active)"`
}

func (this *MarketTemplateFormInfo) TableName() string {
	return "market_template_form_info"
}

func GetAllMarketFormByType(type_id int, is_active bool) (err error, infos []MarketTemplateFormInfo) {
	o := orm.NewOrm()
	qs := o.QueryTable("market_template_form_info")
	if type_id != 0 {
		qs = qs.Filter("type_id", type_id)
	}
	if is_active == true {
		qs = qs.Filter("is_active", true)
	}
	_, err = qs.All(&infos)
	if IsNoLastInsertIdError(err) {
		return nil, infos
	}
	return err, infos
}

func (this *MarketTemplateFormInfo) Active() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("market_template_form_info").Filter("code", this.Code).Update(orm.Params{"is_active": this.IsActive})
	return err
}

func (this *MarketTemplateFormInfo) DelOne() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("market_template_form_info").Filter("code", this.Code).Delete()
	if err == nil {
		o.QueryTable("market_template_form_field").Filter("form_code", this.Code).Delete()
	}
	return err
}

func AddMarketTemplateFormsByJson(jsonStr string, create_userid string) string {
	o := orm.NewOrm()
	form_info := JsonForm{}
	err := json.Unmarshal([]byte(jsonStr), &form_info)
	form_code := uuid.NewV4().String()
	if form_info.FormCode != "" {
		form_code = form_info.FormCode
	}
	if len(form_info.FormCode) > 0 {
		tf_info := MarketTemplateFormInfo{}
		o.QueryTable("market_template_form_info").Filter("code", form_code).One(&tf_info)
		o.QueryTable("market_template_form_info").Filter("code", form_code).Delete()
		o.QueryTable("market_template_form_field").Filter("form_code", form_code).Delete()
	}

	if err != nil {
		fmt.Println(err.Error())
	}

	form := MarketTemplateFormInfo{
		Code:          form_code,
		Title:         form_info.Title,
		TypeId:        form_info.TypeId,
		Remark:        form_info.Remark,
		CreateUserid:  create_userid,
		LabelPosition: form_info.LabelPosition,
		LabelWidth:    form_info.LabelWidth,
		UrlPath:       form_info.URL,
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

		jf := MarketTemplateFormField{
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

func (this *MarketTemplateFormInfo) GetTemplateFormDetailInfo() (err error, infos JsonForm) {
	o := orm.NewOrm()
	form_info := MarketTemplateFormInfo{}
	err = o.QueryTable("market_template_form_info").Filter("code", this.Code).One(&form_info)
	form_fileds := []MarketTemplateFormField{}
	_, err = o.QueryTable("market_template_form_field").Filter("form_code", this.Code).OrderBy("id").All(&form_fileds)
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
		URL:           form_info.UrlPath,
	}
	return err, jf
}
