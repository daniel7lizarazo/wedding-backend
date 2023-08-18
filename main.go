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

type FamiliasResp struct {
	Id                int64
	Id_text           string
	Nombre            string
	Miembro_principal int64
	Nombre_invitacion string
}

type FamiliasCommand struct {
	Nombre            string
	Miembro_principal int64
	Nombre_invitacion string
}

type Asistencia struct {
	Id_text string `json:"id_text"`
	Asiste  bool   `json:"asiste"`
}

type InvitadoId struct {
	Invitado_Id string `json:"invitado_id"`
}

type CancionRequest struct {
	Invitado_Id    string `json:"invitado_id"`
	Nombre_Cancion string `json:"nombre_cancion"`
}

type MensajeRequest struct {
	Invitado_Id string `json:"invitado_id"`
	Mansaje     string `json:"mensaje"`
}

type Prueba struct {
	Input string `json:"input"`
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

	router.GET("/familias", getFamilias)
	router.GET("/familias/:id", getFamiliaById)
	router.GET("/invitados", getInvitados)
	router.GET("/invitados/:id", getInvitadoById)
	router.GET("/invitados/byfamilia/:id", getInvitadoByFamiliaId)
	router.POST("/asistencia/rechazar", rechazarInvitacion)
	router.OPTIONS("/asistencia/rechazar", enableCors)
	router.POST("/asistencia/aceptar", aceptarInvitacion)
	router.OPTIONS("/asistencia/aceptar", enableCors)
	router.POST("/cancion", agregarCancion)
	router.OPTIONS("/cancion", enableCors)
	router.POST("/mensaje", agregarMensaje)
	router.OPTIONS("/mensaje", enableCors)

	router.POST("/prueba1", prueba1)
	router.OPTIONS("/prueba1", enableCors)

	router.POST("/prueba2", prueba2)
	router.OPTIONS("/prueba2", enableCors)

	router.POST("/pruebaHXTrigger", pruebaHXTrigger)
	router.OPTIONS("/pruebaHXTrigger", enableCors)

	router.Run(os.Getenv("LOCALPORT"))

}

func enableCors(gc *gin.Context) {
	gc.Header("Access-Control-Allow-Origin", "*")
	gc.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers, Content-Type, HX-Trigger")
	gc.Header("Content-Type", "application/json")
	gc.Status(http.StatusNoContent)
}

func getInvitados(gc *gin.Context) {
	enableCors(gc)
	invitados, err := getInvitadosDB()

	if err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": "No se encontraron invitados"})
	}

	gc.IndentedJSON(http.StatusOK, invitados)
}

func getInvitadosDB() ([]InvitadoResp, error) {

	var invitados []InvitadoResp

	invResp, err := db.Query("SELECT id, id_text, nombre, nombre_invitacion, asiste FROM Invitados")

	if err != nil {
		return nil, fmt.Errorf("getInvitadosDb %s", err)
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
			return nil, fmt.Errorf("getInvitadosDb %s", err)
		}
		invitados = append(invitados, invitado)
	}

	return invitados, nil
}

func getInvitadoById(gc *gin.Context) {
	enableCors(gc)
	id := gc.Param("id")

	invitado, err := getInvitadoByIdDB(id)

	if err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("No se encontró un invitado con el id %v", id)})
	}

	// return invitado, nil
	gc.IndentedJSON(http.StatusOK, invitado)
}

func getInvitadoByIdDB(id_text string) (InvitadoResp, error) {
	var invitado InvitadoResp

	row := db.QueryRow("SELECT id, id_text, nombre, nombre_invitacion, asiste FROM Invitados WHERE id_text = ?", id_text)

	if err := row.Scan(&invitado.Id, &invitado.Id_text, &invitado.Nombre, &invitado.Nombre_invitacion, &invitado.Asiste); err != nil {
		return invitado, fmt.Errorf("Get invitado ny ID DB %s", err)
	}

	return invitado, nil
}

func getFamilias(gc *gin.Context) {
	enableCors(gc)
	familias, err := getFamiliasDb()

	if err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": "No se encontraron familias"})
	}

	gc.IndentedJSON(http.StatusOK, familias)
}

func getFamiliasDb() ([]FamiliasResp, error) {

	var familias []FamiliasResp

	famResp, err := db.Query("SELECT id, id_text, nombre, miembro_principal, nombre_invitacion FROM Familias")

	if err != nil {
		return nil, fmt.Errorf("getFamiliasDb %s", err)
	}

	defer famResp.Close()

	for famResp.Next() {
		var familia FamiliasResp
		if err := famResp.Scan(
			&familia.Id,
			&familia.Id_text,
			&familia.Nombre,
			&familia.Miembro_principal,
			&familia.Nombre_invitacion); err != nil {
			return nil, fmt.Errorf("getFamiliasDb %s", err)
		}
		familias = append(familias, familia)
	}

	return familias, nil
}

func getFamiliaById(gc *gin.Context) {
	enableCors(gc)
	id := gc.Param("id")

	familia, err := getFamiliaByIdDB(id)

	if err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("No se encontró una familia con el id %v", id)})
	}

	gc.IndentedJSON(http.StatusOK, familia)
}

