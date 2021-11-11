package buisness

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"io/ioutil"
	"net/http"
	"restaurant/meals"
	"restaurant/repository"
	"time"
)

type Processing struct {
	repo *repository.DishRepo
}

var SecretKey = []byte("281fna2jf2baa")

func NewProcess(db *sql.DB) *Processing{
	return &Processing{repo: repository.NewDishRepo(db)}
}

func generateJWT(role string) (string,error){
	token:=jwt.New(jwt.SigningMethodHS256)
	claims:=token.Claims.(jwt.MapClaims)

	fmt.Println("Yr role is",role)
	claims["authorized"] = true
	claims["role"] = role
	claims["exp"] = time.Now().Add(30 * time.Minute).Unix()

	signedToken, err := token.SignedString(SecretKey)
	if err != nil{
		return "", err
	}
	return signedToken, nil
}

func (p *Processing) isAutorized(handler http.HandlerFunc) http.HandlerFunc{
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
	}
}

func (p *Processing) Add(r *http.Request) ([]byte,error){
	data, readErr:=ioutil.ReadAll(r.Body)
	if readErr != nil{
		return nil, readErr
	}
	var dish meals.Dish
	err:=json.Unmarshal(data,&dish)
	if err!=nil{
		return nil,err
	}
	dbAnsw, dbErr:= p.repo.AddNewValue(&dish)
	if dbErr!=nil{
		return nil, dbErr
	}
	coded, jerr:= json.Marshal(dbAnsw)
	if jerr!=nil{
		return nil,jerr
	}
	return coded,nil
}

func (p *Processing) Delete(r *http.Request) ([]byte, error){
	dbAnsw, dbErr:=p.repo.DeleteValue(r.URL.Path[8:])
	if dbErr!=nil{
		return nil, dbErr
	}
	coded, jerr:= json.Marshal(dbAnsw)
	if jerr!=nil{
		return nil, jerr
	}
	return coded, nil
}

func (p *Processing) Update(r *http.Request) ([]byte, error){
	data, readErr:=ioutil.ReadAll(r.Body)
	if readErr != nil{
		return nil,readErr
	}
	var dish meals.Dish
	err:=json.Unmarshal(data,&dish)
	if err!=nil{
		return nil,err
	}
	dbAnsw, dbErr:=p.repo.UpdateValue(&dish)
	if dbErr!=nil{
		return nil,dbErr
	}
	coded, jerr:= json.Marshal(dbAnsw)
	if jerr!=nil{
		return nil, jerr
	}
	return coded,nil
}

func (p *Processing) ReadAll() ([]byte, error){
	allMenu, err:=p.repo.GetMenu()
	if err != nil{
		return nil, err
	}
	coded, jerr:= json.Marshal(allMenu)
	if jerr!=nil{
		return nil, jerr
	}
	return coded, nil
}

func (p *Processing) CreateUser(r *http.Request) ([]byte, error){
	var newUser meals.User
	err:= json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil{
		return nil, err
	}

	dbAnsw, dbErr:=p.repo.AddNewUser(newUser)
	if dbErr!=nil{
		return nil,dbErr
	}

	coded, jerr:= json.Marshal(dbAnsw)
	if jerr!=nil{
		return nil, jerr
	}
	return coded,nil
}

func (p *Processing) Autorization(r *http.Request) ([]byte,string,error){
	var user meals.User
	err:=json.NewDecoder(r.Body).Decode(&user)
	if err!=nil{
		return nil,"", err
	}

	dbAnsw, isAdmin, dbErr:=p.repo.GetUserFromBase(user)
	if dbErr!=nil{
		return nil,"",dbErr
	}
	if isAdmin == "admin"{
		token, tokenErr:=generateJWT(isAdmin)
		if tokenErr!= nil{
			return nil,"", tokenErr
		}
		coded, jerr:=json.Marshal(dbAnsw)
		if jerr!=nil{
			return nil,"", jerr
		}
		return coded, token,nil
	}
	coded, jerr:=json.Marshal(dbAnsw)
	if jerr!=nil{
		return nil,"", jerr
	}
	return coded,"",nil
}