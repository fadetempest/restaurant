package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"restaurant/buisness"
)

type Server struct {
	ctx context.Context
	Address string
	DatabaseUrl string
}

type Handler struct {
	r *mux.Router
	process *buisness.Processing
}

var SecretKey = []byte("281fna2jf2baa")

func NewServer(ctx context.Context, address string, database string) *Server{
	return &Server{
		ctx: ctx,
		Address: address,
		DatabaseUrl: database,
	}
}

func (s *Server) Run() error{
	db, err:= openDb(s.DatabaseUrl)
	if err != nil{
		log.Println("Error while opening DataBase")
	}
	defer db.Close()

	rp:= &Handler{
		r:    mux.NewRouter(),
		process: buisness.NewProcess(db),
	}


	rp.r.HandleFunc("/menu", rp.ReadDishes)
	rp.r.HandleFunc("/add", rp.isAutorized(rp.AddDish))
	rp.r.HandleFunc("/delete/{id}", rp.isAutorized(rp.DeleteDish))
	rp.r.HandleFunc("/update", rp.isAutorized(rp.UpdateDish))
	rp.r.HandleFunc("/signup", rp.SignUp)
	rp.r.HandleFunc("/signin", rp.SignIn)

	srv:=&http.Server{
		Addr: s.Address,
		Handler: rp.r,
	}

	log.Printf("Server is running on %s", s.Address)
	return srv.ListenAndServe()
}

func openDb(baseUrl string) (*sql.DB, error){
	db, err:=sql.Open("postgres", baseUrl)
	if err!=nil{
		return nil, err
	}
	if err:=db.Ping();err!=nil{
		return nil,err
	}
	return db,nil
}

func (h *Handler) AddDish(w http.ResponseWriter, r *http.Request){
	resp, err:= h.process.Add(r)
	if err!=nil{
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (h *Handler) DeleteDish(w http.ResponseWriter, r *http.Request){
	resp, err:=h.process.Delete(r)
	if err!=nil{
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (h *Handler) UpdateDish(w http.ResponseWriter, r *http.Request){
	resp, err:=h.process.Update(r)
	if err!=nil{
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (h *Handler) ReadDishes(w http.ResponseWriter, r *http.Request){
	resp, err:=h.process.ReadAll()
	if err!=nil{
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request){
	resp, err:=h.process.CreateUser(r)
	if err!=nil{
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request){
	resp,token,err:=h.process.Autorization(r)
	if err!=nil{
		w.WriteHeader(http.StatusForbidden)
		log.Println(err)
		return
	}
	if token != ""{
		w.Header().Set("Token", token)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (h *Handler) isAutorized(handler http.HandlerFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode("No token found")
			return
		}
		token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error in parsing")
			}
			return SecretKey, nil
		})

		if err != nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(err)
			return
		}
		if claims, ok:=token.Claims.(jwt.MapClaims); ok && token.Valid{
			if claims["role"]=="admin"{
				r.Header.Set("Role", "admin")
				handler.ServeHTTP(w, r)
				return
			}
		}
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("NOOO"))
	}
}
