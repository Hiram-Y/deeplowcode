package models

import (
	"github.com/astaxie/beego/orm"
	"strings"
)

func init() {
	//orm.RegisterModel(new(AuthPermanentCodes))
	orm.RegisterModel(new(AuthCorpInfo))
	orm.RegisterModel(new(Admin))
	//orm.RegisterModel(new(AuthInfo))
	//orm.RegisterModel(new(AuthUserInfo))
	orm.RegisterModel(new(Department))
	orm.RegisterModel(new(UserInfo))
	orm.RegisterModel(new(TypeInfo))
	orm.RegisterModel(new(Log))
	orm.RegisterModel(new(NodeConfig))
	orm.RegisterModel(new(DocPan))
	orm.RegisterModel(new(FormInfo))
	orm.RegisterModel(new(FormField))
	orm.RegisterModel(new(TemplateFormInfo))
	orm.RegisterModel(new(TemplateFormField))
	orm.RegisterModel(new(Workflow))
	orm.RegisterModel(new(Task))
	orm.RegisterModel(new(MainData))
	orm.RegisterModel(new(TaskPublic))
	orm.RegisterModel(new(AssignData))
	orm.RegisterModel(new(MarketType))
	orm.RegisterModel(new(MarketTemplateFormField))
	orm.RegisterModel(new(MarketTemplateFormInfo))
	orm.RegisterModel(new(MarketAdmin))
	orm.RegisterModel(new(LoginLog))
}

func IsNoLastInsertIdError(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "no LastInsertId available") {
		return true
	} else if strings.Contains(err.Error(), "LastInsertId is not supported") {
		return true
	} else if strings.Contains(err.Error(), "no row found") {
		return true
	}
	return false
}
