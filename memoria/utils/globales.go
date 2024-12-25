package utils

// STRUCTS //
import (
	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"sync"
)

type Memoria struct {
	MemoriaSistema map[int32]contextoDePid // El pid sirve como key que accese al contexto del pid
	MemoriaUsuario []byte                  // Espacio de usuario contiguo
}

type contextoDePid struct {
	Estado       string             `json:"estado"` //estado general del proceso
	Base, Limite uint32             //base y limite del proceso
	tidsDePid    map[int32]Contexto //el tid se usa como key para acceder al contexot
}

type Registros struct {
	AX, BX, CX, DX, EX, FX, GX, HX, PC uint32
	//Base, Limite                       uint32
}

type Contexto struct {
	ContextoRegistros Registros `json:"contextoRegistros"`
	Estado            string    `json:"estado"` // estado de un proceso 2 veces???
	Instrucciones     []string
}

type ContextoDeCPU struct {
	ContextoRegistros Registros `json:"contextoRegistros"`
	Base              uint32    `json:"base"`
	Limite            uint32    `json:"limite"`
}

type SolicitudContexto struct {
	Pid int32 `json:"pid"`
	Tid int32 `json:"tid"`
}

type ProcessRequestMemory struct {
	Tamaño int   `json:"tamaño"`
	Pid    int32 `json:"pid"`
}

type ResponseMemory struct {
	Codigo int   `json:"Codigo"`
	Pid    int32 `json:"pid"`
}

type Particion struct {
	base    int
	limite  int
	ocupado bool
	pid     int32
}

type ThreadExitMemory struct {
	Pid int32 `json:"pid"`
	Tid int32 `json:"tid"`
}

type LineaDeInstruccion struct {
	Instruccion     string `json:"instruccion"`
	DireccionLogica uint32
	//DireccionFisica uint32
}

type RespuestaContexto struct {
	Pid      int32         `json:"pid"`
	Tid      int32         `json:"tid"`
	Contexto ContextoDeCPU `json:"contexto"`
}

type RespuestaFetch struct {
	Pid         int32  `json:"pid"`
	Tid         int32  `json:"tid"`
	Instruccion string `json:"instruccion"`
}

type SolicitudDeFetch struct {
	Pid int32 `json:"pid"`
	Tid int32 `json:"tid"`
	PC  int32 `json:"pc"`
}

type ThreadRequestMemory struct {
	Pid       int32  `json:"pid"`
	Tid       int32  `json:"tid"`
	Pathseudo string `json:"archivo"`
}

type ProcessExitMemory struct {
	Pid int32 `json:"pid"`
}

type ProcessExitRequestCpu struct {
	Pid int32 `json:"pid"`
	Tid int32 `json:"tid"`
}

type FileRequest struct {
	Nombre    string `json:"nombre"`
	Tamaño    int    `json:"tamaño"`
	Contenido []byte `json:"contenido"`
}

type DumpMemoryRequest struct {
	Pid       int32  `json:"pid"`
	Tid       int32  `json:"tid"`
	Nombre    string `json:"nombre"`
	Tamaño    int    `json:"tamaño"`
	Contenido string `json:"contenido"`
}

type DumpMemoryResponse struct {
	Pid     int32 `json:"pid"`
	Tid     int32 `json:"tid"`
	Success bool  `json:"success"`
}

type CompactacionRequest struct {
	Estado string `json:"estado"`
}

type SolicitudLectura struct {
	Pid             int32  `json:"pid"`
	Tid             int32  `json:"tid"`
	DireccionFisica uint32 `json:"direccionfisica"`
}

var (
	MemoryConfig *globals.Config = utils_general.IniciarConfig("configs/config.json", &globals.Config{}).(*globals.Config)

	memoria Memoria = Memoria{MemoriaSistema: make(map[int32]contextoDePid), MemoriaUsuario: make([]byte, MemoryConfig.Memory_Size)}

	particiones []Particion

	muAsignarParticion sync.Mutex
)
