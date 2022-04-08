package controllers

import (
	"DeepWorkload/lib/pinyin"
	"DeepWorkload/models"
	"DeepWorkload/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego"
	"os"
	"strconv"
	"strings"
)

type UserController struct {
	beego.Controller
}

func (this *UserController) CheckUserId() {
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	u_info := &models.UserInfo{
		CorpId: corpid,
		UserId: userid,
	}
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data":    u_info.CheckUserIdExistInThisCorp(),
	}
	this.ServeJSON()
}

func (this *UserController) CheckMobile() {
	corpid := this.GetString("corpid")
	mobile := this.GetString("mobile")
	u_info := &models.UserInfo{
		CorpId: corpid,
		Mobile: mobile,
	}
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data":    u_info.CheckMobileExistInThisCorp(),
	}
	this.ServeJSON()
}

func (this *UserController) GetOneUserInfoDetail() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	u_info := &models.UserInfo{
		UserId: userid,
		CorpId: corpid,
	}
	err, info := u_info.GetUserInfoDetail()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    info,
		}
	}
	this.ServeJSON()

}

func (this *UserController) AddOneUser() {
	departids := this.GetString("departids")
	d_ids := []int{}
	json.Unmarshal([]byte(departids), &d_ids)

	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	username := this.GetString("username")
	mobile := this.GetString("mobile")
	uInfo := &models.UserInfo{
		CorpId:    corpid,
		UserId:    userid,
		Name:      username,
		DepartIds: d_ids,
		Mobile:    mobile,
	}
	err := uInfo.AddUserInfo()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *UserController) AddUserInfoByExcel() {
	doc_code := this.GetString("doc_code")
	corpid := this.GetString("corpid")
	departid, _ := this.GetInt("departid")
	err, docInfo := models.GetDocinfo(doc_code)
	fmt.Println(err)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": "没有该文件的Code",
		}
	} else {
		err, infos := getUserInfoFromExcel(docInfo.Path, corpid, departid)
		if err != nil {
			code := ErrDataBase
			if strings.Contains(err.Error(), "args error may be empty") {
				code = ErrExcelMaybeEmpty
			}
			this.Data["json"] = map[string]interface{}{
				"errcode": code,
				"message": err.Error(),
			}
		} else {
			userinfo := &models.UserInfo{
				CorpId: corpid,
			}
			_, uinfos := userinfo.GetAllUserList()
			if isUserIdrep(infos, uinfos) {
				this.Data["json"] = map[string]interface{}{
					"errcode": ErrUserReap,
					"message": "userid repeat",
				}
			} else {
				err = models.AddUserInfo(infos)
				this.Data["json"] = map[string]interface{}{
					"errcode": OK,
					"message": "OK",
				}
			}
		}
	}
	this.ServeJSON()
}

func (this *UserController) ExportAllUserInfo() {
	corpid := this.GetString("corpid")
	a_user := &models.UserInfo{
		CorpId: corpid,
	}
	f := excelize.NewFile()
	sheetName := "通信录"
	index := f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(index)

	titles := []string{"工号", "姓名", "手机号", "部门"}
	for index, each := range titles {
		file_index := fmt.Sprintf("%s%d", string(65+index), 1)
		f.SetCellStr(sheetName, file_index, each)
	}
	_, u_infos := a_user.GetAllUserList()
	depart_infos := models.GetAllDepartmentInfos(corpid)
	for index, each := range u_infos {
		u_departs := utils.SqlStringValue(each.DepartId)
		departs := []string{}
		for _, each_d := range u_departs {
			classname := getalldepartname(each_d, depart_infos, []string{})
			departs = append(departs, strings.Join(classname, "/"))
		}
		depart_info := strings.Join(departs, "，")
		f.SetCellStr(sheetName, fmt.Sprintf("%s%d", "A", index+2), each.UserId)
		f.SetCellStr(sheetName, fmt.Sprintf("%s%d", "B", index+2), each.Name)
		f.SetCellStr(sheetName, fmt.Sprintf("%s%d", "C", index+2), each.Mobile)
		f.SetCellStr(sheetName, fmt.Sprintf("%s%d", "D", index+2), depart_info)
	}
	file_name := "通信录" + corpid + ".xlsx"
	path := "uploads/temp/" + file_name
	f.SaveAs(path)
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data":    "/temp/" + file_name,
	}
	this.ServeJSON()
}

