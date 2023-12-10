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

type InvitadoFamilia struct {
	Id_text         string
	Nombre          string
	Id_text_familia sql.NullString
	Nombre_familia  sql.NullString
	Asiste          sql.NullBool
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

	router.OPTIONS("/familias/presentacion/:id", enableCors)
	router.GET("/familias/presentacion/:id", getPresentacionFamiliaById)

	router.GET("/invitados", getInvitados)
	router.GET("/invitados/:id", getInvitadoById)
	router.GET("/invitados/byfamilia/:id", getInvitadoByFamiliaId)

	router.OPTIONS("/invitados/presentacion/:id", enableCors)
	router.GET("/invitados/presentacion/:id", getPresentacionInvitadoById)

	router.OPTIONS("/invitados/tabla-rsvp", enableCors)
	router.GET("/invitados/tabla-rsvp", getTablaRsvp)

	router.OPTIONS("/verificarInvitado/:id", enableCors)
	router.GET("/verificarInvitado/:id", verificarInvitado)

	router.OPTIONS("/verificarFamilia/:id", enableCors)
	router.GET("/verificarFamilia/:id", verificarFamilia)

	router.POST("/asistencia/rechazar", rechazarInvitacion)
	router.OPTIONS("/asistencia/rechazar", enableCors)
	router.POST("/asistencia/aceptar", aceptarInvitacion)
	router.OPTIONS("/asistencia/aceptar", enableCors)
	router.POST("/cancion", agregarCancion)
	router.OPTIONS("/cancion", enableCors)
	router.POST("/mensaje", agregarMensaje)
	router.OPTIONS("/mensaje", enableCors)

	router.Run(os.Getenv("LOCALPORT"))

}

