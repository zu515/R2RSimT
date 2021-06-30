package db

import (
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/zhishi/R2RSimT/internal/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
)

var Redis *redis.Client
var Mongo *mongo.Client
var FileConfig *config.FileConfig
var MongoCollection *mongo.Collection

var SerialNum string

const (
	BASE_KEY                = "R2RSimT:"
	SERIAL_KEY              = "SERIAL"
	SLAVER_KEY              = "SLAVER"
	CHANNEL_CONTROLLER_KEY  = "CHANNEL_CONTROLLER"
	CHANNEL_SERVER_NODE_KEY = "CHANNEL_SERVER_NODE"

	DISTRIBUTED_NODELIST_KEY = "DISTRIBUTED_NODELIST"
	DEPLOYED_NODELIST_KEY    = "DEPLOYED_NODELIST_KEY"

	NODE_ADRESS_MAP_KEY = "NODE_ADRESS_MAP"
	//计算端口映射
	NODE_CLIENT_PORT_MAP_KEY = "NODE_CLIENT_PORT_MAP"

	CONTROLLER_KEY = "CONTROLLER"
)

const (
	SERVER_STARTED int = iota
	SERVER_DEPLOYED
	SERVER_CONNECTED
	SERVER_CONVERGENCE

	CONTROLLER_DISTRIBUTE
	CONTROLLER_INIT
)
const (
	C2S           = "c2s"
	S2C           = "s2c"
	STATES_CHANGE = "states_change"
	STATES        = "states"
)

type Message struct {
	Type     string
	States   map[string]int
	NodeInfo string
}

func init() {
	FileConfig = ReadProConfigFile()
	Redis,_ = initRedis()
	Mongo, _ = initMongo()

}

func ReadProConfigFile()(*config.FileConfig) {
	fileConfig, err := config.ReadProConfigfile("configs/project.conf", "yaml")
	if err != nil {
		log.WithFields(log.Fields{
			"Topic": "ReadConfigFile",
		}).Errorf("err:%v", err)
	}
	return fileConfig
}
