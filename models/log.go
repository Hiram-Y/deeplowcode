package models

import (
	"DeepWorkload/utils"
	"github.com/astaxie/beego/orm"
	"time"
)

type Log struct {
	Id         int       `orm:"column(id);pk;auto"`
	CorpId     string    `orm:"column(corpid)"`
	UserId     string    `orm:"column(userid)"`
	UserName   string    `orm:"-"`
	TypeId     int       `orm:"column(type_id)"`
	TypeDesc   string    `orm:"column(type_desc)"`
	TaskCode   string    `orm:"column(task_code)"`
	DataCode   string    `orm:"column(data_code)"`
	CreateTime time.Time `orm:"auto_now_add;type(datetime);column(create_time)"`
	Origin     string    `orm:"column(origin);type(json)"`
	ToValue    string    `orm:"column(to_value);type(json)"`
	Remark     string    `orm:"column(remark)"`
}

func (this *Log) TableName() string {
	return "log"
}

var MapLogType = map[int]string{
	1:  "添加",
	2:  "编辑",
	3:  "添加分配",
	4:  "删除分配",
	5:  "编辑分配",
	6:  "审批通过",
	7:  "审批驳回",
	8:  "提交审批",
	9:  "撤回",
	10: "提交",
	11: "导入",
}

func (this *Log) GetAllLogByDataCode() (error, []Log) {
	o := orm.NewOrm()
	l_infos := []Log{}
	_, err := o.QueryTable("log").Filter("data_code", this.DataCode).OrderBy("-id").All(&l_infos)
	userids := []string{}
	for _, each := range l_infos {
		userids = append(userids, each.UserId)
	}
	userids = utils.RemoveRepByMap(userids)
	if len(userids) > 0 {
		u_infos := []UserInfo{}
		o.QueryTable("user_info").Filter("userid__in", userids).Filter("corpid", l_infos[0].CorpId).All(&u_infos)
		for index, each := range l_infos {
			for _, each_u := range u_infos {
				if each.UserId == each_u.UserId {
					each.UserName = each_u.Name
				}
			}
			l_infos[index] = each
		}
	}
	return err, l_infos
}

func AddOneLog(userid, data_code string, type_id int, remark string) {
	o := orm.NewOrm()
	m_info := MainData{}
	o.QueryTable("main_data").Filter("code", data_code).One(&m_info)
	l_info := &Log{
		UserId:   userid,
		DataCode: data_code,
		CorpId:   m_info.CorpId,
		TaskCode: m_info.TaskCode,
		Remark:   remark,
		TypeId:   type_id,
		TypeDesc: MapLogType[type_id],
	}
	o.Insert(l_info)
}
