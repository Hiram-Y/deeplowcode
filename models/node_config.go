package models

import (
	"DeepWorkload/utils"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	uuid "github.com/satori/go.uuid"
	"strconv"
)

type NodeConfig struct {
	NodeCode      string     `orm:"column(node_code);pk"`
	WorkflowCode  string     `orm:"column(workflow_code)"`
	ParentCode    string     `orm:"column(parent_code)"`
	IsHeader      bool       `orm:"column(is_header)"`
	IsCondition   bool       `orm:"column(is_condition)"`
	NodeName      string     `orm:"column(node_name)" json:"nodeName"`
	UserList      string     `orm:"column(user_list);type(json)"`
	NodeUserList  []NodeUser `orm:"-" json:"nodeUserList"`
	NodeType      int        `orm:"column(node_type)" json:"type"`
	PriorityLevel int        `orm:"column(priority_level)" json:"priorityLevel"`
}

type NodeUser struct {
	TargetId   interface{} `json:"targetId"`
	Type       int         `json:"type"`
	Name       string      `json:"name"`
	Department string      `json:"department"`
}

type NodesTree struct {
	NodeConfig
	ChildNode     *NodesTree  `json:"childNode"`
	ConditionNode []NodesTree `json:"conditionNodes"`
}

func (this *NodeConfig) TableName() string {
	return "node_config"
}

func NodeTreeToInsertDB(nodeTree NodesTree, workflow_code string) error {
	InsertNode(nodeTree, workflow_code, "", false)
	return nil
}

func InsertNode(nodeTree NodesTree, workflow_code, patent_code string, is_condition bool) {
	node_code := uuid.NewV4().String()
	ul, _ := json.Marshal(nodeTree.NodeUserList)
	nc := NodeConfig{
		NodeCode:      node_code,
		WorkflowCode:  workflow_code,
		NodeName:      nodeTree.NodeName,
		UserList:      string(ul),
		NodeType:      nodeTree.NodeType,
		IsHeader:      true,
		PriorityLevel: nodeTree.PriorityLevel,
	}
	if patent_code != "" {
		nc.IsHeader = false
		nc.ParentCode = patent_code
	}
	if is_condition {
		nc.IsCondition = true
	}
	o := orm.NewOrm()
	o.Insert(&nc)
	if nodeTree.ChildNode != nil {
		InsertNode(*nodeTree.ChildNode, workflow_code, node_code, false)
	}

	for _, each := range nodeTree.ConditionNode {
		InsertNode(each, workflow_code, node_code, true)
	}
}

//func TestNodeTreeToArray(nodeTree NodesTree,workflow_code,patent_code string,is_condition bool, ncs *[]NodeConfig)  {
//	node_code := utils.Random4Id()
//	ul,_:=json.Marshal(nodeTree.NodeUserList)
//	nc := NodeConfig{
//		NodeCode:node_code,
//		WorkflowCode:workflow_code,
//		NodeName:nodeTree.NodeName,
//		UserList:string(ul),
//		NodeType:nodeTree.NodeType,
//		IsHeader:true,
//	}
//	if patent_code != ""{
//		nc.IsHeader = false
//		nc.ParentCode = patent_code
//	}
//	if is_condition{
//		nc.IsCondition = true
//	}
//	*ncs = append(*ncs, nc)
//	if nodeTree.ChildNode != nil{
//		TestNodeTreeToArray(*nodeTree.ChildNode,workflow_code,node_code,false,ncs)
//	}
//
//	for _,each := range nodeTree.ConditionNode{
//		TestNodeTreeToArray(each,workflow_code,node_code,true,ncs)
//	}
//}
//
//func TestNodeTreeToUserList(nodeJson string)  {
//	nodeTree := NodesTree{}
//	json.Unmarshal([]byte(nodeJson),&nodeTree)
//	ncs := []NodeConfig{}
//	TestNodeTreeToArray(nodeTree,utils.Random4Id(),"",false,&ncs)
//	p_infos := getParentNode(ncs)
//	tree := workflow2NodeTree(ncs,p_infos)
//	user := QYWXUserInfo{
//		UserId:"150",
//		DepartIds:[]int{324},
//	}
//	userlist := []WorkflowUser{}
//	NodeTreeToUserList(tree,user, &userlist)
//}

