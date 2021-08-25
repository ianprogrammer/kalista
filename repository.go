package pkg

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
)


func GetContractFiles() {
	fmt.Println("ok")
	client := github.NewClient(nil)
	fmt.Println(client)
	repository, _, err := client.Repositories.Get(context.Background(),"ianprogrammer","contracts-repository")

	fmt.Println(err)
	if err == nil {
		fmt.Println(repository)
	}
}
