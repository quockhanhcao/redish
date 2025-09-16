package data_structure

var dataStore = make(map[string]string)

func AddToSet(key, value string) {
	dataStore[key] = value
}

func GetFromSet(key string) string {
	return dataStore[key]
}
