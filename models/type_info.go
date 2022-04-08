package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
)

type TypeInfo struct {
	Id           int    `orm:"column(id);pk;auto"`
	CorpId       string `orm:"column(corpid)"`
	Content      string `orm:"column(content)"`
	TypeDesc     string `orm:"column(type_desc)"`
	CreateUserId string `orm:"column(create_userid)"`
}

func (this *TypeInfo) TableName() string {
	return "type_info"
}

func (this *TypeInfo) AddOneType() (error, int64) {
	o := orm.NewOrm()
	id, err := o.Insert(this)
	return err, id
}

func (this *TypeInfo) CheckSystemType() bool {
	o := orm.NewOrm()
	return o.QueryTable("type_info").Filter("corpid", this.CorpId).Filter("content", this.Content).Exist()
}

func (this *TypeInfo) DelOneTypeInfo() error {
	o := orm.NewOrm()
	t_info := TypeInfo{}
	o.QueryTable("type_info").Filter("id", this.Id).One(&t_info)
	_, err := o.QueryTable("type_info").Filter("id", this.Id).Delete()
	if err == nil {
		if t_info.TypeDesc == "任务类型" {
			o.QueryTable("task").Filter("corpid", this.CorpId).Filter("type_id", this.Id).Update(orm.Params{"type_id": 2})
		} else {
			o.QueryTable("template_form_info").Filter("corpid", this.CorpId).Filter("type_id", this.Id).Update(orm.Params{"type_id": 1})
		}
	}
	return err
}

func (this *TypeInfo) EditTypeInfo() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("type_info").Filter("id", this.Id).Update(orm.Params{"content": this.Content})
	return err
}

func (this *TypeInfo) GetAllTypeInfoByTypeDesc() (error, []TypeInfo) {
	o := orm.NewOrm()
	infos := []TypeInfo{}
	qs := o.QueryTable("type_info").Filter("corpid", this.CorpId).Filter("type_desc", this.TypeDesc).
		OrderBy("-id", "create_userid", "type_desc")
	if this.CreateUserId != "" {
		qs = qs.Filter("create_userid", this.CreateUserId)
	}
	_, err := qs.All(&infos)
	if this.TypeDesc == "规则类型" {
		infos = append(infos, TypeInfo{
			Id:       1,
			CorpId:   this.CorpId,
			TypeDesc: this.TypeDesc,
			Content:  "其他",
		})
	} else {
		infos = append(infos, TypeInfo{
			Id:       2,
			CorpId:   this.CorpId,
			TypeDesc: this.TypeDesc,
			Content:  "其他",
		})
	}
	return err, infos
}

func TypeInfoById(corpid, userid, type_desc string) map[int]TypeInfo {
	o := orm.NewOrm()
	t_info := []TypeInfo{}
	qs := o.QueryTable("type_info").Filter("corpid", corpid).Filter("type_desc", type_desc)
	if userid != "" {
		qs = qs.Filter("create_userid", userid)
	}
	_, err := qs.All(&t_info)
	if err != nil {
		fmt.Println(err.Error())
	}
	ty_map := map[int]TypeInfo{}
	for _, each := range t_info {
		ty_map[each.Id] = each
	}
	return ty_map
}
