#!/bin/bash

# Validar que se pasen exactamente 4 parámetros
if [ "$#" -ne 4 ]; then
    echo "Uso: $0 <ip_kernel> <ip_cpu> <ip_memory> <ip_filesystem>"
    exit 1
fi

# Asignar parámetros a variables
ip_kernel=$1
ip_cpu=$2
ip_memory=$3
ip_filesystem=$4

# Definir la ruta base (un nivel arriba de la carpeta scripts)
ruta_base=$(dirname "$0")/..

# Definir las carpetas principales relativas a la ruta base
carpetas=("kernel" "cpu" "memoria" "filesystem")

# Recorremos cada carpeta
for carpeta in "${carpetas[@]}"; do
    config_dir="$ruta_base/$carpeta/configs"
    
    # Verificar si la subcarpeta 'configs' existe
    if [ -d "$config_dir" ]; then
        # Buscar todos los archivos .json en la carpeta
        for archivo in "$config_dir"/*.json; do
            if [ -f "$archivo" ]; then
                # Reemplazar las IPs en el archivo .json
                sed -i \
                    -e "s/\"ip_kernel\": \".*\"/\"ip_kernel\": \"$ip_kernel\"/" \
                    -e "s/\"ip_cpu\": \".*\"/\"ip_cpu\": \"$ip_cpu\"/" \
                    -e "s/\"ip_memory\": \".*\"/\"ip_memory\": \"$ip_memory\"/" \
                    -e "s/\"ip_filesystem\": \".*\"/\"ip_filesystem\": \"$ip_filesystem\"/" \
                    "$archivo"
                
                echo "Actualizado: $archivo"
            fi
        done
    else
        echo "Advertencia: La carpeta $config_dir no existe."
    fi
done

echo "Actualización completa."
