package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tarm/serial"
)

const (
	portName       = "COM7" // Reemplaza con el nombre del puerto adecuado, por ejemplo, "/dev/ttyUSB0" en Linux
	baudRate       = 9600   // Ajusta la velocidad de baudios según la configuración del HC-05
	readDelay      = 100 * time.Millisecond
	retryDelay     = 1 * time.Second
	inactivityTime = 3 * time.Second
)

var (
	upgrader      = websocket.Upgrader{} // Configuración del upgrader de websockets
	clients       = make(map[*websocket.Conn]bool)
	value         int
	lastMessage   time.Time
	serialTimeout = false
)

func main() {
	// Iniciar la función concurrente para la comunicación serial
	go serialCommunication()

	// Iniciar la función concurrente para el servidor WebSocket
	go startWebSocketServer()

	// Esperar por señal de finalización (por ejemplo, presionar Ctrl+C)
	<-make(chan struct{})
}

func serialCommunication() {
	for {
		// Intentar establecer la conexión serial
		ser, err := openSerialPort()
		if err != nil {
			log.Println("Error al abrir el puerto serie:", err)
			time.Sleep(retryDelay)
			continue
		}

		fmt.Println("Conexión serial establecida con éxito.")
		lastMessage = time.Now()
		serialTimeout = false

		for {
			buf := make([]byte, 1)
			n, err := ser.Read(buf)
			if err != nil {
				log.Println("Error al leer los datos del puerto serie:", err)
				ser.Close()
				serialTimeout = true
				break
			}
			if n > 0 {
				fmt.Println("Dato recibido (en bytes):", buf[0])
				value = int(buf[0])

				broadcastMessage(value)
				lastMessage = time.Now()
			}

			// Verificar inactividad
			if time.Since(lastMessage) > inactivityTime {
				log.Println("Inactividad detectada. Cerrando conexión serial.")
				ser.Close()
				serialTimeout = true
				break
			}
		}
	}
}

func openSerialPort() (*serial.Port, error) {
	// Configuración del puerto serie
	c := &serial.Config{Name: portName, Baud: baudRate, ReadTimeout: readDelay}
	ser, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}
	return ser, nil
}

func startWebSocketServer() {
	// Configuración de rutas de HTTP
	http.HandleFunc("/", handleWebSocket)

	// Iniciar servidor en el puerto 3000
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal("Error al iniciar el servidor:", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Actualizar cabeceras para permitir conexiones de origen cruzado (CORS)
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// Upgrade de la conexión HTTP a una conexión WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error al actualizar la conexión a WebSocket:", err)
		return
	}
	defer conn.Close()

	// Agregar el cliente a la lista de clientes conectados
	clients[conn] = true

	fmt.Println("Cliente conectado:", conn.RemoteAddr())

	for {
		// Esperar por mensajes del cliente (no se utiliza en este ejemplo)
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error al leer mensaje del cliente:", err)
			deleteClient(conn)
			break
		}
	}

	fmt.Println("Cliente desconectado:", conn.RemoteAddr())
}

func broadcastMessage(message int) {
	// Enviar el mensaje a todos los clientes conectados
	for client := range clients {
		err := client.WriteJSON(message)
		if err != nil {
			log.Println("Error al enviar mensaje al cliente:", err)
			deleteClient(client)
		}
	}
}

func deleteClient(client *websocket.Conn) {
	// Eliminar el cliente de la lista de clientes conectados
	delete(clients, client)
}
