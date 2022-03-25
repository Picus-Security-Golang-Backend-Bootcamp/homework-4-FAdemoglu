package helper

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/FAdemoglu/homeworkfourtwo/internal/domain/author"
	"github.com/FAdemoglu/homeworkfourtwo/internal/domain/book"
)

func ReadCsvToBookSlice(fileName string) ([]book.Book, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer f.Close() //Başına defer koyulan yer fonksiyon bitince en son satırdır yani burada bütün işlemlerimiz bitince okuma kısmını kapatmış olacağız.

	reader := csv.NewReader(f)

	lines, err := reader.ReadAll()

	if err != nil {
		return nil, err
	}

	var result []book.Book

	for _, line := range lines[1:] {
		stockCode, _ := strconv.Atoi(line[1])
		isbnNumber, _ := strconv.Atoi(line[2])
		pageNumber, _ := strconv.Atoi(line[3])
		price, _ := strconv.Atoi(line[4])
		stockCount, _ := strconv.Atoi(line[5])
		isDeleted, _ := strconv.ParseBool(line[7])
		AuthorId, _ := strconv.Atoi(line[6])
		data := book.Book{
			BookName:   line[0],
			StockCode:  stockCode,
			ISBNNumber: isbnNumber,
			PageNumber: pageNumber,
			Price:      price,
			StockCount: stockCount,
			AuthorID:   AuthorId,
			IsDeleted:  isDeleted,
		}

		result = append(result, data)
	}

	return result, nil

}

func ReadCsvToAuthorSlice(fileName string) ([]author.Author, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer f.Close() //Başına defer koyulan yer fonksiyon bitince en son satırdır yani burada bütün işlemlerimiz bitince okuma kısmını kapatmış olacağız.

	reader := csv.NewReader(f)

	lines, err := reader.ReadAll()

	if err != nil {
		return nil, err
	}

	var result []author.Author

	for _, line := range lines[1:] {
		Id, _ := strconv.Atoi(line[0])
		data := author.Author{
			AuthorId: Id,
			Name:     line[1],
		}

		result = append(result, data)
	}

	return result, nil
}

type ErrorRequest struct {
	Status int
	Msg    string
}

func (er *ErrorRequest) Error() string {
	return er.Msg
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "" {
		value := r.Header.Get("Content-Type")
		if value == "" || value != "application/json" {
			msg := "Content-Type header is not application/json"
			return &ErrorRequest{Status: http.StatusUnsupportedMediaType, Msg: msg}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &ErrorRequest{Status: http.StatusBadRequest, Msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &ErrorRequest{Status: http.StatusBadRequest, Msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &ErrorRequest{Status: http.StatusBadRequest, Msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &ErrorRequest{Status: http.StatusBadRequest, Msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &ErrorRequest{Status: http.StatusBadRequest, Msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &ErrorRequest{Status: http.StatusRequestEntityTooLarge, Msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return &ErrorRequest{Status: http.StatusBadRequest, Msg: msg}
	}

	return nil
}
