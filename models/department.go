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

type Department struct {
	Id               int    `orm:"pk;auto;column(id)"`
	Corpid           string `orm:"column(corpid)"`
	DepartmentId     int    `orm:"column(department_id)"`
	Department       string `orm:"column(department)"`
	ParentId         int    `orm:"column(parent_id)"`
	DepartmentPinYin string `orm:"column(department_pinyin)"`
}

type Departs struct {
	Id         string `json:"id"`
	DepartId   string `json:"depart_id"`
	Department string
	IsDepart   bool
	ParentId   int
	Children   []Departs
}
type Counts struct {
	Count int `orm:"column(count)"`
}

func (this *Department) TableName() string {
	return "department"
}

func (this *Department) AddDepartment() error {
	o := orm.NewOrm()
	de_info := Department{}
	err := o.QueryTable("department").Filter("corpid", this.Corpid).OrderBy("-department_id").One(&de_info)
	if err == nil {
		this.DepartmentId = de_info.DepartmentId + 1
	}
	this.DepartmentPinYin, _ = pinyin.New(this.Department).Split("").Convert()
	_, err = o.Insert(this)
	con, _ := redisgo.New(redisgo.Options{})
	cache_key := this.Corpid + "department"
	con.Del(cache_key)
	return err
}

func (this *Department) DelDepartment() error {
	o := orm.NewOrm()

	is_exist := o.QueryTable("department").Filter("corpid", this.Corpid).Filter("parent_id", this.DepartmentId).Exist()
	sql := fmt.Sprintf("select count(*) as count from user_info where corpid = '%s' and %d = ANY(departid)", this.Corpid, this.DepartmentId)
	c_info := Counts{}
	o.Raw(sql).QueryRow(&c_info)

	if (!is_exist) && (c_info.Count == 0) && (this.DepartmentId != 1) {
		_, err := o.QueryTable("department").Filter("corpid", this.Corpid).Filter("department_id", this.DepartmentId).Delete()
		con, _ := redisgo.New(redisgo.Options{})
		cache_key := this.Corpid + "department"
		con.Del(cache_key)
		return err
	}
	return errors.New("can not be deleted")
}

func (this *Department) UpdateDepartment(parentid, department string) error {
	o := orm.NewOrm()
	op := orm.Params{}
	if parentid != "" {
		op["parent_id"] = parentid
	}
	if department != "" {
		op["department"] = department
		this.DepartmentPinYin, _ = pinyin.New(this.Department).Split("").Convert()
		op["department_pinyin"] = this.DepartmentPinYin
	}
	con, _ := redisgo.New(redisgo.Options{})
	cache_key := this.Corpid + "department"
	con.Del(cache_key)
	_, err := o.QueryTable("department").Filter("corpid", this.Corpid).Filter("department_id", this.DepartmentId).Update(op)
	return err
}

func GetAllDepartmentInfos(corpid string) (de_info []Department) {
	o := orm.NewOrm()
	o.QueryTable("department").
		Filter("corpid", corpid).All(&de_info)
	return de_info
}

func GetDepartmentInfo(corpid string) (tree []Departs, err error) {
	o := orm.NewOrm()
	con, _ := redisgo.New(redisgo.Options{})
	cache_key := corpid + "department"
	err = con.GetObject(cache_key, &tree)
	if err == nil {
		return tree, err
	}
	var departs_info []Department
	_, err = o.QueryTable("department").
		Filter("corpid", corpid).All(&departs_info)
	if err == nil {
		parent_infos := getParentListData(departs_info)
		Treedata := data2TreeData(departs_info, parent_infos)
		con.Set(cache_key, Treedata, 0)
		return Treedata, nil
	}
	return nil, err
}

func getParentListData(departs_info []Department) (parent_infos []Departs) {
	for _, value := range departs_info {
		if value.ParentId == 0 {
			temp := Departs{
				Id:         strconv.Itoa(value.DepartmentId),
				ParentId:   value.ParentId,
				DepartId:   strconv.Itoa(value.DepartmentId),
				Department: value.Department,
				IsDepart:   true,
			}
			parent_infos = append(parent_infos, temp)
		}
	}
	return parent_infos
}

func data2TreeData(departs_info []Department, parent_infos []Departs) (TreeData []Departs) {
	for p, value := range parent_infos {
		childrenArray := []Departs{}
		for _, d_value := range departs_info {
			departid, _ := strconv.Atoi(value.DepartId)
			if d_value.ParentId == departid {
				t_depart := Departs{
					Id:         strconv.Itoa(d_value.DepartmentId),
					ParentId:   d_value.ParentId,
					DepartId:   strconv.Itoa(d_value.DepartmentId),
					Department: d_value.Department,
					IsDepart:   true,
				}
				childrenArray = append(childrenArray, t_depart)
			}
		}
		parent_infos[p].Children = childrenArray
		if len(childrenArray) > 0 {
			data2TreeData(departs_info, childrenArray)
		}
	}

	return parent_infos
}

func getAllSonDepartment(department_id int, departs []Department) (err error, depart_ids []int) {
	depart_infos := []Department{}
	for _, each := range departs {
		if each.ParentId == department_id {
			depart_infos = append(depart_infos, each)
		}
	}
	for _, each := range depart_infos {
		depart_ids = append(depart_ids, each.DepartmentId)
		if len(depart_ids) > 0 {
			_, sondeparts := getAllSonDepartment(each.DepartmentId, departs)
			depart_ids = append(depart_ids, sondeparts...)
		}
	}
	return err, depart_ids
}

func GetAllSonDepartment(corpid string, department_id int) (err error, departs []int) {
	o := orm.NewOrm()
	depart_infos := []Department{}

	o.QueryTable("department").Filter("corpid", corpid).All(&depart_infos)
	err, departs = getAllSonDepartment(department_id, depart_infos)
	return err, departs
}

func GetAllSonDepartmentId(corpid string, departids []int) []int {
	o := orm.NewOrm()
	depart_infos := []Department{}
	o.QueryTable("department").Filter("corpid", corpid).All(&depart_infos)
	departs := []int{}
	for _, each := range departids {
		_, temp := getAllSonDepartment(each, depart_infos)
		departs = append(departs, temp...)
	}
	departid := []int{}
	for _, each := range departs {
		departid = append(departid, each)
	}
	return utils.RemoveRepeatInt(departid)
}

func GetAllDepartNameByDepartIds(corpid string, departids []int) (err error, departs map[int]string) {
	o := orm.NewOrm()
	depart_info := []Department{}
	if len(departids) == 0 {
		return nil, nil
	}
	o.QueryTable("department").Filter("corpid", corpid).Filter("department_id__in", departids).All(&depart_info)
	departs = map[int]string{}
	for _, each := range depart_info {
		departs[each.DepartmentId] = each.Department
	}
	return err, departs
}

func GetDepartIdsByDepartName(departname string, depart_infos []Department) int {
	de := Department{}
	departnames := strings.Split(departname, "/")
	for index, each := range departnames {
		is_in := false
		for _, each_d := range depart_infos {
			if each_d.Department == each && each_d.ParentId == de.DepartmentId {
				de = each_d
				is_in = true
				break
			}
		}
		if index == len(departname)-1 && is_in == true {
			return de.DepartmentId
		}
	}
	return 0
}
