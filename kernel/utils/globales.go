package utils

import (
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"sync"
)

// STRUCTS //
type ProcessCreateRequestCpu struct {
	Archivo   string `json:"archivo"`
	Tamaño    int    `json:"tamaño"`
	Prioridad int32  `json:"prioridad"`
}

type ProcessDumpRequestMemory struct {
	Pid int32 `json:"pid"`
	Tid int32 `json:"tid"`
}

type ProcessFinishMemory struct {
	Pid int32 `json:"pid"`
}

type ThreadCreateRequestCpu struct {
	Archivo   string `json:"archivo"`
	Prioridad int32  `json:"prioridad"`
}

type ThreadJoinRequestCpu struct {
	Tid int32 `json:"tid"`
}

type ThreadCancelRequestCpu struct {
	Tid int32 `json:"tid"`
}

type ProcessRequestMemory struct {
	Tamaño int   `json:"tamaño"`
	Pid    int32 `json:"pid"`
}

type ThreadRequestMemory struct {
	Archivo string `json:"archivo"`
	Pid     int32  `json:"pid"`
	Tid     int32  `json:"tid"`
}

type ThreadFinishMemory struct {
	Pid int32 `json:"pid"`
	Tid int32 `json:"tid"`
}

type MutexRequestCpu struct {
	Recurso string `json:"recurso"`
}

type IORequestCpu struct {
	Tiempo int `json:"tiempo"`
}

type ExecuteRequestCPU struct {
	Pid int32 `json:"pid"`
	Tid int32 `json:"tid"`
}

type RequestVacia struct {
	Syscall string `json:"syscall"`
}

type ResponseMemory struct {
	Codigo int   `json:"codigo"`
	Pid    int32 `json:"pid"`
}

type ThreadReturnedRequestCpu struct {
	Tid    int32  `json:"tid"`
	Motivo string `json:"motivo"`
}

type InterruptionRequestCpu struct {
	Pid int32 `json:"pid"`
	Tid int32 `json:"tid"`
}

type ProcessDumpResponseMemory struct {
	Pid     int32 `json:"pid"`
	Tid     int32 `json:"tid"`
	Success bool  `json:"success"`
}

type Mutex struct {
	Recurso       string
	Asignado      bool
	Dueño         *Tcb
	HilosEnEspera []*Tcb
}

type Pcb struct {
	Pid     int32
	Tids    []int32
	Mutexes []*Mutex
}

type Tcb struct {
	Tid        int32
	Prioridad  int32
	Pid        int32
	Bloqueados []*Tcb
}

type Params struct {
	Archivo   string
	Tamaño    int
	Prioridad int32
	Pid       int32
	Tid       int32
	Hilo      Tcb
}

type ParamsInitMemoria struct {
	Archivo       string
	Tamaño        int
	PrioridadTid0 int32
}

type CompactacionRequest struct {
	Estado string `json:"estado"`
}

// VARIABLES GLOBALES //
var (
	global_pid          int32                               = 0
	KernelConfig        *globals.Config                     = utils_general.IniciarConfig("configs/config.json", &globals.Config{}).(*globals.Config)
	hiloEnEjecucion     *Tcb                                // Lleva un registro constante del hilo que está en ejecución
	canalIO             = make(chan struct{}, 1)            // Sólo 1 a la vez en IO
	paramsInitMemoria   = make(map[int32]ParamsInitMemoria) // Mapa temporal para guardar los parámetros de inicialización
	planificadorPausado bool
	tengoElControl      bool = true
)

// SEMÁFOROS //
var (
	mutexGlobalPid         sync.Mutex
	mutexColaNew           sync.Mutex
	mutexColaReady         sync.Mutex
	mutexColaBlocked       sync.Mutex
	mutexColaExit          sync.Mutex
	hiloSolicitado         = sync.NewCond(&sync.Mutex{})
	mutexPcbs              sync.Mutex
	mutexHiloEnEjecucion   sync.Mutex
	mutexHilosEnEspera     sync.Mutex
	mutexColasMultinivel   sync.Mutex
	mutexIO                sync.Mutex
	mutexParamsInitMemoria sync.Mutex
	pauseChan              = make(chan struct{}, 1)
	hiloEjecutando         = sync.NewCond(&sync.Mutex{})
)

// Colas //
var (
	colaNew         []Pcb
	colaReady       []Tcb
	colaBlocked     []Tcb
	colaExit        []Tcb
	pcbs            []Pcb                 // Array global para almacenar los PCBs
	colasMultinivel = make(map[int][]Tcb) //colasMultinivel[2].isEmpty()
)
