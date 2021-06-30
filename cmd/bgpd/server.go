package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-systemd/daemon"
	"github.com/jessevdk/go-flags"
	"github.com/kr/pretty"
	log "github.com/sirupsen/logrus"
	api "github.com/zhishi/R2RSimT/api"
	"github.com/zhishi/R2RSimT/internal/pkg/version"
	"github.com/zhishi/R2RSimT/pkg/config"
	"github.com/zhishi/R2RSimT/pkg/db"
	"github.com/zhishi/R2RSimT/pkg/etcd"
	"github.com/zhishi/R2RSimT/pkg/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"

)

const (
	RIB_ADD = "RIB_ADD"
	RIB_DEL = "RIB_DEL"
)

var ActionMap = map[bool]string{
	false: RIB_ADD,
	true:  RIB_ADD,
}

type ServerNode struct {
	MessageChan      chan db.Message
	RibMonitorChan   chan MonitorMessage
	StartedChan      chan bool
	StartedServerNum int64
	States           int
	Addr             string
}

func nodeStart(wg sync.WaitGroup, asn uint32, serverNode *ServerNode) {
	defer wg.Done()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)

	var opts struct {
		ConfigFile      string `short:"f" long:"config-file" description:"specifying a config file"`
		ConfigType      string `short:"t" long:"config-type" description:"specifying config type (toml, yaml, json)" default:"toml"`
		LogLevel        string `short:"l" long:"log-level" description:"specifying log level"`
		LogPlain        bool   `short:"p" long:"log-plain" description:"use plain format for logging (json by default)"`
		UseSyslog       string `short:"s" long:"syslog" description:"use syslogd"`
		Facility        string `long:"syslog-facility" description:"specify syslog facility"`
		DisableStdlog   bool   `long:"disable-stdlog" description:"disable standard logging"`
		CPUs            int    `long:"cpus" description:"specify the number of CPUs to be used"`
		GrpcHosts       string `long:"api-hosts" description:"specify the hosts that gobgpd listens on" default:":50051"`
		GracefulRestart bool   `short:"r" long:"graceful-restart" description:"flag restart-state in graceful-restart capability"`
		Dry             bool   `short:"d" long:"dry-run" description:"check configuration"`
		PProfHost       string `long:"pprof-host" description:"specify the host that gobgpd listens on for pprof" default:"localhost:6060"`
		PProfDisable    bool   `long:"pprof-disable" description:"disable pprof profiling"`
		UseSdNotify     bool   `long:"sdnotify" description:"use sd_notify protocol"`
		TLS             bool   `long:"tls" description:"enable TLS authentication for gRPC API"`
		TLSCertFile     string `long:"tls-cert-file" description:"The TLS cert file"`
		TLSKeyFile      string `long:"tls-key-file" description:"The TLS key file"`
		Version         bool   `long:"version" description:"show version number"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}
	opts.GrpcHosts = ":" + strconv.Itoa(int(server.TopologyConfig.NodeMap[asn].GrpcPort))
	opts.PProfHost = net.JoinHostPort(serverNode.Addr, strconv.Itoa(int(server.TopologyConfig.NodeMap[asn].PProfHost)))

	opts.LogPlain = true

	if opts.Version {
		fmt.Println("gobgpd version", version.Version())
		os.Exit(0)
	}

	if opts.CPUs == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		if runtime.NumCPU() < opts.CPUs {
			log.Errorf("Only %d CPUs are available but %d is specified", runtime.NumCPU(), opts.CPUs)
			os.Exit(1)
		}
		runtime.GOMAXPROCS(opts.CPUs)
	}

	opts.PProfDisable = true
	if !opts.PProfDisable {
		go func() {
			log.Println(http.ListenAndServe(opts.PProfHost, nil))
		}()
	}

	switch opts.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	log.SetLevel(log.DebugLevel)
	if opts.DisableStdlog {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stdout)
	}

	if opts.UseSyslog != "" {
		if err := addSyslogHook(opts.UseSyslog, opts.Facility); err != nil {
			log.Error("Unable to connect to syslog daemon, ", opts.UseSyslog)
		}
	}

	if opts.LogPlain {
		if opts.DisableStdlog {
			log.SetFormatter(&log.TextFormatter{
				DisableColors: true,
			})
		}
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}

	if opts.Dry {
		c, err := config.ReadConfigFile(opts.ConfigFile, opts.ConfigType)
		if err != nil {
			log.WithFields(log.Fields{
				"Topic": "Config",
				"Error": err,
			}).Fatalf("Can't read config file %s", opts.ConfigFile)
		}
		log.WithFields(log.Fields{
			"Topic": "Config",
		}).Info("Finished reading the config file")
		if opts.LogLevel == "debug" {
			pretty.Println(c)
		}
		os.Exit(0)
	}

	maxSize := 256 << 20
	grpcOpts := []grpc.ServerOption{grpc.MaxRecvMsgSize(maxSize), grpc.MaxSendMsgSize(maxSize)}
	if opts.TLS {
		creds, err := credentials.NewServerTLSFromFile(opts.TLSCertFile, opts.TLSKeyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials: %v", err)
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}
	v, _ := json.Marshal(server.TopologyConfig.NodeMap[asn])
	log.Info("-----------------gobgpd started--------------------", asn, string(v))

	bgpServer := server.NewBgpServer(server.GrpcListenAddress(opts.GrpcHosts), server.GrpcOption(grpcOpts))
	go bgpServer.Serve()

	if opts.UseSdNotify {
		if status, err := daemon.SdNotify(false, daemon.SdNotifyReady); !status {
			if err != nil {
				log.Warnf("Failed to send notification via sd_notify(): %s", err)
			} else {
				log.Warnf("The socket sd_notify() isn't available")
			}
		}
	}

	if err := bgpServer.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: GetGlobalConfig(asn),
	}); err != nil {
		log.Fatalf("failed to set global config: %s", err)
		return
	}
	confFile, _ := config.ReadProConfigFile("configs/project.conf", "yaml")
	grpcAddr := net.JoinHostPort(server.TopologyConfig.NodeMap[asn].Adress, strconv.Itoa(int(server.TopologyConfig.NodeMap[asn].GrpcPort)))
	go etcd.Register(confFile.Etcd.Host, strconv.Itoa(int(asn)), grpcAddr, 5)
	startedServerNum := atomic.AddInt64(&serverNode.StartedServerNum, -1)
	if startedServerNum <= 0 {
		log.WithFields(log.Fields{
			"Topic": "ServerDeployed",
			"Key":   serverNode.Addr,
		}).Info("states change to deployed")
		serverNode.States = db.SERVER_DEPLOYED
		msg := db.Message{
			Type:     db.STATES_CHANGE,
			States:   map[string]int{db.STATES: db.SERVER_DEPLOYED},
			NodeInfo: serverNode.Addr,
		}
		serverNode.MessageChan <- msg
	}

	go StartMoniter(bgpServer, asn, serverNode)

	if opts.ConfigFile == "" {
		//阻塞在这里
		<-sigCh
		stopServer(bgpServer, opts.UseSdNotify)
		return
	}

	signal.Notify(sigCh, syscall.SIGHUP)

	initialConfig, err := config.ReadConfigFile(opts.ConfigFile, opts.ConfigType)
	if err != nil {
		log.WithFields(log.Fields{
			"Topic": "Config",
			"Error": err,
		}).Fatalf("Can't read config file %s", opts.ConfigFile)
	}
	log.WithFields(log.Fields{
		"Topic": "Config",
	}).Info("Finished reading the config file")

	currentConfig, err := config.InitialConfig(context.Background(), bgpServer, initialConfig, opts.GracefulRestart)
	if err != nil {
		log.WithFields(log.Fields{
			"Topic": "Config",
			"Error": err,
		}).Fatalf("Failed to apply initial configuration %s", opts.ConfigFile)
	}
	for sig := range sigCh {
		if sig != syscall.SIGHUP {
			stopServer(bgpServer, opts.UseSdNotify)
			return
		}

		log.WithFields(log.Fields{
			"Topic": "Config",
		}).Info("Reload the config file")
		newConfig, err := config.ReadConfigFile(opts.ConfigFile, opts.ConfigType)
		if err != nil {
			log.WithFields(log.Fields{
				"Topic": "Config",
				"Error": err,
			}).Warningf("Can't read config file %s", opts.ConfigFile)
			continue
		}

		currentConfig, err = config.UpdateConfig(context.Background(), bgpServer, currentConfig, newConfig)
		if err != nil {
			log.WithFields(log.Fields{
				"Topic": "Config",
				"Error": err,
			}).Warningf("Failed to update config %s", opts.ConfigFile)
			continue
		}
	}
}

func stopServer(bgpServer *server.BgpServer, useSdNotify bool) {
	bgpServer.StopBgp(context.Background(), &api.StopBgpRequest{})
	if useSdNotify {
		daemon.SdNotify(false, daemon.SdNotifyStopping)
	}
}

func StartMoniter(bgpServer *server.BgpServer, asn uint32, serverNode *ServerNode) {
	for _, ipType := range []*api.Family{ipv4UC, ipv6UC} {
		family, err := checkAddressFamily(ipType)
		if err != nil {
			log.WithFields(log.Fields{
				"Topic": "StartMoniter",
				"Key":   "checkAddressFamily",
			}).Error("GetFreePorts err:", err)
		}
		ctx, _ := context.WithCancel(context.Background())
		bgpServer.MonitorTable(ctx, &api.MonitorTableRequest{
			TableType: api.TableType_GLOBAL,
			Family:    family,
			Current:   true,
		}, func(p *api.Path) {
			pathRecord := MonitorMessage{
				Asn:      asn,
				Action:   ActionMap[p.IsWithdraw],
				PathInfo: p,
			}
			serverNode.RibMonitorChan <- pathRecord
		})
	}

}
