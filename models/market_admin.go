package models

import "github.com/astaxie/beego/orm"

type MarketAdmin struct {
	UserId string `orm:"column(userid);pk"`
	Passwd string `orm:"column(passwd)"`
}

func (this *MarketAdmin) TableName() string {
	return "market_admin"
}

func (this *MarketAdmin) Login() bool {
	o := orm.NewOrm()
	return o.QueryTable("market_admin").Filter("userid", this.UserId).Filter("passwd", this.Passwd).Exist()
}
