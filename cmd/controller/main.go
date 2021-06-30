package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	api "github.com/zhishi/R2RSimT/api"
	"github.com/zhishi/R2RSimT/pkg/config"
	"github.com/zhishi/R2RSimT/pkg/db"
	"github.com/zhishi/R2RSimT/pkg/etcd"
	"github.com/zhishi/R2RSimT/pkg/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"strconv"
)



var clientMap map[uint32] api.GobgpApiClient
var serialNum string

type Controller struct {
	MessageChan chan db.Message
	States      int
}

func main() {
	controller := &Controller{
		MessageChan: make(chan db.Message),
		States:      db.SERVER_STARTED,
	}
	clientMap = make(map[uint32]api.GobgpApiClient)
	go db.Subscribe(controller.MessageChan, "",true)
	serialNum = strconv.Itoa(int(db.Redis.Incr(db.BASE_KEY + db.SERIAL_KEY).Val()))

	topologyNodes := Topologyprocess()
	distributedNodes := NodesDistribute(topologyNodes)
	for {
		select {
		case msg := <-controller.MessageChan:
			switch msg.Type {
			case db.STATES_CHANGE:
				switch msg.States[db.STATES] {
				case db.SERVER_DEPLOYED:
					fmt.Println("部署成功")
					if msg.NodeInfo != ""{
						AddPeerNode(distributedNodes[msg.NodeInfo])
					}
				}
			}
		}
	}
}

func SendServerStartMsg(msg server.BGPServerStartMsg)  {
	server.BgpServerStartChan<- msg
}

func AddPeerNode(topologyNodes []*server.BGPNodeConfig){
	log.WithFields(log.Fields{
		"Topic": "AddPeerNode",
	}).Info("AddPeerNode" )
	for _,node := range topologyNodes{
		client, _ := GetClient(node.ASN)
		for _,peer := range node.Peer{
			go AddPeer(client, peer)
		}
	}
}


func GetClient(asn uint32) (client api.GobgpApiClient, err error)  {
	if clientMap[asn] != nil{
		return clientMap[asn], nil
	}
	confFile, _ := config.ReadProConfigFile("configs/project.conf", "yaml")
	r := etcd.NewResolver(confFile.Etcd.Host)
	resolver.Register(r)
	conn, err := grpc.Dial(r.Scheme()+":///" + strconv.Itoa(int(asn)), grpc.WithBalancerName("round_robin"), grpc.WithInsecure())
	if err != nil {
		log.WithFields(log.Fields{
			"Topic": "GetClient",
		}).Error("GetClient,err:", err)
		return nil, err
	}
	client = api.NewGobgpApiClient(conn)
	clientMap[asn] = client
	return client, nil
}




