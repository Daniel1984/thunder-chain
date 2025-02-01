package peernode

type Node struct {
	IP       string `json:"ip"`
	Port     uint64 `json:"port"`
	IsActive bool   `json:"is_active"`
}
