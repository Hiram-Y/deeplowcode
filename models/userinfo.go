package models

import (
	"DeepWorkload/lib/pinyin"
	"DeepWorkload/lib/redisgo"
	"DeepWorkload/utils"
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
	"strconv"
	"strings"
)

type UserInfo struct {
	Id         int    `orm:"column(id);pk;auto"`
	CorpId     string `orm:"column(corpid)"`
	UserId     string `orm:"column(userid)"`
	Name       string `orm:"column(name)"`
	NamePinyin string `orm:"column(name_pinyin)"`
	Mobile     string `orm:"column(mobile)"`
	Email      string `orm:"column(email)"`
	DepartId   string `orm:"column(departid)"`
	DepartIds  []int  `orm:"-"`
	Passwd     string `orm:"column(passwd)"`
	OpenId     string `orm:"column(openid)"`
	IsSign     bool   `orm:"column(is_sign)"`
}

type ReTurnUserInfo struct {
	Corpid      string         `json:"c"`
	Userid      string         `json:"u"`
	Mobile      string         `json:"m"`
	Name        string         `json:"n"`
	Departid    []int          `json:"de"`
	OpenId      string         `json:"op"`
	Departments map[int]string `json:"des"`
}

type FUserInfo struct {
	Userid   string
	Departid int
	Name     string
}

func (this *UserInfo) TableName() string {
	return "user_info"
}

type CorpAndUser struct {
	AuthCorpInfo
	UserInfo
	IsAdmin         bool
	IsPasswdSetting bool
}

func GetLoginUserInfoByMobile(mobile string) (int, []CorpAndUser) {
	o := orm.NewOrm()
	if o.QueryTable("user_info").Filter("mobile", mobile).Exist() {
		return 0, getCorpAndUser(mobile, "")
	}
	return -1, []CorpAndUser{}
}

func GetLoginUserInfoByOpenId(openid string) (int, []CorpAndUser) {
	o := orm.NewOrm()
	if o.QueryTable("user_info").Filter("openid", openid).Exist() {
		return 0, getCorpAndUserByOpenid(openid, "")
	}
	return -1, []CorpAndUser{}
}

func (this *UserInfo) SetPasswd() error {
	o := orm.NewOrm()
	pa := orm.Params{
		"passwd": this.Passwd,
	}
	_, err := o.QueryTable("user_info").Filter("userid", this.UserId).Filter("corpid", this.CorpId).Update(pa)
	return err
}

func (this *UserInfo) GetUserInfo() interface{} {
	o := orm.NewOrm()
	type fullUserInfo struct {
		UserInfo
		Departments []string
	}
	u_info := UserInfo{}
	o.QueryTable("user_info").Filter("userid", this.UserId).Filter("corpid", this.CorpId).One(&u_info)
	departids := utils.SqlStringValue(u_info.DepartId)
	de_infos := []Department{}
	department := []string{}
	o.QueryTable("department").Filter("corpid", this.CorpId).Filter("department_id__in", departids).All(&de_infos)
	for _, each := range de_infos {
		department = append(department, each.Department)
	}
	return fullUserInfo{u_info, department}
}

func (this *UserInfo) UpdatePersonalInfo(userid string) error {
	o := orm.NewOrm()
	pa := orm.Params{}
	if this.Name != "" {
		pa["name"] = this.Name
		pa["name_pinyin"], _ = pinyin.New(this.Name).Split("").Convert()
	}
	if this.Passwd != "" {
		pa["passwd"] = this.Passwd
	}
	if this.Mobile != "" {
		pa["mobile"] = this.Mobile
	}
	if userid != "" {
		pa["userid"] = this.UserId
	}
	if len(pa) == 0 {
		return errors.New("没有修改信息")
	}
	_, err := o.QueryTable("user_info").Filter("userid", this.UserId).Filter("corpid", this.CorpId).Update(pa)
	con, _ := redisgo.New(redisgo.Options{})
	con.Del(this.CorpId + "userilist")
	return err
}

