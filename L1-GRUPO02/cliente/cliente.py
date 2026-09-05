import socket
import threading
import time

HOST = "127.0.0.1"
PUERTO = 8080
token = None
conexion_tcp = None
heartbeat_activo = False

def registrar_usuario():
    username = input("Ingrese username: ")
    password = input("Ingrese password: ")
    cuerpo = f"username={username}&password={password}"
    
    #Petición HTTP
    solicitud = (
        "POST /register HTTP/1.1\r\n"
        f"Host: {HOST}:{PUERTO}\r\n"
        "Content-Type: application/x-www-form-urlencoded\r\n"
        f"Content-Length: {len(cuerpo.encode('utf-8'))}\r\n"
        "Connection: close\r\n"
        "\r\n"
        f"{cuerpo}"
    )
    
    #Crear socket tcp
    cliente = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    cliente.connect((HOST, PUERTO))
    cliente.sendall(solicitud.encode("utf-8"))

    respuesta = b""

    while True:
        datos = cliente.recv(1024)

        if not datos:
            break

        respuesta += datos

    cliente.close()

    print("\nRespuesta del servidor:")
    print(respuesta.decode("utf-8"))



def iniciar_sesion():
    username = input("Ingrese username: ")
    password = input("Ingrese password: ")

    global conexion_tcp
    conexion_tcp = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    conexion_tcp.connect((HOST, 9000)) #conecta al puerto 9000 para login
    mensaje = f"LOGIN {username} {password}\n"
    conexion_tcp.sendall(mensaje.encode("utf-8"))
    respuesta = conexion_tcp.recv(1024).decode("utf-8").strip()

    print("\nRespuesta del servidor:")
    print(respuesta)

    if respuesta.startswith("OK"):
        partes = respuesta.split(" ")
        token = partes[1]
        puerto_udp = int(partes[2])
        print("Token guardado:", token)
        print("Puerto UDP:", puerto_udp)        
        #Heartbeat en un hilo
        hilo = threading.Thread(
            target=heartbeat_automatico,
            args=(token, puerto_udp),
            daemon=True
        )

        hilo.start()
        #Receptor en un hilo
        hilo_receptor = threading.Thread(
            target=escuchar_servidor,
            daemon=True
        )

        hilo_receptor.start()
        return token

    return None

def enviar_heartbeat(token, puerto_udp):
    cliente_udp = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    mensaje = f"HEARTBEAT {token}"

    cliente_udp.sendto(
        mensaje.encode("utf-8"),
        (HOST, puerto_udp)
    )

    respuesta, _ = cliente_udp.recvfrom(1024)
    cliente_udp.close()
#Heartbeat cada 3s
def heartbeat_automatico(token, puerto_udp):
    global heartbeat_activo
    heartbeat_activo = True
     
    while heartbeat_activo:
        enviar_heartbeat(token, puerto_udp)
        time.sleep(3)

    print("Heartbeat detenido")

def enviar_mensaje(token):
    global conexion_tcp

    if token is None:
        print("Primero debe iniciar sesion")
        return

    texto = input("Escribir el mensaje: ")
    mensaje = f"MSG {token} {texto}\n"
    conexion_tcp.sendall(mensaje.encode("utf-8"))

def escuchar_servidor():
    global conexion_tcp

    while True:
        try:
            respuesta = conexion_tcp.recv(1024)

            if not respuesta:
                break

            mensaje = respuesta.decode("utf-8").strip()
            print("\n" + mensaje)

        except:
            break

def ver_historial():
    cliente = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    cliente.connect((HOST, 8080))

    solicitud = (
        "GET /history HTTP/1.1\r\n"
        f"Host: {HOST}:8080\r\n"
        "Connection: close\r\n"
        "\r\n"
    )

    cliente.sendall(solicitud.encode("utf-8"))
    respuesta = b""

    while True:
        datos = cliente.recv(1024)

        if not datos:
            break

        respuesta += datos

    cliente.close()
    texto = respuesta.decode("utf-8")
    #Separar encabezados HTTP del contenido
    partes = texto.split("\r\n\r\n", 1)

    if len(partes) == 2:
        print("\n--- HISTORIAL ---")
        print(partes[1])

    else:
        print("Error leyendo historial")

def detener_heartbeat():
    global heartbeat_activo

    if heartbeat_activo:
        heartbeat_activo = False
        print("Heartbeat detenido")
    else:
        print("El heartbeat no está activo")

while True:
    print("\n--- MENU ---")
    print("1. Registrar usuario")
    print("2. Iniciar sesión")
    print("3. Enviar mensaje")
    print("4. Ver Historial")
    print("5. Detener heartbeat")
    print("6. Salir")

    opcion = input("Seleccione una opción: ")

    if opcion == "1":
        registrar_usuario()

    elif opcion == "2":
        token = iniciar_sesion()

    elif opcion == "3":
        enviar_mensaje(token)
    
    elif opcion == "4":
        ver_historial()

    elif opcion == "5":
        detener_heartbeat()

    elif opcion == "6":
        print("Saliendo...")
        break

    else:
        print("Opción incorrecta")