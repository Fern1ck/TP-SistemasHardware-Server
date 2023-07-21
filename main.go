package main

import (
	"database/sql"
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
	Temperatura_Celcius float32
	Dispositivo           string
	Humedad 			int
	Fecha               string
}

func main() {
	// Load in the `.env` file
	err := godotenv.Load()
	if err != nil {
		log.Print("failed to load env", err)
	}

	DSN, present := os.LookupEnv("DSN")
	if !present {
        log.Fatal("DSN: not present\n")
        return
    }

	log.Print(DSN)
	// Open a connection to the database
	db, err = sql.Open("mysql", DSN)
	if err != nil {
		log.Fatal("failed to open db connection", err)
	}
	// Build router & define routes
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/registros/:dispositivo", GetRegistrosDelDispositivo)
	router.GET("/dispositivos", GetDispositivos)
	router.POST("/registro", CreateRegistro)
	router.DELETE("/registros", DeleteRegistros)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Not found"})
	})

	// Run the router
	router.Run()
}

// no usado actualmente
func GetRegistros(c *gin.Context) {
	query := "SELECT Dispositivo, Temperatura_Celcius, convert_tz(Fecha, '+00:00' ,'-03:00') FROM Registros order by Fecha desc LIMIT 30"
	res, err := db.Query(query)
	defer res.Close()
	if err != nil {
		log.Fatal("(GetRegistros) db.Query", err)
	}

	registros := []Registro{}
	for res.Next() {
		var registro Registro
		var fecha sql.RawBytes
		err := res.Scan(&registro.Dispositivo, &registro.Temperatura_Celcius, &fecha)
		if err != nil {
			log.Fatal("(GetRegistros) res.Scan", err)
		}
		registro.Fecha = (string(fecha[:]))
		registros = append(registros, registro)
	}

	c.JSON(http.StatusOK, registros)
}

func GetRegistrosDelDispositivo(c *gin.Context) {
	dispositivo := c.Param("dispositivo")
	query := "SELECT ROUND(Temperatura_Celcius, 0), Humedad, convert_tz(Fecha, '+00:00' ,'-03:00') FROM Registros where Dispositivo = ? order by Fecha desc LIMIT 30"
	res, err := db.Query(query, dispositivo)
	defer res.Close()
	if err != nil {
		log.Fatal("(GetRegistrosDelDispositivo) db.Query", err)
	}

	registros := []Registro{}
	for res.Next() {
		var registro Registro
		var fecha sql.RawBytes
		err := res.Scan(&registro.Temperatura_Celcius, &registro.Humedad, &fecha)
		if err != nil {
			log.Fatal("(GetRegistrosDelDispositivo) res.Scan", err)
		}
		registro.Fecha = string(fecha[:])
		registros = append(registros, registro)
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
	var newRegistro Registro
	err := c.BindJSON(&newRegistro)
	if err != nil {
		log.Fatal("(CreateRegistro) c.BindJSON", err)
	}

	var query string
	var res sql.Result
	query = `INSERT INTO Registros (Temperatura_Celcius, Humedad, Dispositivo) VALUES (?, ?, ?)`
	res, err = db.Exec(query, newRegistro.Temperatura_Celcius, newRegistro.Humedad, newRegistro.Dispositivo)

	if err != nil {
		log.Fatal("(CreateRegistro) db.Exec", err, res)
	}

	c.JSON(http.StatusOK, newRegistro)
}

func DeleteRegistros(c *gin.Context) {
	query := `DELETE FROM Registros`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("(DeleteRegistros) db.Exec", err)
	}

	c.Status(http.StatusOK)
}