func getFamiliaByIdDB(id_familia string) (FamiliasResp, error) {
	var familia FamiliasResp

	row := db.QueryRow("SELECT id, id_text, nombre, miembro_principal, nombre_invitacion FROM Familias WHERE id_text = ?", id_familia)

	if err := row.Scan(
		&familia.Id,
		&familia.Id_text,
		&familia.Nombre,
		&familia.Miembro_principal,
		&familia.Nombre_invitacion); err != nil {
		return familia, fmt.Errorf("get familia by id DB %s", err)
	}

	return familia, nil
}

func getInvitadoByFamiliaId(gc *gin.Context) {
	enableCors(gc)
	id := gc.Param("id")

	invitados, err := getInvitadosByFamiliaIdDB(id)

	if err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": "No se encontraron invitados"})
	}

	gc.IndentedJSON(http.StatusOK, invitados)
}

func getInvitadosByFamiliaIdDB(id string) ([]InvitadoResp, error) {

	var invitados []InvitadoResp

	invResp, err := db.Query("SELECT i.id, i.id_text, i.nombre, i.nombre_invitacion, asiste FROM WeddingDB.Invitados i INNER JOIN WeddingDB.Familias f ON i.id_familia = f.id WHERE f.id_text = ?", id)

	if err != nil {
		return nil, fmt.Errorf("getInvitadosDb %s", err)
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
			return nil, fmt.Errorf("getInvitadosDb %s", err)
		}
		invitados = append(invitados, invitado)
	}

	return invitados, nil
}

func updateMultiplesInvitadosAsistencia(gc *gin.Context) {
	enableCors(gc)
	var listaAsistencia []Asistencia

	if err := gc.BindJSON(&listaAsistencia); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Lista de asistencia invalida \n %s", err)})
		return
	}

	listaAsiste := "UPDATE Invitados SET asiste = 1 WHERE id_text IN ("
	listaNoAsiste := "UPDATE Invitados SET asiste = 0 WHERE id_text IN ("

	for _, asistencia := range listaAsistencia {
		if asistencia.Asiste {
			if listaAsiste != "UPDATE Invitados SET asiste = 1 WHERE id_text IN (" {
				listaAsiste = listaAsiste + ", "
			}
			listaAsiste = listaAsiste + "'" + asistencia.Id_text + "'"
			continue
		}
		if listaNoAsiste != "UPDATE Invitados SET asiste = 0 WHERE id_text IN (" {
			listaNoAsiste = listaNoAsiste + ", "
		}

		listaNoAsiste = listaNoAsiste + "'" + asistencia.Id_text + "'"
	}

	listaAsiste = listaAsiste + ");"
	listaNoAsiste = listaNoAsiste + ");"

	if listaAsiste != "UPDATE Invitados SET asiste = 1 WHERE id_text IN ();" {
		if _, err := db.Exec(listaAsiste); err != nil {
			gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Hubo un problema realizando la actualización de los que SI asisten.\n %s \n %s", err, listaAsiste)})
			return
		}
	}

	if listaNoAsiste != "UPDATE Invitados SET asiste = 0 WHERE id_text IN ();" {
		if _, err := db.Exec(listaNoAsiste); err != nil {
			gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Hubo un problema realizando la actualización de los que NO asisten.\n %s \n %s", err, listaNoAsiste)})
			return
		}
	}

	gc.IndentedJSON(http.StatusOK, gin.H{"message": "Asistencia actualizada"})

}

func rechazarInvitacion(gc *gin.Context) {
	enableCors(gc)
	var invitado InvitadoId

	if err := gc.BindJSON(&invitado); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Id del invitado invalido \n %s", err)})
		return
	}

	if _, err := db.Exec(fmt.Sprintf("UPDATE Invitados SET asiste = 0 WHERE id_text = '%s'", invitado.Invitado_Id)); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Hubo un problema realizando la actualización.\n %s \n %s", err, invitado)})
		return
	}

	htmlStr := fmt.Sprintf(`<button type="button" id="rechazar" name="invitado_id" value="%s" class="rechazarSeleccionado"
                    hx-get="/components/rechazar-button-%s" hx-trigger="click from:button#aceptar[value='%s']" hx-swap="outerHTML">
						<span>Rechazado</span>
                    </button>`, invitado.Invitado_Id, invitado.Invitado_Id, invitado.Invitado_Id)

	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlStr))
	// gc.Status(http.StatusNoContent)

}

func aceptarInvitacion(gc *gin.Context) {
	enableCors(gc)
	var invitado InvitadoId

	if err := gc.BindJSON(&invitado); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Id del invitado invalido \n %s", err)})
		return
	}

	if _, err := db.Exec(fmt.Sprintf("UPDATE Invitados SET asiste = 1 WHERE id_text = '%s'", invitado.Invitado_Id)); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Hubo un problema realizando la actualización.\n %s \n %s", err, invitado)})
		return
	}

	htmlStr := fmt.Sprintf(`<button type="button" id="aceptar" name="invitado_id" value="%s" class="aceptarSeleccionado" 
				hx-get="/components/aceptar-button-%s" hx-trigger="click from:button#rechazar[value='%s']" hx-swap="outerHTML">
					<span>Aceptado</span>
				</button>`, invitado.Invitado_Id, invitado.Invitado_Id, invitado.Invitado_Id)

	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlStr))
	// gc.Status(http.StatusNoContent)
}

