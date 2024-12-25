# preguntas para hacer el sábado: 

	1. podemos tener la lista de tcbs en el pcb en lugar de la lista de tids? para vincularlos con su prioridad
	RESP : si

	2. de dónde salen los parámetros que recibe kernel para el proceso inicial?
	RESP: Linea de comandos

	3. tener los registros de cpu en un struct o todos como variables globales? 
	RESP: MEJOR CON STRUCT

	4. para el proceso inicial que crea kernel, hace falta validar que memoria tenga espacio? es decir, nosotros le solicitamos la inicialización del proceso a memoria para que le asigne sus recursos, pero considerando que es el primer proceso de todos, siempre va a haber espacio disponible para este en memoria, por lo que no haria falta validar si la respuesta de memoria es afirmativa, ya que este proceso nunca se encolará en NEW por ser el primero de todos, simplemente habria que avisarle a memoria que asigne los recursos para este
	RESP: DEBE PASAR POR NEW DE TODOS MODOS, YA QUE PODRÍA PASAR QUE EL PROCESO INICIAL PESE MÁS QUE TODO EL DISPONIBLE EN MEMORIA

	5. el proceso inicial que crea kernel pasa directo a READY?
	RESP: DEBE PASAR POR NEW DE TODOS MODOS

	6. la creacion del proceso inicial ocurre ni bien se inicia el modulo kernel o alguien debe solicitarsela a traves de una api?
	RESP: APENAS SE LEVANTA, FUNCION INIT Y CON OS.ARGS AGARRAMOS EL ARCHIVO QUE LE PASAMOS AL INVOCAR EL EJECUTABLE

	7. la secuencia para crear un proceso es?: 
      	1. crear pcb
      	2. crear proceso con dicho pcb y asociarle los parametros recibidos
      	3. dentro de creacion de pcb contemplamos la creacion del tid 0 pero sin asignale la prioridad(tener en cuenta que ya nos llego de cpu la prioridad)
      	4. encolar en NEW el pcb
      	5. planificar a largo plazo la cola de new (por fifo) -> agarrar el primero que esta en la cola y pedir memorias
            	1. respuesta not success->esperar a la finalizacion de otro proceso y volver a ejecutar (al hacer esto mandamos el proceso al final de la cola new?)
            	2. respuesta success ->se le asigna prioridad y se envia a ready
	RESP: Si

	8. para inicializar el proceso la memoria solo necesita tamaño y pid o algo mas?	
	RESP:  si, el path se manda cuando creo el hilo junto con el pid 
   
	9. si hay un proceso muy pesado en new que no puede ser inicializado y el mismo debe esperar, este debe quedarse esperando al principio de la cola y generar una especie de cuello de botella o lo mandamos al final de new para que no retrase a otros procesos tal vez mas livianos
    RESP: se queda donde esta y genera un cuello de botella a proposito
   
	10. Si tanto los procesos como los hilos tienen pseudocodigo, que es lo que se ejecuta, los hilos o los procesos? 
    RESP: el pseudocodigo del proceso con process create es el de su tid 0, y el de thread create es el de ese hilo en particular. El proceso llega hasta new, luego se planifican hilos sin importar de que proceso son

	11. Que recibe cpu de kernel
	RESP: solo necesita tid y pid

	12. Cuando ocurre un PROCESS_EXIT, quién debe informarle a memoria la finalización del proceso, cpu o kernel? o ambos?
	RESP: Kernel.

	13. En memoria se almacena el pcb?
	RESP: No

	14. Dónde ponemos los "go"?
	RESP: Donde se necesito que un hilo se bloquee haciendo algo mientras otra cosa sigue ejecutando.

	15: Todas las syscalls son bloqueantes? La duda surge porque no sabríamos si un hilo al hacer MUTEX_LOCK debería bloquearse luego de que se le asigne, lo cual no nos hace sentido.
	RESP: No

	16: Qué pasa si al final del archivo no hay un THREAD_EXIT, o en caso del tid 0, no hay un PROCESS_EXIT?
	RESP: Lo va a haber siempre.

	17: Cuando en memoria tengo las 2 particiones fija y dinamica. Tengo que tener dos listas por separado? o puedo usar la misma. Si lo hago por separado no se como hacer el array de tamaño fijo
	RESP: usar una sola y la trato de forma distinta.