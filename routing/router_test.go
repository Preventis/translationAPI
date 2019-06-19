package routing

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"preventis.io/translationApi/model"
)

func setupTestEnvironment() *gin.Engine {
	db = model.StartDB("sqlite3", ":memory:")

	eng := model.Language{IsoCode: "en", Name: "English"}
	ger := model.Language{IsoCode: "de", Name: "German"}
	es := model.Language{IsoCode: "es", Name: "Spanish"}
	db.Create(&eng)
	db.Create(&ger)
	db.Create(&es)

	proj1 := model.Project{Name: "Shared", BaseLanguage: eng, Languages: []model.Language{ger, eng}}
	proj2 := model.Project{Name: "Base", BaseLanguage: ger, Languages: []model.Language{ger}}
	archivedProj := model.Project{Name: "Archived", BaseLanguage: ger, Archived: true}
	db.Create(&proj1)
	db.Create(&proj2)
	db.Create(&archivedProj)

	key1 := model.StringIdentifier{Identifier: "key1", Project: proj1}
	key2 := model.StringIdentifier{Identifier: "key2", Project: proj1}
	key3 := model.StringIdentifier{Identifier: "key2", Project: proj2}
	db.Create(&key1)
	db.Create(&key2)
	db.Create(&key3)

	translation1 := model.Translation{Translation: "translation1", Identifier: key1, Language: ger}
	translation2 := model.Translation{Translation: "\"translation2\"", Identifier: key2, Language: ger}
	translation3 := model.Translation{Translation: "translation2", Identifier: key3, Language: ger, Approved: true}

	db.Create(&translation1)
	db.Create(&translation2)
	db.Create(&translation3)

	revision1 := model.Revision{Translation: translation1, RevisionTranslation: translation1.Translation, Approved: translation1.Approved}
	revision2 := model.Revision{Translation: translation2, RevisionTranslation: translation2.Translation, Approved: translation2.Approved}
	revision3 := model.Revision{Translation: translation3, RevisionTranslation: translation3.Translation, Approved: translation3.Approved}

	db.Create(&revision1)
	db.Create(&revision2)
	db.Create(&revision3)

	saltedBytes := []byte("password")
	hashedBytes, _ := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)

	hash := string(hashedBytes[:])

	admin := model.User{Username: "admin1", Password: hash, Mail: "admin1@example.com", Admin: true}
	user := model.User{Username: "user1", Password: hash, Mail: "user1@example.com", Admin: false}

	db.Create(&admin)
	db.Create(&user)

	router := setupRouter()
	return router
}
