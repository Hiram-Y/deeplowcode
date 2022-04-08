package controllers

import (
	"DeepWorkload/conf"
	"DeepWorkload/lib/redisgo"
	"DeepWorkload/models"
	"DeepWorkload/utils/graphics"
	"encoding/base64"
	"github.com/astaxie/beego"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	uuid "github.com/satori/go.uuid"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type CommonController struct {
	beego.Controller
}

func (this *CommonController) Options() {
	this.Data["json"] = map[string]interface{}{"status": 200, "message": "ok", "moreinfo": ""}
	this.ServeJSON()
}

func getqiuniuToken(key string) string {
	token_key := key + "QINIU_UPLOAD_TOKEN"
	con, _ := redisgo.New(redisgo.Options{})
	token, err := con.GetString(token_key)
	if err == nil {
		return token
	} else {
		mac := qbox.NewMac(conf.QN_AK, conf.QN_SK)
		putPolicy := storage.PutPolicy{
			Scope: conf.QN_BUCKET_D,
		}
		putPolicy.Expires = 7200 //示例2小时有效期
		upToken := putPolicy.UploadToken(mac)
		con.Set(token_key, upToken, 7200)
		return upToken
	}
}

func (this *CommonController) GetQiNiuToken() {
	bucket := conf.QN_BUCKET_D
	token := getqiuniuToken(bucket)
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"errmsg":  "OK",
		"data":    token,
	}
	this.ServeJSON()
}

func (this *CommonController) QiNiuDownloadUrlByKey() {
	key := this.GetString("key")
	mac := qbox.NewMac(conf.QN_AK, conf.QN_SK)
	deadline := time.Now().Add(time.Second * 3600).Unix() //1小时有效期
	privateAccessURL := storage.MakePrivateURL(mac, conf.QN_URL, key, deadline)
	this.Data["json"] = map[string]interface{}{
		"errcode": OK,
		"errmsg":  "OK",
		"data":    privateAccessURL,
	}
	this.ServeJSON()
}

func (this *CommonController) UploadFile() {
	file, moreFile, _ := this.GetFile("file")
	userid := this.Ctx.Input.Query("userid")
	defer file.Close()
	shotname := moreFile.Filename
	fileName := strconv.FormatInt(time.Now().UnixNano(), 16)
	filePath := filepath.Join(conf.WorkingDirectory, "uploads/docs", fileName+shotname)
	url_path := "/docs/" + fileName + shotname
	path := filepath.Dir(filePath)
	os.MkdirAll(path, os.ModePerm)
	_ = this.SaveToFile("file", filePath)
	doc := &models.DocPan{
		Type:    "docs",
		Path:    filePath,
		UrlPath: url_path,
		DocCode: uuid.NewV4().String(),
		DocName: shotname,
		Userid:  userid,
	}
	err := doc.InsertDocPan()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode": OK,
			"message": "OK",
			"type":    doc.Type,
			"urlpath": doc.UrlPath,
			"docname": shotname,
			"doccode": doc.DocCode,
		}
	}
	this.ServeJSON()
}

func (this *CommonController) UploadImageBase64() {
	basestr := this.Ctx.Input.Query("basestr")
	userid := this.Ctx.Input.Query("userid")
	shotname := this.Ctx.Input.Query("shotname")
	binary_data, _ := base64.StdEncoding.DecodeString(basestr)
	fileName := "avatar_" + strconv.FormatInt(time.Now().UnixNano(), 16)
	filePath := filepath.Join(conf.WorkingDirectory, "uploads/images", fileName+userid+".jpg")
	thumbPath := filepath.Join(conf.WorkingDirectory, "uploads/images", fileName+userid+"thumb"+".jpg")
	path := filepath.Dir(filePath)
	os.MkdirAll(path, os.ModePerm)
	ioutil.WriteFile(filePath, binary_data, 0666)
	go saveThumbnailImage(filePath, thumbPath)
	url_path := "/images/" + fileName + userid + ".jpg"
	thumb_url_path := "/images/" + fileName + userid + "thumb" + ".jpg"
	doc := &models.DocPan{
		Type:    "image",
		Path:    filePath,
		UrlPath: url_path,
		DocCode: uuid.NewV4().String(),
		DocName: shotname,
		Userid:  userid,
	}
	err := doc.InsertDocPan()
	if err != nil {
		this.Data["json"] = map[string]interface{}{
			"errcode": ErrDataBase,
			"message": err.Error(),
		}
	} else {
		this.Data["json"] = map[string]interface{}{
			"errcode":      OK,
			"message":      "OK",
			"type":         doc.Type,
			"urlpath":      doc.UrlPath,
			"thumburlpath": thumb_url_path,
			"docname":      shotname,
			"doccode":      doc.DocCode,
		}
	}
	this.ServeJSON()
}
func saveThumbnailImage(path, savepath string) {
	file, _ := os.Open(path)
	img, _ := jpeg.Decode(file)
	graphics.ImageThumbnailSaveFile(img, 800, 800, savepath)
}