func isUserIdrep(infos, tempinfos []models.UserInfo) bool {
	userids := []string{}
	t_userids := []string{}
	for _, each := range infos {
		userids = append(userids, each.UserId)
	}
	for _, each := range tempinfos {
		t_userids = append(t_userids, each.UserId)
	}

	for _, each_u := range userids {
		for _, each_t_u := range t_userids {
			if each_u == each_t_u {
				return true
			}
		}
	}
	return false
}

func getalldepartname(departid int, d_infos []models.Department, class_name []string) []string {
	if departid == 0 {
		return class_name
	} else {
		var parent_id int
		for _, each := range d_infos {
			if each.DepartmentId == departid {
				class_name = append(class_name, each.Department)
				parent_id = each.ParentId
			}
		}
		return getalldepartname(parent_id, d_infos, class_name)
	}
}

func getUserInfoFromExcel(path, corpid string, departid int) (error, []models.UserInfo) {
	f, err := excelize.OpenFile(path)
	datas := []models.UserInfo{}
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	rows := f.GetRows("Sheet1")
	Userids := []string{}
	Mobiles := []string{}
	depart_infos := models.GetAllDepartmentInfos(corpid)
	for _, row := range rows[2:] {
		if row[0] != "" {
			Userids = append(Userids, row[0])
			Mobiles = append(Mobiles, row[2])
			depart_desc := strings.Split(row[3], "；")
			departids := []int{}
			for _, each := range depart_desc {
				de_id := models.GetDepartIdsByDepartName(each, depart_infos)
				departids = append(departids, de_id)
			}
			DeId := ""
			if len(departids) == 0 || utils.IsExistInt(departids, 0) {
				DeId = fmt.Sprintf("{%d}", departid)
			} else {
				departidstr := []string{}
				for _, each := range departids {
					departidstr = append(departidstr, fmt.Sprintf("%d", each))
				}
				d_ids := strings.Join(departidstr, ",")
				DeId = fmt.Sprintf("{%s}", d_ids)
			}
			name_p, _ := pinyin.New(row[1]).Split("").Convert()
			datas = append(datas, models.UserInfo{
				CorpId:     corpid,
				UserId:     row[0],
				Name:       row[1],
				Mobile:     row[2],
				DepartId:   DeId,
				NamePinyin: name_p,
			})
		}
	}
	Userids = utils.RemoveRepByMap(Userids)
	Mobiles = utils.RemoveRepByMap(Mobiles)
	go os.Remove(path)
	if (len(Userids) == len(datas)) && (len(Mobiles) == len(datas)) {
		return nil, datas
	}

	return errors.New("excel data repeat"), nil
}

func (this *UserController) DelOneUser() {
	self := this.GetString("self")
	corpid := this.GetString("corpid")
	userid := this.GetString("userid")
	if self == userid {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDelAdmin,
			"message": "不能删除自己",
		}
	} else {
		u_info := &models.UserInfo{
			CorpId: corpid,
			UserId: userid,
		}
		err := u_info.DelUserInfo()
		if err != nil {
			this.Data["json"] = map[string]interface{}{
				"errcode": ErrDataBase,
				"message": err.Error(),
			}
		} else {
			this.Data["json"] = map[string]interface{}{
				"errcode": OK,
				"message": "OK",
			}
		}
	}
	this.ServeJSON()
}

func (this *UserController) DelUsers() {
	self := this.GetString("self")
	corpid := this.GetString("corpid")
	userids := this.GetString("userids")
	uids := []string{}
	json.Unmarshal([]byte(userids), &uids)
	self_in := false
	for _, each := range uids {
		if self == each {
			this.Data["json"] = map[string]interface{}{
				"errcode": ErrDelAdmin,
				"message": "不能删除自己",
			}
			self_in = true
			break
		}
	}
	if self_in == false {
		uInfo := &models.UserInfo{
			CorpId: corpid,
		}
		err := uInfo.PatchDelUserInfo(uids)
		if err != nil {
			this.Data["json"] = map[string]interface{}{
				"errcode": ErrDataBase,
				"message": err.Error(),
			}
		} else {
			this.Data["json"] = map[string]interface{}{
				"errcode": OK,
				"message": "OK",
			}
		}
	}
	this.ServeJSON()
}

