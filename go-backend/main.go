package main

import (
	"context"
	"fmt"
)

func main() {
	conn, err := connectDB()
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())

	fmt.Println("Connected to PostgreSQL successfully!")
}
