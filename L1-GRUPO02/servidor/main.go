package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Ruta
const archivoUsuarios = "datos/usuarios.csv"

var mutexSesiones sync.Mutex

func crearArchivoUsuarios() {
	os.MkdirAll("datos", 0755)

	if _, err := os.Stat(archivoUsuarios); os.IsNotExist(err) {
		archivo, err := os.Create(archivoUsuarios)

		if err != nil {
			fmt.Println("Error creando usuarios.csv")
			return
		}

		defer archivo.Close()
		escritor := csv.NewWriter(archivo)
		defer escritor.Flush()

		escritor.Write([]string{
			"username",
			"password",
			"fecha_registro",
		})
	}
}

// Busca si el usuario esta en el csv
func usuarioExiste(username string) bool {
	archivo, err := os.Open(archivoUsuarios)

	if err != nil {
		return false
	}

	defer archivo.Close()
	lector := csv.NewReader(archivo)
	filas, err := lector.ReadAll()

	if err != nil {
		return false
	}

	for i, fila := range filas {

		if i == 0 {
			continue
		}

		if fila[0] == username {
			return true
		}
	}

	return false
}

// Guarda el usuario en el csv
func guardarUsuario(username string, password string) {
	archivo, err := os.OpenFile(
		archivoUsuarios,
		os.O_APPEND|os.O_WRONLY,
		0644,
	)

	if err != nil {
		fmt.Println("Error abriendo usuarios.csv")
		return
	}

	defer archivo.Close()
	escritor := csv.NewWriter(archivo)
	defer escritor.Flush()
	fecha := time.Now().Format(time.RFC3339)

	escritor.Write([]string{
		username,
		password,
		fecha,
	})
}

// Recibe el POST para registrar el usuario
func registrarUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	//Lee los datos enviados por el cliente
	err := r.ParseForm()

	if err != nil {
		http.Error(w, "Solicitud incorrecta", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := strings.TrimSpace(r.FormValue("password"))

	if username == "" || password == "" {
		http.Error(w, "Faltan username o password", http.StatusBadRequest)
		return
	}

	if usuarioExiste(username) {
		http.Error(w, "Usuario ya existe", http.StatusConflict)
		return
	}

	guardarUsuario(username, password)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Usuario registrado correctamente")
}

func mostrarHistorial(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	archivo, err := os.Open("datos/historial.csv")

	if err != nil {
		http.Error(w, "No se pudo abrir historial.csv", http.StatusInternalServerError)
		return
	}

	defer archivo.Close()
	lector := csv.NewReader(archivo)
	filas, err := lector.ReadAll()

	if err != nil {
		http.Error(w, "Error leyendo historial.csv", http.StatusInternalServerError)
		return
	}

	for _, fila := range filas {
		fmt.Fprintln(w, strings.Join(fila, ","))
	}
}

func validarUsuario(username string, password string) bool {
	archivo, err := os.Open(archivoUsuarios)

	if err != nil {
		return false
	}

	defer archivo.Close()
	lector := csv.NewReader(archivo)
	filas, err := lector.ReadAll()

	if err != nil {
		return false
	}

	for i, fila := range filas {
		if i == 0 {
			continue
		}

		if fila[0] == username && fila[1] == password {
			return true
		}
	}

	return false
}

func main() {
	crearArchivoUsuarios()
	http.HandleFunc("/register", registrarUsuario)
	http.HandleFunc("/history", mostrarHistorial)
	//Inicia el servidor TCP y UDP
	go iniciarServidorTCP()
	go iniciarServidorUDP()
	go revisarSesiones()
	fmt.Println("Servidor HTTP iniciado en puerto 8080")
	//Inicia el servidor HTTP
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Error iniciando servidor:", err)
	}
}
