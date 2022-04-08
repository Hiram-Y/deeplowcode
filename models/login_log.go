package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type LoginLog struct {
	Id         int       `orm:"column(id);pk;auto"`
	CorpId     string    `orm:"column(corpid)"`
	UserId     string    `orm:"column(userid)"`
	CreateTime time.Time `orm:"auto_now_add;type(datetime);column(create_time)"`
	Ip         string    `orm:"column(ip)"`
	City       string    `orm:"column(city)"`
}

func (this *LoginLog) TableName() string {
	return "login_log"
}

func (this *LoginLog) AddOne() {
	o := orm.NewOrm()
	o.Insert(this)
}
