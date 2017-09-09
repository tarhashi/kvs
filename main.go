package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

func Key(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["key"]
}

func main() {
	router := mux.NewRouter()
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	bucketName := []byte("test")

	// CREATE
	router.HandleFunc("/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := []byte(Key(r))
		err := db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists(bucketName)
			if err != nil {
				return err
			}
			bufbody := new(bytes.Buffer)
			bufbody.ReadFrom(r.Body)
			body := bufbody.Bytes()
			err2 := b.Put(key, body)
			if err2 != nil {
				return err2
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(500)
		} else {
			w.WriteHeader(201)
		}
	}).Methods("POST", "PUT")
	// GET
	router.HandleFunc("/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := []byte(Key(r))
		var val []byte
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketName)
			val = b.Get(key)
			if val == nil {
				w.WriteHeader(404)
			} else {
				w.WriteHeader(200)
				fmt.Fprintf(w, string(val))
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(500)
		}
	}).Methods("GET")
	// DELETE
	router.HandleFunc("/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := []byte(Key(r))
		err := db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketName)
			if b == nil {
				w.WriteHeader(404)
				return nil
			}
			val := b.Get(key)
			if val == nil {
				w.WriteHeader(404)
				return nil
			}
			err := b.Delete(key)
			if err != nil {
				return err
			}
			w.WriteHeader(204)
			return nil
		})
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(500)
		}
	}).Methods("DELETE")
	http.ListenAndServe(":8080", router)
}
