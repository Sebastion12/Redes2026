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

		if len(fila) < 3 {
			continue
		}

		username := fila[0]
		ultimoHeartbeat := fila[2]
		fecha, err := time.Parse(time.RFC3339, ultimoHeartbeat)

		if err != nil {
			continue
		}

		tiempoSinHeartbeat := time.Since(fecha)

		if tiempoSinHeartbeat > 60*time.Second {
			fmt.Println("Sesion eliminada:", username)
			continue
		}

		if tiempoSinHeartbeat > 30*time.Second {
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
