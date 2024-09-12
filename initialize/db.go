package initialize

import (
	"fmt"
	"go-tiny/model"
	"log"
	"strconv"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() *gorm.DB {
	dsn := getDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	RegisterTables(db)
	return db
}

// RegisterTables 注册数据库表
func RegisterTables(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.User{},
		&model.Blog{},
		&model.FileUploadAndDownload{},
	)
	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}
}

// getDSN 获取数据源名称
func getDSN() string {
	parseTimeStr := strconv.FormatBool(viper.GetBool("database.parseTime"))
	locStr := viper.GetString("database.loc")

	fmt.Println("------------------------------")
	fmt.Println(viper.GetString("database.username"))
	fmt.Println(viper.GetString("database.password"))
	fmt.Println(viper.GetString("database.host"))
	fmt.Println(viper.GetString("database.port"))
	fmt.Println(viper.GetString("database.dbname"))
	fmt.Println("------------------------------")

	return viper.GetString("database.username") + ":" +
		viper.GetString("database.password") + "@tcp(" +
		viper.GetString("database.host") + ":" +
		viper.GetString("database.port") + ")/" +
		viper.GetString("database.dbname") + "?" +
		"charset=" + viper.GetString("database.charset") + "&parseTime=" +
		parseTimeStr + "&loc=" + locStr
}
