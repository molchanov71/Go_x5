package main

import (
	"errors"
	"fmt"
	"lesson5/student/practice"
)

func main() {
	bn := practice.BaseNotifier{Channel: "канал 1"}
	n1 := practice.EmailNotifier{BaseNotifier: bn, Address: "test1@example.com"}
	n2 := practice.PushNotifier{BaseNotifier: bn, DeviceID: "test1"}
	err := practice.Dispatch([]practice.Notifier{&n1, &n2}, "заказ готов через Dispatch")
	if err != nil {
		fmt.Printf("необработанная ошибка: %v", err)
	} else {
		fmt.Println("n1:")
		for _, s := range n1.Messages {
			fmt.Println(s)
		}
		fmt.Println("\nn2:")
		for _, s := range n2.Messages {
			fmt.Println(s)
		}
		fmt.Println("")
	}

	err = practice.Dispatch([]practice.Notifier{&n1, &n2}, "")
	if err != nil && errors.Is(err, practice.ErrEmptyMessage) {
		fmt.Println("корректная обработка пустого сообщения")
		fmt.Println("n1:")
		for _, s := range n1.Messages {
			fmt.Println(s)
		}
		fmt.Println("\nn2:")
		for _, s := range n2.Messages {
			fmt.Println(s)
		}
		fmt.Println("")
	} else {
		fmt.Println("некорректная обработка пустого сообщения")
	}
	badN := practice.PushNotifier{BaseNotifier: bn}
	err = practice.Dispatch([]practice.Notifier{&badN}, "непустое сообщение")
	var target *practice.SendError
	if err != nil && errors.As(err, &target) {
		fmt.Printf("обработана ошибка SendError: Channel = \"%s\", Reason = \"%s\"\n", target.Channel, target.Reason)
	} else {
		fmt.Printf("некорректная обработка SendError, ошибка: %v\n", err)
	}

	errs := practice.SafeDispatch([]practice.Notifier{&practice.BrokenNotifier{BaseNotifier: bn}, &n2}, "заказ готов через SafeDispatch")
	if errs != nil && len(errs) == 1 && errors.Is(errs[0], practice.ErrNotifierPanic) {
		fmt.Println("корректная обработка ErrNotifierPanic")
	} else {
		fmt.Println("некорректная обработка ErrNotifierPanic")
	}
	fmt.Println("n1:")
	for _, s := range n1.Messages {
		fmt.Println(s)
	}
	fmt.Println("")
}
