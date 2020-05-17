package grpc_proxy_router

import (
	"fmt"
	"github.com/e421083458/go_gateway/dao"
	"github.com/e421083458/go_gateway/grpc_proxy_middleware"
	"github.com/e421083458/go_gateway/reverse_proxy"
	"github.com/e421083458/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"log"
	"net"
)

var grpcServerList = []*warpGrpcServer{}

type warpGrpcServer struct {
	Addr string
	*grpc.Server
}

func GrpcServerRun() {
	serviceList := dao.ServiceManagerHandler.GetGrpcServiceList()
	for _, serviceItem := range serviceList {
		tempItem := serviceItem
		go func(serviceDetail *dao.ServiceDetail) {
			addr := fmt.Sprintf(":%d", serviceDetail.GRPCRule.Port)
			rb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				log.Fatalf(" [INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
				return
			}
			lis, err := net.Listen("tcp", addr)
			if err != nil {
				log.Fatalf(" [INFO] GrpcListen %v err:%v\n", addr, err)
			}
			grpcHandler := reverse_proxy.NewGrpcLoadBalanceHandler(rb)
			s := grpc.NewServer(
				grpc.ChainStreamInterceptor(
					grpc_proxy_middleware.GrpcFlowCountMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcFlowLimitMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcJwtAuthTokenMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcJwtFlowCountMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcJwtFlowLimitMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcWhiteListMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcBlackListMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcHeaderTransferMiddleware(serviceDetail),
				),
				grpc.CustomCodec(proxy.Codec()),
				grpc.UnknownServiceHandler(grpcHandler))

			grpcServerList = append(grpcServerList, &warpGrpcServer{
				Addr:   addr,
				Server: s,
			})
			log.Printf(" [INFO] grpc_proxy_run %v\n", addr)
			if err := s.Serve(lis); err != nil {
				log.Fatalf(" [INFO] grpc_proxy_run %v err:%v\n", addr, err)
			}
		}(tempItem)
	}
}

func GrpcServerStop() {
	for _, grpcServer := range grpcServerList {
		grpcServer.GracefulStop()
		log.Printf(" [INFO] grpc_proxy_stop %v stopped\n", grpcServer.Addr)
	}
}