func enableCors(gc *gin.Context) {
	gc.Header("Access-Control-Allow-Origin", "*")
	gc.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers, Content-Type, hx-current-url, hx-request, hx-target, hx-trigger")
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

func crearBotonAceptado(invitadoId string) string {
	return fmt.Sprintf(`

	<button type="button" id="aceptar%s" name="invitado_id" value="%s" class="aceptarSeleccionado">
		<span>Aceptado</span>
	</button>

	`, invitadoId, invitadoId)
}

func crearBotonRechazado(invitadoId string) string {
	return fmt.Sprintf(`
	
	<button type="button" id="rechazar%s" name="invitado_id" value="%s" class="rechazarSeleccionado">
		<span>Rechazado</span>
	</button>

	`, invitadoId, invitadoId)
}

func crearBotonAceptar(invitadoId string) string {
	return fmt.Sprintf(`

	<button type="button" id="aceptar%s" name="invitado_id" value="%s" class="aceptar"
		hx-post="https://wedding-back.fly.dev/asistencia/aceptar" 
		hx-select="#aceptar%s" 
		hx-swap="outerHTML" 
		hx-select-oob="#rechazar%s" 
		hx-indicator="#svg-load%s, #aceptar-svg%s"
		hx-ext="json-enc">

			<svg id="aceptar-svg%s" class="aceptar-svg response-svg" version="1.1" viewBox="0 0 167.13 173.09" xmlns="http://www.w3.org/2000/svg">
			<defs>
			<clipPath id="21d8b0576d-0">
			<path d="m550.25 1h473.75v492h-473.75z"/>
			</clipPath>
			</defs>
			<metadata>
			<rdf:RDF>
			<cc:Work rdf:about="">
				<dc:format>image/svg+xml</dc:format>
				<dc:type rdf:resource="http://purl.org/dc/dcmitype/StillImage"/>
				<dc:title/>
			</cc:Work>
			</rdf:RDF>
			</metadata>
			<g transform="translate(-143.57 -26.461)">
			<g transform="matrix(.35278 0 0 .35278 -50.548 25.877)" clip-path="url(#21d8b0576d-0)">
			<path d="m1006.1 1.6562-21.289 14.574c-111.31 76.172-225.3 233.87-282.17 362.08-8.7852-12.301-11.438-16.375-19.949-30.051-34.766-46.977-78.848-87.293-99.766-102.18l-32.684 13.664 1.0352 1.9414c22.566 24.711 62.164 79.543 121.84 200.37 3.6602 5.2188 7.5625 10.34 12.004 15.117 11.68 12.512 23.652 15.145 31.641 15.145h0.0117c22.781 0 36.215-17.992 45.465-39.055 59.77-221.92 201.01-374.69 262.09-427.02z" fill="#f9eae6"/>
			</g>
			</g>
			</svg>

			<svg class="loader-svg htmx-indicator" fill="#fff" version="1.1" id="svg-load%s" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
			viewBox="0 0 40 40" enable-background="new 0 0 0 0" xml:space="preserve">
			<circle fill="#fff" stroke="none" cx="6" cy="20" r="6">
				<animate
				attributeName="opacity"
				dur="1s"
				values="0;1;0"
				repeatCount="indefinite"
				begin="0.1"/>    
			</circle>
			<circle fill="#fff" stroke="none" cx="26" cy="20" r="6">
				<animate
				attributeName="opacity"
				dur="1s"
				values="0;1;0"
				repeatCount="indefinite" 
				begin="0.2"/>       
			</circle>
			<circle fill="#fff" stroke="none" cx="46" cy="20" r="6">
				<animate
				attributeName="opacity"
				dur="1s"
				values="0;1;0"
				repeatCount="indefinite" 
				begin="0.3"/>     
			</circle>
			</svg>
		</button>

	`, invitadoId, invitadoId, invitadoId, invitadoId, invitadoId, invitadoId, invitadoId, invitadoId)
}

func crearBotonRechazar(invitadoId string) string {
	return fmt.Sprintf(`
	
	<button type="button" id="rechazar%s" name="invitado_id" value="%s" class="rechazar"
		hx-post="https://wedding-back.fly.dev/asistencia/rechazar" 
		hx-select="#rechazar%s" 
		hx-swap="outerHTML" 
		hx-select-oob="#aceptar%s" 
		hx-indicator="#svg-load%s, #rechazar-svg%s"
		hx-ext="json-enc">
			
			<svg id="rechazar-svg%s" class="rechazar-svg response-svg" version="1.1" viewBox="0 0 181.44 181.43" xmlns="http://www.w3.org/2000/svg">
			<defs>
			<clipPath id="577a602fa9">
			<path d="m3 4h516v514.39h-516z"/>
			</clipPath>
			</defs>
			<metadata>
			<rdf:RDF>
			<cc:Work rdf:about="">
				<dc:format>image/svg+xml</dc:format>
				<dc:type rdf:resource="http://purl.org/dc/dcmitype/StillImage"/>
				<dc:title/>
			</cc:Work>
			</rdf:RDF>
			</metadata>
			<g transform="translate(34.777 -51.84)">
			<g transform="matrix(.35278 0 0 .35278 -36.182 50.389)" clip-path="url(#577a602fa9)">
			<path d="m317.32 261.29 189.34-189.36c15.516-15.516 15.516-40.672 0-56.184-15.516-15.516-40.668-15.516-56.18 0l-189.34 189.36-189.34-189.36c-15.516-15.516-40.664-15.516-56.18 0-15.512 15.516-15.512 40.668 0 56.184l189.34 189.36-189.34 189.36c-15.512 15.516-15.512 40.668 0 56.18 7.7578 7.7578 17.922 11.637 28.09 11.637s20.332-3.8789 28.09-11.637l189.34-189.36 189.34 189.36c7.7578 7.7578 17.922 11.637 28.09 11.637s20.332-3.8789 28.09-11.637c15.516-15.516 15.516-40.668 0-56.18z" fill="#f9eae6"/>
			</g>
			</g>
			</svg>

			<svg class="loader-svg htmx-indicator" fill="#fff" version="1.1" id="svg-load%s" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
			viewBox="0 0 40 40" enable-background="new 0 0 0 0" xml:space="preserve">
			<circle fill="#fff" stroke="none" cx="6" cy="20" r="6">
				<animate
				attributeName="opacity"
				dur="1s"
				values="0;1;0"
				repeatCount="indefinite"
				begin="0.1"/>    
			</circle>
			<circle fill="#fff" stroke="none" cx="26" cy="20" r="6">
				<animate
				attributeName="opacity"
				dur="1s"
				values="0;1;0"
				repeatCount="indefinite" 
				begin="0.2"/>       
			</circle>
			<circle fill="#fff" stroke="none" cx="46" cy="20" r="6">
				<animate
				attributeName="opacity"
				dur="1s"
				values="0;1;0"
				repeatCount="indefinite" 
				begin="0.3"/>     
			</circle>
			</svg>

		</button>

	`, invitadoId, invitadoId, invitadoId, invitadoId, invitadoId, invitadoId, invitadoId, invitadoId)
}

func crearFilaInvitado(invitado InvitadoResp) (fila string) {

	nombreInvitado := fmt.Sprintf("<span> %s </span>", invitado.Nombre)
	if !invitado.Asiste.Valid {
		return "<li>" + nombreInvitado + crearBotonAceptar(invitado.Id_text) + crearBotonRechazar(invitado.Id_text) + "</li>"
	}
	if invitado.Asiste.Bool {
		return "<li>" + nombreInvitado + crearBotonAceptado(invitado.Id_text) + crearBotonRechazar(invitado.Id_text) + "</li>"
	}
	return "<li>" + nombreInvitado + crearBotonAceptar(invitado.Id_text) + crearBotonRechazado(invitado.Id_text) + "</li>"
}

func verificarInvitado(gc *gin.Context) {
	enableCors(gc)
	invitadoId := gc.Param("id")
	invitado, err := getInvitadoByIdDB(invitadoId)
	if err != nil {
		gc.Status(http.StatusNotFound)
		return
	}

	filaInvitado := crearFilaInvitado(invitado)
	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(filaInvitado))
}

