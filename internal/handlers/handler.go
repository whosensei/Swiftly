package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github/whosensei/shortenn/internal/auth"
	"github/whosensei/shortenn/internal/database"
	"github/whosensei/shortenn/internal/model"
	"github/whosensei/shortenn/internal/utils"
	"net/http"
	"os"
)

type UserHandler struct {
	DB *sql.DB
}

func (h *UserHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {

	var u model.User_request

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Failed to get the body", http.StatusBadRequest)
		return
	}

	id := utils.GenerateId()

	if userID == "" {
		h.AnonymousShorten(w,r,u.Long_url,id)
	}else{
		h.AuthenticatedShorten(w,r,u.Long_url,id)
	}

	// Short_url := utils.Url_shorten(id, u.Long_url)
	// data := model.URL{Id: id, Long_url: u.Long_url, Short_url: Short_url}

	// if err := database.URL_Add(h.DB, data); err != nil {
	// 	fmt.Println("Failed to add to database")
	// }
	// redirect_link := fmt.Sprintf("%s/%s","https://swftly.dev",Short_url)

	// w.Header().Set("content-type", "application/json")
	// w.WriteHeader(http.StatusCreated)
	// json.NewEncoder(w).Encode(model.Api_response{
	// 	Success: true,
	// 	Message: "task executed successfully",
	// 	Data:    redirect_link,
	// })
}

func (h *UserHandler) AnonymousShorten(w http.ResponseWriter , r* http.Request , longurl string , id string){

	token := r.Header.Get("Anonymous_Token");
	//check ratelimits

	Short_url := utils.Url_shorten(id,longurl)

	//add to database, Short_url,long_url,token,expires_at,created_at

	baseurl := os.Getenv("baseurl");
	if baseurl == ""{
		baseurl = "https://localhost:8080"
	}

	response := model.ShortenResponse{
		Data: fmt.Sprintf("%s/%s",baseurl,Short_url),
		Anonymous_Token: token,
		//add rest
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func(h *UserHandler) AuthenticatedShorten(w http.ResponseWriter ,r *http.Request , longurl string, id string){
	userID := auth.GetUserId(r)

	//using userID find uuid for user,
	// add links for that uuid in url table;

	Short_url := utils.Url_shorten(id,longurl)

	baseurl := os.Getenv("baseurl");
	if baseurl == ""{
		baseurl = "https://localhost:8080"
	}

	response := model.ShortenResponse{
		Data: fmt.Sprintf("%s/%s",baseurl,Short_url),
		//add rest
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	
}

func (h *UserHandler) Redirect_to_website(w http.ResponseWriter, r *http.Request) {

	shorturl := r.PathValue("shorturl")
	longurl := database.Redirect(h.DB, shorturl)

	fmt.Println(longurl)
	if longurl == "" {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, longurl, http.StatusFound)
}

func (h *UserHandler) Gettallmaps(w http.ResponseWriter, r *http.Request) {
	data := database.Getallurls(h.DB)
	w.Header().Set("content-type", "application-json")
	json.NewEncoder(w).Encode(data)
}
