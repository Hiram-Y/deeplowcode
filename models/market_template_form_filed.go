package models

import (
	"DeepWorkload/utils"
	"DeepWorkload/utils/formula"
	"encoding/json"
	"github.com/astaxie/beego/orm"
)

type MarketTemplateFormField struct {
	Id            int           `orm:"column(id);auto;pk"`
	FormCode      string        `orm:"column(form_code)"`
	CmpType       string        `orm:"column(cmp_type)" json:"cmpType"`
	Label         string        `orm:"column(label)" json:"label"`
	Tag           string        `orm:"column(tag)" json:"tag"`
	TagIcon       string        `orm:"column(tag_icon)" json:"tagIcon"`
	Placeholder   string        `orm:"column(placeholder)" json:"placeholder"`
	Clearable     bool          `orm:"column(clearable)" json:"clearable"`
	Required      bool          `orm:"column(required)" json:"required"`
	Idx           int           `orm:"column(idx)" json:"idx"`
	RenderKey     int64         `orm:"column(render_key)" json:"render_key"`
	JoinGather    bool          `orm:"column(join_gather)" json:"join_gather"`
	IsCounted     bool          `orm:"column(is_counted)" json:"is_counted"`
	EnableGreater bool          `orm:"column(enable_greater)" json:"enable_greater"`
	EnableAssign  bool          `orm:"column(enable_assign)" json:"enable_assign"`
	DecimalPoint  int           `orm:"column(decimal_point)" json:"decimal_point"`
	OutcomeState  int           `orm:"column(outcome_state)" json:"outcome_state"`
	StrList       string        `orm:"column(str_list)" json:"str_list"`
	StrLists      []string      `orm:"-"`
	StrIdList     string        `orm:"column(str_id_list)" json:"str_id_list"`
	StrIdLists    []string      `orm:"-"`
	Conf          string        `orm:"column(conf) ;type(json) " json:"conf "`
	FFConf        FormFieldConf `orm:"-"`
	FiledCode     string        `orm:"column(filed_code)" json:"filed_code"`
}

func (this *MarketTemplateFormField) TableName() string {
	return "market_template_form_field"
}

func (this *MarketTemplateFormField) CountFormulaMethod(field_info string) (error, []FormFieldInfo) {
	fd_info := []FormFieldInfo{}
	json.Unmarshal([]byte(field_info), &fd_info)
	o := orm.NewOrm()
	ff_info := []MarketTemplateFormField{}
	o.QueryTable("market_template_form_field").Filter("form_code", this.FormCode).OrderBy("id").All(&ff_info)
	re_infos := []FormFieldInfo{}
	for _, each := range ff_info {
		if each.IsCounted {
			formula_method := utils.StringValueToStrArray(each.StrIdList)
			if len(formula_method) != 0 {
				err, real_formula := realFormulaMethod(fd_info, formula_method)
				if err == nil {
					v, err := formula.Count(real_formula, each.OutcomeState, each.DecimalPoint)
					if err != nil {
						return err, []FormFieldInfo{}
					}
					f := FormFieldInfo{
						FieldCode:     each.FiledCode,
						FieldLabel:    each.Label,
						JoinGather:    each.JoinGather,
						IsCounted:     each.IsCounted,
						EnableAssign:  each.EnableAssign,
						EnableGreater: each.EnableGreater,
						Value:         v,
					}
					re_infos = append(re_infos, f)
				}
			}
		}
	}
	return nil, re_infos
}
