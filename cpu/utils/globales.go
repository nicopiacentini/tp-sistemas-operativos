package utils

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"net/http"
	"sync"
)

// STRUCTS //

// Registros del CPU (4 bytes cada uno)
type Registros struct {
	PC, AX, BX, CX, DX, EX, FX, GX, HX uint32
}

type solicitudKernel struct {
	Pid int32 `json:"pid"`
	Tid int32 `json:"tid"`
}

type Contexto struct {
	ContextoRegistros Registros `json:"contextoRegistros"`
	Base              uint32    `json:"base"`
	Limite            uint32    `json:"limite"`
}

type RespuestaContexto struct {
	Pid      int32    `json:"pid"`
	Tid      int32    `json:"tid"`
	Contexto Contexto `json:"contexto"`
}

type SolicitudMemoria struct {
	Pid int32 `json:"pid"`
	Tid int32 `json:"tid"`
}

type SolicitudFetch struct {
	Pid int32  `json:"pid"`
	Tid int32  `json:"tid"`
	Pc  uint32 `json:"pc"`
}

type RespuestaFetch struct {
	Pid         int32  `json:"pid"`
	Tid         int32  `json:"tid"`
	Instruccion string `json:"instruccion"`
}

type RespuestaMemoria struct {
	Instruccion     string `json:"instruccion"`
	DireccionLogica uint32
	DireccionFisica uint32
}

// Solicitud para enviar Syscall al kernel
type ProcessCreateRequestCpu struct {
	Archivo   string `json:"archivo"`
	Tamaño    int    `json:"tamaño"`
	Prioridad int    `json:"prioridad"`
}

type ThreadCreateRequestCpu struct {
	Archivo   string `json:"archivo"`
	Prioridad int    `json:"prioridad"`
}

type ThreadJoinRequestCpu struct {
	Tid int `json:"tid"`
}

type ThreadCancelRequestCpu struct {
	Tid int `json:"tid"`
}

type MutexRequestCpu struct {
	Recurso string `json:"recurso"`
}

type IORequestCpu struct {
	Tiempo int `json:"tiempo"`
}

type Interrupcion struct {
	Pid int32 `json:"pid"`
	Tid int32 `json:"tid"`
}

type RespuestaKernel struct {
	Tid    int32  `json:"tid"`
	Motivo string `json:"motivo"`
}

type PidTid struct {
	Pid int32 `json:"pid"`
	Tid int32 `json:"tid"`
}

type RequestVacia struct {
	Syscall string `json:"syscall"`
}

type SolicitudLectura struct {
	Pid             int32  `json:"pid"`
	Tid             int32  `json:"tid"`
	DireccionFisica uint32 `json:"direccionfisica"`
}

// VARIABLES //
var (
	CpuConfig            *globals.Config = utils_general.IniciarConfig("configs/config.json", &globals.Config{}).(*globals.Config)
	pidTidRecibidos                      = make(map[int32]int32)
	client                               = &http.Client{}
	interrupciones                       = make(map[PidTid]bool)
	ErrSegmentationFault                 = errors.New("segmentation fault: acceso inválido a la memoria")
	tablaContextos                       = make(map[PidTid]*Contexto)
	muInterrupt          sync.Mutex
)
