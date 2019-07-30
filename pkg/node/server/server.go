/**
 * Copyright 2019 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"context"
	"net"

	"github.com/IBM/ibm-block-csi-driver-operator/pkg/node/informer"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/node/iscsi"
	pb "github.com/IBM/ibm-block-csi-driver-operator/pkg/node/nodeagent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var nodeName string

// server is used to implement nodeagent.NodeAgentServer.
type server struct {
	nodeName string
}

func (s *server) GetNodeInfo(ctx context.Context, in *pb.GetNodeInfoRequest) (*pb.GetNodeInfoReply, error) {
	if in.GetName() != s.nodeName {
		return nil, status.Error(codes.InvalidArgument, "node name mismatch")
	}

	iqns, err := informer.GetNodeIscsiIQNs()
	if err != nil {
		return nil, status.Convert(err).Err()
	}
	wwpns, err := informer.GetNodeFcWWPNs()
	if err != nil {
		return nil, status.Convert(err).Err()
	}
	return &pb.GetNodeInfoReply{Node: &pb.Node{
		Name:  nodeName,
		Iqns:  iqns,
		Wwpns: wwpns,
	}}, nil
}

func (s *server) IscsiLogin(ctx context.Context, in *pb.IscsiLoginRequest) (*pb.IscsiLoginReply, error) {
	err := iscsi.DiscoverAndLoginPortals(in.GetTargets())
	if err != nil {
		return nil, status.Convert(err).Err()
	}
	return &pb.IscsiLoginReply{}, nil
}

func (s *server) IscsiLogout(ctx context.Context, in *pb.IscsiLogoutRequest) (*pb.IscsiLogoutReply, error) {
	err := iscsi.DiscoverAndLogoutPortals(in.GetTargets())
	if err != nil {
		return nil, status.Convert(err).Err()
	}
	return &pb.IscsiLogoutReply{}, nil
}

func Serve(address, nodeName string) error {
	conn, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	pb.RegisterNodeAgentServer(s, &server{nodeName: nodeName})
	if err := s.Serve(conn); err != nil {
		return err
	}
	return nil
}
