package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type DocPan struct {
	Id         int       `orm:"pk;auto;column(id)"`
	Type       string    `orm:"column(type)"`
	Path       string    `orm:"column(path)"`
	UrlPath    string    `orm:"column(url_path)"`
	DocCode    string    `orm:"column(doc_code)"`
	CreateTime time.Time `orm:"column(create_time);auto_now;type(datetime)"`
	DocName    string    `orm:"column(doc_name)"`
	Userid     string    `orm:"column(userid)"`
}

func (this *DocPan) TableName() string {
	return "doc_pan"
}

func (this *DocPan) InsertDocPan() (err error) {
	o := orm.NewOrm()
	_, err = o.Insert(this)
	return err
}

func GetDocinfo(docCode string) (err error, docInfo DocPan) {
	o := orm.NewOrm()
	err = o.QueryTable("doc_pan").Filter("doc_code", docCode).One(&docInfo)
	return err, docInfo
}
