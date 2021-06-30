package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/zhishi/R2RSimT/pkg/config"
	"github.com/zhishi/R2RSimT/pkg/db"
	"github.com/zhishi/R2RSimT/pkg/server"
	"strconv"
)

func Topologyprocess() []*server.BGPNodeConfig {
	nodes := []*server.BGPNodeConfig{&server.BGPNodeConfig{
			ASN:        111,
			RouterID:   "1.1.1.1",
			Peer: []uint32{222},
			RibTableSize: 0,
			RoaTableSize: 0,
			States: db.CONTROLLER_INIT,

		}, &server.BGPNodeConfig{
			ASN:        222,
			RouterID:   "2.2.2.2",
			Peer: []uint32{111},
			RibTableSize: 0,
			RoaTableSize: 0,
			States:       db.CONTROLLER_INIT,
		},
	}
	return nodes
}

func NodesDistribute(nodes []*server.BGPNodeConfig) map[string][]*server.BGPNodeConfig {
	var DistributedNodes map[string][]*server.BGPNodeConfig
	DistributedNodes = make(map[string][]*server.BGPNodeConfig)
	nodesSize := len(nodes)
	confFile, _ := config.ReadProConfigFile("configs/project.conf", "yaml")
	serverList := confFile.BGPServers
	per := nodesSize / len(serverList)
	redisPipe := db.Redis.Pipeline()
	for indexNum, server := range serverList{
		distributionNode := nodes[Max(indexNum, 0)*per : Min(len(nodes), indexNum+1)*per]
		DistributedNodes[server] = distributionNode
		var asnList []uint32
		for _,node := range distributionNode{
			node.States = db.CONTROLLER_DISTRIBUTE
			asnList = append(asnList,node.ASN)
			data, _ := json.Marshal(node)
			redisPipe.Set(db.BASE_KEY+serialNum+":" + server + ":" + strconv.Itoa(int(node.ASN)), data, 0)
			redisPipe.HSet(db.BASE_KEY + serialNum + ":" + db.NODE_ADRESS_MAP_KEY, strconv.Itoa(int(node.ASN)), server)
		}
		data, _ := json.Marshal(asnList)
		redisPipe.Set(db.BASE_KEY + serialNum+":" + server + ":" + db.DISTRIBUTED_NODELIST_KEY, data, 0)
	}
	_, err := redisPipe.Exec()
	if err != nil{
		log.WithFields(log.Fields{
			"Topic":      "topology",
		}).Error("redisPipe.Exec(),err:",err)
	}
	DistributePublished()
	return DistributedNodes
}

func DistributePublished()  {
	log.WithFields(log.Fields{
		"Topic": "DistributePublished",
	}).Info("DistributePublished")
	msg := db.Message{
		Type: db.STATES_CHANGE,
		States: map[string]int{db.STATES:db.CONTROLLER_DISTRIBUTE},
		NodeInfo: db.CONTROLLER_KEY,
	}
	confFile, _ := config.ReadProConfigFile("configs/project.conf", "yaml")
	serverList := confFile.BGPServers
	for _,serverAddr := range serverList{
		db.Publish(db.CHANNEL_SERVER_NODE_KEY+":"+serverAddr, msg)
	}
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