func verificarFamilia(gc *gin.Context) {

	enableCors(gc)
	familiaId := gc.Param("id")
	familia, err := getInvitadosByFamiliaIdDB(familiaId)
	if err != nil {
		gc.Status(http.StatusNotFound)
		return
	}

	var filaInvitados string
	for _, invitado := range familia {
		filaInvitados = filaInvitados + crearFilaInvitado(invitado)
	}

	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(filaInvitados))
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

	htmlStr := crearBotonAceptado(invitado.Invitado_Id) + crearBotonRechazar(invitado.Invitado_Id)
	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlStr))
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

	htmlStr := crearBotonAceptar(invitado.Invitado_Id) + crearBotonRechazado(invitado.Invitado_Id)

	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlStr))
}

func agregarCancion(gc *gin.Context) {
	enableCors(gc)
	var cancionRequest CancionRequest

	if err := gc.BindJSON(&cancionRequest); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("La petición es incorrecta \n %s", err)})
		return
	}

	_, errInv := getInvitadoByIdDB(cancionRequest.Invitado_Id)
	_, errFam := getFamiliaByIdDB(cancionRequest.Invitado_Id)

	if errInv != nil && errFam != nil {
		htmlStr := "<input type='text' id='cancion-input' name='nombre_cancion' value='' placeholder='ID Invitado Incorrecto' required>"
		gc.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(htmlStr))
		return
	}

	if _, err := db.Exec("INSERT INTO Canciones (id_invitado, fecha, nombre_cancion) VALUES(?, current_timestamp(), ?)", cancionRequest.Invitado_Id, cancionRequest.Nombre_Cancion); err != nil {
		htmlStr := "<input type='text' id='cancion-input' name='nombre_cancion' value='' placeholder='Error. Intentalo de nuevo' required>"
		gc.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(htmlStr))
		return
	}

	htmlStr := "<input type='text' id='cancion-input' name='nombre_cancion' value='' placeholder='¡Gracias! Agrega otra ...' required>"

	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlStr))
}

func agregarMensaje(gc *gin.Context) {
	enableCors(gc)
	var mensajeRequest MensajeRequest

	if err := gc.BindJSON(&mensajeRequest); err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("La petición es incorrecta \n %s", err)})
		return
	}

	_, errInv := getInvitadoByIdDB(mensajeRequest.Invitado_Id)
	_, errFam := getFamiliaByIdDB(mensajeRequest.Invitado_Id)

	if errInv != nil && errFam != nil {
		htmlStr := "<textarea name='mensaje' id='mensaje-textarea' placeholder='Error, intentalo de nuevo' rows='30' required></textarea>"
		gc.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(htmlStr))
		return
	}

	if _, err := db.Exec("INSERT INTO Mensajes (id_invitado, fecha, contenido) VALUES(?, current_timestamp(), ?)", mensajeRequest.Invitado_Id, mensajeRequest.Mansaje); err != nil {
		htmlStr := "<textarea name='mensaje' id='mensaje-textarea' placeholder='Error, intentalo de nuevo' rows='30' required></textarea>"
		gc.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(htmlStr))
		return
	}

	htmlStr := "<textarea name='mensaje' id='mensaje-textarea' placeholder='¡Gracias por tu mensaje! Puedes ingresar otro' rows='30' required></textarea>"

	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlStr))
}