func getCorpAndUserByOpenid(openid, passwd string) []CorpAndUser {
	o := orm.NewOrm()
	u_info := []UserInfo{}
	qs := o.QueryTable("user_info").Filter("openid", openid)
	if passwd != "" {
		qs = qs.Filter("passwd", passwd)
	}
	qs.All(&u_info)
	if len(u_info) == 0 {
		return []CorpAndUser{}
	}
	corpids := []string{}
	userids := []string{}
	for _, each := range u_info {
		corpids = append(corpids, each.CorpId)
		userids = append(userids, each.UserId)
	}
	au := []AuthCorpInfo{}
	o.QueryTable("auth_corp_info").Filter("corpid__in", corpids).All(&au)
	cu := []CorpAndUser{}
	ads := []Admin{}
	o.QueryTable("admin").Filter("corpid__in", corpids).Filter("userid__in", userids).All(&ads)
	for _, each := range au {
		for _, each_u := range u_info {
			if each.Corpid == each_u.CorpId {
				tcu := CorpAndUser{
					each,
					each_u,
					false,
					false,
				}
				for _, each_a := range ads {
					if each_a.UserId == each_u.UserId && each_u.CorpId == each_a.CorpId {
						tcu.IsAdmin = true
						break
					}
				}
				if each_u.Passwd != "" {
					tcu.IsPasswdSetting = true
				}
				cu = append(cu, tcu)
				break
			}
		}
	}
	return cu
}

func getCorpAndUser(mobile, passwd string) []CorpAndUser {
	o := orm.NewOrm()
	u_info := []UserInfo{}
	qs := o.QueryTable("user_info").Filter("mobile", mobile)
	if passwd != "" {
		qs = qs.Filter("passwd", passwd)
	}
	qs.All(&u_info)
	if len(u_info) == 0 {
		return []CorpAndUser{}
	}
	corpids := []string{}
	userids := []string{}
	for _, each := range u_info {
		corpids = append(corpids, each.CorpId)
		userids = append(userids, each.UserId)
	}
	au := []AuthCorpInfo{}
	o.QueryTable("auth_corp_info").Filter("corpid__in", corpids).All(&au)
	cu := []CorpAndUser{}
	ads := []Admin{}
	o.QueryTable("admin").Filter("corpid__in", corpids).Filter("userid__in", userids).All(&ads)
	for _, each := range au {
		for _, each_u := range u_info {
			if each.Corpid == each_u.CorpId {
				tcu := CorpAndUser{
					each,
					each_u,
					false,
					false,
				}
				for _, each_a := range ads {
					if each_a.UserId == each_u.UserId && each_u.CorpId == each_a.CorpId {
						tcu.IsAdmin = true
						break
					}
				}
				if each_u.Passwd != "" {
					tcu.IsPasswdSetting = true
				}
				cu = append(cu, tcu)
				break
			}
		}
	}
	return cu
}

func BindUserInfo(corp_info []CorpAndUser, open_id string) {
	o := orm.NewOrm()
	for _, each := range corp_info {
		o.QueryTable("user_info").Filter("userid", each.UserId).Filter("corpid", each.Corpid).Update(
			orm.Params{"openid": open_id},
		)
	}
}

func (this *UserInfo) CheckPasswd(ip string) (int, []CorpAndUser) {
	o := orm.NewOrm()
	if o.QueryTable("user_info").Filter("mobile", this.Mobile).Filter("passwd", this.Passwd).Exist() {
		corp_info := getCorpAndUser(this.Mobile, this.Passwd)
		go func() {
			for _, each := range corp_info {
				lo := LoginLog{
					CorpId: each.CorpId,
					UserId: each.UserId,
					Ip:     ip,
				}
				lo.AddOne()
			}
		}()
		return 0, corp_info
	}
	return -1, []CorpAndUser{}
}

func (this *UserInfo) CheckOpenid() bool {
	o := orm.NewOrm()
	return o.QueryTable("userinfo").Filter("corpid", this.CorpId).
		Filter("userid", this.UserId).Filter("openid", this.OpenId).Exist()
}

