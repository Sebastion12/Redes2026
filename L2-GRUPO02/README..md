A topología.pkt sólo le añadí cuadrados a los distintos AS para lectura más fácil



Añadido un docx y su PDF del cálculo de todas las LANs, el pdf para poder pasarlo cómo contexto más fácil. al final del documento está un resumen de todas las direcciones que hay que configurar

en topología, realizadas conexiones iniciales de 2 routers, les falta enrutamiento entre ellos



Conexiones router-hosts de todas menos AS30 que pide DHCP listas, con direcciones acordes al pdf de LANs



Completado OSPF de AS10, revisar que funcionen, fue de la forma:

enable

configure terminal

router ospf 1
### (escribir un network para la red actual, y una red por cada conexión a otro router con la dirección de los cables,, con la submask que es el inverso de las masks, tomando de ejemplo a Deadlock:)

network 172.16.24.0 0.0.7.255 area 0

network 1.0.0.0 0.255.255.255 area 0

network 2.0.0.0 0.255.255.255 area 0

end



Revisar conexión de Base server con Cypher, porque la linea apunta al router en vez de la interfaz

