package globals

type Config struct {
	Port               int    `json:"port"`
	Ip_memory          string `json:"ip_memory"`
	Port_memory        int    `json:"port_memory"`
	Mount_dir          string `json:"mount_dir"`
	Block_size         int    `json:"block_size"`
	Block_count        int    `json:"block_count"`
	Block_access_delay int    `json:"block_access_delay"`
	Log_level          string `json:"log_level"`
}
