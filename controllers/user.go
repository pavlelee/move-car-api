package controllers

import (
	"errors"

	"time"

	"github.com/Liv1020/move-car/components"
	"github.com/Liv1020/move-car/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type user struct{}

// User 用户
var User = user{}

// Update Update
func (t *user) Update(c *gin.Context) {
	auth := components.GetAuthFromClaims(c)

	form := new(form)
	err := c.BindJSON(form)
	if err != nil {
		components.ResponseError(c, 1, err)
		return
	}

	if err := form.Validate(); err != nil {
		components.ResponseError(c, 1, err)
		return
	}

	db := components.App.DB()

	qr := new(models.Qrcode)
	if err := db.Where("id = ?", form.QrCode).Last(qr).Error; err != nil {
		components.ResponseError(c, 1, err)
		return
	}

	u := new(models.User)
	if err := db.Where("id = ?", auth.ID).Last(u).Error; err != nil {
		components.ResponseError(c, 1, err)
		return
	}

	u.Mobile = form.Mobile
	u.PlateNumber = form.PlateNumber
	if err := db.Save(u).Error; err != nil {
		components.ResponseError(c, 1, err)
		return
	}

	qr.UserID = u.ID
	if err := db.Save(qr).Error; err != nil {
		components.ResponseError(c, 1, err)
		return
	}

	components.ResponseSuccess(c, u)
}

type form struct {
	QrCode      string `json:"qr_code"`
	Mobile      string `json:"mobile"`
	PlateNumber string `json:"plate_number"`
	Code        string `json:"code"`
}

// Validate Validate
func (t *form) Validate() error {
	if t.Mobile == "" {
		return errors.New("手机号码不能为空")
	}
	if t.Code == "" {
		return errors.New("验证码不能为空")
	}
	db := components.App.DB()
	sc := new(models.SmsCode)
	if err := db.Where("mobile = ? AND expired_at > ? AND is_valid = 0", t.Mobile, time.Now()).Last(sc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("请重新发送验证")
		}

		return err
	}
	if sc.Code != t.Code {
		return errors.New("验证码错误")
	}
	sc.IsValid = 1
	if err := db.Save(sc).Error; err != nil {
		return err
	}
	if t.PlateNumber == "" {
		return errors.New("车牌号不能为空")
	}
	if t.QrCode == "" {
		return errors.New("二维码不能为空")
	}

	return nil
}
