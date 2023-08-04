package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)
var db *sql.DB

type Registro struct {
	Dispositivo string
	Fecha       string
	ID          int
	Datos       []Dato
}

type Dato struct {
	ID       int
	Variable string
	Valor    string
}

type RegistroData struct {
	Dispositivo string
	Datos       map[string]interface{}
}

func main() {
	// Load in the `.env` file
	err := godotenv.Load()
	if err != nil {
		log.Print("failed to load env", err)
	}

	//set gin on release or debug mode
	ENV, present := os.LookupEnv("ENV")
	if !present {
        log.Fatal("ENV: not present\n")
        return
    }
	if ENV == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	if ENV == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	DSN, present := os.LookupEnv("DSN")
	if !present {
        log.Fatal("DSN: not present\n")
        return
    }

	// Open a connection to the database
	db, err = sql.Open("mysql", DSN)
	if err != nil {
		log.Fatal("failed to open db connection", err)
	}
	// Build router & define routes
	router := gin.Default()

	router.Use(cors.Default())
	router.GET("/registros/:dispositivo", GetRegistrosDelDispositivo)
	router.GET("/registrosJSON", ExportJSON)
	router.GET("/dispositivos", GetDispositivos)
	router.POST("/registro", CreateRegistro)
	router.DELETE("/registros", DeleteRegistros)
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Not found"})
	})

	// Run the router
	router.Run()
}

func ExportJSON(ctx *gin.Context) {
	query := `
		SELECT r.ID, r.Dispositivo, convert_tz(r.Fecha, '+00:00', '-03:00') AS Converted_Fecha,
		d.ID_Dato AS Dato_ID, d.Variable, d.Valor
		FROM Registros r 
		LEFT JOIN Datos d ON r.ID = d.ID_Registro
		ORDER BY r.Fecha ASC`
	res, err := db.Query(query)

	defer res.Close()
	if err != nil {
		log.Fatal("(GetRegistrosDelDispositivo) db.Query", err)
	}

	registros := []Registro{}
	currentRegistro := Registro{}
	currentDato := Dato{}

	for res.Next() {
		var fecha sql.RawBytes
		var datoID sql.NullInt64
		err := res.Scan(&currentRegistro.ID, &currentRegistro.Dispositivo, &fecha, &datoID, &currentDato.Variable, &currentDato.Valor)
		if err != nil {
			log.Fatal("(GetRegistrosDelDispositivo) res.Scan", err)
		}

		// Check if the current row is a new registro
		if currentRegistro.Fecha != string(fecha[:]) {
			if currentRegistro.Fecha != "" {
				// Append the current registro to the registros slice
				registros = append(registros, currentRegistro)
			}

			// Create a new registro with the current data
			currentRegistro = Registro{
				ID:          currentRegistro.ID,
				Dispositivo: currentRegistro.Dispositivo,
				Fecha:       string(fecha[:]),
				Datos:       []Dato{},
			}
		}

		// Check if the current row has a dato entry
		if datoID.Valid {
			// Append the current dato to the datos slice of the current registro
			currentDato.ID = int(datoID.Int64)
			currentRegistro.Datos = append(currentRegistro.Datos, currentDato)
		}
	}

	// Append the last registro to the registros slice
	if currentRegistro.Fecha != "" {
		registros = append(registros, currentRegistro)
	}

	jsonData, err := json.MarshalIndent(registros, "", "  ")
	if err != nil {
	    log.Fatal("(ExportJSON) json.MarshalIndent", err)
	}

	ctx.Header("Content-Disposition", "attachment; filename=registros.json")
	ctx.Header("Accept-Length", fmt.Sprintf("%d", len(jsonData)))
	ctx.Data(http.StatusOK, "application/json", jsonData)	
}

