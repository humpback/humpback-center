package cluster

func filterEnginesIPList(engines []*Engine, ipList []string) []*Engine {

	if len(ipList) == 0 {
		return engines
	}

	for _, ip := range ipList {
		for i, engine := range engines {
			if ip == engine.IP {
				engines = append(engines[:i], engines[i+1:]...)
				break
			}
		}
	}
	return engines
}
