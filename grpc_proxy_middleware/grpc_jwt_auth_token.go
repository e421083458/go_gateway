package grpc_proxy_middleware

import (
	"github.com/e421083458/go_gateway/dao"
	"github.com/e421083458/go_gateway/public"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"strings"
)

//jwt auth token
func GrpcJwtAuthTokenMiddleware(serviceDetail *dao.ServiceDetail) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error{
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error{
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return errors.New("miss metadata from context")
		}
		authToken:=""
		auths:=md.Get("authorization")
		if len(auths)>0{
			authToken = auths[0]
		}
		token:=strings.ReplaceAll(authToken,"Bearer ","")
		appMatched:=false
		if token!=""{
			claims,err:=public.JwtDecode(token)
			if err!=nil{
				return errors.WithMessage(err,"JwtDecode")
			}
			appList:=dao.AppManagerHandler.GetAppList()
			for _,appInfo:=range appList{
				if appInfo.AppID==claims.Issuer{
					md.Set("app",public.Obj2Json(appInfo))
					appMatched = true
					break
				}
			}
		}
		if serviceDetail.AccessControl.OpenAuth==1 && !appMatched{
			return errors.New("not match valid app")
		}
		if err := handler(srv, ss);err != nil {
			log.Printf("GrpcJwtAuthTokenMiddleware failed with error %v\n", err)
			return err
		}
		return nil
	}
}
