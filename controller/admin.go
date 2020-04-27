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
)

//AdminRegister admin路由注册
func AdminRegister(router *gin.RouterGroup) {
	admin := AdminController{}
	router.GET("/admin_info", admin.AdminInfo)
	router.POST("/change_pwd", admin.ChangePwd)
}

type AdminController struct {
}

// ListPage godoc
// @Summary 登陆信息
// @Description 登陆信息
// @Tags 管理员接口
// @ID /admin/admin_info
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/admin_info [get]
func (admin *AdminController) AdminInfo(c *gin.Context) {
	session := sessions.Default(c)
	adminInfoStr := session.Get(public.AdminInfoSessionKey)
	sessionInfo := &dto.AdminSession{}
	if err := json.Unmarshal([]byte(adminInfoStr.(string)), sessionInfo); err != nil {
		middleware.ResponseError(c, 200, err)
		return
	}
	output := &dto.AdminInfoOutput{
		ID:           sessionInfo.ID,
		Name:         sessionInfo.UserName,
		Avatar:       "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
		Introduction: "I am a super administrator",
		Roles:        []string{"admin"},
		LoginTime:    sessionInfo.LoginTime,
	}
	middleware.ResponseSuccess(c, output)
	return
}

func (admin *AdminController) ChangePwd(c *gin.Context) {
	params := &dto.AdminChangePwdInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	session := sessions.Default(c)
	adminInfoStr := session.Get(public.AdminInfoSessionKey)
	sessionInfo := &dto.AdminSession{}
	if err := json.Unmarshal([]byte(adminInfoStr.(string)), sessionInfo); err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	search := &dao.GatewayAdmin{
		ID: sessionInfo.ID,
	}
	adminInfo, err := search.Find(c, lib.GORMDefaultPool, search)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	savePassword := public.GenSaltPassword(params.Password, adminInfo.Salt)
	adminInfo.Password = savePassword
	if err := adminInfo.Save(c, lib.GORMDefaultPool); err != nil {
		middleware.ResponseError(c, 2004, err)
		return
	}
	middleware.ResponseSuccess(c, "")
	return
}
