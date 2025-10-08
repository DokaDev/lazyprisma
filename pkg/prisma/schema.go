package prisma

import "os"

type SchemaChecker struct{}

func NewSchemaChecker() *SchemaChecker {
	return &SchemaChecker{}
}

func (s *SchemaChecker) Check() bool {
	_, err := os.Stat("prisma/schema.prisma")
	return err == nil
}
