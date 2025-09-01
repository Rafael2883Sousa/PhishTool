package models

import "github.com/jinzhu/gorm"

func GetDB() *gorm.DB { return db } // usa a instância global já inicializada em models