func (this *UserInfo) UnBindUserInfo() error {
	o := orm.NewOrm()
	op := orm.Params{}
	op["openid"] = ""
	_, err := o.QueryTable("userinfo").Filter("corpid", this.CorpId).
		Filter("openid", this.OpenId).Filter("userid", this.UserId).Update(op)
	return err
}

func (this *UserInfo) CheckMobileExistInThisCorp() bool {
	o := orm.NewOrm()
	return o.QueryTable("user_info").Filter("corpid", this.CorpId).Filter("mobile", this.Mobile).Exist()
}

func (this *UserInfo) CheckUserIdExistInThisCorp() bool {
	o := orm.NewOrm()
	return o.QueryTable("user_info").Filter("corpid", this.CorpId).Filter("userid", this.UserId).Exist()
}

func (this *UserInfo) AddUserInfo() (err error) {
	o := orm.NewOrm()
	if this.CheckMobileExistInThisCorp() || this.CheckUserIdExistInThisCorp() {
		return errors.New("手机号或用户ID重复")
	}
	de := utils.SqlArrayValue(this.DepartIds)
	this.NamePinyin, _ = pinyin.New(this.Name).Split("").Convert()
	_, err = o.Raw("INSERT INTO user_info (corpid, userid, name,name_pinyin, departid,mobile) VALUES (?,?,?,?,?,?)",
		this.CorpId, this.UserId, this.Name, this.NamePinyin, de, this.Mobile).Exec()
	con, _ := redisgo.New(redisgo.Options{})
	con.Del(this.CorpId + "userilist")
	return err
}

func (this *UserInfo) GetUserInfoDetail() (err error, userinfo ReTurnUserInfo) {
	o := orm.NewOrm()
	u_info := UserInfo{}
	qs := o.QueryTable("user_info").Filter("userid", this.UserId).Filter("corpid", this.CorpId)
	err = qs.One(&u_info)
	departids := utils.StringValueToIntArray(u_info.DepartId)
	err, departinfo := GetAllDepartNameByDepartIds(u_info.CorpId, departids)
	tempdepartinfo := map[int]string{}
	for _, each_id := range departids {
		if departinfo[each_id] != "" {
			tempdepartinfo[each_id] = departinfo[each_id]
		}
	}
	each_info := ReTurnUserInfo{
		Userid:      u_info.UserId,
		Corpid:      u_info.CorpId,
		Name:        u_info.Name,
		Mobile:      u_info.Mobile,
		OpenId:      u_info.OpenId,
		Departid:    departids,
		Departments: tempdepartinfo,
	}
	return err, each_info
}

func ClearQyWxUserAndDepartmentByCorpId(corpid string) {
	o := orm.NewOrm()
	o.QueryTable("user_info").Filter("corpid", corpid).Delete()
	o.QueryTable("department").Filter("corpid", corpid).Delete()
}

func (this *UserInfo) FormatUserDepartId() error {
	if len(this.DepartIds) == 0 {
		return errors.New("Deparid Array is nil")
	}
	s := make([]string, len(this.DepartIds))
	for i, v := range this.DepartIds {
		s[i] = strconv.Itoa(int(v))
	}
	this.DepartId = "{" + strings.Join(s, ",") + "}"
	return nil
}

//func (this *UserInfo) AddUserInfo()   {
//	o := orm.NewOrm()
//	err:=this.FormatUserDepartId()
//	if err == nil{
//		if !o.QueryTable("user_info").Filter("corpid",this.CorpId).Filter("userid",this.UserId).Exist(){
//			this.DepartId = utils.IntArrayToStr(this.DepartIds)
//			o.Insert(this)
//			con, _ := redisgo.New(redisgo.Options{})
//			con.Del(this.CorpId + "userilist")
//		}
//	}else {
//		fmt.Println(err.Error())
//	}
//}

func (this *UserInfo) UpdateUserInfoId(userid string) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("user_info").Filter("corpid", this.CorpId).Filter("userid", this.UserId).Update(orm.Params{"userid": userid})
	con, _ := redisgo.New(redisgo.Options{})
	con.Del(this.CorpId + "userilist")
	return err
}

