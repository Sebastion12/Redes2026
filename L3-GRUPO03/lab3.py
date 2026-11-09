#!/usr/bin/env python3
"""Cliente base — Laboratorio 2, Redes de Computadores.

Cliente del sistema de registro (HTTP), autenticación y mensajería (TCP) y
presencia (UDP) del Laboratorio 1, usado en la Etapa 1 del Laboratorio 2.

Solo usa la biblioteca estándar (Python 3.10+). Toda la comunicación HTTP se
hace con sockets TCP, igual que en el Laboratorio 1.

Ejecución:  python3 cliente.py

Comandos dentro del chat:
  <texto>     envía un mensaje (MSG) a todos los conectados
  /acceso     envía SOLICITUD_ACCESO_WEB para obtener la clave del portal
  /historial  muestra el historial (GET /history)
  /salir      cierra la sesión (LOGOUT)
"""
# contraseña usuario 2:  aasdas

import os
import socket
import sys
import threading
import time

# ===================================================================
# Configuración de red
# ===================================================================
# >>> TODO: Complete la dirección IP del Servidor Central (la entrega el cuerpo docente). || SERVER_IP = "COMPLETAR"
SERVER_IP = os.environ.get("LAB2_SERVER", "192.168.1.157")
# <<< TODO
# >>> TODO: Complete el puerto del servicio HTTP de registro. || HTTP_PORT = 0
HTTP_PORT = 8080
# <<< TODO
# >>> TODO: Complete el puerto del servidor TCP de chat. || TCP_PORT = 0
TCP_PORT = 9000
# <<< TODO
# El puerto UDP del heartbeat NO se configura aquí: lo informa el servidor
# en la respuesta OK <token> <puerto_udp>.

HEARTBEAT_INTERVALO = 3  # segundos


# ===================================================================
# Construcción de mensajes del protocolo
# ===================================================================

def construir_registro_http(usuario: str, clave: str) -> bytes:
    """Solicitud HTTP completa (línea de petición + encabezados + cuerpo) para
    registrar un usuario en POST /register usando un formulario
    (application/x-www-form-urlencoded). Recuerde: cada línea del encabezado
    termina en \\r\\n, una línea vacía separa encabezados y cuerpo, y
    Content-Length es el largo del cuerpo en BYTES."""
    # >>> TODO: Construya la solicitud HTTP POST /register (Host, Content-Type, Content-Length, Connection: close y el cuerpo username=...&password=...).
    cuerpo = f"username={usuario}&password={clave}".encode("utf-8")
    encabezados = (
        "POST /register HTTP/1.1\r\n"
        f"Host: {SERVER_IP}:{HTTP_PORT}\r\n"
        "User-Agent: cliente-lab2-python\r\n"
        "Content-Type: application/x-www-form-urlencoded\r\n"
        f"Content-Length: {len(cuerpo)}\r\n"
        "Connection: close\r\n"
        "\r\n"
    )
    return encabezados.encode("utf-8") + cuerpo
    # <<< TODO


def construir_login(usuario: str, clave: str) -> str:
    """Comando de inicio de sesión del protocolo TCP (Lab 1)."""
    # >>> TODO: Construya el comando LOGIN (no olvide el delimitador de fin de mensaje).
    return f"LOGIN {usuario} {clave}\n"
    # <<< TODO


def construir_heartbeat(token: str) -> bytes:
    """Contenido del datagrama UDP de presencia (Lab 1)."""
    # >>> TODO: Construya el datagrama HEARTBEAT con el token de sesión.
    return f"HEARTBEAT {token}".encode("utf-8")
    # <<< TODO


def construir_mensaje(token: str, texto: str) -> str:
    """Comando para publicar un mensaje en el chat (Lab 1)."""
    # >>> TODO: Construya el comando MSG con el token y el texto.
    return f"MSG {token} {texto}\n"
    # <<< TODO


def construir_solicitud_acceso(token: str) -> str:
    """Mensaje de la Etapa 1 del Laboratorio 2 para obtener la clave web."""
    # >>> TODO: Construya el mensaje que solicita la clave de acceso web (ver enunciado, Etapa 1).
    return construir_mensaje(token, "SOLICITUD_ACCESO_WEB")
    # <<< TODO


def construir_logout(token: str) -> str:
    return f"LOGOUT {token}\n"


# ===================================================================
# HTTP sobre sockets TCP
# ===================================================================

def _verificar_config():
    if SERVER_IP in ("", "COMPLETAR") or not HTTP_PORT or not TCP_PORT:
        sys.exit("Configuración incompleta: complete SERVER_IP, HTTP_PORT y TCP_PORT.")


