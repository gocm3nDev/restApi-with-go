package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type customer struct {
	Id     int    `json:"id"`
	Ad     string `json:"name"`
	Soyad  string `json:"surname"`
	Sehir  string `json:"city"`
	Bakiye int    `json:"balance"`
}

func ConnectDB() *sql.DB {
	const (
		host     = "localhost"
		port     = 5432
		user     = "postgres"
		password = "427973aefe@root"
		dbname   = "dbUrun"
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Connection failed: ", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Ping failed: ", err)
	}
	fmt.Println("Connection established as successfully to postgres.")

	return db
}

func getCustomerById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "invalid id"})
		return
	}

	db := ConnectDB()
	defer db.Close()

	var cust customer

	err = db.QueryRow("SELECT id, ad, soyad, sehir, bakiye FROM musteri WHERE id = $1", id).Scan(&cust.Id, &cust.Ad, &cust.Soyad, &cust.Sehir, &cust.Bakiye)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"result": "no matching users"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"result": "query error"})
		return
	}

	c.JSON(http.StatusOK, cust)
}

func getAllCustomers(c *gin.Context) {
	db := ConnectDB()
	defer db.Close() // If at the end of the main function sql.DB still opened, defer will close.

	rows, err := db.Query("select id, ad, soyad, sehir, bakiye from musteri")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query error"})
		return
	}
	defer rows.Close()

	var customers []customer

	for rows.Next() {
		var cust customer
		err := rows.Scan(&cust.Id, &cust.Ad, &cust.Soyad, &cust.Sehir, &cust.Bakiye)

		if err != nil {
			log.Println("An error occured while reading rows: ", err)
			continue
		}
		customers = append(customers, cust)
	}

	c.JSON(http.StatusOK, customers)
}

func removeCustomer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid id"})
		return
	}

	db := ConnectDB()
	defer db.Close()

	row, err := db.Exec("DELETE FROM musteri WHERE id = $1", id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	rowsAffected, err := row.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get affected rows"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "no record found to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted successfully"})
}

func main() {
	router := gin.Default()

	router.GET("/customers", getAllCustomers)
	router.GET("/customers/:id", getCustomerById)
	router.DELETE("/customers/:id", removeCustomer)

	router.Run(":8080") // http://localhost:8080/customers
}
