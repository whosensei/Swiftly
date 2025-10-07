package store

var user_store = make(map[string]string)

func Add_mapping(shorturl string, longurl string){
	user_store[shorturl] = longurl
}

func Redirect(shorturl string) string{
	val := user_store[shorturl]
	return val
}

func Getallmaps() map[string]string {
	return user_store
}

