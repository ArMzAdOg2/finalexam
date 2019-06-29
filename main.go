package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ArMzAdOg2/finalexam/model"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func createTable() *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("error", err.Error())
	}
	_, err = db.Exec("select * from customer")
	if err != nil {
		createTb := `
		CREATE TABLE customer(
			id SERIAL PRIMARY KEY,
			name TEXT,
			email TEXT,
			status TEXT
		);
		`
		_, err = db.Exec(createTb)
		if err != nil {
			log.Fatal("error", err.Error())
		}
	}
	return db
}

func getDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func handlerError(err error, c *gin.Context) bool {
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return false
	}
	return true
}

func insertCustomer(c *gin.Context) {
	customer := model.Customer{}
	err := c.ShouldBindJSON(&customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	stmt, err := DB.Prepare("insert into customer (name,email,status) values ($1,$2,$3) returning id;")
	row := stmt.QueryRow(customer.Name, customer.Email, customer.Status)
	var id int
	err = row.Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	customer.ID = id
	c.JSON(201, customer)
}

func getCustomerByID(c *gin.Context) {
	customer := model.Customer{}
	id := c.Param("id")

	stmt, err := DB.Prepare("select * from customer where id = $1")
	row := stmt.QueryRow(id)
	err = row.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, customer)
}

func getCustomer(c *gin.Context) {
	customers := []model.Customer{}
	stmt, err := DB.Prepare("select id,name,email,status from customer;")
	rows, err := stmt.Query()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	for rows.Next() {
		customer := model.Customer{}
		err = rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		customers = append(customers, customer)
	}
	c.JSON(http.StatusOK, customers)
}

func updateCustomerByID(c *gin.Context) {
	customer := model.Customer{}
	id := c.Param("id")
	err := c.ShouldBindJSON(&customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	stmt, err := DB.Prepare("update customer set name=$2 , email=$3, status=$4 where id = $1")
	_, err = stmt.Query(id, customer.Name, customer.Email, customer.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, customer)
}

func deteleCustomerByID(c *gin.Context) {
	id := c.Param("id")
	stmt, err := DB.Prepare("delete from customer where id=$1")
	_, err = stmt.Query(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "customer deleted",
	})
}

func middleWare(c *gin.Context) {
	token := c.GetHeader("Authorization")
	fmt.Println(token)
	if token != "token2019" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": http.StatusText(http.StatusUnauthorized),
		})
		return
	}
	c.Next()
}

func main() {
	DB = createTable()
	r := setupRouter()
	r.Run(":2019")
	defer DB.Close()
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(middleWare)
	r.POST("/customers", insertCustomer)
	r.GET("/customers/:id", getCustomerByID)
	r.GET("/customers", getCustomer)
	r.PUT("/customers/:id", updateCustomerByID)
	r.DELETE("/customers/:id", deteleCustomerByID)
	return r
}
