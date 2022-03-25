package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/FAdemoglu/homeworkfourtwo/helper"
	"github.com/FAdemoglu/homeworkfourtwo/infrastructure"
	"github.com/FAdemoglu/homeworkfourtwo/internal/domain/author"
	"github.com/FAdemoglu/homeworkfourtwo/internal/domain/book"
	"github.com/gorilla/mux"
)

var (
	bookRepository   *book.BookRepository
	authorRepository *author.AuthorRepository
)

func init() {
	db := infrastructure.NewMySQLDB("root:Furkan7937.@tcp(127.0.0.1:3306)/authorandbook?parseTime=true&loc=Local")
	bookRepository = book.NewBookRepository(db)
	authorRepository = author.NewAuthorRepository(db)
	bookRepository.Migration()
	authorRepository.Migration()
	bookRepository.InsertSampleData()
	authorRepository.InsertSampleData()
	books, _ := helper.ReadCsvToBookSlice("../resources/books.csv")
	bookRepository.InsertCsvDatas(books)
	authors, _ := helper.ReadCsvToAuthorSlice("../resources/authors.csv")
	authorRepository.InsertCsvDatas(authors)
	bookRepository.InsertSampleData()
}
func main() {
	r := mux.NewRouter()

	r.Use(loggingMiddleware)
	r.Use(authenticationMiddleware)
	r.HandleFunc("/search", SearchWithQueryParam)
	r.HandleFunc("/booksandauthors", GetAllListWithAuthor)
	s := r.PathPrefix("/books").Subrouter()
	s.HandleFunc("", GetAllList).Methods("GET")
	s.HandleFunc("/remove/{id:[0-9]+}", DeleteBook).Methods("DELETE")
	s.HandleFunc("/create", CreateBook).Methods("POST")
	s.HandleFunc("/buy", BuyBook).Methods("PUT")

	srv := &http.Server{
		Addr:         "0.0.0.0:8090",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	ShutdownServer(srv, time.Second*10)

}

func BuyBook(w http.ResponseWriter, r *http.Request) {
	Id := r.FormValue("Id")
	Quantity := r.FormValue("quantity")
	id, _ := strconv.Atoi(Id)
	quantity, _ := strconv.Atoi(Quantity)
	if quantity < 0 || id < 0 {
		http.Error(w, "Id veya miktar 0'dan küçük olamaz", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	err := bookRepository.Update(id, quantity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte("Kitap satma işlemi başarılı"))
}

func SearchWithQueryParam(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query().Get("search")
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "GET" {
		http.Error(w, "You can use GET method only", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	books := bookRepository.SearchByAuthorAndBookName(param)
	resp, _ := json.Marshal(books)
	w.Write(resp)
}

func CreateBook(w http.ResponseWriter, r *http.Request) {
	var b book.Book
	err := helper.DecodeJSONBody(w, r, &b)
	if err != nil {
		var mr *helper.ErrorRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	bookRepository.Create(b)
	w.Write([]byte("Kitap başarıyla eklendi"))
}

func GetAllList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "GET" {
		http.Error(w, "You can use GET method only", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	books := bookRepository.FindAll()
	resp, _ := json.Marshal(books)
	w.Write(resp)
}

func GetAllListWithAuthor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "GET" {
		http.Error(w, "You can use GET method only", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	books, _ := bookRepository.GetAllBooksWithAuthorInformation()
	resp, _ := json.Marshal(books)
	w.Write(resp)
}

func DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "DELETE" {
		http.Error(w, "You can use DELETE method only", http.StatusBadRequest)
		return
	}
	d := vars["id"]
	Id, _ := strconv.Atoi(d)
	if Id < 0 {
		http.Error(w, "Id can not be lower than 0", http.StatusBadRequest)
		return
	}
	errorNotFound := bookRepository.DeleteById(Id)
	if errorNotFound != nil {
		http.Error(w, "Bu id ile kitap bulunamadı", http.StatusNotFound)
		return
	}

	resp, _ := json.Marshal(d)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Query())
		next.ServeHTTP(w, r)
	})
}

func authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if strings.HasSuffix(r.URL.Path, "/create") {
			if token == "dXNlckBleGFtcGxlLmNvbTpzZWNyZXQ=" {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, "Token not found", http.StatusUnauthorized)
			}
		} else {
			next.ServeHTTP(w, r)
		}

	})
}

func ShutdownServer(srv *http.Server, timeout time.Duration) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}
