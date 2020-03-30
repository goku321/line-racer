package main

type node struct {
	ID     int    `json:"id"`
	IPAddr int    `json:"ip_addr"`
	Port   string `json:"port"`
	Status string `json:"status"`
}

type data struct {
	Source  node
	Dest    node
	Message []int
}
