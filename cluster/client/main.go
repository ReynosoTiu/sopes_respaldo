package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	pb "client/grpc"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Informacion representa los datos de una solicitud
type Informacion struct {
	AlbumTitulo string `json:"album"`
	Anio        string `json:"year"`
	Artista     string `json:"name"`
	Posicion    string `json:"rank"`
}

// ManejadorInsertarDatos procesa las peticiones POST para insertar datos
func ManejadorInsertarDatos(w http.ResponseWriter, req *http.Request) {
	var informacion Informacion
	if errorDecodificar := json.NewDecoder(req.Body).Decode(&informacion); errorDecodificar != nil {
		http.Error(w, errorDecodificar.Error(), http.StatusBadRequest)
		return
	}

	conexionGRPC, errorConexion := grpc.Dial("producer:3001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if errorConexion != nil {
		log.Fatalf("No se pudo establecer conexión con el servidor gRPC: %v", errorConexion)
	}
	defer conexionGRPC.Close()

	clienteGRPC := pb.NewGetInfoClient(conexionGRPC)
	respuesta, errorGRPC := clienteGRPC.ReturnInfo(context.Background(), &pb.RequestId{
		Album:  informacion.AlbumTitulo,
		Year:   informacion.Anio,
		Artist: informacion.Artista,
		Ranked: informacion.Posicion,
	})
	if errorGRPC != nil {
		log.Fatalf("Error durante la llamada al servicio gRPC: %v", errorGRPC)
	}

	json.NewEncoder(w).Encode(respuesta)
}

func iniciarServidor() {
	router := mux.NewRouter()
	router.HandleFunc("/insert", ManejadorInsertarDatos).Methods("POST")

	servidor := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(servidor.ListenAndServe())
}

func main() {
	iniciarServidor()
}
