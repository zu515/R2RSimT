//
// Copyright (C) 2014-2017 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"github.com/zhishi/R2RSimT/pkg/db"
	"github.com/zhishi/R2RSimT/pkg/server"
	_ "net/http/pprof"
	"sync"
)



func main()  {
	serverNode := &ServerNode{
		MessageChan : make(chan db.Message),
		StartedChan : make(chan bool),
		States: db.SERVER_STARTED,
		Addr: GetOutboundIP().String(),
	}
	defer close(serverNode.MessageChan)
	defer close(serverNode.StartedChan)

	go db.Subscribe(serverNode.MessageChan, serverNode.Addr, false)
	var wg = sync.WaitGroup{}
	server.BgpServerStartChan = make(chan server.BGPServerStartMsg)
	server.TopologyConfig = server.NewBGPTopologyConfig()
	for {
		select {
		case msg := <-serverNode.MessageChan:
			switch msg.Type {
			case db.STATES_CHANGE:
				switch msg.States[db.STATES] {
				case db.CONTROLLER_DISTRIBUTE:
					fmt.Println("节点开始部署")
					db.SerialNum = db.Redis.Get(db.BASE_KEY + db.SERIAL_KEY).Val()
					GetDistributeNode(serverNode.Addr)
					serverNode.StartedServerNum = int64(len(server.TopologyConfig.NodeMap))
					for _, n := range server.TopologyConfig.NodeMap {
						wg.Add(1)
						go nodeStart(wg, n.ASN, serverNode)
					}
				case db.SERVER_DEPLOYED:
					UpdateStates(serverNode.Addr)
					DePloyedPublished(serverNode.Addr)
				}
			}
		case ribMsg := <-serverNode.RibMonitorChan:
			go UpdateRib(ribMsg)
		}
	}
	wg.Wait()
}



//func stopServer(bgpServer *server.BgpServer, useSdNotify bool) {
//	bgpServer.StopBgp(context.Background(), &api.StopBgpRequest{})
//	if useSdNotify {
//		daemon.SdNotify(false, daemon.SdNotifyStopping)
//	}
//}





