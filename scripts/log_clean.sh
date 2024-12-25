#!/bin/bash

# Lista de archivos de log a vaciar
logs=("../kernel/kernel.log" "../cpu/cpu.log" "../memoria/memoria.log" "../filesystem/filesystem.log")

for log in "${logs[@]}"; do
    # Vaciar el contenido del archivo sin eliminarlo
    > "$log"
    echo "Archivo $log ha sido vaciado."
done