func agregarCancion(gc *gin.Context) {
	enableCors(gc)
	var cancionRequest CancionRequest

	if err := gc.BindJSON(&cancionRequest); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("La petición es incorrecta \n %s", err)})
		return
	}

	_, errInv := getInvitadoByIdDB(cancionRequest.Invitado_Id)
	if errInv == nil {
		if _, err := db.Exec("INSERT INTO Canciones (id_invitado, fecha, nombre_cancion) VALUES(?, current_timestamp(), ?)", cancionRequest.Invitado_Id, cancionRequest.Nombre_Cancion); err != nil {
			gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Hubo un problema realizando la inserción de la canción.\n %s \n %s", err, cancionRequest)})
			return
		}

		gc.Status(http.StatusNoContent)

		return
	}

	_, errFam := getFamiliaByIdDB(cancionRequest.Invitado_Id)
	if errFam != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("No se encuentra un registro que concuerde con este id %s", cancionRequest.Invitado_Id)})
		return
	}

	if _, err := db.Exec("INSERT INTO Canciones (id_invitado, fecha, nombre_cancion) VALUES(?, current_timestamp(), ?)", cancionRequest.Invitado_Id, cancionRequest.Nombre_Cancion); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Hubo un problema realizando la inserción de la canción.\n %s \n %s", err, cancionRequest)})
		return
	}

	gc.Status(http.StatusNoContent)

}

func agregarMensaje(gc *gin.Context) {
	enableCors(gc)
	var mensajeRequest MensajeRequest

	if err := gc.BindJSON(&mensajeRequest); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("La petición es incorrecta \n %s", err)})
		return
	}

	_, errInv := getInvitadoByIdDB(mensajeRequest.Invitado_Id)
	if errInv == nil {
		if _, err := db.Exec("INSERT INTO Mensajes (id_invitado, fecha, contenido) VALUES(?, current_timestamp(), ?)", mensajeRequest.Invitado_Id, mensajeRequest.Mansaje); err != nil {
			gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Hubo un problema realizando la inserción del mensaje.\n %s \n %s", err, mensajeRequest)})
			return
		}

		gc.Status(http.StatusNoContent)

		return
	}

	_, errFam := getFamiliaByIdDB(mensajeRequest.Invitado_Id)
	if errFam != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("No se encuentra un registro que concuerde con este id %s", mensajeRequest.Invitado_Id)})
		return
	}

	if _, err := db.Exec("INSERT INTO Mensajes (id_invitado, fecha, contenido) VALUES(?, current_timestamp(), ?)", mensajeRequest.Invitado_Id, mensajeRequest.Mansaje); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Hubo un problema realizando la inserción del mensaje.\n %s \n %s", err, mensajeRequest)})
		return
	}

	gc.Status(http.StatusNoContent)

}

func prueba1(gc *gin.Context) {
	enableCors(gc)
	var inputRequest Prueba

	if err := gc.BindJSON(&inputRequest); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("La petición es incorrecta \n %s", err)})
		return
	}

	fmt.Printf("El id de invitado es  %s", inputRequest.Input)

	strBoton := fmt.Sprintf("<button name='input' value='%s' hx-post='http://localhost:8080/prueba2' hx-ext='json-enc' hx-swap='outerHTML'  class='aceptar'>a probar 2</button>", inputRequest.Input)

	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(strBoton))
}

func prueba2(gc *gin.Context) {
	enableCors(gc)
	var inputRequest Prueba

	if err := gc.BindJSON(&inputRequest); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("La petición es incorrecta \n %s", err)})
		return
	}

	fmt.Printf("El id de invitado es  %s", inputRequest.Input)

	strBoton := fmt.Sprintf("<button name='input' value='%s' hx-post='http://localhost:8080/prueba1' hx-ext='json-enc' hx-swap='outerHTML' class='aceptarSeleccionado'>a probar 1</button>", inputRequest.Input)

	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(strBoton))
}

func pruebaHXTrigger(gc *gin.Context) {
	enableCors(gc)
	var inputRequest Prueba

	if err := gc.BindJSON(&inputRequest); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("La petición es incorrecta \n %s", err)})
		return
	}

	// gc.Header("hX-TriGger", "pruebaEvent")
	gc.Writer.Header()["HX-Trigger"] = []string{"{\"pruebaEvent\": \"\"}"}
	// gc.Writer.Header()["HX-Trigger"] = []string{struct{ eventoPrueba string }{eventoPrueba: ""}}

	// header := gin.H{"HX-Trigger2": "pruebaEvent"}

	// gc.ShouldBindHeader(header)
	gc.Status(http.StatusOK)
	fmt.Printf("estos son los headers %+v", gc.Writer.Header())
	// gc.IndentedJSON(http.StatusNoContent, gin.H{"HX-Trigger": "pruebaEvent"})
}
