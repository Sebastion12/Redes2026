package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

func revisarUnaVez() {
	mutexSesiones.Lock()
	defer mutexSesiones.Unlock()
	archivo, err := os.Open("datos/sesiones.csv")

	if err != nil {
		return
	}

	lector := csv.NewReader(archivo)
	lector.FieldsPerRecord = -1
	filas, err := lector.ReadAll()
	archivo.Close()

	if err != nil {
		return
	}

	nuevasFilas := [][]string{}

	for i, fila := range filas {
		if i == 0 {
			nuevasFilas = append(nuevasFilas, fila)
			continue
		}

		if len(fila) < 5 {
			continue
		}

		username := fila[1]
		timestampCreacion := fila[2]
		timestampHeartbeat := fila[3]
		creacion, err := time.Parse(
			time.RFC3339,
			timestampCreacion,
		)

		if err != nil {
			continue
		}

		heartbeat, err := time.Parse(
			time.RFC3339,
			timestampHeartbeat,
		)

		if err != nil {
			continue
		}

		tiempoSesion := time.Since(creacion)
		tiempoSinHeartbeat := time.Since(heartbeat)

		if tiempoSesion > 10*time.Minute {
			fmt.Println("Sesion expirada por TTL:", username)
			continue
		}

		if tiempoSinHeartbeat > 60*time.Second {
			fmt.Println("Sesion eliminada:", username)
			continue
		}

		if tiempoSinHeartbeat > 30*time.Second {
			fila[4] = "INACTIVO"
			fmt.Println("Sesion inactiva:", username)
		}

		nuevasFilas = append(nuevasFilas, fila)
	}

	archivo, err = os.Create("datos/sesiones.csv")

	if err != nil {
		return
	}

	escritor := csv.NewWriter(archivo)
	escritor.WriteAll(nuevasFilas)
	escritor.Flush()
	archivo.Close()
}

func revisarSesiones() {
	for {
		time.Sleep(5 * time.Second)
		revisarUnaVez()
	}
}
