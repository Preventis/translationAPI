package routing

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"preventis.io/translationApi/model"
)

type simpleProjectDTO struct {
	Id           uint
	Name         string
	BaseLanguage model.Language
	Languages    []model.Language
}

func getAllActiveProjects(c *gin.Context) {
	var projects []model.Project
	db.Preload("BaseLanguage").Preload("Languages").Find(&projects)
	result := convertProjectListToDTO(projects, false)
	c.JSON(200, result)
}

func getAllArchivedProjects(c *gin.Context) {
	var projects []model.Project
	db.Find(&projects)
	result := convertProjectListToDTO(projects, true)
	c.JSON(200, result)
}

func convertProjectListToDTO(projects []model.Project, archived bool) []simpleProjectDTO {
	var result []simpleProjectDTO
	result = []simpleProjectDTO{}
	for _, e := range projects {
		if archived == e.Archived {
			var languages = e.Languages
			if languages == nil {
				languages = []model.Language{}
			}
			p := simpleProjectDTO{e.ID, e.Name, e.BaseLanguage, languages}
			result = append(result, p)
		}
	}
	return result
}

type projectDTO struct {
	Id           uint
	Name         string
	BaseLanguage model.Language
	Languages    []model.Language
	Identifiers  []identifierDTO
}

type identifierDTO struct {
	Id           uint
	Identifier   string
	Translations []translationDTO
}

type translationDTO struct {
	Id                uint
	Translation       string
	Language          string
	Approved          bool
	ImprovementNeeded bool
}

func getProject(c *gin.Context) {
	id := c.Param("id")
	var project model.Project
	if err := db.Where("id = ?", id).
		Preload("Languages").
		Preload("BaseLanguage").
		Preload("Identifiers").
		Preload("Identifiers.Translations").
		First(&project).
		Error; err != nil {
		c.AbortWithStatus(404)
		fmt.Println(err)
		return
	} else {
		c.JSON(200, projectToDTO(project))
	}
}

func projectToDTO(project model.Project) projectDTO {
	identifiers := []identifierDTO{}
	for _, e := range project.Identifiers {
		identifiers = append(identifiers, identifierToDTO(e))
	}
	return projectDTO{Id: project.ID, Name: project.Name, BaseLanguage: project.BaseLanguage, Languages: project.Languages, Identifiers: identifiers}
}

type projectValidation struct {
	Name    string `form:"name" json:"name" xml:"name"  binding:"required"`
	IsoCode string `form:"baseLanguageCode" json:"baseLanguageCode" xml:"baseLanguageCode"  binding:"required"`
}

func createProject(c *gin.Context) {
	var json projectValidation
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		println(err.Error())
		return
	}

	var projects []model.Project
	db.Where("Name = ?", json.Name).Find(&projects)
	if len(projects) > 0 {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Project with same name already exists"})
		fmt.Println("Project with same name already exists")
		return
	}

	var project model.Project
	project.Name = json.Name
	var baseLang model.Language
	if err := db.Where("iso_code = ?", json.IsoCode).First(&baseLang).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		fmt.Println(err)
		return
	}
	project.BaseLanguage = baseLang
	project.Languages = []model.Language{baseLang}
	project.Archived = false

	if dbc := db.Create(&project); dbc.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": dbc.Error.Error()})
		return
	}
	c.JSON(201, projectToDTO(project))
}

type projectRenameValidation struct {
	Name string `form:"name" json:"name" xml:"name"  binding:"required"`
}

func renameProject(c *gin.Context) {
	id := c.Param("id")
	var json projectRenameValidation
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var project model.Project
	if err := db.Where("id = ?", id).First(&project).Error; err != nil {
		c.AbortWithStatus(404)
		fmt.Println(err)
		return
	} else {
		project.Name = json.Name
		db.Save(&project)
		c.JSON(200, projectToDTO(project))
	}
}

func archiveProject(c *gin.Context) {
	id := c.Param("id")
	var project model.Project
	if err := db.Where("id = ?", id).First(&project).Error; err != nil {
		c.AbortWithStatus(404)
		fmt.Println(err)
		return
	} else {
		project.Archived = true
		db.Save(&project)
		c.JSON(200, projectToDTO(project))
	}
}

type projectLanguageValidation struct {
	IsoCode string `form:"languageCode" json:"languageCode" xml:"languageCode"  binding:"required"`
}

func addLanguageToProject(c *gin.Context) {
	id := c.Param("id")
	var json projectLanguageValidation
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var project model.Project
	if err := db.Where("id = ?", id).
		Preload("Languages").
		Preload("BaseLanguage").
		Preload("Identifiers").
		Preload("Identifiers.Translations").
		First(&project).Error; err != nil {
		c.AbortWithStatus(404)
		fmt.Println(err)
		return
	} else {
		if containsLanguage(json.IsoCode, project) {
			c.AbortWithStatusJSON(409, gin.H{"error": "Project already contains language"})
			fmt.Println("language already present in project")
			return
		}
		var lang model.Language
		if err := db.Where("iso_code = ?", json.IsoCode).First(&lang).Error; err != nil {
			c.AbortWithStatus(404)
			fmt.Println(err)
			return
		}
		db.Model(&project).Association("Languages").Append(lang)
		db.Save(&project)
		c.JSON(200, projectToDTO(project))
	}
}

func containsLanguage(lang string, project model.Project) bool {
	var containsLanguage = false
	for _, e := range project.Languages {
		if e.IsoCode == lang {
			containsLanguage = true
		}
	}
	return containsLanguage
}

func setBaseLanguage(c *gin.Context) {
	id := c.Param("id")
	var json projectLanguageValidation
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var project model.Project
	if err := db.Where("id = ?", id).First(&project).Error; err != nil {
		c.AbortWithStatus(404)
		fmt.Println(err)
		return
	} else {
		var baseLang model.Language
		if err := db.Where("iso_code = ?", json.IsoCode).First(&baseLang).Error; err != nil {
			c.AbortWithStatus(404)
			fmt.Println(err)
			return
		}
		project.BaseLanguage = baseLang

		var containsLang = false
		for _, lang := range project.Languages {
			if lang.IsoCode == baseLang.IsoCode {
				containsLang = true
			}
		}

		if !containsLang {
			project.Languages = append(project.Languages, baseLang)
		}

		db.Save(&project)
		c.JSON(200, projectToDTO(project))
	}
}
