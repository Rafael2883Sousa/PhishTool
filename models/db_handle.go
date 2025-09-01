package models

import "github.com/jinzhu/gorm"

// Expose the global *gorm.DB from this package (GORM v1).
func DB() *gorm.DB { return db }
