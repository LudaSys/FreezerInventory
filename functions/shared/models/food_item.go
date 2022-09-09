package models

type FoodItem struct {
	ItemId          string `json:"itemId"`
	Name            string `json:"name"`
	StorageLocation string `json:"storageLocation"`
}
