package globals

type Config struct {
	Port             int    `json:"port"`
	Memory_Size      int    `json:"memory_size"`
	Instruction_Path string `json:"instruction_path"`
	Response_Delay   int    `json:"response_delay"`
	Ip_Kernel        string `json:"ip_kernel"`
	Port_Kernel      int    `json:"port_kernel"`
	Ip_CPU           string `json:"ip_cpu"`
	Port_CPU         int    `json:"port_cpu"`
	Ip_Filesystem    string `json:"ip_filesystem"`
	Port_Filesystem  int    `json:"port_filesystem"`
	Scheme           string `json:"scheme"`
	Search_Algorithm string `json:"search_algorithm"`
	Partitions       []int  `json:"partitions"`
	Log_level        string `json:"log_level"`
	Fin_de_Linea     string `json:"fin_de_linea"`
}
