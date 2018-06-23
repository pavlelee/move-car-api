package controllers

import (
	"fmt"

	"github.com/Liv1020/move-car/components"
	"github.com/Liv1020/move-car/middlewares"
	"github.com/Liv1020/move-car/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/json-iterator/go"
	mpoauth "gopkg.in/chanxuehong/wechat.v2/mp/oauth2"
	"gopkg.in/chanxuehong/wechat.v2/oauth2"
)

type oauth struct{}

// Oauth 用户
var Oauth = oauth{}

// Index Index
func (t oauth) Index(c *gin.Context) {
	conf := components.App.Config

	callUrl := conf.Wechat.OAuthUrl
	fmt.Println(callUrl)
	authUrl := mpoauth.AuthCodeURL(conf.Wechat.AppID, callUrl, "snsapi_userinfo", "STATE")

	c.Redirect(302, authUrl)
}

// Code Code
func (t oauth) Code(c *gin.Context) {
	code := c.Query("code")
	conf := components.App.Config

	p := mpoauth.NewEndpoint(conf.Wechat.AppID, conf.Wechat.AppSecret)
	cli := &oauth2.Client{
		Endpoint: p,
	}

	token, err := cli.ExchangeToken(code)
	if err != nil {
		components.ResponseError(c, 1, err)
		return
	}

	info, err := mpoauth.GetUserInfo(token.AccessToken, token.OpenId, mpoauth.LanguageZhCN, nil)
	if err != nil {
		components.ResponseError(c, 1, err)
		return
	}

	db := components.App.DB()

	u := new(models.User)
	if err := db.Where("openid = ?", info.OpenId).Last(u).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			components.ResponseError(c, 1, err)
			return
		}
	}

	u.OpenID = info.OpenId
	u.Nickname = info.Nickname
	u.Sex = info.Sex
	u.City = info.City
	u.Province = info.Province
	u.Country = info.Country
	u.HeadImageUrl = info.HeadImageURL
	if err := db.Save(u).Error; err != nil {
		components.ResponseError(c, 1, err)
		return
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	iby, err := json.Marshal(&info)
	if err != nil {
		components.ResponseError(c, 1, err)
		return
	}

	c.SetCookie("wechat", string(iby), int(token.ExpiresIn), "", "", false, false)

	// 设置jwt token
	appToken := middlewares.JwtMiddleware.TokenGenerator(fmt.Sprintf("%d", u.ID))
	c.SetCookie("token", appToken, int(token.ExpiresIn), "", "", false, false)

	c.Redirect(302, "http://mc.liv1020.com:8080")
}
