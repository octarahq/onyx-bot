package logs

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Code string

func GenerateCode(date time.Time, module, submodule string) Code {
	prefix := fmt.Sprintf("%s%s", strings.ToUpper(string(module[0])), strings.ToUpper(string(submodule[0])))
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 5)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return Code(fmt.Sprintf("%s-%s", prefix, string(b)))
}
