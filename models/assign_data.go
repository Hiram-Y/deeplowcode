package models

import (
	"DeepWorkload/utils"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	"math"
	"math/rand"
	"time"
)

type AssignData struct {
	Id         int       `orm:"column(id)"`
	ToUserId   string    `orm:"column(to_userid)"`
	ToUserName string    `orm:"-"`
	ToValue    float64   `orm:"column(to_value)"`
	CreateTime time.Time `orm:"auto_now_add;type(datetime);column(create_time)"`
	DataCode   string    `orm:"column(data_code)"`
	FieldCode  string    `orm:"column(field_code)"`
	FormCode   string    `orm:"column(form_code)"`
	TaskCode   string    `orm:"column(task_code)"`
}

func (this *AssignData) TableName() string {
	return "assign_data"
}

func (this *AssignData) InsertOneAssignData() error {
	o := orm.NewOrm()
	_, err := o.Insert(this)
	go this.CheckAssignDataAndUpdateMainDataState()
	return err
}

func InsertMultiAssignData(ass_data []AssignData) {
	o := orm.NewOrm()
	if len(ass_data) > 0 {
		o.InsertMulti(len(ass_data), ass_data)
		ass_data[0].CheckAssignDataAndUpdateMainDataState()
	}
}

func (this *AssignData) CheckAssignDataAndUpdateMainDataState() {
	o := orm.NewOrm()
	m_info := MainData{}
	err := o.QueryTable("main_data").Filter("code", this.DataCode).One(&m_info)
	f_infos := []FormFieldInfo{}
	err = json.Unmarshal([]byte(m_info.FormFieldInfo), &f_infos)
	if err != nil {
		fmt.Println(err.Error())
	}
	as_infos := []AssignData{}
	asf_infos := map[string]float64{}
	_, err = o.QueryTable("assign_data").Filter("data_code", this.DataCode).All(&as_infos)
	for _, each := range as_infos {
		asf_infos[each.FieldCode] += each.ToValue
	}
	is_assign_done := true
	for _, each := range f_infos {
		for k, v := range asf_infos {
			if each.FieldCode == k {
				if each.EnableGreater == true {
					if each.Value.(float64) > v {
						is_assign_done = false
						break
					}
				} else {
					if each.Value.(float64) != v {
						is_assign_done = false
						break
					}
				}
			}
		}
	}
	if is_assign_done == true {
		o.QueryTable("main_data").Filter("code", this.DataCode).Update(orm.Params{"assign_state": 2})
	} else {
		o.QueryTable("main_data").Filter("code", this.DataCode).Update(orm.Params{"assign_state": 1})
	}
}

func formateAssignDataUserName(ass_data []AssignData) []AssignData {
	userids := []string{}
	for _, each := range ass_data {
		userids = append(userids, each.ToUserId)
	}
	if len(userids) == 0 {
		return []AssignData{}
	}
	o := orm.NewOrm()
	u_infos := []UserInfo{}
	o.QueryTable("user_info").Filter("userid__in", userids).All(&u_infos)
	for index, each := range ass_data {
		for _, each_u := range u_infos {
			if each.ToUserId == each_u.UserId {
				each.ToUserName = each_u.Name
			}
		}
		ass_data[index] = each
	}
	return ass_data
}

type AssignDataF struct {
	AssignData
	FormCode       string `orm:"column(form_code)"`
	FormName       string `orm:"column(form_name)"`
	FieldName      string `orm:"column(field_name)"`
	FromUserId     string `orm:"column(from_userid)"`
	FormUserName   string `orm:"column(from_username)"`
	State          int    `orm:"column(main_state)"`
	FormFieldInfo  string `orm:"column(form_field_info);type(json)"`
	FormFieldInfos []FormFieldInfo
}

func AssignDataToUserSelf(task_code, userid, corpid, form_code string) (error, []AssignDataF) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select assign_data.id, assign_data.to_userid, assign_data.to_value, assign_data.field_code,"+
		"assign_data.data_code, assign_data.create_time, main_data.form_code, main_data.create_userid as from_userid , "+
		"main_data.state as main_state, main_data.form_field_info as form_field_info, "+
		"(select name from user_info where main_data.create_userid = user_info.userid and user_info.corpid = main_data.corpid ) as from_username , "+
		"(select title from form_info where code = main_data.form_code) as form_name, "+
		" (select label from form_field where form_code = main_data.form_code and filed_code = assign_data.field_code) as  field_name   "+
		"from assign_data, main_data where assign_data.data_code = main_data.code and main_data.state in (3,5) and main_data.assign_state = 2 and "+
		"main_data.task_code = '%s' and assign_data.to_userid = '%s' and main_data.corpid = '%s'", task_code, userid, corpid)
	if form_code != "" {
		sql += fmt.Sprintf(" and main_data.form_code = '%s'", form_code)
	}
	as_infos := []AssignDataF{}
	_, err := o.Raw(sql).QueryRows(&as_infos)
	for index, each := range as_infos {
		json.Unmarshal([]byte(each.FormFieldInfo), &each.FormFieldInfos)
		as_infos[index] = each
	}

	return err, as_infos
}

