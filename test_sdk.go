// +build ignore

package main

import (
	"fmt"
	"github.com/huifurepo/bspay-go-sdk/BsPaySdk"
)

func main() {
	// 测试SDK的方法
	sdk, err := BsPaySdk.NewBsPay(false, "./test_config.json")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	// 打印SDK类型信息
	fmt.Printf("SDK initialized successfully!\n")
	fmt.Printf("SDK Type: %T\n", sdk)
	fmt.Printf("SysID: %s\n", sdk.Msc.SysId)
	fmt.Printf("ProductID: %s\n", sdk.Msc.ProductId)
}