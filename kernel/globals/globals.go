package globals

type Config struct {
	Ip_memory           string `json:"ip_memory"`
	Port_memory         int    `json:"port_memory"`
	Ip_cpu              string `json:"ip_cpu"`
	Port_cpu            int    `json:"port_cpu"`
	Scheduler_algorithm string `json:"scheduler_algorithm"`
	Quantum             int    `json:"quantum"`
	Log_level           string `json:"log_level"`
	Port            	int    `json:"port"`
}