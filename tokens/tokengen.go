package tokens

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/kravi0/BizGrowth-backend/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email    string
	MobileNo string
	Name     string
	Uid      string
	jwt.StandardClaims
}

var SellerData *mongo.Collection = database.UserData(database.Client, "Seller")
var SECRET_KEY = os.Getenv("SECRET_KRY")

func TokenGenerator(email string, MobileNo string, Name string, uid string) (signedToken string, signedRefreshToken string, err error) {

	claims := &SignedDetails{
		Email:    email,
		MobileNo: MobileNo,
		Name:     Name,
		Uid:      uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}
	refreshclaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}
	refreshtoken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshclaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panicln(err)
		return
	}
	return token, refreshtoken, err

}
func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "the token was invalid"
		return
	}

	if claims.StandardClaims.ExpiresAt < time.Now().Local().Unix() {
		msg = "token is already expired"
		return
	}

	return claims, msg

}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userid string) {

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})
	updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updatedat", Value: updated_at})

	upsert := true
	filter := bson.M{"user_id": userid}

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := SellerData.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updateObj}}, &opt)
	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}

}
