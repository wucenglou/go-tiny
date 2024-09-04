package initialize

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigName("config") // 设置配置文件名称
	viper.AddConfigPath(".")      // 设置配置文件所在目录
	viper.SetConfigType("yaml")   // 设置配置文件类型

	err := viper.ReadInConfig() // 查找并读取配置文件
	if err != nil {             // 处理读取错误
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	viper.WatchConfig() // 监听配置文件的变化
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file changed: %s\n", e.Name)
	})

	log.Println("Loaded configuration from", viper.ConfigFileUsed())

	// 打印配置项，确保它们被正确加载
	log.Println("Server port:", viper.GetString("server.port"))
	log.Println("Database driver:", viper.GetString("database.driver"))
	log.Println("Redis addr:", viper.GetString("redis.addr"))
	log.Println("JWT expire_hours:", viper.GetInt("jwt.expire_hours"))
}

func GetConfig() *viper.Viper {
	return viper.GetViper()
}
