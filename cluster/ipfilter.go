package cluster

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