func (this *UserInfo) UpdateUserInfo() error {
	o := orm.NewOrm()
	pa := orm.Params{}
	if this.Name != "" {
		pa["name"] = this.Name
		this.NamePinyin, _ = pinyin.New(this.Name).Split("").Convert()
		pa["name_pinyin"] = this.NamePinyin

	}
	if this.Mobile != "" {
		pa["mobile"] = this.Mobile
	}

	if len(this.DepartIds) > 0 {
		pa["departid"] = utils.IntArrayToStr(this.DepartIds)
	}
	if len(pa) > 0 {
		_, err := o.QueryTable("user_info").Filter("corpid", this.CorpId).Filter("userid", this.UserId).Update(pa)
		con, _ := redisgo.New(redisgo.Options{})
		con.Del(this.CorpId + "userilist")
		return err
	}

	return nil
}

func (this *UserInfo) DelUserInfo() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("user_info").Filter("userid", this.UserId).Filter("corpid", this.CorpId).Delete()
	con, _ := redisgo.New(redisgo.Options{})
	con.Del(this.CorpId + "userilist")
	return err
}

func (this *UserInfo) PatchDelUserInfo(userids []string) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("user_info").Filter("corpid", this.CorpId).Filter("userid__in", userids).Delete()
	con, _ := redisgo.New(redisgo.Options{})
	con.Del(this.CorpId + "userilist")
	return err
}

func (this *UserInfo) GetUserMapByUserIds(userids []string) map[string][]int {
	o := orm.NewOrm()
	infos := []UserInfo{}
	o.QueryTable("user_info").Filter("corpid", this.CorpId).All(&infos)
	for index, each := range infos {
		each.DepartIds = utils.StringValueToIntArray(each.DepartId)
		infos[index] = each
	}
	re_info := map[string][]int{}
	for _, each := range userids {
		for _, each_f := range infos {
			if each == each_f.UserId {
				re_info[each] = each_f.DepartIds
				break
			}
		}
	}
	return re_info
}

func GetUserNameMapByUseridAndCorpId(userids []string, corpid string) map[string]string {
	o := orm.NewOrm()
	u_infos := []UserInfo{}
	o.QueryTable("user_info").Filter("corpid", corpid).Filter("userid__in", userids).All(&u_infos)
	uMap := map[string]string{}
	for _, each := range u_infos {
		uMap[each.UserId] = each.Name
	}
	return uMap
}

func GetUserDepartIds(corpid, userid string) []int {
	o := orm.NewOrm()
	info := UserInfo{}
	o.QueryTable("user_info").Filter("corpid", corpid).Filter("userid", userid).One(&info)
	if info.UserId == "" {
		return []int{}
	} else {
		return utils.StringValueToIntArray(info.DepartId)
	}
}

func GetAllUserList(corpid string) []FUserInfo {
	o := orm.NewOrm()
	con, _ := redisgo.New(redisgo.Options{})
	Users := []UserInfo{}
	UserInfos := []FUserInfo{}
	cache_key := corpid + "userilist"
	c_err := con.GetObject(cache_key, &UserInfos)
	if c_err == nil {
		return UserInfos
	}
	var list []orm.ParamsList
	_, err := o.Raw("SELECT userid,departid,name FROM user_info WHERE corpid = ?", corpid).ValuesList(&list)
	if err != nil {
		fmt.Println(err)
	}
	for _, each := range list {
		userinfo := UserInfo{
			UserId:    each[0].(string),
			DepartIds: utils.StringValueToIntArray(each[1].(string)),
			Name:      each[2].(string),
		}
		Users = append(Users, userinfo)
	}

	for _, each := range Users {
		for _, each_departid := range each.DepartIds {
			u_info := FUserInfo{
				Userid:   each.UserId,
				Departid: each_departid,
				Name:     each.Name,
			}
			UserInfos = append(UserInfos, u_info)
		}
	}
	con.Set(cache_key, UserInfos, -1)
	return UserInfos
}

