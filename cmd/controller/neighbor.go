package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	api "github.com/zhishi/R2RSimT/api"
	"github.com/zhishi/R2RSimT/pkg/db"
	"github.com/zhishi/R2RSimT/pkg/server"
	"io"
	"context"
	"strconv"
)

func GetNeighbor(client api.GobgpApiClient) {
	ctx := context.Background()
	stream, err := client.ListPeer(ctx, &api.ListPeerRequest{
		Address:          "",
		EnableAdvertised: false,
	})
	if err != nil {
		fmt.Println(err)
	}
	l := make([]*api.Peer, 0, 1024)
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(err)
		}
		l = append(l, r.Peer)
		fmt.Printf("%+v", r.Peer.Conf.String())
		fmt.Println()
	}
}
func AddPeer(client api.GobgpApiClient,peerAsn uint32) error {
	serverAddr := db.Redis.HGet(db.BASE_KEY + serialNum + ":"+db.NODE_ADRESS_MAP_KEY, strconv.Itoa(int(peerAsn))).Val()
	var peerNode server.BGPNodeConfig
	peerNodeStr := db.Redis.Get(db.BASE_KEY+ serialNum+":"+serverAddr + ":"+strconv.Itoa(int(peerAsn))).Val()
	json.Unmarshal([]byte(peerNodeStr), &peerNode)
	ctx := context.Background()
	peer := &api.Peer{
		Conf:  &api.PeerConf{},
		State: &api.PeerState{},
		Transport: &api.Transport{},
	}
	peer.Conf.NeighborAddress = peerNode.Adress
	peer.Conf.PeerAs = peerNode.ASN
	peer.Transport.RemotePort = peerNode.ServerPort
	_, err := client.AddPeer(ctx, &api.AddPeerRequest{
		Peer: peer,
	})
	if err != nil{
		log.WithFields(log.Fields{
			"Topic": "AddPeer",
			"Key":   peerAsn,
		}).Error("AddPeer err:", err)
		return err
	}
	return nil
}