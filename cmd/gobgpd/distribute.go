package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	api "github.com/zhishi/R2RSimT/api"
	"github.com/zhishi/R2RSimT/pkg/db"
	"github.com/zhishi/R2RSimT/pkg/server"
	"net"
	"strconv"
)

func GetDistributeNode(addr string) () {
	//serialNum := db.Redis.Get(db.BASE_KEY + db.SERIAL_KEY)
	distributeNodeList := db.Redis.Get(db.BASE_KEY+ db.SerialNum+":"+ addr + ":" + db.DISTRIBUTED_NODELIST_KEY)
	fmt.Println("取出来", db.BASE_KEY +db.SerialNum+":"+addr+":"+db.DISTRIBUTED_NODELIST_KEY)
	fmt.Println("取出来长度", len(distributeNodeList.Val()))

	var asnList []uint32
	var nodes = make([]server.BGPNodeConfig,0)
	json.Unmarshal([]byte(distributeNodeList.Val()),&asnList)
	freePorts ,err := GetFreePorts(len(asnList) * 3)
	serverNodes := splitArray(freePorts, int64(len(asnList)))
	if err != nil{
		log.WithFields(log.Fields{
			"Topic": "GetDistributeNode",
			"Key":   addr,
			"Data": freePorts,
		}).Error("GetFreePorts err:", err)
	}
	for index,asn := range asnList{
		//redisPipe.Set(db.BASE_KEY+serialNum.Val()+":"+server+":"+strconv.Itoa(int(node.ASN)), data, 0)
		var distributeNode server.BGPNodeConfig
		json.Unmarshal([]byte(db.Redis.Get(db.BASE_KEY+db.SerialNum+":"+addr+":"+strconv.Itoa(int(asn))).Val()), &distributeNode)
		distributeNode.ServerPort = uint32(serverNodes[index][0])
		distributeNode.GrpcPort = uint32(serverNodes[index][1])
		distributeNode.PProfHost = uint32(serverNodes[index][2])
		distributeNode.Adress = addr
		server.TopologyConfig.NodeMap[distributeNode.ASN] = &distributeNode
		nodes = append(nodes, distributeNode)
	}
}

//
func UpdateStates(addr string)  {
	serialNum := db.Redis.Get(db.BASE_KEY + db.SERIAL_KEY)
	redisPipe := db.Redis.Pipeline()
	var asnList []uint32
	for index,_ := range server.TopologyConfig.NodeMap{
		node := server.TopologyConfig.NodeMap[index]
		data, _ := json.Marshal(node)
		redisPipe.Set(db.BASE_KEY+serialNum.Val()+":"+addr+":"+strconv.Itoa(int(node.ASN)), data, 0)
		asnList = append(asnList,node.ASN)
	}
	data, _ := json.Marshal(asnList)
	redisPipe.Set(db.BASE_KEY+serialNum.Val()+":" + addr + ":" + db.DEPLOYED_NODELIST_KEY, data, 0)
	_, err := redisPipe.Exec()
	if err != nil {
		log.WithFields(log.Fields{
			"Topic": "topology",
		}).Error("redisPipe.Exec(),err:", err)
	}
	return
}

func DePloyedPublished(addr string) {
	log.WithFields(log.Fields{
		"Topic": "DePloyedPublished",
	}).Info("DePloyedPublished")
	msg := db.Message{
		Type:   db.STATES_CHANGE,
		States: map[string]int{db.STATES: db.SERVER_DEPLOYED},
		NodeInfo: addr,
	}
	db.Publish(db.CHANNEL_CONTROLLER_KEY, msg)
}

func GetFreePorts(count int) ([]int, error) {
	var ports []int
	for i := 0; i < count; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			return nil, err
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return nil, err
		}
		defer l.Close()
		ports = append(ports, l.Addr().(*net.TCPAddr).Port)
	}
	return ports, nil
}
func GetGlobalConfig(ASN uint32) *api.Global {
	//routerId := topologyConfig.NodeMap[ASN].RouterID
	//addr :=
	//
	//
	//addr := strings.Split(portNewConfig.ServerAddressAndPortMap[routerId], ":")[0]
	//listenPort ,_ :=  strconv.ParseInt(strings.Split(portNewConfig.ServerAddressAndPortMap[routerId], ":")[1], 10, 64)

	return &api.Global{
		As:              ASN,
		RouterId:        server.TopologyConfig.NodeMap[ASN].RouterID,
		ListenPort:      int32(server.TopologyConfig.NodeMap[ASN].ServerPort),
		ListenAddresses: []string{server.TopologyConfig.NodeMap[ASN].Adress},
		//As:               c.Config.As,
		//RouterId:         c.Config.RouterId,
		//ListenPort:       c.Config.Port,
		//ListenAddresses:  c.Config.LocalAddressList,
		//Families:         families,
		//UseMultiplePaths: c.UseMultiplePaths.Config.Enabled,
		//RouteSelectionOptions: &api.RouteSelectionOptionsConfig{
		//	AlwaysCompareMed:         c.RouteSelectionOptions.Config.AlwaysCompareMed,
		//	IgnoreAsPathLength:       c.RouteSelectionOptions.Config.IgnoreAsPathLength,
		//	ExternalCompareRouterId:  c.RouteSelectionOptions.Config.ExternalCompareRouterId,
		//	AdvertiseInactiveRoutes:  c.RouteSelectionOptions.Config.AdvertiseInactiveRoutes,
		//	EnableAigp:               c.RouteSelectionOptions.Config.EnableAigp,
		//	IgnoreNextHopIgpMetric:   c.RouteSelectionOptions.Config.IgnoreNextHopIgpMetric,
		//	DisableBestPathSelection: c.RouteSelectionOptions.Config.DisableBestPathSelection,
		//},
		//DefaultRouteDistance: &api.DefaultRouteDistance{
		//	ExternalRouteDistance: uint32(c.DefaultRouteDistance.Config.ExternalRouteDistance),
		//	InternalRouteDistance: uint32(c.DefaultRouteDistance.Config.InternalRouteDistance),
		//},
		//Confederation: &api.Confederation{
		//	Enabled:      c.Confederation.Config.Enabled,
		//	Identifier:   c.Confederation.Config.Identifier,
		//	MemberAsList: c.Confederation.Config.MemberAsList,
		//},
		//GracefulRestart: &api.GracefulRestart{
		//	Enabled:             c.GracefulRestart.Config.Enabled,
		//	RestartTime:         uint32(c.GracefulRestart.Config.RestartTime),
		//	StaleRoutesTime:     uint32(c.GracefulRestart.Config.StaleRoutesTime),
		//	HelperOnly:          c.GracefulRestart.Config.HelperOnly,
		//	DeferralTime:        uint32(c.GracefulRestart.Config.DeferralTime),
		//	NotificationEnabled: c.GracefulRestart.Config.NotificationEnabled,
		//	LonglivedEnabled:    c.GracefulRestart.Config.LongLivedEnabled,
		//},
		//ApplyPolicy: applyPolicy,
	}
}

//数组平分
func splitArray(arr []int, num int64) ([][]int) {
	max := int64(len(arr))
	if max < num {
		return nil
	}
	var segmens = make([][]int, 0)
	quantity := max / num
	end := int64(0)
	for i := int64(1); i <= num; i++ {
		qu := i * quantity
		if i != num {
			segmens = append(segmens, arr[i-1+end:qu])
		} else {
			segmens = append(segmens, arr[i-1+end:])
		}
		end = qu - i
	}
	return segmens
}