func (this *UserController) EditUser() {
	departids := this.GetString("departids")
	d_ids := []int{}
	json.Unmarshal([]byte(departids), &d_ids)
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	username := this.GetString("username")
	mobile := this.GetString("mobile")
	uInfo := &models.UserInfo{
		CorpId:    corpid,
		UserId:    userid,
		Name:      username,
		Mobile:    mobile,
		DepartIds: d_ids,
	}
	err := uInfo.UpdateUserInfo()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}

func (this *UserController) GetAllUserByDepartId() {
	corpid := this.GetString("corpid")
	department_id, _ := strconv.Atoi(this.GetString("department_id"))
	page_size, _ := this.GetInt("page_size")
	page_index, _ := this.GetInt("page_index")
	_, departs := models.GetAllSonDepartment(corpid, department_id)
	departs = append(departs, department_id)
	err, u_infos, count := models.GetUserInfoByDepartIds(departs, corpid, page_size, page_index)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data": map[string]interface{}{
				"users": u_infos,
				"count": count,
			},
		}
	}
	this.ServeJSON()
}

func (this *UserController) SearchUser() {
	query_data := this.GetString("query_data")
	depart_id, _ := this.GetInt("depart_id")
	corpid := this.GetString("corpid")
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data":    models.Search(query_data, corpid, depart_id),
	}
	this.ServeJSON()
}

func (this *UserController) DelOneAdminUser() {
	userid := this.GetString("userid")
	admin_userid := this.GetString("admin_userid")
	if userid == admin_userid {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDelAdmin,
			"message": "不能删除自己",
		}
	} else {
		corpid := this.GetString("corpid")
		admin := &models.Admin{
			CorpId: corpid,
			UserId: admin_userid,
		}
		err := admin.DelOne()
		if err != nil {
			this.Data["json"] = map[string]interface{}{
				"errcode": ErrDataBase,
				"message": err.Error(),
			}
		} else {
			this.Data["json"] = map[string]interface{}{
				"errcode": OK,
				"message": "OK",
			}
		}
	}
	this.ServeJSON()
}

func (this *UserController) AddAdminUsers() {
	admin_userids := this.GetString("admin_userids")
	uids := []string{}
	json.Unmarshal([]byte(admin_userids), &uids)
	corpid := this.GetString("corpid")
	admin := &models.Admin{
		CorpId: corpid,
	}
	go admin.PatchInsert(uids)
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
	}
	this.ServeJSON()
}

func (this *UserController) GetAllAdminUserInfo() {
	corpid := this.GetString("corpid")
	ad := &models.Admin{
		CorpId: corpid,
	}
	data := ad.GetAllAdminInfo()
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"message": "OK",
		"data":    data,
	}
	this.ServeJSON()
}

func (this *UserController) GetAllUserNameAndUserId() {
	corpid := this.GetString("corpid")
	u := &models.UserInfo{
		CorpId: corpid,
	}
	err, info := u.GetAllUserNameAndUserId()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"data":    info,
		}
	}
	this.ServeJSON()
}

func (this *UserController) UpdatePersonalData() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	mobile := this.GetString("mobile")
	passwd := this.GetString("passwd")
	change_userid := this.GetString("change_userid")
	name := this.GetString("name")
	u_info := &models.UserInfo{
		UserId: userid,
		CorpId: corpid,
		Mobile: mobile,
		Passwd: passwd,
		Name:   name,
	}
	err := u_info.UpdatePersonalInfo(change_userid)
	if err != nil {
		errcode := ErrDataBase
		if strings.Contains(err.Error(), "没有") {
			errcode = ErrUserUpdate
		}
		this.Data["json"] = map[string]interface{}{
			"errcode": errcode,
			"message": err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
		}
	}
	this.ServeJSON()
}
