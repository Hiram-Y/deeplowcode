package models

import (
	"DeepWorkload/conf"
	"DeepWorkload/utils"
	"DeepWorkload/utils/formula"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego/orm"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"strconv"
	"time"
)

type FormField struct {
	Id            int           `orm:"column(id);auto;pk"`
	FormCode      string        `orm:"column(form_code)" json:"form_code"`
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
	StrList       string        `orm:"column(str_list)" json:"strlist"`
	StrLists      []string      `orm:"-"`
	StrIdList     string        `orm:"column(str_id_list)" json:"str_id_list"`
	StrIdLists    []string      `orm:"-"`
	Conf          string        `orm:"column(conf) ;type(json) " json:"conf "`
	FFConf        FormFieldConf `orm:"-"`
	FiledCode     string        `orm:"column(filed_code)" json:"filed_code"`
}

type FormFieldInfo struct {
	FieldCode     string      `json:"field_code"`
	FieldLabel    string      `json:"field_label"`
	JoinGather    bool        `json:"join_gather"`
	IsCounted     bool        `json:"is_counted"`
	EnableGreater bool        `json:"enable_greater"`
	EnableAssign  bool        `json:"enable_assign"`
	Value         interface{} `json:"value"`
	OptLabel      string      `json:"opt_label"`
	AssignData    []Assign    `json:"assign_data"`
	StrList       []string    `json:"str_list"`
}

type Assign struct {
	ToUserId   string
	ToValue    float64
	ToUserName string
}

func (this *FormField) TableName() string {
	return "form_field"
}

func (this *FormField) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(this)
	return err
}

func CheckFormulaMethod(formula_methods []string) error {
	return formula.CheckFormula(formula_methods)
}

func (this *FormField) CountFormulaMethod(field_info string) (error, []FormFieldInfo) {
	fd_info := []FormFieldInfo{}
	json.Unmarshal([]byte(field_info), &fd_info)
	o := orm.NewOrm()
	ff_info := []FormField{}
	o.QueryTable("form_field").Filter("form_code", this.FormCode).OrderBy("id").All(&ff_info)
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

func CreateFormUploadExcel(form_code string) (error, string) {
	o := orm.NewOrm()
	f_info := FormInfo{}
	err := o.QueryTable("form_info").Filter("code", form_code).One(&f_info)
	if err != nil {
		return err, ""
	}
	ff_infos := []FormField{}
	o.QueryTable("form_field").Filter("form_code", form_code).All(&ff_infos)
	f := excelize.NewFile()
	index := f.GetActiveSheetIndex()
	sheetName := f.GetSheetName(index)
	style, _ := f.NewStyle(`{"font":{"bold":true,"family":"Times New Roman","size":22}}`)
	f.SetCellValue(sheetName,
		"A1",
		`
		不能在该Excel表中对信息类别进行增加、删除或修改！如需填写时间,请填写为文本格式，示例格式：2020/11/12`,
	)
	f.SetCellStyle(sheetName, "A1", "A1", style)
	f.MergeCell(sheetName, "A1", "Z1")
	f.SetRowHeight(sheetName, 1, 70)
	f.SetCellValue(sheetName, "A2", "工号")
	f.SetCellValue(sheetName, "B2", "姓名")
	for _, each := range ff_infos {
		if !utils.IsExistStr(conf.NotExportIcon, each.TagIcon) {
			if each.TagIcon != "upload" && each.TagIcon != "image" {
				f_code := utils.GetNextLetters(utils.GetNextLetters(each.FiledCode))
				f.SetCellValue(sheetName, fmt.Sprintf("%s%d", f_code, 2), each.Label)
			}
		}
	}
	file_name := f_info.Title + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ".xlsx"
	path := "uploads/docs/" + file_name
	err = f.SaveAs(path)
	url_path := "docs/" + file_name
	return err, url_path
}

func formatFormFieldValueByExcel(ff_infos []FormField, values []string) []FormFieldInfo {
	ffs := []FormFieldInfo{}
	for _, each := range ff_infos {
		temp_tf := FormFieldInfo{
			FieldCode:     each.FiledCode,
			FieldLabel:    each.Label,
			JoinGather:    each.JoinGather,
			IsCounted:     each.IsCounted,
			EnableAssign:  each.EnableAssign,
			EnableGreater: each.EnableGreater,
			StrList:       utils.StringValueToStrArray(each.StrList),
		}
		for index, each_v := range values {
			letter_code := utils.GetLetterByIdx(index + 1)
			if !utils.IsExistStr(conf.NotExportIcon, each.TagIcon) {
				if each.FiledCode == letter_code {
					temp_tf.Value = each_v
					if each.TagIcon == "number" {
						f_value, _ := strconv.ParseFloat(each_v, 64)
						temp_tf.Value = f_value
					}
					if each.TagIcon == "select_number" {
						json.Unmarshal([]byte(each.Conf), &each.FFConf)
						for _, each_opt := range each.FFConf.Options {
							if each_opt["label"] == each_v {
								fmt.Println(temp_tf.Value)
								fmt.Println(each_opt["value"])

								temp_tf.Value = each_opt["value"]
								temp_tf.OptLabel = fmt.Sprintf("%s__%s", decimal.NewFromFloat(each_opt["value"].(float64)).String(), each_opt["label"])
								break
							}
						}
					}
				}
			}
		}
		ffs = append(ffs, temp_tf)
	}
	return ffs
}

func GetMainDatasByExcel(path, corpid, task_code, form_code string) (uids []UserInfo, m_datas []MainData) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	datas := []MainData{}
	rows := f.GetRows("Sheet1")
	ff_infos := []FormField{}
	o := orm.NewOrm()
	o.QueryTable("form_field").Filter("form_code", form_code).All(&ff_infos)
	all_uids := GetAllUserIdByCorpId(corpid)
	for _, each := range rows[2:] {
		if each[0] != "" && utils.IsExistStr(all_uids, each[0]) {
			ff := formatFormFieldValueByExcel(ff_infos, each[2:])
			ff_s, _ := json.Marshal(ff)
			datas = append(datas, MainData{
				Code:          uuid.NewV4().String(),
				CorpId:        corpid,
				FormFieldInfo: string(ff_s),
				FormCode:      form_code,
				CreateUserId:  each[0],
				TaskCode:      task_code,
				State:         6,
			})
		} else {
			if each[0] != "" {
				uids = append(uids, UserInfo{
					UserId: each[0],
					Name:   each[1],
				})
			}

		}
	}
	return uids, datas
}

func realFormulaMethod(fields []FormFieldInfo, formula_method []string) (error, []string) {
	for index, each := range formula_method {

		for _, each_f := range fields {
			if each == each_f.FieldCode {
				if each_f.Value == nil {
					return errors.New("no value"), []string{}
				}
				formula_method[index] = fmt.Sprintf("%f", each_f.Value.(float64))
			}
		}
	}
	return nil, formula_method
}
