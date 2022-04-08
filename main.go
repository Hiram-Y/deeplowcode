package main

import (
	"DeepWorkload/conf"
	"DeepWorkload/models"
	_ "DeepWorkload/routers"
	"DeepWorkload/utils/AbstractAPI"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/plugins/cors"
	"github.com/astaxie/beego/toolbox"
	_ "github.com/lib/pq"
	"runtime"
	"time"
)

func init() {
	orm.Debug = true
	_ = orm.RegisterDriver("postgres", orm.DRPostgres)
	_ = orm.RegisterDataBase(
		"default",
		"postgres",
		"user=postgres password=xxx dbname=DeepWorkload host=127.0.0.1 port=5432 sslmode=disable")
	orm.SetMaxIdleConns("default", 1000) // 最大打开的连接数
	orm.SetMaxOpenConns("default", 500)
}

func autoStartTask() *toolbox.Task {
	tk := toolbox.NewTask("autoTurnOn", "0 00 8 * * *", func() error {
		o := orm.NewOrm()
		datelayout := "2006-01-02"
		m_infos := []models.Task{}
		o.QueryTable("task").Filter("state", 1).All(&m_infos)

		for _, each := range m_infos {
			tm := time.Unix(each.StartDate, 0)
			if tm.Format(datelayout) == time.Now().Format(datelayout) {
				o.QueryTable("task").Filter("code", each.Code).Update(orm.Params{"state": 0})
			}
		}

		pts := []models.TaskPublic{}
		o.QueryTable("task_public").Filter("state", 1).All(&pts)
		for _, each := range pts {
			tm := time.Unix(each.StartDate, 0)
			if tm.Format(datelayout) == time.Now().Format(datelayout) {
				o.QueryTable("task_public").Filter("code", each.Code).Update(orm.Params{"state": 0})
			}
		}

		return nil
	})
	return tk
}
func autoEndTask() *toolbox.Task {
	tk := toolbox.NewTask("autoTurnOff", "0 00 23 * * *", func() error {
		o := orm.NewOrm()
		datelayout := "2006-01-02"
		m_infos := []models.Task{}
		o.QueryTable("task").Filter("state", 0).All(&m_infos)
		for _, each := range m_infos {
			tm := time.Unix(each.EndDate, 0)
			if tm.Format(datelayout) == time.Now().Format(datelayout) {
				o.QueryTable("task").Filter("code", each.Code).Update(orm.Params{"state": 2})
			}
		}

		pts := []models.TaskPublic{}
		o.QueryTable("task_public").Filter("state", 1).All(&pts)
		for _, each := range pts {
			tm := time.Unix(each.StartDate, 0)
			if tm.Format(datelayout) == time.Now().Format(datelayout) {
				o.QueryTable("task_public").Filter("code", each.Code).Update(orm.Params{"state": 0})
			}
		}

		return nil
	})
	return tk
}

func autoSendAuditMessage() *toolbox.Task {
	tk := toolbox.NewTask("autoSendAudit", "0 00 10 * * *", func() error {
		u_map := models.GetAuditUseridCount()
		wx_api := AbstractAPI.NewWXAuthAPI(conf.WxAppId, conf.WxSecret, "")
		nowstr := time.Now().Format("2006-01-02 15:04:05")
		for _, each := range u_map {
			for k, v := range each {
				content := fmt.Sprintf("您有%d条新的待审批表单，请移步至【数据审核-待审核数据】查看处理。\n ", v)
				message_data := map[string]interface{}{
					"first":    map[string]interface{}{"value": content},
					"keyword1": map[string]interface{}{"value": "审批提醒"},
					"keyword2": map[string]interface{}{"value": nowstr},
				}

				go wx_api.SendMessage(k, conf.WxTemplateMessageId, "", message_data)
			}
		}

		return nil
	})
	return tk
}

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
	}
	beego.SetStaticPath("/docs", "uploads/docs")
	beego.SetStaticPath("/images", "uploads/images")
	beego.SetStaticPath("/temp", "uploads/temp")
	beego.InsertFilter("*",
		beego.BeforeRouter, cors.Allow(&cors.Options{
			AllowOrigins: []string{"https://server.deepgrid.cn"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"Origin", "Authorization",
				"Access-Control-Allow-Origin", "Access-Control-Allow-Headers",
				"Content-Type",
				"x-access-timestamp", "X-Access-Signature", "web-token"},
			ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
			AllowCredentials: true,
		}))

	autost := autoStartTask()
	toolbox.AddTask("autoStart", autost)
	autoed := autoEndTask()
	toolbox.AddTask("autoEnd", autoed)
	autosendau := autoSendAuditMessage()
	toolbox.AddTask("autoSendAu", autosendau)
	toolbox.StartTask()
	runtime.GOMAXPROCS(runtime.NumCPU())
	beego.Run()
}
