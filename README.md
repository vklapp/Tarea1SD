Vicente Klapp 202173014-1

Instrucciones:

  Division de las MV:
    - 10.10.28.59 : proxy.db
    - 10.10.28.60 : APIRESTFULL (server)
    - 10.10.28.61 : Cliente

Para poder ejecutar el server necesita acceder maquina virtual con ese rol, acceder a la carpeta Tarea1SD y ejecutar el comando "go run server.go" en la maquina virtual ya esta instalado todo lo necesario
Para poder ejecutar el cliente necesita acceder a la maquina virtual determinada para ese rol, acceder a la acarpeta Tarea1SD y ejecutar el comando "go run cliente.go" la maquina virtual tiene instalado
el paquete go.

Consideraciones:
- Los mod.go son para poder ejecutar los comando go de instalacion. fijarse tambien que la base de datos se tuvo que montar en la maquina de servidor debido a que SQLite necesita trabajar de forma local
  por lo que se monto la Base de datos referenciando al proxy.db de la maquina virtual 10.10.28.59.
  
- Considerar que en las carpetas Tarea1SD estan todos los archivos debido a que cuando conecte github se sincronizo. Tambien considerar que el cliente manda el request a la maquina virtual 10.10.28.60
  y la base de datos tambien hace referencia a la maquina 10.10.28.59 por lo que no funcionarian si se ocupa el programa en otra maquina.
  
- Al momento de ejecutar la quinta opcion, la respuesta se demora aproximadamente 58sg.

