package controllers

import (
	"DeepWorkload/models"
	"fmt"
	"github.com/astaxie/beego"
	"strconv"
	"strings"
)

type DepartmentController struct {
	beego.Controller
}

func (this *DepartmentController) GetAllDepartment() {
	corpid := this.GetString("corpid")
	TreeData, err := models.GetDepartmentInfo(corpid)
	if err == nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"errmsg":  "OK",
			"data":    TreeData,
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": err.Error(),
		}
	}
	this.ServeJSON()
}

func (this *DepartmentController) AddDepartment() {
	corpid := this.GetString("corpid")
	department := this.GetString("department")
	parentid, _ := strconv.Atoi(this.GetString("parentid"))
	departinfo := &models.Department{
		Corpid:     corpid,
		Department: department,
		ParentId:   parentid,
	}
	err := departinfo.AddDepartment()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"errmsg":  err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"errmsg":  "OK",
		}
	}
	this.ServeJSON()
}

func (this *DepartmentController) DelDepartment() {
	corpid := this.GetString("corpid")
	department_id, _ := strconv.Atoi(this.GetString("department_id"))
	departinfo := &models.Department{
		Corpid:       corpid,
		DepartmentId: department_id,
	}
	err := departinfo.DelDepartment()
	if err != nil {
		if strings.Contains(err.Error(), "can not be deleted") {
			this.Data["json"] = map[string]interface{}{
				"errcode": ErrDepartDEl,
				"message": err.Error(),
			}
		} else {
			this.Data["json"] = map[string]interface{}{
				"errcode": ErrDataBase,
				"message": err.Error(),
			}
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"errmsg":  "OK",
		}
	}
	this.ServeJSON()
}

func (this *DepartmentController) EditDepartment() {
	corpid := this.GetString("corpid")
	department_id, _ := strconv.Atoi(this.GetString("department_id"))
	parent_id := this.GetString("parent_id")
	department := this.GetString("department")
	departinfo := &models.Department{
		Corpid:       corpid,
		DepartmentId: department_id,
	}
	err := departinfo.UpdateDepartment(parent_id, department)
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"errmsg":  "OK",
		}
	}
	this.ServeJSON()
}

func (this *DepartmentController) GetAllDepartmentAndUser() {
	corpid := this.GetString("corpid")
	TreeData, err := models.GetDepartmentInfo(corpid)
	Userinfos := models.GetAllUserList(corpid)
	FTreeData := userJoinDepart(TreeData, Userinfos)
	if err == nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"errmsg":  "OK",
			"data":    FTreeData,
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": err.Error(),
		}
	}
	this.ServeJSON()
}

func userJoinDepart(Tree []models.Departs, Userinfos []models.FUserInfo) []models.Departs {
	for index, each := range Tree {
		for _, each_u := range Userinfos {
			departid, _ := strconv.Atoi(each.DepartId)
			if each_u.Departid == departid {
				d_info := models.Departs{
					Id:         each_u.Userid,
					DepartId:   fmt.Sprintf("%s_%d", each_u.Userid, departid),
					Department: each_u.Name,
					ParentId:   departid,
				}
				each.Children = append(each.Children, d_info)
			}
			Tree[index].Children = each.Children
		}

		userJoinDepart(each.Children, Userinfos)
	}
	return Tree
}
