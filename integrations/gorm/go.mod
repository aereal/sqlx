module github.com/aereal/sqlx/integrations/gorm

go 1.26.1

require (
	github.com/aereal/sqlx v1.14.0
	gorm.io/driver/sqlite v1.5.6
	gorm.io/gorm v1.25.10
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
)

replace github.com/aereal/sqlx => ../../
