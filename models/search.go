package models

import (
	"DeepWorkload/utils"
	"fmt"
	"github.com/astaxie/beego/orm"
	"time"
)

func FuzzSearch(corpid, userid string, startTime, endTime time.Time, task_codes, content []string) (err error, m_infos []MainData) {
	if len(content) == 0 {
		return nil, []MainData{}
	}
	sql := fmt.Sprintf("select * from main_data where corpid = '%s'", corpid)
	if len(task_codes) != 0 {
		sql += fmt.Sprintf(" and task_code in ( %s )", utils.StringArrayToINArray(task_codes))
	}
	if userid != "" {
		sql += fmt.Sprintf(" and create_userid = '%s'", userid)
	}
	if !startTime.IsZero() {
		sql += fmt.Sprintf(" and create_time >= '%s'", startTime)
	}
	if !endTime.IsZero() {
		sql += fmt.Sprintf(" and create_time <= '%s'", endTime)
	}
	if len(content) != 0 {
		sql += " and ("
		for index, each := range content {
			if index == 0 {
				sql += fmt.Sprintf(" form_field_info :: TEXT like '%%%s%%'", each)
			} else {
				sql += fmt.Sprintf(" or form_field_info :: TEXT like '%%%s%%'", each)
			}
		}
		sql += " )"
	}
	o := orm.NewOrm()
	_, err = o.Raw(sql).QueryRows(&m_infos)
	m_infos = formatMainData(m_infos)
	return
}
