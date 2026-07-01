package app

import "github.com/taka1156/brite/internal/entity"

type HelpCommand interface {
	Help()
}

type InitializeCommand interface {
	Initialize(clientConfig entity.ClientConfig)
}

type SetupCommand interface {
	Setup(clientConfig entity.ClientConfig)
}

type AddCommand interface {
	Add(clientConfig entity.ClientConfig)
}

type ConvertCommand interface {
	Convert(clientConfig entity.ClientConfig, jsonNames entity.JsonNames)
}

type PublishCommand interface {
	Publish(clientConfig entity.ClientConfig)
}