func GetUserCountByDepartIds(departids []int, corpid string) int {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select count(*) as count from user_info where corpid = '%s'", corpid)
	if len(departids) > 0 {
		sql += " and ("
		for index, each := range departids {
			if index == 0 {
				sql += fmt.Sprintf("%d = any(departid)", each)
			} else {
				sql += fmt.Sprintf("or %d = any(departid)", each)
			}
		}
		sql += ")"
	}
	cc := Counts{}
	o.Raw(sql).QueryRow(&cc)
	return cc.Count
}

func AddUserInfo(infos []UserInfo) (err error) {
	o := orm.NewOrm()
	if len(infos) > 0 {
		_, err = o.InsertMulti(len(infos), infos)
		con, _ := redisgo.New(redisgo.Options{})
		con.Del(infos[0].CorpId + "userilist")
	}
	return err
}

func (this *UserInfo) GetAllUserList() (error, []UserInfo) {
	o := orm.NewOrm()
	Users := []UserInfo{}
	_, err := o.QueryTable("user_info").Filter("corpid", this.CorpId).All(&Users)
	return err, Users
}

func GetAllUserIdByCorpId(corpid string) []string {
	o := orm.NewOrm()
	Users := []UserInfo{}
	o.QueryTable("user_info").Filter("corpid", corpid).All(&Users)
	uids := []string{}
	for _, each := range Users {
		uids = append(uids, each.UserId)
	}
	return uids
}

type UserNameAndUserId struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}

func (this *UserInfo) GetAllUserNameAndUserId() (error, []UserNameAndUserId) {
	o := orm.NewOrm()
	Users := []UserInfo{}
	_, err := o.QueryTable("user_info").Filter("corpid", this.CorpId).All(&Users)
	un := []UserNameAndUserId{}
	for _, each := range Users {
		un = append(un, UserNameAndUserId{
			Uid:  each.UserId,
			Name: each.Name,
		})
	}
	return err, un
}

func GetUserInfoByDepartIds(departids []int, corpid string, pagesize, pageindex int) (err error, userinfo []ReTurnUserInfo, count int) {
	o := orm.NewOrm()
	count = GetUserCountByDepartIds(departids, corpid)
	sql := fmt.Sprintf("select * from user_info where corpid = '%s'  ", corpid)
	if len(departids) > 0 {
		sql += " and ("
		for index, each := range departids {
			if index == 0 {
				sql += fmt.Sprintf("%d = any(departid)", each)
			} else {
				sql += fmt.Sprintf("or %d = any(departid)", each)
			}
		}
		sql += ")"
	}
	if pagesize != 0 {
		sql += fmt.Sprintf("limit %d offset %d", pagesize, (pageindex-1)*pagesize)
	}
	u_info := []UserInfo{}
	_, err = o.Raw(sql).QueryRows(&u_info)
	wxu_info := []ReTurnUserInfo{}
	err, departinfo := GetAllDepartNameByDepartIds(corpid, departids)

	for _, each_u := range u_info {
		departs := utils.SqlStringValue(each_u.DepartId)
		is_exist := false
		tempdepartinfo := map[int]string{}
		for _, each_id := range departs {
			if departids != nil && departinfo[each_id] != "" {
				is_exist = true
				tempdepartinfo[each_id] = departinfo[each_id]
			}
		}
		if is_exist {
			each_info := ReTurnUserInfo{
				Userid:      each_u.UserId,
				Corpid:      corpid,
				Name:        each_u.Name,
				Mobile:      each_u.Mobile,
				OpenId:      each_u.OpenId,
				Departid:    departs,
				Departments: tempdepartinfo,
			}
			wxu_info = append(wxu_info, each_info)
		}
	}
	return err, wxu_info, count
}

