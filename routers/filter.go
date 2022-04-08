package routers

import (
	"DeepWorkload/conf"
	"DeepWorkload/utils"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"strconv"
	"time"
)

func RouterFilter() {
	var check = func(ctx *context.Context) {
		if ctx.Request.Method != "OPTIONS" {
			Timestamp := ctx.Request.Header.Get("X-Access-Timestamp")
			Signature := ctx.Request.Header.Get("X-Access-Signature")
			signature := utils.MD5("timestamp=" + Timestamp + "&" + "token=" + conf.Token)
			timestamp, _ := strconv.Atoi(Timestamp[:10])
			datetime := time.Unix(int64(timestamp), 0)
			subM := time.Now().Sub(datetime)
			if subM.Seconds() > 600 || subM.Seconds() < -600 {
				ctx.ResponseWriter.WriteHeader(403)
			} else {
				if signature != Signature {
					ctx.ResponseWriter.WriteHeader(403)
				}
			}
		}
	}
	beego.InsertFilter("/api/*", beego.BeforeExec, check)
}
