package cluster

import "github.com/humpback/gounits/utils"

func filterAppendIPList(engine *Engine, ipList []string) []string {

	if engine != nil {
		if ret := utils.Contains(engine.IP, ipList); !ret {
			ipList = append(ipList, engine.IP)
		}
	}
	return ipList
}

func filterIPList(engines []*Engine, ipList []string) []*Engine {

	if len(ipList) == 0 {
		return engines
	}

	p := []*Engine{}
	for _, engine := range engines {
		found := false
		for _, ip := range ipList {
			if engine.IP == ip {
				found = true
				break
			}
		}
		if !found {
			p = append(p, engine)
		}
	}
	return p
}