func Search(data, corpid string, depart_id int) map[string]interface{} {
	o := orm.NewOrm()
	u_sql := fmt.Sprintf("select * from user_info where corpid = '%s' and ( userid like '%%%s%%' or name like '%%%s%%' "+
		" or name_pinyin like '%%%s%%' or mobile like '%%%s%%' )", corpid, data, data, data, data)
	if depart_id != 0 {
		u_sql += " and ("
		_, son_departids := GetAllSonDepartment(corpid, depart_id)
		son_departids = append(son_departids, depart_id)
		for index, each := range son_departids {
			if index == 0 {
				u_sql += fmt.Sprintf(" %d = any(departid)", each)
			} else {
				u_sql += fmt.Sprintf(" or %d = any(departid)", each)
			}
		}
		u_sql += " )"
	}
	de := []Departs{}
	if depart_id == 0 {
		d_infos := []Department{}
		d_sql := fmt.Sprintf("select * from department where  corpid = '%s' "+
			"and (department like  '%%%s%%' or department_pinyin like '%%%s%%')", corpid, data, data)
		o.Raw(d_sql).QueryRows(&d_infos)
		for _, each := range d_infos {
			de = append(de,
				Departs{
					DepartId:   fmt.Sprintf("%d", each.DepartmentId),
					Department: each.Department,
					ParentId:   each.ParentId,
					IsDepart:   true,
				})
		}
	}
	u_infos := []UserInfo{}
	o.Raw(u_sql).QueryRows(&u_infos)
	departids := []int{}
	for _, each := range u_infos {
		de_ids := utils.StringValueToIntArray(each.DepartId)
		departids = append(departids, de_ids...)
	}
	departids = utils.RemoveRepeatInt(departids)
	wxu_info := []ReTurnUserInfo{}
	_, departinfo := GetAllDepartNameByDepartIds(corpid, departids)
	for _, each_u := range u_infos {
		departs := utils.SqlStringValue(each_u.DepartId)
		is_exist := false
		tempdepartinfo := map[int]string{}
		for _, each_id := range departs {
			if departids != nil && departinfo[each_id] != "" {
				is_exist = true
				tempdepartinfo[each_id] = departinfo[each_id]
			}
		}
		if is_exist {
			each_info := ReTurnUserInfo{
				Userid:      each_u.UserId,
				Corpid:      corpid,
				Name:        each_u.Name,
				Mobile:      each_u.Mobile,
				OpenId:      each_u.OpenId,
				Departid:    departs,
				Departments: tempdepartinfo,
			}
			wxu_info = append(wxu_info, each_info)
		}
	}

	return map[string]interface{}{
		"user":   wxu_info,
		"depart": de,
	}
}

func (this *UserInfo) GetAllUserInfoByUserId() []string {
	o := orm.NewOrm()
	u_info := UserInfo{}
	userids := []string{}
	o.QueryTable("user_info").Filter("corpid", this.CorpId).Filter("userid", this.UserId).One(&u_info)
	if u_info.UserId == "" {
		return userids
	}
	departids := utils.StringValueToIntArray(u_info.DepartId)
	sql := "select * from user_info "
	for index, each := range departids {
		if index == 0 {
			sql += fmt.Sprintf("where %d = any(departid)", each)
		} else {
			sql += fmt.Sprintf(" or %d = any(departid)", each)
		}
	}
	u_infos := []UserInfo{}
	o.Raw(sql).QueryRows(&u_infos)
	for _, each := range u_infos {
		userids = append(userids, each.UserId)
	}
	return userids
}

func UseridsAndDepartidsToOpenIds(userids []string, departids []int, corpid string) []string {
	sql := fmt.Sprintf("select * from user_info where corpid = '%s' ", corpid)
	if len(departids) > 0 {
		sql += " and ("
		for index, each := range departids {
			if index == 0 {
				sql += fmt.Sprintf("%d = any(departid)", each)
			} else {
				sql += fmt.Sprintf("or %d = any(departid)", each)
			}
		}
		sql += ")"
	}
	o := orm.NewOrm()
	u_infos := []UserInfo{}
	o.Raw(sql).QueryRows(&u_infos)
	openids := []string{}
	for _, each := range u_infos {
		if each.OpenId != "" {
			openids = append(openids, each.OpenId)
		}
	}
	o.QueryTable("user_info").Filter("corpid", corpid).Filter("userid__in", userids).All(&u_infos)
	for _, each := range u_infos {
		if each.OpenId != "" {
			openids = append(openids, each.OpenId)
		}
	}
	openids = utils.RemoveRepByMap(openids)
	return openids
}
