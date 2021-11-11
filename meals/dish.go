package meals

type Dish struct {
	ID int `json:"id"`
	Description string `json:"description"`
	Composition string `json:"composition"`
	Price int32 `json:"price"`
}
