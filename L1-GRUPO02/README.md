# Integrantes

- Sebastián Castillo — 202273583-K
- Sebastián Césped — 202273603-K
- Ricardo Ruz - 202373567-1


## Descripción

Este proyecto implementa un sistema cliente-servidor para registro, autenticación, presencia y mensajería en tiempo real, utilizando los protocolos HTTP, TCP y UDP.
La arquitectura está compuesta por Servidor (implementado en Go), Cliente (implementado en python), HTTP (utilizado para el registro de usuarios y consulta del historial), TCP (utilizado para la autenticación y el envío/retransmisión de mensajes), UDP (utilizado para el envío periódico de mensajes de heartbeat), Archivos CSV (utilizados para la persistencia de usuarios, sesiones e historial).
El servidor permite atender múltiples clientes simultáneamente mediante concurrencia y mantiene un registro de las sesiones activas.

- main.go: inicia el servidor HTTP y coordina los componentes del servidor.
- servidor_tcp.go: implementa autenticación, sesiones, recepción de mensajes y broadcast.
- servidor_udp.go: implementa la recepción de heartbeats.
- watchdog.go: supervisa la expiración e inactividad de las sesiones.
- cliente.py: implementa el cliente HTTP, TCP y UDP, además de la interfaz de consola.
- usuarios.csv: almacena los usuarios registrados.
- sesiones.csv: almacena las sesiones generadas.
- historial.csv: almacena los mensajes enviados.

## Requisitos

Para ejecutar el proyecto se requiere Go 1.20 o superior y Python 3.10 o superior. El laboratorio utiliza únicamente bibliotecas estándar de Go y Python.

## Puertos utilizados

- Servidor HTTP: Protocolo TCP, Puerto 8080.
- Servidor TCP: Protocolo TCP, Puerto 9000.
- Servidor UDP: Protocolo UDP, Puerto 9001.
- El cliente se conecta por defecto a Host: 127.0.0.1

## Ejecutar Servidores

Desde la carpeta `servidor` ejecutar: 
```bash 
go run *.go 
```

Al iniciar correctamente deberían aparecer mensajes similares a:
- Servidor HTTP iniciado en puerto 8080
- Servidor TCP iniciado en puerto 9000
- Servidor UDP iniciado en puerto 9001
El servidor ejecuta simultáneamente los servicios HTTP, TCP y UDP.

## Ejecutar Cliente

Desde la carpeta `cliente` ejecutar:
```bash
python3 cliente.py
```

El cliente permite:

1. Registrar usuario
2. Iniciar sesión
3. Enviar mensaje
4. Ver historial
5. Detener heartbeat
6. Salir

## Protocolos

1. HTTP Puerto 8080:

Registrar usuario: 
POST /register 
username=<usuario>&password=<contraseña>

Respuestas:
201 Created - Registro exitoso 
400 Bad Request - Datos faltantes 
409 Conflict - Usuario ya registrado

Consultar:
GET /history

2. TCP Puerto 9000:

Inicio de sesión:
LOGIN <username> <password>\n
Respuesta exitosa: OK <token> <puerto_udp>\n
Credenciales incorrectas: ERROR INVALID CREDENTIALS\n

Enviar mensaje:
MSG <token> <mensaje>\n
Respuesta exitosa: ACK\n
El servidor retransmite a los demás clientes: INCOMING <usuario> <mensaje>\n
Errores: 
ERROR INVALID TOKEN\n
ERROR SESSION EXPIRED\n

3. UDP Puerto 9001:

El cliente envía un heartbeat cada 3 segundos: HEARTBEAT <token>
El servidor actualiza el último heartbeat de la sesión. Si no se recibe el primer heartbeat durante 30 segundos, la sesión se invalida. Si pasan más de 60 segundos desde el último heartbeat, la sesión se revoca y se cierra la conexión TCP.

## Archivos CSV

Los datos se almacenan en:

1. usuarios.csv:
username
password
fecha_registro

2. sesiones.csv:
token
username
timestamp_creacion
timestamp_ultimo_heartbeat
estado

3. historial.csv:
timestamp
username
mensaje

## Prueba rápida

Para comprobar el funcionamiento:

1. Ejecutar el servidor.
2. Ejecutar dos clientes.
3. Registrar ambos usuarios.
4. Iniciar sesión en ambos clientes.
5. Verificar los tokens y heartbeats.
6. Enviar un mensaje desde un cliente.
7. Comprobar el ACK y el broadcast en el otro cliente.
8. Detener el heartbeat de un cliente.
9. Esperar más de 60 segundos y comprobar la revocación de la sesión.
