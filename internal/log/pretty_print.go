package log

import (
	"encoding/json"
	"fmt"
)

// PrettyPrint prints any struct in a human-readable format.
func PrettyPrint(i interface{}) {
	marshaled, _ := json.MarshalIndent(i, "", "  ")
	fmt.Println(string(marshaled))
}
