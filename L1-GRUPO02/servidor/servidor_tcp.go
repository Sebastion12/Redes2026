package main

import (
	"bufio"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const PUERTO_TCP = "9000"

var clientesConectados = make(map[string]net.Conn)

func generarToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func manejarCliente(conn net.Conn) {
	lector := bufio.NewReader(conn)

	for {
		mensaje, err := lector.ReadString('\n')

		if err != nil {
			fmt.Println("Cliente desconectado")
			conn.Close()
			return
		}

		mensaje = strings.TrimSpace(mensaje)
		fmt.Println("Recibido:", mensaje)
		partes := strings.SplitN(mensaje, " ", 3)

		if len(partes) < 2 {
			fmt.Fprintln(conn, "ERROR formato incorrecto")
			continue
		}

		comando := partes[0]

		// LOGIN
		if comando == "LOGIN" {

			if len(partes) != 3 {
				fmt.Fprintln(conn, "ERROR formato LOGIN incorrecto")
				return
			}

			username := partes[1]
			password := partes[2]

			if !validarUsuario(username, password) {
				fmt.Fprintln(conn, "ERROR usuario o password incorrectos")
				return
			}

			token := generarToken()
			guardarSesion(username, token)
			clientesConectados[token] = conn
			fmt.Fprintln(conn, "OK", token)
			continue
		}

		// MSG
		if comando == "MSG" {

			if len(partes) != 3 {
				fmt.Fprintln(conn, "ERROR formato MSG incorrecto")
				return
			}

			token := partes[1]
			texto := partes[2]
			username := obtenerUsuarioPorToken(token)

			if username == "" {
				fmt.Fprintln(conn, "ERROR token invalido")
				return
			}

			fmt.Println("Mensaje de", username+":", texto)
			guardarMensaje(username, texto)
			enviarBroadcast(token, username, texto)
			fmt.Fprintln(conn, "OK mensaje recibido")
			continue
		}

		fmt.Fprintln(conn, "ERROR comando desconocido")
	}
}

func guardarSesion(username string, token string) {
	//Proteger sesiones.csv
	mutexSesiones.Lock()
	defer mutexSesiones.Unlock()

	archivo, err := os.OpenFile(
		"datos/sesiones.csv",
		os.O_APPEND|os.O_WRONLY,
		0644,
	)

	if err != nil {
		fmt.Println("Error abriendo sesiones.csv")
		return
	}

	defer archivo.Close()
	escritor := csv.NewWriter(archivo)
	defer escritor.Flush()
	fecha := time.Now().Format(time.RFC3339)

	escritor.Write([]string{
		username,
		token,
		fecha,
	})
}

func obtenerUsuarioPorToken(token string) string {
	//Proteger sesiones.csv
	mutexSesiones.Lock()
	defer mutexSesiones.Unlock()
	archivo, err := os.Open("datos/sesiones.csv")

	if err != nil {
		return ""
	}

	defer archivo.Close()
	lector := csv.NewReader(archivo)
	filas, err := lector.ReadAll()

	if err != nil {
		return ""
	}

	for i, fila := range filas {
		if i == 0 {
			continue
		}

		if len(fila) < 3 {
			continue
		}

		if fila[1] == token {
			return fila[0]
		}
	}

	return ""
}

func guardarMensaje(username string, texto string) {
	archivo, err := os.OpenFile(
		"datos/historial.csv",
		os.O_APPEND|os.O_WRONLY,
		0644,
	)

	if err != nil {
		fmt.Println("Error abriendo historial.csv")
		return
	}

	defer archivo.Close()
	escritor := csv.NewWriter(archivo)
	defer escritor.Flush()
	fecha := time.Now().Format(time.RFC3339)

	escritor.Write([]string{
		fecha,
		username,
		texto,
	})
}

func enviarBroadcast(tokenEmisor string, username string, texto string) {
	mensaje := fmt.Sprintf("MSG %s: %s\n", username, texto)

	for token, conexion := range clientesConectados {
		if token != tokenEmisor {
			fmt.Fprint(conexion, mensaje)
		}
	}
}

func iniciarServidorTCP() {
	listener, err := net.Listen("tcp", ":"+PUERTO_TCP)

	if err != nil {
		fmt.Println("Error iniciando servidor TCP:", err)
		return
	}

	defer listener.Close()
	fmt.Println("Servidor TCP iniciado en puerto", PUERTO_TCP)

	for {

		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error aceptando cliente:", err)
			continue
		}

		go manejarCliente(conn)
	}
}
