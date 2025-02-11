/*
Copyright (c) Facebook, Inc. and its affiliates.
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

// Package swx_proxy provides a thin client for using swx proxy service.
// This can be used by apps to discover and contact the service, without knowing about
// the RPC implementation.
package hlr_proxy

import (
	"errors"
	"fmt"

	"magma/feg/cloud/go/protos"
	"magma/feg/cloud/go/protos/hlr"
	"magma/feg/gateway/registry"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Wrapper for GRPC Client
// functionality
type hlrProxyClient struct {
	hlr.HlrProxyClient
	cc *grpc.ClientConn
}

// getHlrProxyClient is a utility function to get a RPC connection to the
// HLR Proxy service
func getHlrProxyClient() (*hlrProxyClient, error) {
	var conn *grpc.ClientConn
	var err error
	conn, err = registry.GetConnection(registry.HLR_PROXY)
	if err != nil {
		errMsg := fmt.Sprintf("HLR Proxy client initialization error: %s", err)
		glog.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	return &hlrProxyClient{
		hlr.NewHlrProxyClient(conn),
		conn,
	}, err
}

const (
	resyncRandEnd = hlr.AuthInfoReq_ResyncInfo_RAND_LEN
	resyncAuthEnd = hlr.AuthInfoReq_ResyncInfo_RAND_LEN + hlr.AuthInfoReq_ResyncInfo_AUTH_LEN
)

// Authenticate - HLR equivalent of SWX Authenticate
func Authenticate(ctx context.Context, req *protos.AuthenticationRequest) (*protos.AuthenticationAnswer, error) {
	cli, err := getHlrProxyClient()
	if err != nil {
		return nil, err
	}
	hlrAns, err := cli.AuthInfo(ctx, &hlr.AuthInfoReq{
		UserName:                req.GetUserName(),
		NumRequestedUmtsVectors: req.GetSipNumAuthVectors(),
		ResyncInfo: &hlr.AuthInfoReq_ResyncInfo{
			Rand: req.GetResyncInfo()[:resyncRandEnd],
			Autn: req.GetResyncInfo()[resyncRandEnd:resyncAuthEnd]},
	})
	res := &protos.AuthenticationAnswer{
		UserName:       req.GetUserName(),
		SipAuthVectors: []*protos.AuthenticationAnswer_SIPAuthVector{}}
	if err != nil {
		return res, err
	}
	if hlrAns.GetErrorCode() != hlr.ErrorCode_SUCCESS {
		return res, fmt.Errorf("HLR Error: %s for User: %s", hlrAns.GetErrorCode().String(), req.GetUserName())
	}
	for _, v := range hlrAns.GetUmtsVectors() {
		res.SipAuthVectors = append(res.SipAuthVectors, &protos.AuthenticationAnswer_SIPAuthVector{
			AuthenticationScheme: protos.AuthenticationScheme_EAP_AKA,
			RandAutn:             append(v.GetRand(), v.GetAutn()...),
			Xres:                 v.GetXres(),
			ConfidentialityKey:   v.GetCk(),
			IntegrityKey:         v.GetIk(),
		})
	}
	return res, nil
}

// Register HLR equivalent of SWX register
func Register(_ context.Context, req *protos.RegistrationRequest) (*protos.RegistrationAnswer, error) {
	return &protos.RegistrationAnswer{SessionId: req.SessionId}, nil
}

// Deregister HLR equivalent of SWX deregister
func Deregister(_ context.Context, req *protos.RegistrationRequest) (*protos.RegistrationAnswer, error) {
	return &protos.RegistrationAnswer{SessionId: req.SessionId}, nil
}
