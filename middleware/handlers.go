package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-postgres/models" 
	"log"
	"net/http" 
	"github.com/google/uuid"

	"os"      
	"io"
	"strings"
	"mime"
	"path/filepath"
	"strconv" 

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type response struct {
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

func createConnection() *sql.DB {
	db, err := sql.Open("postgres", "user=postgres2 dbname=postgres host=database-2.ctoftsuzye6s.ap-northeast-1.rds.amazonaws.com sslmode=require password=Admin123 sslrootcert=certs/ap-northeast-1-bundle.pem")
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")
	return db
}

func CreateAddress(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(10 << 20)
    if err != nil {
        http.Error(w, "Unable to parse form", http.StatusBadRequest)
        return
    }
	defer r.Body.Close()
	var addressBook models.AddressBook
    file, handler, err := r.FormFile("image_path")
    if err != nil {
        http.Error(w, "Unable to get file from form", http.StatusBadRequest)
        return
    }
    defer file.Close()
    newFileName := uuid.New().String() + filepath.Ext(handler.Filename)
    imagePath := filepath.Join("uploads", newFileName)
    destination, err := os.Create(imagePath)

    if err != nil {
        http.Error(w, "Unable to create file", http.StatusInternalServerError)
        return
    }
    defer destination.Close()

    _, err = io.Copy(destination, file)

    if err != nil {
        http.Error(w, "Unable to copy file", http.StatusInternalServerError)
        return
    }
    addressBook.ImagePath = imagePath
	addressBook.Name = r.FormValue("name")
	addressBook.Mobile = r.FormValue("mobile")
	addressBook.Address = r.FormValue("address")

    insertID := insertAddress(addressBook)

    res := response{
        ID:      insertID,
        Message: "Address Book created successfully",
    }
    json.NewEncoder(w).Encode(res)
}

func GetAddressbyId(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}
	addressBook, err := getAddressBookbyId(int64(id))

	if err != nil {
		log.Fatalf("Unable to get user. %v", err)
	}
	json.NewEncoder(w).Encode(addressBook)
}

func GetAlladdressBook(w http.ResponseWriter, r *http.Request) {

	addressBooks, err := getAllAddressBooks()
	if err != nil {
		log.Fatalf("Unable to get all address. %v", err)
	}
	json.NewEncoder(w).Encode(addressBooks)
}

func UpdateAddress(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
        http.Error(w, "Unable to parse form", http.StatusBadRequest)
        return
    }
	defer r.Body.Close()
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}

	var addressBook models.AddressBook

	file, handler, err := r.FormFile("image_path")
    if err != nil {
        http.Error(w, "Unable to get file from form", http.StatusBadRequest)
        return
    }
    defer file.Close()

    newFileName := uuid.New().String() + filepath.Ext(handler.Filename)
    imagePath := filepath.Join("uploads", newFileName)
    destination, err := os.Create(imagePath)

    if err != nil {
        http.Error(w, "Unable to create file", http.StatusInternalServerError)
        return
    }
    defer destination.Close()
    _, err = io.Copy(destination, file)
    if err != nil {
        http.Error(w, "Unable to copy file", http.StatusInternalServerError)
        return
    }
	addressBook.ImagePath = imagePath
	addressBook.Name = r.FormValue("name")
	addressBook.Mobile = r.FormValue("mobile")
	addressBook.Address = r.FormValue("address")
	if err != nil {
		log.Fatalf("Unable to decode the request body.  %v", err)
	}
	updatedRows := updateaddress(int64(id), addressBook)

	msg := fmt.Sprintf("Address updated successfully. Total rows/record affected %v", updatedRows)
	res := response{
		ID:      int64(id),
		Message: msg,
	}
	json.NewEncoder(w).Encode(res)
}

func DeleteAddress(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}
	deletedRows := deleteaddress(int64(id))
	msg := fmt.Sprintf("Address deleted successfully. Total rows/record affected %v", deletedRows)
	res := response{
		ID:      int64(id),
		Message: msg,
	}
	json.NewEncoder(w).Encode(res)
}

//------------------------- handler functions ----------------

func insertAddress(addressBook models.AddressBook) int64 {
	db := createConnection()
	defer db.Close()
	sqlStatement := `INSERT INTO addressbook (name, mobile, address, image_path) VALUES ($1, $2, $3, $4) RETURNING id`

	var id int64
	err := db.QueryRow(sqlStatement, addressBook.Name, addressBook.Mobile, addressBook.Address, addressBook.ImagePath).Scan(&addressBook.ID)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}
	fmt.Printf("Inserted a single record %v", id)
	return id
}

func getAddressBookbyId(id int64) (models.AddressBook, error) {
	db := createConnection()
	defer db.Close()
	var addressBook models.AddressBook
	sqlStatement := `SELECT * FROM addressbook WHERE id=$1`
	row := db.QueryRow(sqlStatement, id)
	err := row.Scan(&addressBook.ID, &addressBook.Name, &addressBook.Mobile, &addressBook.Address, &addressBook.ImagePath)
	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return addressBook, nil
	case nil:
		return addressBook, nil
	default:
		log.Fatalf("Unable to scan the row. %v", err)
	}
	return addressBook, err
}

func getAllAddressBooks() ([]models.AddressBook, error) {
	db := createConnection()
	defer db.Close()
	var addressBooks []models.AddressBook
	sqlStatement := `SELECT * FROM addressbook`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var addressBook models.AddressBook
		err = rows.Scan(&addressBook.ID, &addressBook.Name, &addressBook.Mobile, &addressBook.Address, &addressBook.ImagePath)
		if err != nil {
			log.Fatalf("Unable to scan the row. %v", err)
		}
		addressBooks = append(addressBooks, addressBook)
	}

	// return empty user on error
	return addressBooks, err
}

func updateaddress(id int64, addressBook models.AddressBook) int64 {
	db := createConnection()
	defer db.Close()
	sqlStatement := `UPDATE addressbook SET name=$2, mobile=$3, address=$4, image_path=$5 WHERE id=$1`
	res, err := db.Exec(sqlStatement, id, addressBook.Name, addressBook.Mobile, addressBook.Address, addressBook.ImagePath)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}
	fmt.Printf("Total rows/record affected %v", rowsAffected)
	return rowsAffected
}

func deleteaddress(id int64) int64 {
	db := createConnection()
	defer db.Close()
	sqlStatement := `DELETE FROM addressbook WHERE id=$1`
	res, err := db.Exec(sqlStatement, id)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}
	fmt.Printf("Total rows/record affected %v", rowsAffected)
	return rowsAffected
}

func ServeImage(w http.ResponseWriter, r *http.Request) {
    imagePath := strings.TrimPrefix(r.URL.Path, "/images/")
	filePath := "" + imagePath
    imageFile, err := os.Open(filePath)
    if err != nil {
        http.Error(w, "Image not found", http.StatusNotFound)
        return
    }
    defer imageFile.Close()
    contentType := mime.TypeByExtension(filepath.Ext(filePath))
    if contentType == "" {
        contentType = "application/octet-stream"
    }
    w.Header().Set("Content-Type", contentType)
    _, err = io.Copy(w, imageFile)
    if err != nil {
        http.Error(w, "Error serving image", http.StatusInternalServerError)
    }
}
