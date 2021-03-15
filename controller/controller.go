package controller

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"example.com/hungry-server/config/db"
	"example.com/hungry-server/model"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

var secret_key = os.Getenv("HungryTestSecretKey")

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user model.User
	var response model.ResponseResult
	var result model.User

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &user)

	if err != nil {
		response.Error = err.Error()
		json.NewEncoder(w).Encode(response)
		return
	}

	collection, err := db.GetDBCollection()

	if err != nil {
		response.Error = err.Error()
		json.NewEncoder(w).Encode(response)
		return
	}

	err = collection.FindOne(context.TODO(), bson.D{{Key: "username", Value: user.Username}}).Decode(&result)

	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 5)

			if err != nil {
				response.Error = "Error while hashing password. Try again!"
				json.NewEncoder(w).Encode(response)
				return
			}

			user.Password = string(hash)
			_, err = collection.InsertOne(context.TODO(), user)

			if err != nil {
				response.Error = "Error while creating user. Try again!"
				json.NewEncoder(w).Encode(response)
				return
			}

			response.Result = "Registration successful!"
			json.NewEncoder(w).Encode(response)
			return
		}

		response.Error = err.Error()
		json.NewEncoder(w).Encode(response)
		return
	}

	response.Error = "This username already exists!"
	json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user model.User
	var response model.ResponseResult
	var result model.User

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &user)

	if err != nil {
		response.Error = err.Error()
		json.NewEncoder(w).Encode(response)
		return
	}

	collection, err := db.GetDBCollection()

	if err != nil {
		response.Error = err.Error()
		json.NewEncoder(w).Encode(response)
		return
	}

	err = collection.FindOne(context.TODO(), bson.D{{Key: "username", Value: user.Username}}).Decode(&result)

	if err != nil {
		response.Error = "Invalid username or password. Please try again!"
		json.NewEncoder(w).Encode(response)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))

	if err != nil {
		response.Error = "Invalid username or password. Please try again!"
		json.NewEncoder(w).Encode(response)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": result.Username,
	})

	tokenString, err := token.SignedString([]byte(secret_key))

	if err != nil {
		response.Error = "There was an error generating the token. Please try again!"
		json.NewEncoder(w).Encode(response)
		return
	}

	result.Token = tokenString
	result.Password = ""

	json.NewEncoder(w).Encode(result)
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response model.ResponseResult
	var result model.User

	tokenString := r.Header.Get("Authorization")

	if tokenString == "" {
		response.Error = "There is no authorization token!"
		json.NewEncoder(w).Encode(response)
		return
	}

	tokenSplit := strings.Split(tokenString, " ")[1]

	token, err := jwt.Parse(tokenSplit, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret_key), nil
	})

	if err != nil {
		response.Error = err.Error()
		json.NewEncoder(w).Encode(response)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok && token.Valid {
		result.Username = claims["username"].(string)
		result.Token = tokenSplit

		json.NewEncoder(w).Encode(result)
		return
	} else {
		response.Error = err.Error()
		json.NewEncoder(w).Encode(response)
		return
	}
}
