package models

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
)

type Admin struct {
	Id     int    `orm:"pk;auto;column(id)"`
	UserId string `orm:"column(userid)"`
	CorpId string `orm:"column(corpid)"`
}

func (this *Admin) TableName() string {
	return "admin"
}

func (this *Admin) PatchInsert(userids []string) {
	o := orm.NewOrm()
	for _, each := range userids {
		if !o.QueryTable("admin").
			Filter("corpid", this.CorpId).Filter("userid", each).Exist() {
			o.Insert(&Admin{
				UserId: each,
				CorpId: this.CorpId,
			})
		}
	}
}

func (this *Admin) InsertOne() error {
	o := orm.NewOrm()
	if !o.QueryTable("admin").
		Filter("corpid", this.CorpId).Filter("userid", this.UserId).Exist() {
		_, err := o.Insert(this)
		return err
	}
	return errors.New("已经存在")
}

func (this *Admin) DelOne() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("admin").Filter("corpid", this.CorpId).Filter("userid", this.UserId).Delete()
	return err
}

type FAdmin struct {
	Admin
	Name string `orm:"column(name)"`
}

func (this *Admin) GetAllAdminInfo() []FAdmin {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select admin.userid,user_info.name from user_info,admin "+
		"where admin.userid=user_info.userid and admin.corpid = '%s' and admin.corpid = user_info.corpid order by admin.id", this.CorpId)
	f := []FAdmin{}
	o.Raw(sql).QueryRows(&f)
	return f
}

func (this *Admin) CheckAdmin() bool {
	o := orm.NewOrm()
	return o.QueryTable("admin").Filter("corpid", this.CorpId).Filter("userid", this.UserId).Exist()
}
