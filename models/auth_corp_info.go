package models

import (
	"DeepWorkload/lib/pinyin"
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
	uuid "github.com/satori/go.uuid"
)

type AuthCorpInfo struct {
	Id                int    `orm:"pk;auto;column(id)"`
	Corpid            string `orm:"column(corpid)"`
	CorpName          string `orm:"column(corp_name)"`
	CorpType          string `orm:"column(corp_type)"`
	CorpSquareLogoUrl string `orm:"column(corp_square_logo_url)"`
	CorpUserMax       int    `orm:"column(corp_user_max)"`
	CorpAgentMax      int    `orm:"column(corp_agent_max)"`
	CorpFullName      string `orm:"column(corp_full_name)"`
	VerifiedEndTime   int    `orm:"column(verified_end_time)"`
	SubjectType       int    `orm:"column(subject_type)"`
	CorpWxqrcode      string `orm:"column(corp_wxqrcode)"`
	CorpScale         string `orm:"column(corp_scale)"`
	CorpIndustry      string `orm:"column(corp_industry)"`
	CorpSubIndustry   string `orm:"column(corp_sub_industry)"`
	Location          string `orm:"column(location)"`
	Suiteid           string `orm:"column(suiteid)"`
	IsQywx            bool   `orm:"column(is_qywx)"`
}

func (this *AuthCorpInfo) TableName() string {
	return "auth_corp_info"
}

func (this *AuthCorpInfo) SignUpAuthCorpInfo(userid, username, passwd, mobile string) error {
	if SignUpCheck(mobile) {
		return errors.New("手机号已经注册")
	}
	o := orm.NewOrm()
	this.Corpid = uuid.NewV4().String()
	usernamepy, _ := pinyin.New(username).Split("").Convert()
	u_info := UserInfo{
		CorpId:     this.Corpid,
		UserId:     userid,
		Name:       username,
		Passwd:     passwd,
		Mobile:     mobile,
		IsSign:     true,
		DepartId:   "{1}",
		NamePinyin: usernamepy,
	}
	corpnamepy, _ := pinyin.New(this.CorpName).Split("").Convert()
	de := Department{
		Corpid:           this.Corpid,
		Department:       this.CorpName,
		ParentId:         0,
		DepartmentId:     1,
		DepartmentPinYin: corpnamepy,
	}
	admin := Admin{
		UserId: userid,
		CorpId: this.Corpid,
	}

	_, err := o.Insert(this)
	if err != nil {
		fmt.Println(err.Error())
	}
	o.Insert(&de)
	o.Insert(&u_info)
	o.Insert(&admin)
	return err
}

func SignUpCheck(mobile string) bool {
	o := orm.NewOrm()
	return o.QueryTable("user_info").Filter("mobile", mobile).Filter("is_sign", true).Exist()
}
