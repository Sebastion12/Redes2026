import socket

HOST = "127.0.0.1"
PUERTO = 8080


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


registrar_usuario()