def http_crudo(solicitud: bytes) -> tuple[str, str]:
    """Envía una solicitud HTTP ya construida y devuelve (línea de estado, cuerpo)."""
    with socket.create_connection((SERVER_IP, HTTP_PORT), timeout=10) as s:
        s.sendall(solicitud)
        partes = []
        while True:
            d = s.recv(4096)
            if not d:
                break
            partes.append(d)
    resp = b"".join(partes).decode("utf-8", errors="replace")
    cabecera, _, cuerpo = resp.partition("\r\n\r\n")
    return cabecera.split("\r\n", 1)[0], cuerpo


def registrar(usuario, clave):
    estado, cuerpo = http_crudo(construir_registro_http(usuario, clave))
    print(f"[HTTP] {estado}")
    if cuerpo.strip():
        print(f"[HTTP] {cuerpo.strip()}")


def historial():
    req = (f"GET /history HTTP/1.1\r\nHost: {SERVER_IP}:{HTTP_PORT}\r\n"
           "User-Agent: cliente-lab2-python\r\nConnection: close\r\n\r\n").encode()
    estado, cuerpo = http_crudo(req)
    print(f"[HTTP] {estado}")
    print(cuerpo.rstrip() or "(historial vacío)")


# ===================================================================
# Sesión TCP + heartbeat UDP
# ===================================================================

def hilo_heartbeat(token, puerto_udp, detener):
    """Envía HEARTBEAT <token> cada 3 segundos hasta que se pida detener."""
    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as udp:
        while not detener.is_set():
            try:
                udp.sendto(construir_heartbeat(token), (SERVER_IP, puerto_udp))
            except OSError as e:
                print(f"\n[UDP] error enviando heartbeat: {e}")
            detener.wait(HEARTBEAT_INTERVALO)


def hilo_receptor(lector, detener):
    """Imprime todo lo que llega por TCP (respuestas y broadcasts)."""
    for crudo in lector:
        linea = crudo.decode("utf-8", errors="replace").rstrip("\r\n")
        if linea.startswith("ACCESS_KEY "):
            clave = linea.split(" ", 1)[1]
            print(f"\n[SERVIDOR] Clave de acceso web: {clave}")
            print(f"[SERVIDOR] Ingrésela en http://{SERVER_IP}:{HTTP_PORT}/mision")
        elif linea.startswith("INCOMING "):
            _, emisor, texto = (linea.split(" ", 2) + ["", ""])[:3]
            print(f"\n<{emisor}> {texto}")
        else:
            print(f"\n[SERVIDOR] {linea}")
        print("> ", end="", flush=True)
    print("\n[TCP] El servidor cerró la conexión.")
    detener.set()


def sesion_chat(usuario, clave):
    sock = socket.create_connection((SERVER_IP, TCP_PORT), timeout=10)
    sock.settimeout(None)
    lector = sock.makefile("rb")
    sock.sendall(construir_login(usuario, clave).encode("utf-8"))
    respuesta = lector.readline().decode("utf-8", errors="replace").strip()
    partes = respuesta.split()
    if len(partes) != 3 or partes[0] != "OK":
        print(f"[TCP] Inicio de sesión rechazado: {respuesta or '(sin respuesta)'}")
        sock.close()
        return
    token, puerto_udp = partes[1], int(partes[2])
    print(f"[TCP] Sesión iniciada. Token: {token} — heartbeat UDP al puerto {puerto_udp}")

    detener = threading.Event()
    threading.Thread(target=hilo_heartbeat, args=(token, puerto_udp, detener), daemon=True).start()
    threading.Thread(target=hilo_receptor, args=(lector, detener), daemon=True).start()

    print("Escriba mensajes. Comandos: /acceso  /historial  /salir")
    try:
        while not detener.is_set():
            texto = input("> ").strip()
            if detener.is_set():
                break
            if not texto:
                continue
            if texto == "/salir":
                sock.sendall(construir_logout(token).encode("utf-8"))
                time.sleep(0.5)
                break
            if texto == "/historial":
                historial()
            elif texto == "/acceso":
                sock.sendall(construir_solicitud_acceso(token).encode("utf-8"))
            else:
                sock.sendall(construir_mensaje(token, texto).encode("utf-8"))
    except (EOFError, KeyboardInterrupt, OSError):
        pass
    finally:
        detener.set()
        try:
            sock.close()
        except OSError:
            pass


def main():
    _verificar_config()
    print(f"Servidor Central: {SERVER_IP} (HTTP {HTTP_PORT}, TCP {TCP_PORT})")
    while True:
        print("\n[1] Registrar usuario   [2] Iniciar sesión   [3] Ver historial   [0] Salir")
        try:
            op = input("Opción: ").strip()
        except (EOFError, KeyboardInterrupt):
            return
        if op == "0":
            return
        if op == "3":
            historial()
            continue
        if op not in ("1", "2"):
            continue
        usuario = input("Usuario (grupoXX): ").strip()
        clave = input("Contraseña: ").strip()
        try:
            if op == "1":
                registrar(usuario, clave)
            else:
                sesion_chat(usuario, clave)
        except OSError as e:
            print(f"[RED] Error de conexión: {e}")


if __name__ == "__main__":
    main()