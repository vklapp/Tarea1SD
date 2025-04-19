package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// Cambia la IP y puerto por donde esté tu servidor
	server := "10.10.28.60:8080"

	conn, err := net.Dial("tcp", server)
	if err != nil {
		fmt.Println("Error conectando al servidor:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Conectado al servidor:", server)

	// Ejemplo de envío de mensaje
	message := "Hola desde cliente\n"
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error enviando mensaje:", err)
		return
	}

	// Ejemplo de recepción de respuesta
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error recibiendo datos:", err)
		return
	}

	fmt.Println("Respuesta del servidor:", string(buffer[:n]))
}

