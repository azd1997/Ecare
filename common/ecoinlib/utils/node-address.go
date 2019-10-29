package utils


func NodeIsKnown(nodeAddr string, KnownNodeAddrList []string) bool {
	for _, node := range KnownNodeAddrList {
		if node == nodeAddr {
			return true
		}
	}
	return false
}
