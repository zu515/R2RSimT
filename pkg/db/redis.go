package db

import (
	"encoding/json"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"net"
)

func initRedis()(*redis.Client,error)  {
	Redis = redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(FileConfig.Redis.Host, FileConfig.Redis.Port),
		Password: FileConfig.Redis.Password,
		DB:       0,
	})

	pong, err := Redis.Ping().Result()
	log.WithFields(log.Fields{
		"Topic": "RedisInit",
	}).Debugf("redis pong result:%v", pong)

	if err != nil {
		log.WithFields(log.Fields{
			"Topic": "RedisInit",
		}).Errorf("err:%v", err)
		return nil, err
	}
	return Redis,nil
}





func Subscribe(channel chan Message,address string,isController bool) {
	// 订阅channel1这个channel  订阅自己
	var sub *redis.PubSub
	sub = Redis.Subscribe(CHANNEL_SERVER_NODE_KEY + ":" + address)
	if isController {
		sub = Redis.Subscribe(CHANNEL_CONTROLLER_KEY)
	}
	// 读取channel消息
	for {
		iface, err := sub.Receive()
		if err != nil {
			// handle error
		}

		// 检测收到的消息类型
		switch iface.(type) {
		case *redis.Subscription:
			// 订阅成功
		case *redis.Message:
			// 处理收到的消息
			// 这里需要做一下类型转换
			m := iface.(*redis.Message)
			msg := Message{}
			json.Unmarshal([]byte(m.Payload),&msg)
			channel <- msg
		case *redis.Pong:
			// 收到Pong消息
		default:
			// handle error
			return
		}
	}

}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}



func Publish(channelServer string,message interface{})  {
	mesgJson,_:= json.Marshal(message)
	msg := string(mesgJson)
	res := Redis.Publish(channelServer, msg)
	if res.Val() != 1{
		log.WithFields(log.Fields{
			"Topic": "Publish",
			"Key": channelServer,
			"Data": msg,
		}).Error("Publish err:", res)
	}
}
