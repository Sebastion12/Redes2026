package main

import (
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const PUERTO_UDP = "9001"

func actualizarHeartbeat(token string) bool {
	//Proteger sesiones.csv
	mutexSesiones.Lock()
	defer mutexSesiones.Unlock()
	archivo, err := os.Open("datos/sesiones.csv")

	if err != nil {
		return false
	}

	lector := csv.NewReader(archivo)
	lector.FieldsPerRecord = -1
	filas, err := lector.ReadAll()
	archivo.Close()

	if err != nil {
		return false
	}

	encontrado := false

	for i, fila := range filas {
		if i == 0 {
			continue
		}

		// Revisamos que la fila tenga las 3 columnas
		if len(fila) < 5 {
			continue
		}

		//Si el token coincide y el etado es pendiente o activo, se actualiza el heartbeat y cambia estaodo a activo
		if fila[0] == token && (fila[4] == "PENDIENTE" || fila[4] == "ACTIVO") {
			fila[3] = time.Now().Format(time.RFC3339)
			fila[4] = "ACTIVO"
			encontrado = true
		}
	}

	if !encontrado { //si no encuentra token, retorna False
		return false
	}

	archivo, err = os.Create("datos/sesiones.csv")

	if err != nil {
		return false
	}

	defer archivo.Close()
	escritor := csv.NewWriter(archivo)
	escritor.WriteAll(filas)
	escritor.Flush()
	return true
}

func iniciarServidorUDP() {
	direccion, err := net.ResolveUDPAddr("udp", ":"+PUERTO_UDP)

	if err != nil {
		fmt.Println("Error resolviendo dirección UDP:", err)
		return
	}

	// Inicia server UDP
	conexion, err := net.ListenUDP("udp", direccion)

	if err != nil {
		fmt.Println("Error iniciando servidor UDP:", err)
		return
	}

	defer conexion.Close()
	fmt.Println("Servidor UDP iniciado en puerto", PUERTO_UDP)
	buffer := make([]byte, 1024)

	for {
		n, cliente, err := conexion.ReadFromUDP(buffer)

		if err != nil {
			fmt.Println("Error leyendo UDP:", err)
			continue
		}

		mensaje := strings.TrimSpace(string(buffer[:n]))
		fmt.Println("UDP recibido desde", cliente, ":", mensaje)

		if strings.HasPrefix(mensaje, "HEARTBEAT ") {
			partes := strings.Split(mensaje, " ")

			if len(partes) != 2 {
				conexion.WriteToUDP([]byte("ERROR formato incorrecto\n"), cliente)
				continue
			}

			token := partes[1]

			if actualizarHeartbeat(token) {
				conexion.WriteToUDP([]byte("OK\n"), cliente)
			} else {
				conexion.WriteToUDP([]byte("ERROR token invalido\n"), cliente)
			}
		}
	}
}