type Node struct {
	Color string  `json:"color"`
	Label string  `json:"label"`
	Y     float64 `json:"y"`
	X     float64 `json:"x"`
	Id    string  `json:"id"`
	Size  float64 `json:"size"`
	Count int
}

type Edge struct {
	SourceID string `json:"sourceID"`
	TargetID string `json:"targetID"`
}

func sigmoid(x float64) float64 {
	return 1 / (1 + math.Pow(math.E, -x))

}

func GetDependencies(corpid, userid string) ([]Node, []Edge) {
	o := orm.NewOrm()
	colors := []string{"#33B5E5", "#0099CC", "#AA66CC", "#9933CC", "#99CC00", "#669900", "#FFBB33", "#FF8800", "#FF4444", "#CC0000"}
	wh := []float64{1000.00, 1000.00}
	m_datas := []MainData{}
	as_datas := []AssignData{}
	qs := o.QueryTable("main_data").Filter("corpid", corpid)
	qs.All(&m_datas)
	task_code := []string{}
	uids := []string{}
	for _, each := range m_datas {
		task_code = append(task_code, each.TaskCode)
		uids = append(uids, each.CreateUserId)
	}
	task_code = utils.DeleteRepeat(task_code)
	o.QueryTable("assign_data").Filter("task_code__in", task_code).All(&as_datas)
	for _, each := range as_datas {
		uids = append(uids, each.ToUserId)
	}
	uids = utils.DeleteRepeat(uids)
	um := GetUserNameMapByUseridAndCorpId(uids, corpid)
	nodes := map[string]Node{}
	edges := map[string]Edge{}
	rand.Seed(time.Now().Unix())
	for _, each_m := range m_datas {
		for _, each_a := range as_datas {
			if each_m.Code == each_a.DataCode {
				_, m_ex := nodes[each_m.CreateUserId]
				_, a_ex := nodes[each_a.ToUserId]
				if !m_ex {
					nodes[each_m.CreateUserId] = Node{
						Color: colors[rand.Intn(len(colors))],
						Label: um[each_m.CreateUserId],
						Y:     rand.Float64() * wh[0],
						X:     rand.Float64() * wh[1],
						Id:    each_m.CreateUserId,
						Count: 1,
					}
				} else {
					if each_m.CreateUserId != each_a.ToUserId {
						temp := nodes[each_m.CreateUserId]
						temp.Count += 1
						nodes[each_m.CreateUserId] = temp
					}
				}
				if !a_ex {
					nodes[each_a.ToUserId] = Node{
						Color: colors[rand.Intn(len(colors))],
						Label: um[each_a.ToUserId],
						Y:     rand.Float64() * wh[0],
						X:     rand.Float64() * wh[1],
						Id:    each_a.ToUserId,
						Count: 1,
					}
				} else {
					if each_m.CreateUserId != each_a.ToUserId {
						temp := nodes[each_a.ToUserId]
						temp.Count += 1
						nodes[each_a.ToUserId] = temp
					}
				}

				if each_m.CreateUserId != each_a.ToUserId {
					edges[each_m.CreateUserId+each_a.ToUserId] = Edge{
						TargetID: each_a.ToUserId,
						SourceID: each_m.CreateUserId,
					}
				}
			}
		}
	}
	re_nodes := []Node{}
	re_edges := []Edge{}
	if userid != "" {
		for _, each := range edges {
			if each.SourceID == userid || each.TargetID == userid {
				re_edges = append(re_edges, each)
			}
		}
		for key, each := range nodes {
			for _, each_e := range re_edges {
				if key == each_e.TargetID || key == each_e.SourceID {
					re_nodes = append(re_nodes, each)
					break
				}
			}
		}
	} else {
		for _, each := range nodes {
			re_nodes = append(re_nodes, each)
		}

		for _, each := range edges {
			re_edges = append(re_edges, each)
		}
	}
	//size := 250/float64(len(nodes))
	for index, each := range re_nodes {
		//each.Size = float64(each.Count) * size
		//if each.Size < 10{
		//	each.Size = 10
		//}else if each.Size > 300 {
		//	each.Size = 300
		//}
		each.Size = 30
		re_nodes[index] = each
	}
	return re_nodes, re_edges
}
