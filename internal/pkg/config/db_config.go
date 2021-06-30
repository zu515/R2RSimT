package config

import (
	"github.com/spf13/viper"
)

type Redis struct {
	Host 		string `mapstructure:"host"`
	Port 		string `mapstructure:"port"`
	Password 	string `mapstructure:"password"`
}

type Mongo struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DB       string `mapstructure:"db"`
}

type Etcd struct {
	Host string `mapstructure:"host"`

}
type FileConfig struct {
	Redis            Redis `mapstructure:"redis"`
	Mongo            Mongo `mapstructure:"mongo"`
	BGPServers 		 []string `mapstructure:"bgpservers"`
	Etcd 			 Etcd `mapstructure:"etcd"`
}

type BGPServers struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
}



func ReadProConfigfile(path, format string) (*FileConfig, error) {
	// Update config file type, if detectable
	format = detectConfigFileType(path, format)

	config := &FileConfig{}
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType(format)
	var err error
	if err = v.ReadInConfig(); err != nil {
		return nil, err
	}
	if err = v.UnmarshalExact(config); err != nil {
		return nil, err
	}

	return config, nil
}
//
//func ConfigSetToRoutingPolicy(c *BgpConfigSet) *RoutingPolicy {
//	return &RoutingPolicy{
//		DefinedSets:       c.DefinedSets,
//		PolicyDefinitions: c.PolicyDefinitions,
//	}
//}
//
//func UpdatePeerGroupConfig(curC, newC *BgpConfigSet) ([]PeerGroup, []PeerGroup, []PeerGroup) {
//	addedPg := []PeerGroup{}
//	deletedPg := []PeerGroup{}
//	updatedPg := []PeerGroup{}
//
//	for _, n := range newC.PeerGroups {
//		if idx := existPeerGroup(n.Config.PeerGroupName, curC.PeerGroups); idx < 0 {
//			addedPg = append(addedPg, n)
//		} else if !n.Equal(&curC.PeerGroups[idx]) {
//			log.WithFields(log.Fields{
//				"Topic": "Config",
//			}).Debugf("Current peer-group config:%v", curC.PeerGroups[idx])
//			log.WithFields(log.Fields{
//				"Topic": "Config",
//			}).Debugf("New peer-group config:%v", n)
//			updatedPg = append(updatedPg, n)
//		}
//	}
//
//	for _, n := range curC.PeerGroups {
//		if existPeerGroup(n.Config.PeerGroupName, newC.PeerGroups) < 0 {
//			deletedPg = append(deletedPg, n)
//		}
//	}
//	return addedPg, deletedPg, updatedPg
//}
//
//func UpdateNeighborConfig(curC, newC *BgpConfigSet) ([]Neighbor, []Neighbor, []Neighbor) {
//	added := []Neighbor{}
//	deleted := []Neighbor{}
//	updated := []Neighbor{}
//
//	for _, n := range newC.Neighbors {
//		if idx := inSlice(n, curC.Neighbors); idx < 0 {
//			added = append(added, n)
//		} else if !n.Equal(&curC.Neighbors[idx]) {
//			log.WithFields(log.Fields{
//				"Topic": "Config",
//			}).Debugf("Current neighbor config:%v", curC.Neighbors[idx])
//			log.WithFields(log.Fields{
//				"Topic": "Config",
//			}).Debugf("New neighbor config:%v", n)
//			updated = append(updated, n)
//		}
//	}
//
//	for _, n := range curC.Neighbors {
//		if inSlice(n, newC.Neighbors) < 0 {
//			deleted = append(deleted, n)
//		}
//	}
//	return added, deleted, updated
//}
//
//func CheckPolicyDifference(currentPolicy *RoutingPolicy, newPolicy *RoutingPolicy) bool {
//
//	log.WithFields(log.Fields{
//		"Topic": "Config",
//	}).Debugf("Current policy:%v", currentPolicy)
//	log.WithFields(log.Fields{
//		"Topic": "Config",
//	}).Debugf("New policy:%v", newPolicy)
//
//	var result bool
//	if currentPolicy == nil && newPolicy == nil {
//		result = false
//	} else {
//		if currentPolicy != nil && newPolicy != nil {
//			result = !currentPolicy.Equal(newPolicy)
//		} else {
//			result = true
//		}
//	}
//	return result
//}
//
