package globals

type Config struct {
	Ip_memory           string `json:"ip_memory"`
	Port_memory         int    `json:"port_memory"`
	Ip_kernel           string `json:"ip_kernel"`
	Port_kernel         int    `json:"port_kernel"`
	Log_level           string `json:"log_level"`
	Port                int    `json:"port"`
}