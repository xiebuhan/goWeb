package article

import (
	"goWeb/pkg/logger"
	"goWeb/pkg/model"
	"goWeb/pkg/types"
)

// GET通过 ID 获取文章ID
func Get(idstr string) (Article, error) {
	var article Article
	id := types.StringToUint64(idstr)
	if err := model.DB.First(&article, id).Error; err != nil {
		return article, err
	}

	return article, nil
}
//获取全部文章

func GetAll()([]Article,error)  {
	var articles []Article
	if err := model.DB.Find(&articles).Error;err != nil{
		return articles, err
	}
	return articles, nil
}

// 创建文章
func (article *Article) Create() (err error)  {
	if err = model.DB.Create(&article).Error; err != nil {
		logger.LogError(err)
		return err
	}

	return nil
}