func GetRegistrosDelDispositivo(c *gin.Context) {
	dispositivo := c.Param("dispositivo")
	query := `
		SELECT r.ID, r.Dispositivo, convert_tz(r.Fecha, '+00:00', '-03:00') AS Converted_Fecha,
		d.ID_Dato AS Dato_ID, d.Variable, d.Valor
		FROM Registros r 
		LEFT JOIN Datos d ON r.ID = d.ID_Registro
		WHERE r.Dispositivo = ? 
		ORDER BY r.Fecha DESC 
		LIMIT 30;`
	res, err := db.Query(query, dispositivo)
	defer res.Close()
	if err != nil {
		log.Fatal("(GetRegistrosDelDispositivo) db.Query", err)
	}

	registros := []Registro{}
	currentRegistro := Registro{}
	currentDato := Dato{}

	for res.Next() {
		var fecha sql.RawBytes
		var datoID sql.NullInt64
		err := res.Scan(&currentRegistro.ID, &currentRegistro.Dispositivo, &fecha, &datoID, &currentDato.Variable, &currentDato.Valor)
		if err != nil {
			log.Fatal("(GetRegistrosDelDispositivo) res.Scan", err)
		}

		// Check if the current row is a new registro
		if currentRegistro.Fecha != string(fecha[:]) {
			if currentRegistro.Fecha != "" {
				// Append the current registro to the registros slice
				registros = append(registros, currentRegistro)
			}

			// Create a new registro with the current data
			currentRegistro = Registro{
				ID:          currentRegistro.ID,
				Dispositivo: currentRegistro.Dispositivo,
				Fecha:       string(fecha[:]),
				Datos:       []Dato{},
			}
		}

		// Check if the current row has a dato entry
		if datoID.Valid {
			// Append the current dato to the datos slice of the current registro
			currentDato.ID = int(datoID.Int64)
			currentRegistro.Datos = append(currentRegistro.Datos, currentDato)
		}
	}

	// Append the last registro to the registros slice
	if currentRegistro.Fecha != "" {
		registros = append(registros, currentRegistro)
	}

	c.JSON(http.StatusOK, registros)
}

func GetDispositivos(c *gin.Context) {
	query := "SELECT DISTINCT Dispositivo FROM Registros"
	res, err := db.Query(query)
	defer res.Close()
	if err != nil {
		log.Fatal("(GetDispositivos) db.Query", err)
	}
	var listaDispositivos []string	
	for res.Next() {
		var dispositivo string	
		err := res.Scan(&dispositivo)
		if err != nil {
			log.Fatal("(GetDispositivos) res.Scan", err)
		}
		listaDispositivos = append(listaDispositivos, dispositivo)
	}

	c.JSON(http.StatusOK, listaDispositivos)
}

func CreateRegistro(c *gin.Context) {
	var registroData RegistroData
	err := c.BindJSON(&registroData)
	log.Printf("registroData: %v", registroData)
	if err != nil {
		log.Fatal("(CreateRegistro) c.BindJSON", err)
	}

	// Create entry in Registros table
	queryRegistro := "INSERT INTO Registros (Dispositivo) VALUES (?)"
	res, err := db.Exec(queryRegistro, registroData.Dispositivo)
	if err != nil {
		log.Fatal("(CreateRegistro) db.Exec", err)
	}
	registroID, err := res.LastInsertId()
	if err != nil {
		log.Fatal("(CreateRegistro) res.LastInsertId", err)
	}

	// Create entries in Datos table for each key-value pair in datos field
	queryDatos := "INSERT INTO Datos (ID_Registro, Variable, Valor) VALUES (?, ?, ?)"
	for variable, valor := range registroData.Datos {
		_, err := db.Exec(queryDatos, registroID, variable, valor)
		if err != nil {
			log.Fatal("(CreateRegistro) db.Exec", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registro created successfully"})
}

func DeleteRegistros(c *gin.Context) {
	query := `DELETE FROM Registros`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("(DeleteRegistros) db.Exec", err)
	}

	c.Status(http.StatusOK)
}
