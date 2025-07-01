package common

import (
	"fmt"

	"gopkg.in/ini.v1"
)

var CFG map[string]map[string]interface{}

func Load_Config(iniName string) error {
	iniFile, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, iniName) // # 주석 무시
	if err != nil {
		//fmt.Println("\033[41merr :", err.Error(), "\033[0m")
		return err
	}

	CFG = make(map[string]map[string]interface{})
	for _, v := range iniFile.SectionStrings() {
		CFG[v] = make(map[string]interface{})
		for _, k := range iniFile.Section(v).KeyStrings() {
			CFG[v][k] = iniFile.Section(v).Key(k).String()
		}
	}
	return nil
}

func Print_Config() {
	for k, v := range CFG {
		fmt.Printf(" [ %s ] =================================================\n", k)
		for key, val := range v {
			switch value := val.(type) {
			case int:
				fmt.Printf("\t\t[ %s ] = [ %d ]  \n", key, value)
			case string:
				fmt.Printf("\t\t[ %s ] = [ %s ]  \n", key, value)
			default:
				fmt.Printf("\t\t[ %s ] = [ %s ]  \n", key, value.(string))
			}
		}
	}
}