func getPresentacionFamiliaById(gc *gin.Context) {
	enableCors(gc)
	id := gc.Param("id")

	familia, err := getFamiliaByIdDB(id)

	if err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("La petición es incorrecta \n %s", err)})
		return
	}

	htmlStr := fmt.Sprintf(`<h1 class='nombre-presentacion nombre-principal'>%s</h1> 
							<h1 class='nombre-presentacion nombre-secundario'>%s</h1>`, familia.Nombre, familia.Nombre_invitacion)

	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlStr))
}

func getPresentacionInvitadoById(gc *gin.Context) {
	enableCors(gc)
	id := gc.Param("id")

	invitado, err := getInvitadoByIdDB(id)

	if err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("La petición es incorrecta \n %s", err)})
		return
	}

	htmlStr := fmt.Sprintf(`<h1 class='nombre-presentacion nombre-principal'>%s</h1> 
							<h1 class='nombre-presentacion nombre-secundario'>%s</h1>`, invitado.Nombre, invitado.Nombre_invitacion)

	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlStr))
}

func getInvitadosFamiliaDB() ([]InvitadoFamilia, error) {

	var invitadosFamilias []InvitadoFamilia

	invResp, err := db.Query("SELECT inv.id_text, inv.nombre, inv.asiste, fam.id_text, fam.nombre FROM Invitados inv LEFT JOIN Familias fam on inv.id_familia = fam.id")

	if err != nil {
		return nil, fmt.Errorf("getInvitadosFamiliaDB %s", err)
	}

	defer invResp.Close()

	for invResp.Next() {
		var invitadoFam InvitadoFamilia
		if err := invResp.Scan(
			&invitadoFam.Id_text,
			&invitadoFam.Nombre,
			&invitadoFam.Asiste,
			&invitadoFam.Id_text_familia,
			&invitadoFam.Nombre_familia); err != nil {
			return nil, fmt.Errorf("getInvitadosDb %s", err)
		}
		invitadosFamilias = append(invitadosFamilias, invitadoFam)
	}

	return invitadosFamilias, nil
}

func getTablaRsvp(gc *gin.Context) {
	enableCors(gc)

	invitados, err := getInvitadosFamiliaDB()

	if err != nil {
		gc.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Ha sucedido un error por favor intentelo de nuevo \n %s", err)})
		return
	}

	var htmlStr string
	invitadosNo := len(invitados)
	var invitadosSinRespuesta int32
	var invitadosRechazado int32
	var invitadosAceptado int32

	for _, inv := range invitados {
		classStr := getClassAsisteByInv(inv.Asiste)
		htmlStr = htmlStr + fmt.Sprintf(`	
			<div class="asistencia">
				<span class="nombre-asistente">%s</span>
				<span class="familia-asistente">%s</span>
				<span class="asiste %s"></span>
			</div>`,
			inv.Nombre, inv.Nombre_familia.String, classStr)
		if !inv.Asiste.Valid {
			invitadosSinRespuesta += 1
			continue
		}
		if inv.Asiste.Bool {
			invitadosAceptado += 1
			continue
		}
		invitadosRechazado += 1
	}

	htmlStr = fmt.Sprintf(`<div class="totales-invitados">
		<div class="total-data total-invitados">
			<h1>%v</h1>
			<h2>INV</h2>
		</div>
		<div class="total-data total-sin">
			<h1>%v</h1>
			<h2>SIN</h2>
		</div>
		<div class="total-data total-rechazadas">
			<h1>%v</h1>
			<h2>RCH</h2>
		</div>
		<div class="total-data total-aceptadas">
			<h1>%v</h1>
			<h2>ACP</h2>
		</div>
	</div>
	<div class="outter-asistencia-container">
		<div class="asistencia-container" id="asistencia-container"> 
				%s 
		</div>
	</div>`,
		invitadosNo, invitadosSinRespuesta, invitadosRechazado, invitadosAceptado, htmlStr)

	gc.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlStr))

}

func getClassAsisteByInv(asiste sql.NullBool) string {
	if !asiste.Valid {
		return "sin-respuesta"
	}
	if asiste.Bool {
		return "si-asiste"
	}
	return "no-asiste"
}
