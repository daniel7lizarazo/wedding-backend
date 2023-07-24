package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

type InvitadoResp struct {
	Id                int64
	Id_text           string
	Nombre            string
	Nombre_invitacion string
	Asiste            sql.NullBool
}

type InvitadoCommand struct {
	Nombre            string
	Nombre_invitacion string
	Asiste            *bool
}

func main() {
	// Capture connection properties.
	cfg := mysql.Config{
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 os.Getenv("DBADRESS"),
		DBName:               os.Getenv("DBNAME"),
		AllowNativePasswords: true,
	}

	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	fmt.Println("connected!")

	db.SetMaxOpenConns(2)

	router := gin.Default()

	router.GET("/invitados", getInvitados)
	router.GET("/invitados/:id", getInvitadoById)

	router.Run(os.Getenv("LOCALPORT"))

}

func getInvitados(gc *gin.Context) {
	invitados, err := getInvitadosDb()

	if err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": "No se encontraron invitados"})
	}

	gc.IndentedJSON(http.StatusOK, invitados)
}

func getInvitadosDb() ([]InvitadoResp, error) {

	var invitados []InvitadoResp

	invResp, err := db.Query("SELECT id, id_text, nombre, nombre_invitacion, asiste FROM Invitados")

	if err != nil {
		return nil, fmt.Errorf("getInvitados %s", err)
	}

	defer invResp.Close()

	for invResp.Next() {
		var invitado InvitadoResp
		if err := invResp.Scan(
			&invitado.Id,
			&invitado.Id_text,
			&invitado.Nombre,
			&invitado.Nombre_invitacion,
			&invitado.Asiste); err != nil {
			return nil, fmt.Errorf("getInvitados %s", err)
		}
		invitados = append(invitados, invitado)
	}

	return invitados, nil
}

func getInvitadoById(gc *gin.Context) {
	id := gc.Param("id")

	// An invitador slice to hold the data returned.
	var invitado InvitadoResp

	row := db.QueryRow("SELECT id, id_text, nombre, nombre_invitacion, asiste FROM Invitados WHERE id_text = ?", id)

	if err := row.Scan(&invitado.Id, &invitado.Id_text, &invitado.Nombre, &invitado.Nombre_invitacion, &invitado.Asiste); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("No se encontró uninvitado con el id %v", id)})
	}

	// return invitado, nil
	gc.IndentedJSON(http.StatusOK, invitado)
}
