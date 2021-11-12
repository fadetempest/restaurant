package repository

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	meals2 "restaurant/pkg/meals"
)

type DishRepo struct {
	Db *sql.DB
}

func NewDishRepo(db *sql.DB) *DishRepo {
	return &DishRepo{Db: db}
}

func (base *DishRepo) AddNewValue(dish *meals2.Dish) (string, error) {
	searchQuery := `SELECT id FROM meals WHERE id=$1`

	if base.Db.QueryRow(searchQuery, dish.ID).Scan(&dish.ID) == sql.ErrNoRows {
		insertValues := `INSERT INTO meals (id, description, composition, price) VALUES ($1, $2, $3, $4)`
		_, er := base.Db.Exec(insertValues, dish.ID, dish.Description, dish.Composition, dish.Price)
		if er != nil {
			return "", er
		}
		return "Successfully added to the menu", nil
	}
	return fmt.Sprintf("Dish with id=%d already exist", dish.ID), nil
}

func (base *DishRepo) DeleteValue(id string) (string, error) {
	delValue := `DELETE FROM meals WHERE id=$1`
	_, er := base.Db.Exec(delValue, id)
	if er != nil {
		return "", er
	}
	return "Successfully deleted", nil
}

func (base *DishRepo) UpdateValue(dish *meals2.Dish) (string, error) {
	updValue := `UPDATE meals SET description=$1, composition=$2, price=$3 WHERE id=$4`
	_, er := base.Db.Exec(updValue, dish.Description, dish.Composition, dish.Price, dish.ID)
	if er != nil {
		return "", er
	}
	return fmt.Sprintf("Successfully updated dish #%d", dish.ID), nil
}

func (base *DishRepo) GetMenu() ([]meals2.Dish, error) {
	rows, er := base.Db.Query("SELECT * FROM meals ORDER BY id")
	if er != nil {
		log.Fatal("DB operation error")
	}
	defer rows.Close()

	var dishes []meals2.Dish

	for rows.Next() {
		var dish meals2.Dish
		if scanEr := rows.Scan(&dish.ID, &dish.Description, &dish.Composition, &dish.Price); scanEr != nil {
			return dishes, scanEr
		}
		dishes = append(dishes, dish)
	}
	if rowErr := rows.Err(); rowErr != nil {
		return dishes, rowErr
	}
	return dishes, nil
}

func (base *DishRepo) AddNewUser(newUser meals2.User) (string, error) {
	searchQuery := `SELECT login FROM users WHERE login=$1`

	if base.Db.QueryRow(searchQuery, newUser.Login).Scan(&newUser.Login) == sql.ErrNoRows {
		if newUser.Role == "" {
			insertWithoutRole := `INSERT INTO users (login, password) VALUES ($1, $2)`
			_, insEr := base.Db.Exec(insertWithoutRole, newUser.Login, newUser.Password)
			if insEr != nil {
				return "", insEr
			}
			return fmt.Sprintf("Welcome to API %v", newUser.Login), nil
		}
		insertValues := `INSERT INTO users (login,password,role) VALUES ($1, $2, $3)`
		_, er := base.Db.Exec(insertValues, newUser.Login, newUser.Password, newUser.Role)
		if er != nil {
			return "", er
		}
		return fmt.Sprintf("Welcome to API %v", newUser.Login), nil
	}
	return "User with this login already exist", nil
}

func (base *DishRepo) GetUserFromBase(user meals2.User) (string, string, error) {
	searchQuery := `SELECT login, password FROM users WHERE login=$1 AND password=$2`

	if base.Db.QueryRow(searchQuery, user.Login, user.Password).Scan(&user.Login, &user.Password) == sql.ErrNoRows {
		return "User with this Login or Password not found", "", nil
	}
	rows, er := base.Db.Query("SELECT login,role FROM users WHERE login=$1 AND password=$2", user.Login, user.Password)
	if er != nil {
		log.Fatal("DB operation error")
	}
	defer rows.Close()

	type foundedUser struct {
		login string
		role  string
	}

	found := foundedUser{}
	for rows.Next() {
		if scanEr := rows.Scan(&found.login, &found.role); scanEr != nil {
			return "", "", scanEr
		}
	}
	if found.role == "admin" {
		return "Hello admin", found.role, nil
	}
	return "Successfully signIn", "", nil
}
