package main

import (
	"fmt"

	"example.com/orders/internal/config"
	"example.com/orders/internal/handler"
	"example.com/orders/internal/repository"
	"example.com/orders/internal/service"
)

func main() {
	repo := repository.NewMemory()
	orders := service.NewService(repo)
	handler_ := handler.NewHandler(orders)

	address := config.Address()
	if address == "" {
		address = ":8080"
	}

	order, err := handler_.Create("Покупатель")
	if err != nil {
		fmt.Println("ошибка:", err)
		return
	}
	fmt.Printf("сервис слушал бы %s, создан заказ: %+v\n", address, order)
}