func (this *NodeConfig) GetNodeTree() (err error, tree NodesTree) {
	o := orm.NewOrm()
	wf_infos := []NodeConfig{}
	_, err = o.QueryTable("node_config").Filter("workflow_code", this.WorkflowCode).All(&wf_infos)
	if err == nil {
		p_infos := getParentNode(wf_infos)
		tree := workflow2NodeTree(wf_infos, p_infos)
		return nil, tree
	}
	return err, tree
}

func GetWorkflowUserList(user UserInfo, workflow_code string) []MainDataAudit {

	o := orm.NewOrm()
	wf_infos := []NodeConfig{}
	o.QueryTable("node_config").Filter("workflow_code", workflow_code).All(&wf_infos)
	p_infos := getParentNode(wf_infos)
	tree := workflow2NodeTree(wf_infos, p_infos)
	userlist := []MainDataAudit{}
	departs := []Department{}
	o.QueryTable("department").Filter("corpid", user.CorpId).All(&departs)
	NodeTreeToUserList(tree, user, &userlist, departs)
	return userlist
}

func GetWorkflowDepartAndCondition(depart_code, node_code, workflow_code string) []MainDataAudit {
	o := orm.NewOrm()
	wf_infos := []NodeConfig{}
	o.QueryTable("node_config").Filter("workflow_code", workflow_code).All(&wf_infos)
	p_infos := getParentNode(wf_infos)
	tree := workflow2NodeTree(wf_infos, p_infos)
	mapTree := nodeTreeToMap(tree)
	conditions := mapTree[depart_code].(map[string]interface{})["conditions"].([]Condition)
	for _, each := range conditions {
		if each.NodeCode == node_code {
			return each.UserList
		}

	}
	return nil
}

func GetAllWorkflowList(task_code string) (error, map[string]interface{}) {
	o := orm.NewOrm()
	t_info := Task{}
	err := o.QueryTable("task").Filter("code", task_code).One(&t_info)
	if err == nil {
		wf_infos := []NodeConfig{}
		o.QueryTable("node_config").Filter("workflow_code", t_info.WorkflowCode).All(&wf_infos)
		p_infos := getParentNode(wf_infos)
		tree := workflow2NodeTree(wf_infos, p_infos)
		return nil, nodeTreeToMap(tree)
	}
	return err, nil
}

// 每个学院下的多种审批流程
type Condition struct {
	NodeCode      string
	UserList      []MainDataAudit
	ConditionName string
}

func getUserList(tree NodesTree, u_list *[][]NodeUser) {

	err := json.Unmarshal([]byte(tree.UserList), &tree.NodeUserList)
	if err != nil {
		fmt.Println("userlist", err.Error())
	}
	*u_list = append(*u_list, tree.NodeUserList)
	if tree.ChildNode != nil {
		getUserList(*tree.ChildNode, u_list)
	}
}

func nodeUserListToMaindataAudit(u_list [][]NodeUser) []MainDataAudit {
	info := []MainDataAudit{}
	for _, each := range u_list {
		uid_list := []string{}
		for _, u := range each {
			userid := fmt.Sprint(u.TargetId)
			uid_list = append(uid_list, userid)
		}
		m := MainDataAudit{
			UserIds:   uid_list,
			NodeState: 1,
		}
		info = append(info, m)
	}
	return info
}

func nodeTreeToMap(tree NodesTree) map[string]interface{} {

	json.Unmarshal([]byte(tree.UserList), &tree.NodeUserList)
	if tree.IsHeader != true {
		mapNodeTree := map[string]interface{}{}
		if tree.NodeType == 5 {
			for _, each := range tree.ConditionNode {
				nau := []Condition{}
				for _, each_c := range each.ChildNode.ConditionNode {
					u_list := [][]NodeUser{}
					getUserList(*each_c.ChildNode, &u_list)
					nau = append(nau, Condition{
						NodeCode:      each_c.NodeCode,
						ConditionName: each_c.NodeName,
						UserList:      nodeUserListToMaindataAudit(u_list),
					})
				}
				mapNodeTree[each.NodeCode] = map[string]interface{}{
					"depart_name": each.NodeName,
					"conditions":  nau,
				}
			}
			return mapNodeTree
		}
	}
	if tree.ChildNode != nil {
		return nodeTreeToMap(*tree.ChildNode)
	}
	return nil
}

