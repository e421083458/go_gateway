package controller

import (
	"encoding/json"
	"github.com/e421083458/go_gateway/dao"
	"github.com/e421083458/go_gateway/dto"
	"github.com/e421083458/go_gateway/middleware"
	"github.com/e421083458/go_gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"time"
)

//AdminRegister admin_login路由注册
func AdminLoginRegister(router *gin.RouterGroup) {
	admin := AdminLogin{}
	router.POST("/login", admin.Login)
	router.POST("/logout", admin.Logout)
}

type AdminLogin struct {
}

// ListPage godoc
// @Summary 登陆接口
// @Description 登陆接口
// @Tags 管理员接口
// @ID /admin_login/login
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (AdminLogin *AdminLogin) Login(c *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	adminInfo, err := (&dao.GatewayAdmin{}).LoginCheck(c, lib.GORMDefaultPool, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	adminSession := &dto.AdminSession{
		ID:        adminInfo.ID,
		LoginTime: time.Now(),
		UserName:  adminInfo.UserName,
	}
	session := sessions.Default(c)
	adminBts, _ := json.Marshal(adminSession)
	session.Set(public.AdminInfoSessionKey, string(adminBts))
	session.Save()
	//fmt.Println(session.Get(public.AdminInfoSessionKey))
	output := &dto.AdminLoginOutput{Token: adminInfo.UserName}
	middleware.ResponseSuccess(c, output)
	return
}

// Logout godoc
// @Summary 退出接口
// @Description 退出接口
// @Tags 管理员接口
// @ID /admin_login/logout
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [post]
func (AdminLogin *AdminLogin) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	middleware.ResponseSuccess(c, "")
	return
}
