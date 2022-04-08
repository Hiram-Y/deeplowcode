package controllers

import (
	"DeepWorkload/models"
	"encoding/json"
	"github.com/astaxie/beego"
	"time"
)

type DataAnalysis struct {
	beego.Controller
}

func (this *DataAnalysis) Dependencies() {
	userid := this.GetString("userid")
	corpid := this.GetString("corpid")
	n, e := models.GetDependencies(corpid, userid)
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"errmsg":  "OK",
		"data": map[string]interface{}{
			"nodes": n,
			"edges": e,
		},
	}
	this.ServeJSON()
}

func (this *DataAnalysis) Search() {
	corpid := this.GetString("corpid")
	task_code := this.GetString("task_codes")
	task_codes := []string{}
	if task_code != "" {
		json.Unmarshal([]byte(task_code), &task_codes)
	}
	userid := this.GetString("userid")
	start_time := this.GetString("start_time")
	end_time := this.GetString("end_time")
	datelayout := "2006-01-02 15:04:05"
	content := this.GetString("content")
	contents := []string{}
	if content != "" {
		json.Unmarshal([]byte(content), &contents)
	}
	startTime := time.Time{}
	endTime := time.Time{}
	if start_time != "" {
		startTime, _ = time.ParseInLocation(datelayout, start_time, time.Local)
	}
	if end_time != "" {
		endTime, _ = time.ParseInLocation(datelayout, end_time, time.Local)
	}
	err, info := models.FuzzSearch(corpid, userid, startTime, endTime, task_codes, contents)
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
