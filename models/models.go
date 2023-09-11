package models

type AddressBook struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Mobile   string `json:"mobile"`
	Address   string `json:"address"`
	ImagePath   string `json:"image_path"`
}
