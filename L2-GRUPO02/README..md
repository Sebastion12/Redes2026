A topología.pkt sólo le añadí cuadrados a los distintos AS para lectura más fácil



Añadido un docx y su PDF del cálculo de todas las LANs, el pdf para poder pasarlo cómo contexto más fácil. al final del documento está un resumen de todas las direcciones que hay que configurar

en topología, realizadas conexiones iniciales de 2 routers, les falta enrutamiento entre ellos



Conexiones router-hosts de todas menos AS30 que pide DHCP listas, con direcciones acordes al pdf de LANs



### Completado OSPF de AS10, revisar que funcionen, fue de la forma:

```
enable
configure terminal
router ospf 1
### (escribir un network para la red actual, y una red por cada conexión a otro router con la dirección de los cables,
     con la submask que es el inverso de las masks, tomando de ejemplo a Deadlock:)
network 172.16.24.0 0.0.7.255 area 0
network 1.0.0.0 0.255.255.255 area 0
network 2.0.0.0 0.255.255.255 area 0
end
```

### DHCP de Gekko y Viper listos, se llaman GEKKO_LAN y VIPER_LAN


```
# (en exclude-adders, empieza desde la ip de la interfaz, y se excluyen ese y las 9 siguientes direcciones)
Router(config)# ip dhcp excluded-address <ip_inicio> <ip_final>         #ip inicio +9
Router(config)# ip dhcp pool POOL_LAN                                   #nombre de pool
Router(dhcp-config)# network <network_ID> <netmask>                     #direccion de red (la base) y su red
Router(dhcp-config)# default-router <ip_router>                         #direccion de la interfaz

```
Revisar conexión de Base server con Cypher, porque la linea apunta al router en vez de la interfaz

## SI VAN A HACER CAMBIOS A LOS ROUTER:
ejecutar ```Router(config)# do write memory``` o la opcón de save NVRAM en opciones de los routers
se guarden en VRAM y se guarden de forma consistente a futuro


--30/10/26---
Rutas básicas creadas para todo router y sus hosts
falta revisar cómo conectar el server base al router, porque parece que no puede compartir subred con interfaz
conexiones de AS 10 a AS 40 funcionando, de AS50 sólo conecta con AS40, hay que revisar porque no puede salir más allá de ahí
en AS40, cómo estaba Sage, hice un OSPF a la rápida y parece que funciona bien, hay que revisar que se tenga que hacer así y no de otra forma

Ir probando enviar mensajes entre todos los routers primero para ir asegurando las conexiones, y después entre pares de hosts de distintos routers