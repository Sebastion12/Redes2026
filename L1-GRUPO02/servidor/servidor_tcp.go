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
	tokenCliente := ""

	for {
		mensaje, err := lector.ReadString('\n')

		if err != nil {
			fmt.Println("Cliente desconectado")

			if tokenCliente != "" {
				mutexClientes.Lock()
				delete(clientesConectados, tokenCliente)
				mutexClientes.Unlock()
			}

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
				fmt.Fprintln(conn, "ERROR INVALID_CREDENTIALS")
				return
			}

			token := generarToken()
			guardarSesion(username, token)
			mutexClientes.Lock()
			clientesConectados[token] = conn
			mutexClientes.Unlock()
			tokenCliente = token
			fmt.Fprintln(conn, "OK", token, PUERTO_UDP)
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
			username, errorSesion := validarSesion(token)

			if errorSesion == "INVALID_TOKEN" {
				fmt.Fprintln(conn, "ERROR INVALID_TOKEN")
				continue
			}

			if errorSesion == "SESSION_EXPIRED" {
				fmt.Fprintln(conn, "ERROR SESSION_EXPIRED")
				continue
			}

			fmt.Println("Mensaje de", username+":", texto)
			guardarMensaje(username, texto)
			enviarBroadcast(token, username, texto)
			fmt.Fprintln(conn, "ACK")
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
		token,
		username,
		fecha,
		"",
		"PENDIENTE",
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
	lector.FieldsPerRecord = -1
	filas, err := lector.ReadAll()

	if err != nil {
		return ""
	}

	for i, fila := range filas {
		if i == 0 {
			continue
		}

		if len(fila) < 5 {
			continue
		}

		if fila[0] == token {
			return fila[1]
		}
	}

	return ""
}

func guardarMensaje(username string, texto string) {
	//Proteger historial.csv
	mutexHistorial.Lock()
	defer mutexHistorial.Unlock()
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
	mutexClientes.Lock()
	defer mutexClientes.Unlock()
	mensaje := fmt.Sprintf("INCOMIG %s %s\n", username, texto)

	for token, conexion := range clientesConectados {
		if token != tokenEmisor {
			fmt.Fprint(conexion, mensaje)
		}
	}
}

func validarSesion(token string) (string, string) {
	mutexSesiones.Lock()
	defer mutexSesiones.Unlock()
	archivo, err := os.Open("datos/sesiones.csv")

	if err != nil {
		return "", "INVALID_TOKEN"
	}

	defer archivo.Close()
	lector := csv.NewReader(archivo)
	lector.FieldsPerRecord = -1
	filas, err := lector.ReadAll()

	if err != nil {
		return "", "INVALID_TOKEN"
	}

	for i, fila := range filas {
		if i == 0 {
			continue
		}

		if len(fila) < 5 {
			continue
		}

		if fila[0] != token {
			continue
		}

		username := fila[1]
		timestampCreacion := fila[2]
		timestampHeartbeat := fila[3]
		estado := fila[4]

		if estado != "ACTIVO" {
			return "", "SESSION_EXPIRED"
		}

		creacion, err := time.Parse(
			time.RFC3339,
			timestampCreacion,
		)

		if err != nil {
			return "", "SESSION_EXPIRED"
		}

		heartbeat, err := time.Parse(
			time.RFC3339,
			timestampHeartbeat,
		)

		if err != nil {
			return "", "SESSION_EXPIRED"
		}

		if time.Since(creacion) > 10*time.Minute {
			return "", "SESSION_EXPIRED"
		}

		if time.Since(heartbeat) > 60*time.Second {
			return "", "SESSION_EXPIRED"
		}

		return username, ""
	}

	return "", "INVALID_TOKEN"
}

func desconectarCliente(token string) {
	mutexClientes.Lock()
	defer mutexClientes.Unlock()
	conexion, existe := clientesConectados[token]

	if existe {
		conexion.Close()
		delete(clientesConectados, token)
		fmt.Println("Conexión TCP cerrada para token:", token)
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
