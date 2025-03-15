package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/swaggest/openapi-go/openapi3"
)

type apiv1 struct {
	path        string
	method      string
	handler     echo.HandlerFunc
	payload     any
	response    any
	description string
}

func openapiPath(path string) string {
	result := []rune{}
	variable := false
	for _, char := range path {
		if char == ':' {
			result = append(result, '{')
			variable = true
			continue
		}
		if char == '/' {
			if variable {
				result = append(result, '}')
				variable = false
			}
		}
		result = append(result, char)
	}
	if variable {
		result = append(result, '}')
	}
	return string(result)
}

func openapiSpec(e *echo.Echo, definitions []apiv1) error {
	apiGroup := e.Group("/api/v1")
	reflector := openapi3.Reflector{}
	reflector.Spec = &openapi3.Spec{Openapi: "3.0.3"}
	reflector.Spec.Info.
		WithTitle("MaizAI API").
		WithVersion("0.0.1").
		WithDescription("Maizai HTTP API spec")
	for _, definition := range definitions {
		apiGroup.Add(definition.method, definition.path, definition.handler)
		path := fmt.Sprintf("/api/v1%s", openapiPath(definition.path))
		operation, err := reflector.NewOperationContext(definition.method, path)
		if err != nil {
			return err
		}
		if definition.payload != nil {
			operation.AddReqStructure(definition.payload)
		}
		operation.SetDescription(definition.description)
		operation.AddRespStructure(definition.response)
		err = reflector.AddOperation(operation)
		if err != nil {
			return err
		}
	}
	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		return err
	}
	e.GET("/openapi.yaml", func(c echo.Context) error {
		return c.String(http.StatusOK, string(schema))
	})
	return nil
}
