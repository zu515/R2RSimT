package server

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhishi/R2RSimT/pkg/db"
	"github.com/zhishi/R2RSimT/pkg/packet/bgp"
	"net"
	"sync"
)

type Path struct {
	pathAttrs []bgp.PathAttributeInterface
	dels      []bgp.BGPAttrType
	attrsHash uint32
	rejected  bool
	// doesn't exist in the adj
	dropped bool

	// For BGP Nexthop Tracking, this field shows if nexthop is invalidated by IGP.
	IsNexthopInvalid bool
	IsWithdraw       bool
}
type Destination struct {
	nlri          bgp.AddrPrefixInterface
	knownPathList []*Path
}
type Table struct {
	routeFamily  bgp.RouteFamily
	destinations map[string]*Destination
}

type BGPNodeConfig struct {
	ASN        uint32
	RouterID   string
	GrpcPort   uint32
	ServerPort uint32
	PProfHost  uint32
	Adress     string
	Peer       []uint32
	//table table.Table
	States     int
	RibTableSize uint32
	RoaTableSize uint32

}

type BGPTopologyConfig struct {
	NodeMap          map[uint32]*BGPNodeConfig
	RouterId2AsnMap  map[string]uint32
	Asn2RouterIdMap  map[uint32]string
	PeerMap          map[uint32][]*BGPNodeConfig
	ClientConPortMap map[string]string
}

func NewBGPTopologyConfig() *BGPTopologyConfig {
	return &BGPTopologyConfig{
		NodeMap:          make(map[uint32]*BGPNodeConfig),
		RouterId2AsnMap:  make(map[string]uint32),
		Asn2RouterIdMap:  make(map[uint32]string),
		PeerMap:          make(map[uint32][]*BGPNodeConfig),
		ClientConPortMap: make(map[string]string),

	}
}



var TopologyConfig *BGPTopologyConfig

var ClientConPortMapLock sync.RWMutex
func (T *BGPTopologyConfig) GetClientConPortMap( address ,port string)(addressAndPort string)  {
	addressAndPort = db.Redis.HGet(db.BASE_KEY+db.SerialNum+":"+address+":"+db.NODE_CLIENT_PORT_MAP_KEY, port).Val()
	//ClientConPortMapLock.RLock()
	//addressAndPort =  T.ClientConPortMap[net.JoinHostPort(address,port)]
	//ClientConPortMapLock.RUnlock()
	return addressAndPort
}
func (T *BGPTopologyConfig) AddClientConPortMap(address,clientPort, serverPort  string){
	//+":" + strconv.Itoa(int(node.ASN))
	err := db.Redis.HSet(db.BASE_KEY+db.SerialNum+":"+address + ":" +db.NODE_CLIENT_PORT_MAP_KEY, clientPort,serverPort).Err()
	if err != nil{
		log.WithFields(log.Fields{
			"Topic": "GetDistributeNode",
			"Key":   net.JoinHostPort(address,serverPort),
			"Data": net.JoinHostPort(address, serverPort),
		}).Error("AddClientConPortMap err:", err)
	}
	//ClientConPortMapLock.RLock()
	//T.ClientConPortMap[net.JoinHostPort(address, clientPort)] = net.JoinHostPort(address, serverPort)
	//ClientConPortMapLock.RUnlock()
}

type BGPServerStartMsg struct {
	BGPNode BGPNodeConfig
}
var BgpServerStartChan chan BGPServerStartMsg

