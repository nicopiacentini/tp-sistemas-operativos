#!/bin/bash

# Lista de archivos a eliminar
archivos=(
  "../filesystem/mount_dir/bitmap.dat"
  "../filesystem/mount_dir/bloques.dat"
)

# Eliminar los archivos de la lista
for archivo in "${archivos[@]}"; do
  if [ -f "$archivo" ]; then
    echo "Eliminando el archivo: $archivo"
    rm "$archivo"
    if [ $? -eq 0 ]; then
      echo "Archivo $archivo eliminado exitosamente."
    else
      echo "Error al eliminar el archivo $archivo."
    fi
  else
    echo "El archivo $archivo no existe o no es un archivo regular."
  fi
done