func NodeTreeToUserList(tree NodesTree, user UserInfo, maindata_audit *[]MainDataAudit, departs []Department) {
	json.Unmarshal([]byte(tree.UserList), &tree.NodeUserList)

	if tree.IsHeader != true {
		if len(tree.ConditionNode) == 0 {
			u_infos := []string{}
			for _, each := range tree.NodeUserList {
				if each.Type == 1 {
					u_infos = append(u_infos, each.TargetId.(string))
				}
			}
			if (tree.NodeType != 3) && (tree.NodeType != 4) {
				*maindata_audit = append(*maindata_audit, MainDataAudit{NodeState: 1, UserIds: u_infos})
			}
			if tree.ChildNode != nil {
				NodeTreeToUserList(*tree.ChildNode, user, maindata_audit, departs)
			}
		} else {
			userid := user.UserId
			departid := user.DepartIds[0]
			for _, each := range tree.ConditionNode {
				json.Unmarshal([]byte(each.UserList), &each.NodeUserList)
				for _, each_u := range each.NodeUserList {
					if each_u.Type == 3 {
						switch each_u.TargetId.(type) {
						case string:
							t_departid, _ := strconv.Atoi(each_u.TargetId.(string))
							_, de_ids := getAllSonDepartment(t_departid, departs)
							de_ids = append(de_ids, t_departid)
							if utils.IsExistInt(de_ids, departid) {
								NodeTreeToUserList(*each.ChildNode, user, maindata_audit, departs)
							}
						case float64:
							t_departid := int(each_u.TargetId.(float64))
							_, de_ids := getAllSonDepartment(t_departid, departs)
							de_ids = append(de_ids, t_departid)
							if utils.IsExistInt(de_ids, departid) {
								NodeTreeToUserList(*each.ChildNode, user, maindata_audit, departs)
							}
						}
					} else if each_u.Type == 1 {
						if userid == each_u.TargetId.(string) {
							if each.ChildNode != nil {
								NodeTreeToUserList(*each.ChildNode, user, maindata_audit, departs)
							}
						}
					}
				}
			}
			u_infos := []string{}
			for _, each := range tree.NodeUserList {
				if each.Type == 1 {
					u_infos = append(u_infos, each.TargetId.(string))
				}
			}
			if (tree.NodeType != 3) && (tree.NodeType != 4) {
				*maindata_audit = append(*maindata_audit, MainDataAudit{NodeState: 1, UserIds: u_infos})
			}
			if tree.ChildNode != nil {
				NodeTreeToUserList(*tree.ChildNode, user, maindata_audit, departs)
			}
		}
	} else {
		NodeTreeToUserList(*tree.ChildNode, user, maindata_audit, departs)
	}
}

func splitChildNodeTree(node NodeConfig, nodetrees *NodesTree) bool {
	if node.ParentCode == nodetrees.NodeCode {
		if node.IsCondition {
			nodetrees.ConditionNode = append(nodetrees.ConditionNode, NodesTree{NodeConfig: node})
		} else {
			nodetrees.ChildNode = &NodesTree{NodeConfig: node}
		}
		return true
	} else {
		state := false
		for index, _ := range nodetrees.ConditionNode {
			state = splitChildNodeTree(node, &nodetrees.ConditionNode[index])
			if state == true {
				return state
			}
		}
		if nodetrees.ChildNode == nil && state == false {
			return false
		}
		return splitChildNodeTree(node, nodetrees.ChildNode)
	}
}

func workflow2NodeTree(wf_infos []NodeConfig, parent_info NodesTree) NodesTree {
	nf := []NodeConfig{}
	for _, each := range wf_infos {
		json.Unmarshal([]byte(each.UserList), &each.NodeUserList)
		if each.IsHeader == false {
			if !splitChildNodeTree(each, &parent_info) {
				nf = append(nf, each)
			}
		}
	}
	if len(nf) == 0 {
		return parent_info
	}
	return workflow2NodeTree(nf, parent_info)
}

func getParentNode(wf_infos []NodeConfig) (parent_info NodesTree) {
	for _, value := range wf_infos {
		json.Unmarshal([]byte(value.UserList), &value.NodeUserList)
		if value.IsHeader == true {
			parent_info = NodesTree{NodeConfig: value}
		}
	}
	return parent_info
}
