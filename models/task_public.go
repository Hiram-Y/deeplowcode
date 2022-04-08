package models

import (
	"DeepWorkload/utils"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	"time"
)

type TaskPublic struct {
	Code             string            `orm:"column(code);pk"`
	StartDate        int64             `orm:"column(start_date)"`
	EndDate          int64             `orm:"column(end_date)"`
	State            int               `orm:"column(state)"`
	Remark           string            `orm:"column(remark)"`
	PubScope         string            `orm:"column(pub_scope);type(json)"`
	PubScopeInfo     []FrontDepartInfo `orm:"-"`
	RealPubScope     string            `orm:"column(real_pub_scope);type(json)"`
	RealPubScopeInfo UserAndDepart     `orm:"-"`
	CreateTime       time.Time         `orm:"auto_now_add;type(datetime);column(create_time)"`
}

func (this *TaskPublic) TableName() string {
	return "task_public"
}

func (this *TaskPublic) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(this)
	return err
}

func (this *TaskPublic) GetTaskPubInfo() (err error, t_info TaskPublic) {
	o := orm.NewOrm()
	err = o.QueryTable("task_public").Filter("code", this.Code).One(&t_info)
	if t_info.Code == "" {
		return nil, t_info
	}
	json.Unmarshal([]byte(t_info.PubScope), &t_info.PubScopeInfo)
	return err, t_info
}

func (this *TaskPublic) UpdatePubInfo(corpid string) (err error) {
	o := orm.NewOrm()
	t_info := TaskPublic{}
	o.QueryTable("task_public").Filter("code", this.Code).One(&t_info)
	if this.PubScope != "" {
		json.Unmarshal([]byte(this.PubScope), &this.PubScopeInfo)
		this.RealPubScopeInfo = FrontDepartInfoToReal(this.PubScopeInfo, corpid)
		real, _ := json.Marshal(this.RealPubScopeInfo)
		this.RealPubScope = string(real)
	}
	if this.EndDate != 0 {
		this.EndDate = this.EndDate / 1000
	}
	if this.StartDate != 0 {
		this.StartDate = this.StartDate / 1000
	}
	if t_info.Code == "" {
		_, err = o.Insert(this)
	} else {
		pa := orm.Params{}
		if this.Remark != "" {
			pa["remark"] = this.Remark
		}
		if this.StartDate != 0 {
			pa["start_date"] = this.StartDate
		}
		if this.EndDate != 0 {
			pa["end_date"] = this.EndDate
		}
		if this.PubScope != "" {
			pa["pub_scope"] = this.PubScope
			pa["real_pub_scope"] = this.RealPubScope
		}
		pa["state"] = this.State
		_, err = o.QueryTable("task_public").Filter("code", this.Code).Update(pa)
	}
	if IsNoLastInsertIdError(err) {
		return nil
	}
	return err
}

type FullPubTaskInfo struct {
	Code         string `orm:"column(code)"`
	Title        string `orm:"column(title)"`
	TypeId       int    `orm:"column(type_id)"`
	TypeContent  string
	Icon         string `orm:"column(icon);type(json)"`
	Icons        Icons
	StartDate    int64  `orm:"column(start_date)"`
	EndDate      int64  `orm:"column(end_date)"`
	Remark       string `orm:"column(remark)"`
	PubScope     string `orm:"column(pub_scope)"`
	PubScopeInfo []FrontDepartInfo
}

func GetAllPubTask(corpid, userid string, page_size, page_index int) (err error, fp_info []FullPubTaskInfo, count int) {
	o := orm.NewOrm()
	u_info := UserInfo{}
	o.QueryTable("user_info").Filter("userid", userid).
		Filter("corpid", corpid).One(&u_info)
	if u_info.UserId == "" {
		return err, []FullPubTaskInfo{}, 0
	}
	u_departids := utils.StringValueToIntArray(u_info.DepartId)
	sql := "select task.code as code, task.title as title, task.type_id as type_id, task.icon as icon, task_public.start_date as start_date, " +
		"task_public.end_date as end_date, task_public.remark as remark, task_public.pub_scope as pub_scope from task, task_public where "
	c_sql := "select count(*) as count from task, task_public where "
	w_sql := ""

	w_sql += fmt.Sprintf(" task.corpid = '%s' and ( task_public.real_pub_scope :: jsonb -> 'userid' # '%s' ", corpid, userid)
	for _, each := range u_departids {
		w_sql += fmt.Sprintf(" or task_public.real_pub_scope :: jsonb -> 'departid' # '%s'", fmt.Sprint(each))
	}
	w_sql += fmt.Sprintf(" ) and task_public.state = 1 and task.code= task_public.code")
	c := Counts{}
	o.Raw(c_sql + w_sql).QueryRow(&c)
	w_sql += fmt.Sprintf(" order by task_public.create_time desc")
	w_sql += fmt.Sprintf(" limit %d offset %d", page_size, (page_index-1)*page_size)
	_, err = o.Raw(sql + w_sql).QueryRows(&fp_info)
	mti_info := TypeInfoById(corpid, "", "任务类型")
	for index, each := range fp_info {
		each = formatTaskPublic(each, mti_info)
		fp_info[index] = each
	}
	return err, fp_info, c.Count
}

func formatTaskPublic(ft_info FullPubTaskInfo, mti_info map[int]TypeInfo) FullPubTaskInfo {
	json.Unmarshal([]byte(ft_info.PubScope), &ft_info.PubScopeInfo)
	json.Unmarshal([]byte(ft_info.Icon), &ft_info.Icons)
	ft_info.TypeContent = mti_info[ft_info.TypeId].Content
	if ft_info.TypeId == 2 {
		ft_info.TypeContent = "其他"
	}
	return ft_info
}
