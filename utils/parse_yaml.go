package utils

import (
	"log"
	"miHttpServer/config"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseYaml 解析YAML配置文件
func ParseYaml() {
	// 打开 YAML 文件
	file, err := os.Open("./config/config.yaml")
	if err != nil {
		log.Fatalln("读取配置文件失败:", err)
	}
	defer file.Close()

	// 创建解析器
	decoder := yaml.NewDecoder(file)
	// 解析 YAML 数据
	err = decoder.Decode(&config.Configs)
	if err != nil {
		log.Fatalln("解析配置文件失败:", err)
	}
}
