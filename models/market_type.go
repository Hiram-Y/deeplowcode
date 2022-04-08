package models

import (
	"github.com/astaxie/beego/orm"
	"github.com/pkg/errors"
)

type MarketType struct {
	Id       int    `orm:"column(id)"`
	Content  string `orm:"column(content)"`
	ParentId int    `orm:"column(parent_id)"`
}

func (this *MarketType) TableName() string {
	return "market_type"
}

type FMarketType struct {
	Id       int
	Content  string
	ParentId int
	Child    []FMarketType
}

func GetAllTypeInfo() (err error, info []FMarketType) {
	o := orm.NewOrm()
	m_info := []MarketType{}
	_, err = o.QueryTable("market_type").OrderBy("-id").All(&m_info)
	p_info := getParentTypeList(m_info)

	return err, typeData2TreeData(m_info, p_info)
}

func (this *MarketType) AddTypeInfo() error {
	o := orm.NewOrm()
	_, err := o.Insert(this)
	return err
}

func (this *MarketType) DelOneType() error {
	o := orm.NewOrm()
	if o.QueryTable("market_type").Filter("parent_id", this.Id).Exist() {
		return errors.New("存在子类")
	}
	_, err := o.QueryTable("market_type").Filter("id", this.Id).Delete()
	return err
}

func (this *MarketType) UpdateOne() error {
	o := orm.NewOrm()
	_, err := o.QueryTable("market_type").Filter("id", this.Id).Update(orm.Params{"content": this.Content})
	return err
}

func getParentTypeList(m_type []MarketType) (p_infos []FMarketType) {
	for _, value := range m_type {
		if value.ParentId == 1 {
			temp := FMarketType{
				Id:       value.Id,
				ParentId: value.ParentId,
				Content:  value.Content,
			}
			p_infos = append(p_infos, temp)
		}
	}
	return p_infos
}

func typeData2TreeData(m_type []MarketType, p_infos []FMarketType) []FMarketType {
	for p, value := range p_infos {
		childrenArray := []FMarketType{}
		for _, d_value := range m_type {
			if d_value.ParentId == value.Id {
				temp := FMarketType{
					Id:       d_value.Id,
					ParentId: d_value.ParentId,
					Content:  d_value.Content,
				}
				childrenArray = append(childrenArray, temp)
			}
		}
		p_infos[p].Child = childrenArray
	}
	return p_infos
}
