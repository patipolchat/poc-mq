/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
// k6 run --out xk6-kafka=brokers=broker_host:8000,topic=k6

package main

import "mq-poc/cmd"

func main() {
	cmd.Execute()
}

//
