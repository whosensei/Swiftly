package store

import "github/whosensei/shortenn/internal/model"

var user_store = make(map[string][]model.Short_url)

func Add_mapping(id string, u model.Short_url) bool{
	user_store[id] = append(user_store[id], u)
	return true
}

func Get_all_urls_by_id(id string) []model.Short_url{
	return user_store[id]
}

// func Redirect(id string,shorturl string) string {
// 	for i := range user_store{
// 		if i == id{
// 			for _,val := range user_store[id]{
// 				if val.Short_url == shorturl {
// 					return val.Long_url
// 				}
// 			}
// 		}
// 	}
// 	return "No suchh url exists"
// }

