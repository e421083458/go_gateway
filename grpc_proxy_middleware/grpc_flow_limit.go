package grpc_proxy_middleware

import (
	"fmt"
	"github.com/e421083458/go_gateway/dao"
	"github.com/e421083458/go_gateway/public"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"log"
	"strings"
)

func GrpcFlowLimitMiddleware(serviceDetail *dao.ServiceDetail) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error{
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error{
		if serviceDetail.AccessControl.ServiceFlowLimit != 0 {
			serviceLimiter, err := public.FlowLimiterHandler.GetLimiter(
				public.FlowServicePrefix+serviceDetail.Info.ServiceName,
				float64(serviceDetail.AccessControl.ServiceFlowLimit))
			if err != nil {
				return err
			}
			if !serviceLimiter.Allow() {
				return errors.New(fmt.Sprintf("service flow limit %v", serviceDetail.AccessControl.ServiceFlowLimit), )
			}
		}
		peerCtx,ok:=peer.FromContext(ss.Context())
		if !ok{
			return errors.New("peer not found with context")
		}
		peerAddr:=peerCtx.Addr.String()
		addrPos:=strings.LastIndex(peerAddr,":")
		clientIP:=peerAddr[0:addrPos]
		if serviceDetail.AccessControl.ClientIPFlowLimit > 0 {
			clientLimiter, err := public.FlowLimiterHandler.GetLimiter(
				public.FlowServicePrefix+serviceDetail.Info.ServiceName+"_"+clientIP,
				float64(serviceDetail.AccessControl.ClientIPFlowLimit))
			if err != nil {
				return err
			}
			if !clientLimiter.Allow() {
				return errors.New(fmt.Sprintf("%v flow limit %v",clientIP, serviceDetail.AccessControl.ClientIPFlowLimit), )
			}
		}
		if err := handler(srv, ss);err != nil {
			log.Printf("GrpcFlowLimitMiddleware failed with error %v\n", err)
			return err
		}
		return nil
	}
}
