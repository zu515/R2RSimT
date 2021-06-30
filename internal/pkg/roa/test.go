package main

import (
	"fmt"
	"github.com/zhishi/R2RSimT/internal/pkg/config"
	"github.com/zhishi/R2RSimT/internal/pkg/table"
	"github.com/zhishi/R2RSimT/pkg/packet/bgp"
	"net"
	"strconv"
	"strings"
	"time"
)

func strToASParam(str string) *bgp.PathAttributeAsPath {
	toList := func(asstr, sep string) []uint32 {
		as := make([]uint32, 0)
		l := strings.Split(asstr, sep)
		for _, s := range l {
			v, _ := strconv.ParseUint(s, 10, 32)
			as = append(as, uint32(v))
		}
		return as
	}
	var atype uint8
	var as []uint32
	if strings.HasPrefix(str, "{") {
		atype = bgp.BGP_ASPATH_ATTR_TYPE_SET
		as = toList(str[1:len(str)-1], ",")
	} else if strings.HasPrefix(str, "(") {
		atype = bgp.BGP_ASPATH_ATTR_TYPE_CONFED_SET
		as = toList(str[1:len(str)-1], " ")
	} else {
		atype = bgp.BGP_ASPATH_ATTR_TYPE_SEQ
		as = toList(str, " ")
	}

	return bgp.NewPathAttributeAsPath([]bgp.AsPathParamInterface{bgp.NewAs4PathParam(atype, as)})
}

func validateOne(rt *table.ROATable, cidr, aspathStr string) config.RpkiValidationResultType {
	var nlri bgp.AddrPrefixInterface
	ip, r, _ := net.ParseCIDR(cidr)
	length, _ := r.Mask.Size()
	if ip.To4() == nil {
		nlri = bgp.NewIPv6AddrPrefix(uint8(length), ip.String())
	} else {
		nlri = bgp.NewIPAddrPrefix(uint8(length), ip.String())
	}
	attrs := []bgp.PathAttributeInterface{strToASParam(aspathStr)}
	path := table.NewPath(&table.PeerInfo{LocalAS: 65500}, nlri, false, attrs, time.Now(), false)
	ret := rt.Validate(path)
	return ret.Status
}
func main()  {

	roaTable := table.NewROATable()
	roaTable.Add(table.NewROA(bgp.AFI_IP, net.ParseIP("192.168.0.0").To4(), 24, 32, 100, ""))
	roaTable.Add(table.NewROA(bgp.AFI_IP, net.ParseIP("192.168.0.0").To4(), 24, 24, 200, ""))

	r := validateOne(roaTable, "192.168.0.0/24", "100")
	fmt.Println(r)
}
