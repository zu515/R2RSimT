package main

import (
	"fmt"
	api "github.com/zhishi/R2RSimT/api"
	"github.com/zhishi/R2RSimT/internal/pkg/apiutil"
	"github.com/zhishi/R2RSimT/pkg/packet/bgp"
	"io"
	"net"
	"time"
	"context"
)

func GetRib(client api.GobgpApiClient) {
	ctx := context.Background()
	r, _ := client.GetBgp(ctx, &api.GetBgpRequest{})
	fmt.Printf("%v", r.Global.String())
	fmt.Println()
}

func GetGlobalRib(client api.GobgpApiClient) () {
	ctx := context.Background()

	familys := []*api.Family{ipv4UC, ipv6UC}
	for _, family := range familys {
		family, err := checkAddressFamily(family)
		var filter []*api.TableLookupPrefix
		stream, err := client.ListPath(ctx, &api.ListPathRequest{
			TableType: api.TableType_GLOBAL,
			Family:    family,
			Name:      "",
			Prefixes:  filter,
			SortType:  api.ListPathRequest_PREFIX,
		})
		if err != nil {
			fmt.Println(err)
		}
		for {
			r, err := stream.Recv()
			if err == io.EOF {
				break
			} else if err != nil {
				fmt.Println(err)
			}
			//destination := &model.Destination{}
			//destination.Paths = make([]*model.PathAttr, 0)
			//destination.Prefix = r.Destination.Prefix
			for _, path := range r.Destination.Paths {
				fmt.Printf("%+v", path.Nlri.String())
				//destinationPath := &model.PathAttr{}
				//destinationPath.Best = path.Best
				////destinationPath.IsWithdraw = path.IsWithdraw
				////var validation *model.Validation

				//validation := &model.Validation{}
				//validation.State = path.Validation.State
				//validation.Reason = path.Validation.Reason
				//validation.Matched = path.Validation.Matched
				//validation.UnmatchedLength = path.Validation.UnmatchedLength
				//validation.UnmatchedAs = path.Validation.UnmatchedAs
				//
				//destinationPath.Validation = validation
				//
				//if path.Family.Afi == ipv4UC.Afi {
				//	destinationPath.Family = "ipv4"
				//} else {
				//	destinationPath.Family = "ipv6"
				//}
				//destinationPath.SourceAsn = path.SourceAsn
				////destinationPath.SourceId = path.SourceId
				//destinationPath.Filtered = path.Filtered
				////destinationPath.NeighborIp = path.NeighborIp
				//attrs, _ := apiutil.GetNativePathAttributes(r.Destination.Paths[0])
				//aspathstr := func() string {
				//	for _, attr := range attrs {
				//		switch a := attr.(type) {
				//		case *bgp.PathAttributeAsPath:
				//			return bgp.AsPathString(a)
				//		}
				//	}
				//	return ""
				//}()
				//destinationPath.AS_PATH = aspathstr
				//destination.Paths = append(destination.Paths, destinationPath)
			}

		}
	}

}

func AddRib(client api.GobgpApiClient, family string, prefix string) (err error) {
	ctx := context.Background()
	attrs := make([]bgp.PathAttributeInterface, 0, 1)
	f, err := checkAddressFamily(familyMap[family])
	if err != nil {
		fmt.Println(err)
	}
	rf := apiutil.ToRouteFamily(f)

	var nlri bgp.AddrPrefixInterface

	addr, nw, err := net.ParseCIDR(prefix)
	ones, _ := nw.Mask.Size()

	if rf == bgp.RF_IPv4_UC {
		nlri = bgp.NewIPAddrPrefix(uint8(ones), addr.String())
	} else {
		nlri = bgp.NewIPv6AddrPrefix(uint8(ones), addr.String())
	}
	attrs = append(attrs, bgp.NewPathAttributeNextHop("1.1.1.1"))
	attrs = append(attrs, bgp.NewPathAttributeOrigin(0))
	path := apiutil.NewPath(nlri, false, attrs, time.Now())
	r := api.TableType_GLOBAL
	j, err := client.AddPath(ctx, &api.AddPathRequest{
		TableType: r,
		Path:      path,
	})

	//ctx := context.Background()
	//client, err := GetClient(containerID, 50051)
	//if err != nil {
	//	belogs.Error("AddRib() GetClient: err:", err)
	//	return err
	//}
	//attrs := make([]bgp.PathAttributeInterface, 0)
	//var nlri bgp.AddrPrefixInterface
	//addr, nw, err := net.ParseCIDR(prefix)
	//ones, _ := nw.Mask.Size()
	//if family == "ipv4"{
	//	nlri = bgp.NewIPAddrPrefix(uint8(ones), addr.String())
	//} else {
	//	nlri = bgp.NewIPv6AddrPrefix(uint8(ones), addr.String())
	//}
	//attrs = append(attrs, bgp.NewPathAttributeNextHop("0.0.0.0"))
	//attrs = append(attrs, bgp.NewPathAttributeOrigin(uint8(Host.ContainerMap[containerID].Router.ASN)))
	//path := apiutil.NewPath(nlri, false, attrs, time.Now())
	//j, err := client.AddPath(ctx, &api.AddPathRequest{
	//	TableType: api.TableType_GLOBAL,
	//	Path: path,
	//})

	fmt.Println("添加路径", j)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil

}
