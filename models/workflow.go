package models

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	uuid "github.com/satori/go.uuid"
)

type Workflow struct {
	WorkflowCode   string `orm:"column(workflow_code);pk"`
	Name           string `orm:"column(name)"`
	TypeId         int    `orm:"column(type_id)"`
	FlowPermission string `orm:"column(flow_permission);type(json)"`
}

type WorkFlowDef struct {
	WorkflowCode string `json:"workflow_code"`
	Name         string `json:"name"`
	Type         int    `json:"type"`
}

type FlowPermission struct {
	Type     int         `json:"type"`
	TargetId interface{} `json:"targetId"`
}

type WorkflowJson struct {
	WorkflowDef     WorkFlowDef    `json:"workFlowDef"`
	FlowPermissions FlowPermission `json:"flow_permission"`
	NodeConfig      NodesTree      `json:"nodeConfig"`
}

func (this *Workflow) TableName() string {
	return "workflow"
}

func WorkflowJsonToDB(workflow_json string) string {
	wfj := WorkflowJson{}
	err := json.Unmarshal([]byte(workflow_json), &wfj)
	if err != nil {
		fmt.Println("workflow json unmarshal error")
		fmt.Println(err.Error())
	}
	if wfj.WorkflowDef.WorkflowCode != "" {
		o := orm.NewOrm()
		o.QueryTable("workflow").Filter("workflow_code", wfj.WorkflowDef.WorkflowCode).Delete()
		o.QueryTable("node_config").Filter("workflow_code", wfj.WorkflowDef.WorkflowCode).Delete()
	} else {
		wfj.WorkflowDef.WorkflowCode = uuid.NewV4().String()
	}
	fp, _ := json.Marshal(wfj.FlowPermissions)
	wf := Workflow{
		WorkflowCode:   wfj.WorkflowDef.WorkflowCode,
		Name:           wfj.WorkflowDef.Name,
		TypeId:         wfj.WorkflowDef.Type,
		FlowPermission: string(fp),
	}
	wf.Insert()
	NodeTreeToInsertDB(wfj.NodeConfig, wfj.WorkflowDef.WorkflowCode)
	return wfj.WorkflowDef.WorkflowCode
}

func (this *Workflow) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(this)
	return err
}

func CheckWorkflowTypeByTaskCode(task_code string) (error, Workflow) {
	o := orm.NewOrm()
	info := Workflow{}
	task := Task{}
	err := o.QueryTable("task").Filter("code", task_code).One(&task)
	if task.WorkflowCode == "" {
		return nil, info
	}
	err = o.QueryTable("workflow").Filter("workflow_code", task.WorkflowCode).One(&info)
	return err, info
}
